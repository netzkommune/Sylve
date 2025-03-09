// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package serviceInterfaces

type CustomClaims struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	AuthType string `json:"authType"`
}

type AuthServiceInterface interface {
	VerifyTokenInDb(token string) bool
	GetJWTSecret() (string, error)
	CreateJWT(username, password, authType string, remember bool) (string, error)
	RevokeJWT(token string) error
	ValidateToken(tokenString string) (CustomClaims, error)
	ClearExpiredJWTTokens()
	InitSecret(name string, shaRounds int) error

	AuthenticatePAM(username, password string) (bool, error)
}
