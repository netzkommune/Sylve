// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package systemServiceInterfaces

type Process struct {
	User          string `json:"user"`
	PID           string `json:"pid"`
	PercentCPU    string `json:"percent-cpu"`
	PercentMemory string `json:"percent-memory"`
	VirtualSize   string `json:"virtual-size"`
	RSS           string `json:"rss"`
	TerminalName  string `json:"terminal-name"`
	State         string `json:"state"`
	StartTime     string `json:"start-time"`
	CPUTime       string `json:"cpu-time"`
	Command       string `json:"command"`
}

type ProcessInformation struct {
	Process []Process `json:"process"`
}
