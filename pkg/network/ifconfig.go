// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package network

import (
	"bufio"
	"strconv"
	"strings"
	"sylve/pkg/utils"
)

type NetworkInterface struct {
	Name        string   `json:"name"`
	Flags       string   `json:"flags"`
	ParsedFlags []string `json:"parsedFlags"`
	MAC         string   `json:"mac"`
	IPv4        string   `json:"ipv4"`
	Netmask     string   `json:"netmask"`
	Broadcast   string   `json:"broadcast"`
	IPv6        []string `json:"ipv6"`
	Media       string   `json:"media"`
	Status      string   `json:"status"`
	MTU         int      `json:"mtu"`
	Metric      int      `json:"metric"`
	IsUp        bool     `json:"isUp"`
	IsLoopback  bool     `json:"isLoopback"`
	IsRunning   bool     `json:"isRunning"`
}

func ifconfig(args ...string) (string, error) {
	output, err := utils.RunCommand("ifconfig", args...)
	if err != nil {
		return "", err
	}
	return output, nil
}

func extractFlags(flagsLine string) ([]string, string) {
	start := strings.Index(flagsLine, "<")
	end := strings.Index(flagsLine, ">")
	if start != -1 && end != -1 && end > start {
		flagPart := flagsLine[start+1 : end]
		parsed := strings.Split(flagPart, ",")
		return parsed, flagsLine
	}
	return []string{}, flagsLine
}

func ParseAll(output string) ([]NetworkInterface, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))

	var ifaces []NetworkInterface
	var current *NetworkInterface

	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == "" {
			continue
		}

		if !utils.IsIndented(line) {
			if current != nil {
				ifaces = append(ifaces, *current)
			}

			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}

			name := strings.TrimSpace(parts[0])
			flagsRest := strings.TrimSpace(parts[1])
			parsedFlags, rawFlags := extractFlags(flagsRest)

			current = &NetworkInterface{
				Name:        name,
				Flags:       rawFlags,
				ParsedFlags: parsedFlags,
				IsUp:        utils.Contains(parsedFlags, "UP"),
				IsLoopback:  utils.Contains(parsedFlags, "LOOPBACK"),
				IsRunning:   utils.Contains(parsedFlags, "RUNNING"),
			}

			parts = strings.Fields(flagsRest)
			for i := 0; i < len(parts)-1; i++ {
				if parts[i] == "metric" {
					if m, err := strconv.Atoi(parts[i+1]); err == nil {
						current.Metric = m
					}
				}
				if parts[i] == "mtu" {
					if mtu, err := strconv.Atoi(parts[i+1]); err == nil {
						current.MTU = mtu
					}
				}
			}

			continue
		}

		if current == nil {
			continue
		}

		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "ether"):
			current.MAC = strings.TrimSpace(strings.TrimPrefix(line, "ether"))
		case strings.HasPrefix(line, "inet "):
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				current.IPv4 = parts[1]
			}
			for i := 0; i < len(parts)-1; i++ {
				if parts[i] == "netmask" {
					current.Netmask = parts[i+1]
				} else if parts[i] == "broadcast" {
					current.Broadcast = parts[i+1]
				}
			}
		case strings.HasPrefix(line, "inet6"):
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				current.IPv6 = append(current.IPv6, parts[1])
			}
		case strings.HasPrefix(line, "media:"):
			current.Media = strings.TrimSpace(strings.TrimPrefix(line, "media:"))
		case strings.HasPrefix(line, "status:"):
			current.Status = strings.TrimSpace(strings.TrimPrefix(line, "status:"))
		}
	}

	if current != nil {
		ifaces = append(ifaces, *current)
	}

	return ifaces, nil
}

func GetInterfaces() ([]NetworkInterface, error) {
	output, err := ifconfig()
	if err != nil {
		return nil, err
	}

	ifaces, err := ParseAll(output)
	if err != nil {
		return nil, err
	}

	return ifaces, nil
}
