// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package libvirt

import (
	"net/url"
	libvirtServiceInterfaces "sylve/internal/interfaces/services/libvirt"
	"sylve/internal/logger"
	"sync"

	"github.com/digitalocean/go-libvirt"
	"gorm.io/gorm"
)

var _ libvirtServiceInterfaces.LibvirtServiceInterface = (*Service)(nil)

type Service struct {
	DB   *gorm.DB
	Conn *libvirt.Libvirt

	actionMutex sync.Mutex
	crudMutex   sync.Mutex
}

func NewLibvirtService(db *gorm.DB) libvirtServiceInterfaces.LibvirtServiceInterface {
	uri, _ := url.Parse("bhyve:///system")
	l, err := libvirt.ConnectToURI(uri)
	if err != nil {
		logger.L.Fatal().Err(err).Msg("failed to connect to libvirt")
	}

	v, err := l.ConnectGetLibVersion()

	if err != nil {
		logger.L.Fatal().Err(err).Msg("failed to retrieve libvirt version")
	}

	logger.L.Info().Msgf("Libvirt version: %d", v)

	return &Service{
		DB:   db,
		Conn: l,
	}
}

func (s *Service) CheckVersion() error {
	_, err := s.Conn.ConnectGetLibVersion()
	if err != nil {
		return err
	}

	return nil
}
