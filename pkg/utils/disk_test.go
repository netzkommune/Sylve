package utils

import (
	"errors"
	"testing"
)

func TestGetGeomXML(t *testing.T) {
	original := getSysctlBytes
	defer func() { getSysctlBytes = original }()

	tests := []struct {
		name       string
		mockReturn []byte
		mockError  error
		expected   []byte
	}{
		{
			name:       "Successful sysctl read",
			mockReturn: []byte("<geom><class name=\"DISK\"/></geom>"),
			mockError:  nil,
			expected:   []byte("<geom><class name=\"DISK\"/></geom>"),
		},
		{
			name:       "Error from sysctl",
			mockReturn: nil,
			mockError:  errors.New("failed to read"),
			expected:   []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getSysctlBytes = func(key string) ([]byte, error) {
				if key != "kern.geom.confxml" {
					t.Errorf("unexpected sysctl key: %s", key)
				}
				return tt.mockReturn, tt.mockError
			}

			result := GetGeomXML()
			if string(result) != string(tt.expected) {
				t.Errorf("expected %q, got %q", string(tt.expected), string(result))
			}
		})
	}
}

func TestGetDiskTypeFromUUID(t *testing.T) {
	tests := []struct {
		name         string
		inputUUID    string
		defaultValue string
		expected     string
	}{
		{
			name:         "Known UUID (ZFS)",
			inputUUID:    "516E7CBA-6ECF-11D6-8FF8-00022D09712B",
			defaultValue: "Unknown",
			expected:     "ZFS",
		},
		{
			name:         "Known UUID lowercase",
			inputUUID:    "516e7cba-6ecf-11d6-8ff8-00022d09712b",
			defaultValue: "Unknown",
			expected:     "ZFS",
		},
		{
			name:         "Unknown UUID",
			inputUUID:    "00000000-0000-0000-0000-deadbeefdead",
			defaultValue: "Custom",
			expected:     "Custom",
		},
		{
			name:         "Empty UUID",
			inputUUID:    "",
			defaultValue: "Not found",
			expected:     "Not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDiskTypeFromUUID(tt.inputUUID, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetDiskTypeFromUUID(%q) = %q; want %q", tt.inputUUID, result, tt.expected)
			}
		})
	}
}
