// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package info

import (
	"errors"
	infoModels "sylve/internal/db/models/info"
	infoServiceInterfaces "sylve/internal/interfaces/services/info"

	"gorm.io/gorm"
)

var _ infoServiceInterfaces.InfoServiceInterface = (*Service)(nil)

type Service struct {
	DB *gorm.DB
}

func NewInfoService(db *gorm.DB) infoServiceInterfaces.InfoServiceInterface {
	return &Service{
		DB: db,
	}
}

func (s *Service) GetNotes() ([]infoModels.Note, error) {
	var notes []infoModels.Note
	err := s.DB.Find(&notes).Error
	return notes, err
}

func (s *Service) AddNote(title, note string) error {
	return s.DB.Create(&infoModels.Note{Title: title, Content: note}).Error
}

func (s *Service) DeleteNoteByID(id int) error {
	result := s.DB.Delete(&infoModels.Note{}, id)
	if result.RowsAffected == 0 {
		return errors.New("record not found")
	}
	return result.Error
}

func (s *Service) UpdateNoteByID(id int, title, note string) error {
	return s.DB.Model(&infoModels.Note{}).Where("id = ?", id).Updates(infoModels.Note{Title: title, Content: note}).Error
}
