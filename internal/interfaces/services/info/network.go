// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package infoServiceInterfaces

import "time"

type NetworkInterface struct {
	Name    string `json:"name"`
	Flags   string `json:"flags"`
	Network string `json:"network"`
	Address string `json:"address"`

	ReceivedPackets int64 `json:"received-packets"`
	ReceivedErrors  int64 `json:"received-errors"`
	DroppedPackets  int64 `json:"dropped-packets"`
	ReceivedBytes   int64 `json:"received-bytes"`

	SentPackets int64 `json:"sent-packets"`
	SendErrors  int64 `json:"send-errors"`
	SentBytes   int64 `json:"sent-bytes"`

	Collisions int64 `json:"collisions"`
}

type HistoricalNetworkInterface struct {
	SentBytes     int64     `json:"sentBytes"`
	ReceivedBytes int64     `json:"receivedBytes"`
	CreatedAt     time.Time `json:"createdAt"`
}
