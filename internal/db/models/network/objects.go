package networkModels

import "time"

type Object struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Type        string    `json:"type" gorm:"not null"` // "Host", "Network", "Port", "Country", "List"
	IPVersion   string    `json:"ipVersion"`            // for Country/List: "IPv4", "IPv6", "Both"
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	Entries []ObjectEntry `json:"entries" gorm:"foreignKey:ObjectID"`
}

type ObjectEntry struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ObjectID  uint      `json:"objectId" gorm:"index"`
	Value     string    `json:"value"`   // IP, CIDR, port, country code, FQDN, etc.
	Comment   string    `json:"comment"` // optional for clarity
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ObjectResolution struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ObjectID   uint      `json:"objectId" gorm:"index"`
	Source     string    `json:"source"`     // e.g., "example.com" or "AE"
	ResolvedIP string    `json:"resolvedIp"` // actual IP/CIDR resolved
	ExpiresAt  time.Time `json:"expiresAt"`  // TTL expiry
	CreatedAt  time.Time `json:"createdAt"`
}
