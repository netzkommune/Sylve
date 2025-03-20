// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package auth

import (
	"crypto/tls"
	"fmt"
	"os"
	"sylve/internal/config"
	"sylve/internal/db/models"
	"sylve/pkg/crypto"
)

func (s *Service) GetSylveCertificate() (*tls.Config, error) {
	certPath := config.ParsedConfig.TLS.CertFile
	keyPath := config.ParsedConfig.TLS.KeyFile
	var certPEM, keyPEM []byte

	if certPath == "" || keyPath == "" {
		var certRecord, keyRecord models.SystemSecrets

		if err := s.DB.Where("name = ?", "tls_cert").First(&certRecord).Error; err == nil {
			certPEM = []byte(certRecord.Data)
		}

		if err := s.DB.Where("name = ?", "tls_key").First(&keyRecord).Error; err == nil {
			keyPEM = []byte(keyRecord.Data)
		}
	}

	if certPath != "" && keyPath != "" {
		certData, err := os.ReadFile(certPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read certificate file: %w", err)
		}
		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read key file: %w", err)
		}

		certPEM = certData
		keyPEM = keyData
	}

	if len(certPEM) == 0 || len(keyPEM) == 0 {
		var err error
		certPEM, keyPEM, err = crypto.GenerateSelfSignedCertificate()
		if err != nil {
			return nil, fmt.Errorf("failed to generate self-signed certificate: %w", err)
		}

		var certRecord, keyRecord models.SystemSecrets

		certRecord.Name = "tls_cert"
		certRecord.Data = string(certPEM)

		keyRecord.Name = "tls_key"
		keyRecord.Data = string(keyPEM)

		if err := s.DB.Create(&certRecord).Error; err != nil {
			return nil, fmt.Errorf("failed to save certificate: %w", err)
		}

		if err := s.DB.Create(&keyRecord).Error; err != nil {
			return nil, fmt.Errorf("failed to save key: %w", err)
		}
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
}
