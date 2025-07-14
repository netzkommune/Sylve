package auth

import (
	"fmt"
	"sylve/internal/db/models"
)

func (s *Service) ListUsers() ([]models.User, error) {
	var users []models.User
	if err := s.DB.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed_to_list_users: %w", err)
	}
	return users, nil
}

func (s *Service) CreateUser(user *models.User) error {
	if err := s.DB.Create(user).Error; err != nil {
		return fmt.Errorf("failed_to_create_user: %w", err)
	}
	return nil
}

func (s *Service) EditUser(user *models.User) error {
	if err := s.DB.Save(user).Error; err != nil {
		return fmt.Errorf("failed_to_edit_user: %w", err)
	}
	return nil
}

func (s *Service) DeleteUser(userID uint) error {
	if err := s.DB.Delete(&models.User{}, userID).Error; err != nil {
		return fmt.Errorf("failed_to_delete_user: %w", err)
	}
	return nil
}
