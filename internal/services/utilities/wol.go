package utilities

import (
	"net"
	utilitiesModels "sylve/internal/db/models/utilities"
	"sylve/internal/logger"
	"sylve/pkg/utils"
	"time"

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
						// ----- < update your DB here >  -----
						// e.g. s.DB.Create(&models.WolEvent{Mac: mac, Timestamp: time.Now()})
						// fmt.Println("⚡ WOL packet detected for MAC:", mac)
						// ------------------------------------
						s.DB.Create(&utilitiesModels.WoL{
							Mac:    mac,
							Status: "pending",
						})

						logger.L.Info().Msgf("⚡ WOL packet detected for MAC: %s", mac)
					}
				}
			}
		}(handle)
	}

	logger.L.Info().Msg("WoL server started")
	select {}
}
