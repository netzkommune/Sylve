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
