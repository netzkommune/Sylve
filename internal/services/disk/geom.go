// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package disk

import (
	"encoding/xml"
	diskServiceInterfaces "sylve/internal/interfaces/services/disk"
	"sylve/internal/utils"
)

func (s *Service) ParseGeomOutput() (diskServiceInterfaces.Mesh, error) {
	geomOutput := utils.GetGeomXML()

	var mesh diskServiceInterfaces.Mesh

	err := xml.Unmarshal(geomOutput, &mesh)

	if err != nil {
		return diskServiceInterfaces.Mesh{}, err
	}

	return mesh, nil
}
