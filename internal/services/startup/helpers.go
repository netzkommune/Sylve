package startup

import (
	"fmt"
	"strings"
	"sylve/internal/logger"
	"sylve/pkg/pkg"
	"sylve/pkg/rcconf"
	"sylve/pkg/utils"
	sysctl "sylve/pkg/utils/sysctl"
	"sync"
)

func (s *Service) SysctlSync() error {
	intVals := map[string]int32{
		"net.inet.ip.forwarding":      1,
		"net.link.bridge.inherit_mac": 1,
	}

	for k, v := range intVals {
		_, err := sysctl.GetInt64(k)
		if err != nil {
			logger.L.Error().Msgf("Error getting sysctl %s: %v, skipping!", k, err)
			continue
		}

		err = sysctl.SetInt32(k, v)
		if err != nil {
			logger.L.Error().Msgf("Error setting sysctl %s: %v", k, err)
		}
	}

	return nil
}

func (s *Service) InitFirewall() error {
	return nil
}

func (s *Service) FreeBSDCheck() error {
	minMajor := uint64(14)
	minMinor := uint64(3)

	output, err := utils.RunCommand("uname", "-r")
	output = strings.TrimSpace(output)

	if err != nil {
		return fmt.Errorf("failed to run uname command: %w", err)
	}

	parts := strings.Split(output, "-")
	if len(parts) < 1 {
		return fmt.Errorf("unexpected output from uname command: %s", output)
	}

	versionParts := strings.Split(parts[0], ".")
	if len(versionParts) < 2 {
		return fmt.Errorf("unexpected version format: %s", parts[0])
	}

	majorVersion := utils.StringToUint64(versionParts[0])
	minorVersion := utils.StringToUint64(versionParts[1])

	if majorVersion < minMajor || (majorVersion == minMajor && minorVersion < minMinor) {
		return fmt.Errorf("unsupported FreeBSD version: %s, minimum required is %d.%d", output, minMajor, minMinor)
	}

	return nil
}

func (s *Service) CheckPackageDependencies() error {
	requiredPackages := []string{
		"libvirt",
		"bhyve-firmware",
		"smartmontools",
		"tmux",
		"samba419",
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(requiredPackages))

	for _, p := range requiredPackages {
		p := p
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !pkg.IsPackageInstalled(p) {
				errCh <- fmt.Errorf("Required package %s is not installed", p)
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) CheckServiceDependencies() error {
	const rcConfPath = "/etc/rc.conf"

	enabledServices := []string{
		"zfs_enable",
		"libvirtd_enable",
		"dnsmasq_enable",
		"rpcbind_enable",
		"samba_server_enable",
	}

	serviceNames := map[string]string{
		"libvirtd_enable":     "libvirtd",
		"dnsmasq_enable":      "dnsmasq",
		"rpcbind_enable":      "rpcbind",
		"samba_server_enable": "samba_server",
	}

	config, err := rcconf.Parse(rcConfPath)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", rcConfPath, err)
	}

	for _, key := range enabledServices {
		val, ok := config[key]
		if !ok || val != "YES" {
			return fmt.Errorf("required service %s is not enabled in %s", key, rcConfPath)
		}

		if key == "zfs_enable" || key == "samba_server_enable" {
			continue
		}

		service := serviceNames[key]
		if err := ensureServiceRunning(service); err != nil {
			return fmt.Errorf("failed to ensure service %s is running: %w", service, err)
		}
	}

	return nil
}

func (s *Service) CheckLoaderConf() error {
	const loaderConfPath = "/boot/loader.conf"

	required := map[string]string{
		"kern.geom.label.disk_ident.enable": "0",
		"kern.geom.label.gptid.enable":      "0",
		"cryptodev_load":                    "YES",
		"zfs_load":                          "YES",
		"vmm_load":                          "YES",
		"nmdm_load":                         "YES",
		"if_tap_load":                       "YES",
		"if_bridge_load":                    "YES",
		"hw.vmm.iommu.passthrough":          "1",
	}

	config, err := rcconf.Parse(loaderConfPath)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", loaderConfPath, err)
	}

	for key, expected := range required {
		val, ok := config[key]
		if !ok {
			return fmt.Errorf("missing required key in %s: %s", loaderConfPath, key)
		}
		if val != expected {
			return fmt.Errorf("invalid value for %s in %s: got %q, want %q", key, loaderConfPath, val, expected)
		}
	}

	return nil
}

func (s *Service) CheckKernelModules() error {
	requiredModules := []string{
		"vmm",
		"nmdm",
		"if_bridge",
	}

	output, err := utils.RunCommand("kldstat", "-q")

	if err != nil {
		return fmt.Errorf("failed to run kldstat command: %w", err)
	}

	for _, module := range requiredModules {
		if !strings.Contains(output, fmt.Sprintf("%s.ko", module)) {
			return fmt.Errorf("required kernel module %s is not loaded", module)
		}
	}

	return nil
}

func ensureServiceRunning(service string) error {
	_, err := utils.RunCommand("service", service, "status")
	if err == nil {
		return nil
	}

	_, startErr := utils.RunCommand("service", service, "start")
	if startErr != nil {
		return fmt.Errorf("could not start service %s: %w", service, startErr)
	}

	return nil
}
