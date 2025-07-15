package systemServiceInterfaces

import "time"

type FileNode struct {
	ID   string    `json:"id"`
	Date time.Time `json:"date"`
	Type string    `json:"type"`
	Lazy bool      `json:"lazy,omitempty"`
	Size int64     `json:"size,omitempty"`
}
