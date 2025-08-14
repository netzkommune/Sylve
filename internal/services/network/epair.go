// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package network

import (
	"fmt"
	"strings"
	jailModels "sylve/internal/db/models/jail"
	utils "sylve/pkg/utils"

	iface "sylve/pkg/network/iface"
)

func (s *Service) CreateEpair(name string) error {
	output, err := utils.RunCommand("ifconfig", "epair", "create")
	if err != nil {
		return fmt.Errorf("failed to create epair: %w", err)
	}

	epairA := strings.TrimSpace(string(output))
	if epairA == "" {
		return fmt.Errorf("failed to get epair name")
	}

	epairB := strings.TrimSuffix(epairA, "a") + "b"

	_, err = utils.RunCommand("ifconfig", epairA, "name", name+"a")
	if err != nil {
		return fmt.Errorf("failed to rename epair %s to %s: %w", epairA, name+"a", err)
	}

	_, err = utils.RunCommand("ifconfig", epairB, "name", name+"b")
	if err != nil {
		return fmt.Errorf("failed to rename epair %s to %s: %w", epairB, name+"b", err)
	}

	return nil
}

func (s *Service) DeleteEpair(name string) error {
	ifaces, err := iface.List()
	if err != nil {
		return fmt.Errorf("failed to list interfaces: %w", err)
	}

	var epairA string
	for _, iface := range ifaces {
		if strings.HasPrefix(iface.Name, name) {
			if strings.HasSuffix(iface.Name, "a") {
				epairA = iface.Name
			}
		}
	}

	if epairA == "" {
		return fmt.Errorf("epair %s not found", name)
	}

	_, err = utils.RunCommand("ifconfig", epairA, "destroy")

	if err != nil {
		return fmt.Errorf("failed to delete epair %s: %w", epairA, err)
	}

	return nil
}

func (s *Service) SyncEpairs() error {
	var jails []jailModels.Jail
	err := s.DB.Preload("Networks").Find(&jails).Error
	if err != nil {
		return fmt.Errorf("failed to find jails: %w", err)
	}

	ifaces, err := iface.List()
	if err != nil {
		return fmt.Errorf("failed to list interfaces: %w", err)
	}

	for _, jail := range jails {
		for _, network := range jail.Networks {
			networkId := fmt.Sprintf("%d", network.SwitchID)
			epairA := utils.HashIntToNLetters(jail.CTID, 5) + "_" + networkId + "a"

			found := false
			for _, iface := range ifaces {
				if iface.Name == epairA {
					found = true
					break
				}
			}

			if !found {
				if err := s.CreateEpair(utils.HashIntToNLetters(jail.CTID, 5) + "_" + networkId); err != nil {
					return fmt.Errorf("failed to create epair for jail %d network %d: %w", jail.CTID, network.SwitchID, err)
				}
			}
		}
	}

	return nil
}
