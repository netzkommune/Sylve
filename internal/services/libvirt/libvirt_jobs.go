package libvirt

import (
	"fmt"
	"strings"
	utilitiesModels "sylve/internal/db/models/utilities"
	"sylve/internal/logger"
	"time"
)

func (s *Service) WolTasks() {
	for {
		var wols []utilitiesModels.WoL
		if err := s.DB.Find(&wols, "status = ?", "pending").Error; err != nil {
			logger.L.Error().Msgf("failed_to_find_pending_wol_tasks: %v", err)
		} else {
			for _, wol := range wols {
				vm, err := s.FindVmByMac(wol.Mac)
				if err != nil {
					logger.L.Debug().Msgf("Failed to find VM associated with MAC: %s", wol.Mac)

					if strings.Contains(err.Error(), "record not found") {
						if err := s.DB.Model(&wol).Update("status", "not_found").Error; err != nil {
							logger.L.Error().Msgf("failed_to_update_wol_status: %v", err)
						}
						continue
					}
				}

				err = s.LvVMAction(vm, "start")

				if err != nil {
					message := fmt.Sprintf("failed_to_start_vm: %s", err.Error())
					if err := s.DB.Model(&wol).Update("status", message).Error; err != nil {
						logger.L.Error().Msgf("failed_to_update_wol_status: %v", err)
					}
				} else {
					if err := s.DB.Model(&wol).Update("status", "completed").Error; err != nil {
						logger.L.Error().Msgf("failed_to_update_wol_status: %v", err)
					}
				}
			}
		}

		time.Sleep(2 * time.Second)
	}
}
