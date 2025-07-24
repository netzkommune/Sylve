package networkModels

import "time"

type Object struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"uniqueIndex;not null"`
	Type      string    `json:"type" gorm:"not null"` // "Host", "Network", "Port", "Country", "List"
	Comment   string    `json:"description"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Entries     []ObjectEntry      `json:"entries" gorm:"foreignKey:ObjectID"`
	Resolutions []ObjectResolution `json:"resolutions" gorm:"foreignKey:ObjectID"`
}

type ObjectEntry struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ObjectID  uint      `json:"objectId" gorm:"index"`
	Value     string    `json:"value"` // IP, CIDR, port, country code, FQDN, etc.
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ObjectResolution struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ObjectID   uint      `json:"objectId" gorm:"index"`
	ResolvedIP string    `json:"resolvedIp"` // actual IP resolved only in the case of FQDN
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
