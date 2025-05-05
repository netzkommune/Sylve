package rcconf_test

import (
	"os"
	"path/filepath"
	"testing"

	"sylve/pkg/rcconf"
)

func TestParse(t *testing.T) {
	content := `
hostname="freebsd"
ifconfig_em0="inet 192.168.1.10 netmask 255.255.255.0"
defaultrouter="192.168.1.1"
sshd_enable="YES"
`

	tmp := filepath.Join(t.TempDir(), "rc.conf")
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp rc.conf: %v", err)
	}

	conf, err := rcconf.Parse(tmp)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	tests := map[string]string{
		"hostname":      "freebsd",
		"ifconfig_em0":  "inet 192.168.1.10 netmask 255.255.255.0",
		"defaultrouter": "192.168.1.1",
		"sshd_enable":   "YES",
	}

	for k, expected := range tests {
		if val, ok := conf[k]; !ok || val != expected {
			t.Errorf("unexpected value for %s: got %q, want %q", k, val, expected)
		}
	}
}

func TestParse_EdgeCases(t *testing.T) {
	content := `
# A comment
; Another comment

empty_key=
whitespace_key = "value with spaces"  

quoted='single quotes'
doublequoted="double quotes"
no_quotes=plainvalue

invalid_line_without_equal
=missing_key
missing_value=

# trailing whitespace  
trailing_space =    " test "    

`

	tmp := filepath.Join(t.TempDir(), "rc_edge.conf")
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp rc.conf: %v", err)
	}

	conf, err := rcconf.Parse(tmp)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	tests := map[string]string{
		"empty_key":      "",
		"whitespace_key": "value with spaces",
		"quoted":         "single quotes",
		"doublequoted":   "double quotes",
		"no_quotes":      "plainvalue",
		"trailing_space": "test",
		"missing_value":  "", // still counts as key with empty value
	}

	for k, expected := range tests {
		val, ok := conf[k]
		if !ok {
			t.Errorf("expected key %q not found", k)
			continue
		}
		if val != expected {
			t.Errorf("unexpected value for %q: got %q, want %q", k, val, expected)
		}
	}

	// Check that invalid or malformed lines are skipped
	if _, ok := conf["invalid_line_without_equal"]; ok {
		t.Errorf("invalid line without equal should be skipped")
	}
	if _, ok := conf[""]; ok {
		t.Errorf("entry with missing key should be skipped")
	}
}
