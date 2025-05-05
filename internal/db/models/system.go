package models

type DefaultRoutes struct {
	IPv4 string `json:"ipv4"`
	IPv6 string `json:"ipv6"`
}

type System struct {
	ID            int           `json:"id" gorm:"primaryKey"`
	Initialized   bool          `json:"initialized"`
	Hostname      string        `json:"hostname"`
	DefaultRoutes DefaultRoutes `json:"defaultRoutes" gorm:"embedded"`
	ISODir        string        `json:"isoDir"`
}
