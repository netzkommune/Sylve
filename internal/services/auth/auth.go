// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package auth

import (
	"errors"
	"fmt"
	"strings"
	"sylve/internal/db/models"
	serviceInterfaces "sylve/internal/interfaces/services"
	"sylve/internal/logger"
	"sylve/pkg/utils"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ serviceInterfaces.AuthServiceInterface = (*Service)(nil)

type Service struct {
	DB *gorm.DB
}
type JWT struct {
	jwt.RegisteredClaims
	CustomClaims serviceInterfaces.CustomClaims `json:"custom_claims"`
}

func NewAuthService(db *gorm.DB) serviceInterfaces.AuthServiceInterface {
	return &Service{
		DB: db,
	}
}

func (s *Service) GetJWTSecret() (string, error) {
	var secret models.SystemSecrets

	if err := s.DB.Where("name = ?", "JWTSecret").First(&secret).Error; err != nil {
		return "", fmt.Errorf("jwt_secret_not_found")
	}

	return secret.Data, nil
}

func (s *Service) CreateJWT(username, password, authType string, remember bool) (string, error) {
	var user models.User

	if authType == "sylve" {
		if err := s.DB.Where("username = ?", username).First(&user).Error; err != nil {
			return "", fmt.Errorf("invalid_credentials")
		}

		if !utils.CheckPasswordHash(password, user.Password) {
			return "", fmt.Errorf("invalid_credentials")
		}
	} else if authType == "pam" {
		valid, err := s.AuthenticatePAM(username, password)

		if err != nil {
			return "", fmt.Errorf("pam_auth_error")
		}

		if !valid {
			return "", fmt.Errorf("invalid_credentials")
		}

		user.ID = utils.StringToUint(username)
		user.Username = username
	} else {
		return "", fmt.Errorf("invalid_auth_type")
	}

	var expiry time.Time

	if remember {
		expiry = time.Now().Add(time.Hour * 24 * 7)
	} else {
		expiry = time.Now().Add(time.Hour * 24)
	}

	data := JWT{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			ID:        uuid.NewString(),
		},
		CustomClaims: serviceInterfaces.CustomClaims{
			UserID:   user.ID,
			Username: user.Username,
			AuthType: authType,
		},
	}

	secret, err := s.GetJWTSecret()

	if err != nil {
		return "", fmt.Errorf("jwt_secret_not_found")
	}

	token, err := (jwt.NewWithClaims(jwt.SigningMethodHS256, data)).SignedString([]byte(secret))

	if err != nil {
		return "", fmt.Errorf("jwt_signing_failed")
	}

	tokenRecord := models.Token{
		Token:    token,
		AuthType: authType,
		UserID:   user.ID,
		Expiry:   expiry,
	}

	err = s.DB.Create(&tokenRecord).Error
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: tokens.token") {
			if updateErr := s.DB.Model(&tokenRecord).
				Where("token = ?", tokenRecord.Token).
				Updates(models.Token{UserID: tokenRecord.UserID}).Error; updateErr != nil {
				return "", fmt.Errorf("token_update_failed: %v", updateErr)
			}
		} else {
			return "", fmt.Errorf("token_save_failed: %v", err)
		}
	}

	if err != nil {
		return "", fmt.Errorf("hostname_fetch_failed")
	}

	return token, nil
}

func (s *Service) RevokeJWT(token string) error {
	var tokenRecord models.Token

	if err := s.DB.Where("token = ?", token).First(&tokenRecord).Error; err != nil {
		return fmt.Errorf("token_not_found")
	}

	if err := s.DB.Delete(&tokenRecord).Error; err != nil {
		return fmt.Errorf("token_delete_failed")
	}

	var user models.User

	if err := s.DB.Where("id = ?", tokenRecord.UserID).First(&user).Error; err != nil {
		return fmt.Errorf("user_not_found")
	}

	return nil
}

func (s *Service) VerifyTokenInDb(token string) bool {
	var tokenRecord models.Token

	if err := s.DB.Where("token = ?", token).First(&tokenRecord).Error; err != nil {
		logger.L.Error().Msgf("Token not found: %v", err)
		return false
	}

	var user models.User

	if err := s.DB.Where("id = ?", tokenRecord.UserID).First(&user).Error; err != nil {
		logger.L.Error().Msgf("User not found: %v", err)
		return false
	}

	return true
}

func (s *Service) ValidateToken(tokenString string) (serviceInterfaces.CustomClaims, error) {
	secret, err := s.GetJWTSecret()

	if err != nil {
		return serviceInterfaces.CustomClaims{}, fmt.Errorf("jwt_secret_not_found")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWT{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return serviceInterfaces.CustomClaims{}, fmt.Errorf("jwt_invalid")
	}

	claims, ok := token.Claims.(*JWT)

	if !ok || !token.Valid {
		return serviceInterfaces.CustomClaims{}, fmt.Errorf("jwt_invalid")
	}

	if time.Now().After(claims.ExpiresAt.Time) {
		return serviceInterfaces.CustomClaims{}, fmt.Errorf("jwt_expired")
	}

	if !s.VerifyTokenInDb(tokenString) {
		return serviceInterfaces.CustomClaims{}, fmt.Errorf("jwt_not_found_in_db")
	}

	return claims.CustomClaims, nil
}

func (s *Service) InitSecret(name string, shaRounds int) error {
	uuid, err := utils.GetSystemUUID()
	if err != nil {
		return fmt.Errorf("failed to get device UUID: %w", err)
	}

	secret := utils.SHA256(uuid, shaRounds)

	var systemSecret models.SystemSecrets
	err = s.DB.Where("name = ?", name).First(&systemSecret).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newSecret := models.SystemSecrets{
				Name: name,
				Data: secret,
			}
			if err := s.DB.Create(&newSecret).Error; err != nil {
				return fmt.Errorf("failed to create %s: %w", name, err)
			}
			logger.L.Debug().Msgf("Created new %s", name)
		} else {
			return fmt.Errorf("error fetching %s: %w", name, err)
		}
	} else {
		if systemSecret.Data != secret {
			if err := s.DB.Model(&systemSecret).Update("data", secret).Error; err != nil {
				return fmt.Errorf("failed to update %s: %w", name, err)
			}
			logger.L.Debug().Msgf("Updated existing %s", name)
		} else {
			logger.L.Debug().Msgf("%s is up to date", name)
		}
	}

	return nil
}

func (s *Service) GetSecret(name string) (string, error) {
	var secret models.SystemSecrets

	if err := s.DB.Where("name = ?", name).First(&secret).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("secret_not_found")
		} else {
			return "", fmt.Errorf("failed_to_get_secret")
		}
	}

	return secret.Data, nil
}

func (s *Service) UpsertSecret(name string, data string) error {
	var secret models.SystemSecrets

	err := s.DB.Where("name = ?", name).First(&secret).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newSecret := models.SystemSecrets{
				Name: name,
				Data: data,
			}
			if err := s.DB.Create(&newSecret).Error; err != nil {
				return fmt.Errorf("failed_to_create")
			}
		} else {
			return fmt.Errorf("failed_to_fetch")
		}
	} else {
		if secret.Data != data {
			if err := s.DB.Model(&secret).Update("data", data).Error; err != nil {
				return fmt.Errorf("failed_to_update")
			}
		} else {
			return fmt.Errorf("already_upto_date")
		}
	}

	return nil
}

func (s *Service) ClearExpiredJWTTokens() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := s.DB.Where("expiry < ?", time.Now()).Delete(&models.Token{})
			if err.Error != nil {
				logger.L.Error().Msgf("Error deleting expired tokens: %v", err.Error)
			} else {
				if err.RowsAffected > 0 {
					logger.L.Info().Msgf("Cleared %d expired tokens", err.RowsAffected)
				}
			}
		}
	}
}
