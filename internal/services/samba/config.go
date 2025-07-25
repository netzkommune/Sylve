package samba

import (
	"fmt"
	"os"
	"strings"
	sambaModels "sylve/internal/db/models/samba"
	"sylve/pkg/system"
	"sylve/pkg/utils"
	"sylve/pkg/zfs"

	iface "sylve/pkg/network/iface"
)

func (s *Service) GetGlobalConfig() (sambaModels.SambaSettings, error) {
	var settings sambaModels.SambaSettings
	if err := s.DB.First(&settings).Error; err != nil {
		return sambaModels.SambaSettings{}, fmt.Errorf("failed to retrieve Samba settings: %w", err)
	}
	return settings, nil
}

func (s *Service) SetGlobalConfig(unixCharset string,
	workgroup string,
	serverString string,
	interfaces string,
	bindInterfacesOnly bool) error {
	if unixCharset == "" || workgroup == "" || serverString == "" {
		return fmt.Errorf("unixCharset, workgroup, and serverString cannot be empty")
	}

	if interfaces == "" {
		interfaces = "lo0"
	}

	supportedCharsets := utils.GetSupportedCharsets()

	if !utils.StringInSlice(unixCharset, supportedCharsets) {
		return fmt.Errorf("unsupported unixCharset: %s", unixCharset)
	}

	if !utils.IsValidWorkgroup(workgroup) {
		return fmt.Errorf("invalid workgroup name: %s", workgroup)
	}

	if !utils.IsValidServerString(serverString) {
		return fmt.Errorf("invalid server string: %s", serverString)
	}

	interfacesList := strings.Split(interfaces, ",")
	interfacesList = utils.RemoveDuplicates(interfacesList)

	for _, eIface := range interfacesList {
		eIface = strings.TrimSpace(eIface)
		_, err := iface.Get(eIface)
		if err != nil {
			return fmt.Errorf("invalid interface '%s': %w", eIface, err)
		}
	}

	if len(interfacesList) > 0 {
		interfaces = strings.Join(interfacesList, ",")
	} else {
		interfaces = "lo0"
	}

	var settings sambaModels.SambaSettings
	if err := s.DB.First(&settings).Error; err != nil {
		return fmt.Errorf("failed to retrieve Samba settings: %w", err)
	}

	settings.UnixCharset = unixCharset
	settings.Workgroup = workgroup
	settings.ServerString = serverString
	settings.Interfaces = interfaces
	settings.BindInterfacesOnly = bindInterfacesOnly

	if err := s.DB.Save(&settings).Error; err != nil {
		return fmt.Errorf("failed to update Samba settings: %w", err)
	}

	return s.WriteConfig(true)
}

func (s *Service) GlobalConfig() (string, error) {
	settings, err := s.GetGlobalConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get global Samba settings: %w", err)
	}

	var config string
	config += "# === This file is automatically generated by Sylve, don't edit! ===\n"

	config += "[global]\n"
	config += fmt.Sprintf("unix charset = %s\n", settings.UnixCharset)
	config += fmt.Sprintf("workgroup = %s\n", settings.Workgroup)
	config += fmt.Sprintf("server string = %s\n", settings.ServerString)

	interfaces := settings.Interfaces
	if interfaces == "" {
		interfaces = "lo0"
	} else {
		interfaces = strings.ReplaceAll(interfaces, ",", " ")
	}

	config += fmt.Sprintf("interfaces = %s\n", interfaces)

	if settings.BindInterfacesOnly {
		config += "bind interfaces only = yes\n"
	} else {
		config += "bind interfaces only = no\n"
	}

	config += "vfs objects = zfsacl\n"
	config += "inherit acls = yes\n"

	return config, nil
}

func (s *Service) ShareConfig() (string, error) {
	shares := []sambaModels.SambaShare{}
	if err := s.DB.Preload("ReadOnlyGroups").Preload("WriteableGroups").Find(&shares).Error; err != nil {
		return "", fmt.Errorf("failed to retrieve Samba shares: %w", err)
	}

	datasets, err := zfs.Datasets("")

	if err != nil {
		return "", fmt.Errorf("failed to fetch datasets: %v", err)
	}

	var config strings.Builder
	for _, share := range shares {
		var dataset *zfs.Dataset

		for _, ds := range datasets {
			dProps, err := ds.GetAllProperties()
			if err != nil {
				return "", fmt.Errorf("failed to get properties for dataset %s: %v", share.Dataset, err)
			}

			if dProps["guid"] == share.Dataset {
				dataset = ds
				break
			}
		}

		if dataset == nil {
			return "", fmt.Errorf("dataset not found for share %s", share.Name)
		}

		config.WriteString(fmt.Sprintf("[%s]\n", share.Name))
		config.WriteString(fmt.Sprintf("\tpath = %s\n", dataset.Mountpoint))

		if share.GuestOk {
			config.WriteString(fmt.Sprintf("\tguest ok = yes\n"))
		} else {
			config.WriteString(fmt.Sprintf("\tguest ok = no\n"))
		}

		rGroups := make([]string, 0)
		wGroups := make([]string, 0)

		if len(share.ReadOnlyGroups) > 0 {
			for _, group := range share.ReadOnlyGroups {
				rGroups = append(rGroups, group.Name)
			}
		}

		if len(share.WriteableGroups) > 0 {
			for _, group := range share.WriteableGroups {
				wGroups = append(wGroups, group.Name)
			}
		}

		aGroups := utils.JoinStringSlices(rGroups, wGroups)
		writeList := fmt.Sprintf("%s%s", "@", strings.Join(wGroups, " @"))

		if len(aGroups) > 0 {
			config.WriteString(fmt.Sprintf("\tvalid users = %s\n", "@"+strings.Join(aGroups, " @")))
		}

		if share.ReadOnly {
			config.WriteString("\tread only = yes\n")
		}

		if share.GuestOk && !share.ReadOnly {
			config.WriteString("\tforce user = root\n")
		}

		if len(rGroups) > 0 && len(wGroups) > 0 {
			config.WriteString("\tread only = yes\n")
		}

		if len(wGroups) > 0 {
			config.WriteString(fmt.Sprintf("\twrite list = %s\n", writeList))
		}

		config.WriteString(fmt.Sprintf("\tcreate mask = %s\n", share.CreateMask))
		config.WriteString(fmt.Sprintf("\tdirectory mask = %s\n", share.DirectoryMask))
		config.WriteString("\n\n")

		_, err := utils.RunCommand("setfacl", "-b", dataset.Mountpoint)
		if err != nil {
			return "", fmt.Errorf("failed to clear ACLs on mountpoint %s: %w", dataset.Mountpoint, err)
		}

		if len(rGroups) > 0 {
			_, err := utils.RunCommand("setfacl", "-m", fmt.Sprintf("g:%s:read_set:fd:allow", strings.Join(rGroups, ",")), dataset.Mountpoint)
			if err != nil {
				return "", fmt.Errorf("failed to set read ACLs for groups %v on mountpoint %s: %w", rGroups, dataset.Mountpoint, err)
			}
		}

		if len(wGroups) > 0 {
			_, err := utils.RunCommand("setfacl", "-m", fmt.Sprintf("g:%s:modify_set:fd:allow", strings.Join(wGroups, ",")), dataset.Mountpoint)
			if err != nil {
				return "", fmt.Errorf("failed to set write ACLs for groups %v on mountpoint %s: %w", wGroups, dataset.Mountpoint, err)
			}
		}
	}

	return config.String(), nil
}

func (s *Service) WriteConfig(reload bool) error {
	gCfg, err := s.GlobalConfig()
	if err != nil {
		return err
	}

	if gCfg == "" {
		return fmt.Errorf("global configuration is empty")
	}

	shareCfg, err := s.ShareConfig()
	if err != nil {
		return fmt.Errorf("failed to get share configuration: %w", err)
	}

	fullConfig := gCfg + "\n" + shareCfg
	filePath := "/usr/local/etc/smb4.conf"

	if err := os.WriteFile(filePath, []byte(fullConfig), 0644); err != nil {
		return fmt.Errorf("failed to write Samba configuration to %s: %w", filePath, err)
	}

	if reload {
		if err := system.ServiceAction("samba_server", "reload"); err != nil {
			return fmt.Errorf("failed to reload Samba service: %w", err)
		}
	}

	return nil
}
