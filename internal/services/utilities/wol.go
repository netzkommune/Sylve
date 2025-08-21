// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package utilities

import (
	"net"
	"time"

	utilitiesModels "github.com/alchemillahq/sylve/internal/db/models/utilities"
	"github.com/alchemillahq/sylve/internal/logger"
	"github.com/alchemillahq/sylve/pkg/utils"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var lastSeen = make(map[string]time.Time)
var dedupWindow = time.Second

func shouldEmit(payload []byte) bool {
	key := string(payload)
	now := time.Now()

	if t, ok := lastSeen[key]; ok {
		if now.Sub(t) < dedupWindow {
			return false
		}
	}
	lastSeen[key] = now
	return true
}

func (s *Service) StartWOLServer() error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	for _, iface := range ifaces {
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}

		handle, err := pcap.OpenLive(iface.Name, 65535, true, pcap.BlockForever)
		if err != nil {
			continue
		}

		go func(h *pcap.Handle) {
			src := gopacket.NewPacketSource(h, h.LinkType())
			for pkt := range src.Packets() {
				if udpLayer := pkt.Layer(layers.LayerTypeUDP); udpLayer != nil {
					udp := udpLayer.(*layers.UDP)
					if udp.DstPort != 9 && udp.DstPort != 7 {
						continue
					}
					app := pkt.ApplicationLayer()
					if app == nil {
						continue
					}
					payload := app.Payload()
					if len(payload) == 102 && utils.IsWOLPacket(payload) && shouldEmit(payload) {
						mac := utils.FormatMAC(payload[6:12])
						s.DB.Create(&utilitiesModels.WoL{
							Mac:    mac,
							Status: "pending",
						})

						logger.L.Debug().Msgf("⚡ WOL packet detected for MAC: %s", mac)
					}
				}
			}
		}(handle)
	}

	logger.L.Info().Msg("WoL server started")
	select {}
}
