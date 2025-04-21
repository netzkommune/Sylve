// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package utils

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func TestFNVHash(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
	}{
		{
			input:    "hello",
			expected: FNVHash("hello"),
		},
		{
			input:    "world",
			expected: FNVHash("world"),
		},
		{
			input:    "",
			expected: FNVHash(""),
		},
		{
			input:    "12345",
			expected: FNVHash("12345"),
		},
		{
			input:    "special@chars!",
			expected: FNVHash("special@chars!"),
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input=%q", tt.input), func(t *testing.T) {
			result := FNVHash(tt.input)
			if result != tt.expected {
				t.Errorf("FNVHash(%q) = %d; want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHashAndCheckPassword(t *testing.T) {
	password := "supersecret123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	if !CheckPasswordHash(password, hash) {
		t.Error("CheckPasswordHash failed for correct password")
	}

	if CheckPasswordHash("wrongpassword", hash) {
		t.Error("CheckPasswordHash returned true for incorrect password")
	}
}

func TestSHA256(t *testing.T) {
	tests := []struct {
		input    string
		count    int
		expected string
	}{
		{
			input:    "hello",
			count:    1,
			expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			input:    "hello",
			count:    2,
			expected: "9595c9df90075148eb06860365df33584b75bff782a510c6cd4883a419833d50",
		},
		{
			input:    "",
			count:    1,
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			input:    "test",
			count:    0,
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input=%s,count=%d", tt.input, tt.count), func(t *testing.T) {
			result := SHA256(tt.input, tt.count)
			if result != tt.expected {
				t.Errorf("SHA256(%q, %d) = %q; want %q", tt.input, tt.count, result, tt.expected)
			}
		})
	}
}

func TestRemoveSpaces(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "hello world",
			expected: "helloworld",
		},
		{
			input:    "   leading spaces",
			expected: "leadingspaces",
		},
		{
			input:    "trailing spaces   ",
			expected: "trailingspaces",
		},
		{
			input:    "  spaces  in  between  ",
			expected: "spacesinbetween",
		},
		{
			input:    "nospace",
			expected: "nospace",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input=%q", tt.input), func(t *testing.T) {
			result := RemoveSpaces(tt.input)
			if result != tt.expected {
				t.Errorf("RemoveSpaces(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStringToUint(t *testing.T) {
	tests := []struct {
		input    string
		expected uint
	}{
		{
			input:    "hello",
			expected: uint(FNVHash("hello")),
		},
		{
			input:    "world",
			expected: uint(FNVHash("world")),
		},
		{
			input:    "",
			expected: uint(FNVHash("")),
		},
		{
			input:    "12345",
			expected: uint(FNVHash("12345")),
		},
		{
			input:    "special@chars!",
			expected: uint(FNVHash("special@chars!")),
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input=%q", tt.input), func(t *testing.T) {
			result := StringToUint(tt.input)
			if result != tt.expected {
				t.Errorf("StringToUint(%q) = %d; want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateRandomUUID(t *testing.T) {
	uuids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		uuid := GenerateRandomUUID()
		if _, exists := uuids[uuid]; exists {
			t.Errorf("Duplicate UUID generated: %s", uuid)
		}
		uuids[uuid] = true

		if len(uuid) == 0 {
			t.Error("Generated UUID is empty")
		}
	}
}

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		length int
	}{
		{length: 0},
		{length: 1},
		{length: 5},
		{length: 10},
		{length: 50},
		{length: 100},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("length=%d", tt.length), func(t *testing.T) {
			result := GenerateRandomString(tt.length)

			if len(result) != tt.length {
				t.Errorf("GenerateRandomString(%d) = %q; length = %d; want length = %d", tt.length, result, len(result), tt.length)
			}

			for _, char := range result {
				if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", char) {
					t.Errorf("GenerateRandomString(%d) generated invalid character: %q", tt.length, char)
				}
			}
		})
	}

	t.Run("randomness", func(t *testing.T) {
		length := 10
		generatedStrings := make(map[string]bool)
		for i := 0; i < 100; i++ {
			str := GenerateRandomString(length)
			if generatedStrings[str] {
				t.Errorf("Duplicate string generated: %q", str)
			}
			generatedStrings[str] = true
		}
	})
}

func TestStringInSlice(t *testing.T) {
	tests := []struct {
		element  string
		list     []string
		expected bool
	}{
		{
			element:  "apple",
			list:     []string{"apple", "banana", "cherry"},
			expected: true,
		},
		{
			element:  "grape",
			list:     []string{"apple", "banana", "cherry"},
			expected: false,
		},
		{
			element:  "",
			list:     []string{"apple", "banana", "cherry"},
			expected: false,
		},
		{
			element:  "banana",
			list:     []string{},
			expected: false,
		},
		{
			element:  "orange",
			list:     []string{"orange"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("element=%q,list=%v", tt.element, tt.list), func(t *testing.T) {
			result := StringInSlice(tt.element, tt.list)
			if result != tt.expected {
				t.Errorf("StringInSlice(%q, %v) = %v; want %v", tt.element, tt.list, result, tt.expected)
			}
		})
	}
}

func TestStringToUint64(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
	}{
		{
			input:    "12345",
			expected: 12345,
		},
		{
			input:    "0",
			expected: 0,
		},
		{
			input:    "18446744073709551615",
			expected: 18446744073709551615,
		},
		{
			input:    "",
			expected: 0,
		},
		{
			input:    "notanumber",
			expected: 0,
		},
		{
			input:    "-12345",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input=%q", tt.input), func(t *testing.T) {
			result := StringToUint64(tt.input)
			if result != tt.expected {
				t.Errorf("StringToUint64(%q) = %d; want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStringToFloat64(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{
			input:    "123.45",
			expected: 123.45,
		},
		{
			input:    "0",
			expected: 0,
		},
		{
			input:    "-987.65",
			expected: -987.65,
		},
		{
			input:    "1e3",
			expected: 1000,
		},
		{
			input:    "",
			expected: 0,
		},
		{
			input:    "notanumber",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input=%q", tt.input), func(t *testing.T) {
			result := StringToFloat64(tt.input)
			if result != tt.expected {
				t.Errorf("StringToFloat64(%q) = %f; want %f", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRemoveEmptyLines(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "line1\n\nline2\n\nline3\n",
			expected: "line1\nline2\nline3\n",
		},
		{
			input:    "\n\n\nline1\n\nline2\n\n",
			expected: "line1\nline2\n",
		},
		{
			input:    "line1\nline2\nline3",
			expected: "line1\nline2\nline3",
		},
		{
			input:    "\n\n\n",
			expected: "",
		},
		{
			input:    "",
			expected: "",
		},
		{
			input:    "   \n\nline1\n   \n\nline2\n",
			expected: "   \nline1\n   \nline2\n",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input=%q", tt.input), func(t *testing.T) {
			result := RemoveEmptyLines(tt.input)
			if result != tt.expected {
				t.Errorf("RemoveEmptyLines(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseJWT(t *testing.T) {
	claims := jwt.MapClaims{
		"username": "testuser",
		"role":     "admin",
		"exp":      time.Now().Add(time.Hour).Unix(),
		"jti":      "some-jti",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	t.Run("valid token", func(t *testing.T) {
		result, err := ParseJWT(tokenString)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		parsedClaims, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map[string]interface{}, got %T", result)
		}

		if parsedClaims["username"] != "testuser" {
			t.Errorf("expected username 'testuser', got %v", parsedClaims["username"])
		}
		if parsedClaims["role"] != "admin" {
			t.Errorf("expected role 'admin', got %v", parsedClaims["role"])
		}
		if _, exists := parsedClaims["exp"]; exists {
			t.Errorf("expected 'exp' to be removed, but it exists")
		}
		if _, exists := parsedClaims["jti"]; exists {
			t.Errorf("expected 'jti' to be removed, but it exists")
		}
	})

	t.Run("invalid token format", func(t *testing.T) {
		_, err := ParseJWT("thisisnot.valid.jwt")
		if err == nil {
			t.Error("expected error for invalid JWT format, got nil")
		}
	})

	t.Run("not a JWT", func(t *testing.T) {
		_, err := ParseJWT("notatoken")
		if err == nil {
			t.Error("expected error for malformed token, got nil")
		}
	})
}

func TestBytesToSize(t *testing.T) {
	tests := []struct {
		toType   string
		bytes    float64
		expected float64
	}{
		{
			toType:   "KB",
			bytes:    1024,
			expected: 1,
		},
		{
			toType:   "MB",
			bytes:    1048576,
			expected: 1,
		},
		{
			toType:   "GB",
			bytes:    1073741824,
			expected: 1,
		},
		{
			toType:   "TB",
			bytes:    1099511627776,
			expected: 1,
		},
		{
			toType:   "KB",
			bytes:    2048,
			expected: 2,
		},
		{
			toType:   "MB",
			bytes:    2097152,
			expected: 2,
		},
		{
			toType:   "GB",
			bytes:    2147483648,
			expected: 2,
		},
		{
			toType:   "TB",
			bytes:    2199023255552,
			expected: 2,
		},
		{
			toType:   "unknown",
			bytes:    12345,
			expected: 12345,
		},
		{
			toType:   "",
			bytes:    12345,
			expected: 12345,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("toType=%q,bytes=%f", tt.toType, tt.bytes), func(t *testing.T) {
			result := BytesToSize(tt.toType, tt.bytes)
			if result != tt.expected {
				t.Errorf("BytesToSize(%q, %f) = %f; want %f", tt.toType, tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestUnescapeFilepath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{
			input:    "normal/path",
			expected: "normal/path",
			hasError: false,
		},
		{
			input:    "escaped\\040path",
			expected: "escaped path",
			hasError: false,
		},
		{
			input:    "multiple\\040escaped\\041chars",
			expected: "multiple escaped!chars",
			hasError: false,
		},
		{
			input:    "\\011tab\\012newline",
			expected: "\ttab\nnewline",
			hasError: false,
		},
		{
			input:    "invalid\\08code",
			expected: "",
			hasError: true,
		},
		{
			input:    "short\\04",
			expected: "",
			hasError: true,
		},
		{
			input:    "",
			expected: "",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input=%q", tt.input), func(t *testing.T) {
			result, err := UnescapeFilepath(tt.input)
			if (err != nil) != tt.hasError {
				t.Errorf("UnescapeFilepath(%q) error = %v; want error = %v", tt.input, err != nil, tt.hasError)
			}
			if result != tt.expected {
				t.Errorf("UnescapeFilepath(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHumanFormatToSize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint64
	}{
		{"Bytes", "123", 123},
		{"Bytes with B", "123B", 123},
		{"Bytes with b", "123b", 123},
		{"Kilobytes", "1KB", 1 << 10},
		{"Kilobytes lowercase", "1kb", 1 << 10},
		{"Megabytes", "1MB", 1 << 20},
		{"Gigabytes", "1GB", 1 << 30},
		{"Terabytes", "1TB", 1 << 40},
		{"Petabytes", "1PB", 1 << 50},
		{"Decimal KB", "1.5KB", (1 << 10) + (1 << 9)},
		{"Decimal MB", "0.5MB", 1 << 19},
		{"Small decimal", "0.0001GB", 107374},
		{"Space before unit", "1 KB", 1 << 10},
		{"Multiple spaces", "1   MB", 1 << 20},
		{"Tab separator", "1\tGB", 1 << 30},
		{"Zero", "0", 0},
		{"Zero with unit", "0GB", 0},
		{"Large number", "1000000000000000000", 1000000000000000000},
		{"Max uint64", "18446744073709551615", ^uint64(0)},
		{"Empty string", "", 0},
		{"Only spaces", "   ", 0},
		{"Invalid unit", "1XB", 0},
		{"Negative number", "-1KB", 0},
		{"Invalid format", "KB", 0},
		{"Number too large", "18446744073709551616", ^uint64(0)},
		{"Number way too large", "1e100PB", ^uint64(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HumanFormatToSize(tt.input); got != tt.expected {
				t.Errorf("HumanFormatToSize(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
func TestIsIndented(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{
			input:    "    indented line",
			expected: true,
		},
		{
			input:    "\tindented with tab",
			expected: true,
		},
		{
			input:    "not indented",
			expected: false,
		},
		{
			input:    "",
			expected: false,
		},
		{
			input:    " \t mixed spaces and tab",
			expected: true,
		},
		{
			input:    "\nnewline character",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input=%q", tt.input), func(t *testing.T) {
			result := IsIndented(tt.input)
			if result != tt.expected {
				t.Errorf("IsIndented(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}
func TestContains(t *testing.T) {
	tests := []struct {
		slice    []string
		val      string
		expected bool
	}{
		{
			slice:    []string{"apple", "banana", "cherry"},
			val:      "banana",
			expected: true,
		},
		{
			slice:    []string{"apple", "banana", "cherry"},
			val:      "grape",
			expected: false,
		},
		{
			slice:    []string{},
			val:      "apple",
			expected: false,
		},
		{
			slice:    []string{"apple", "banana", "cherry"},
			val:      "",
			expected: false,
		},
		{
			slice:    []string{"", "banana", "cherry"},
			val:      "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("slice=%v,val=%q", tt.slice, tt.val), func(t *testing.T) {
			result := Contains(tt.slice, tt.val)
			if result != tt.expected {
				t.Errorf("Contains(%v, %q) = %v; want %v", tt.slice, tt.val, result, tt.expected)
			}
		})
	}
}
