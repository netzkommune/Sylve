// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"regexp"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func SHA256(input string, count int) string {
	sum := []byte(input)

	for i := 0; i < count; i++ {
		hash := sha256.Sum256(sum)
		sum = hash[:]
	}

	return hex.EncodeToString(sum)
}

func RemoveSpaces(input string) string {
	return strings.ReplaceAll(input, " ", "")
}

func StringToUint(s string) uint {
	hasher := fnv.New64a()
	hasher.Write([]byte(s))
	return uint(hasher.Sum64())
}

func GenerateRandomUUID() string {
	return uuid.New().String()
}

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[uuid.New().String()[i]%byte(len(charset))]
	}
	return string(b)
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func StringToUint64(s string) uint64 {
	r, error := strconv.ParseUint(s, 10, 64)

	if error != nil {
		return 0
	}

	return r
}

func StringToFloat64(s string) float64 {
	r, _ := strconv.ParseFloat(s, 64)
	return r
}

func RemoveEmptyLines(s string) string {
	re := regexp.MustCompile(`(?m)^\s*\n`)
	return re.ReplaceAllString(s, "")
}

func ParseJWT(tokenString string) (any, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT token format")
	}

	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("error extracting claims")
	}

	customClaims := make(map[string]interface{})
	for k, v := range claims {
		if k != "exp" && k != "jti" {
			customClaims[k] = v
		}
	}

	return customClaims, nil
}

func BytesToSize(toType string, bytes float64) float64 {
	switch toType {
	case "KB":
		return bytes / 1024
	case "MB":
		return bytes / 1024 / 1024
	case "GB":
		return bytes / 1024 / 1024 / 1024
	case "TB":
		return bytes / 1024 / 1024 / 1024 / 1024
	default:
		return bytes
	}
}

/*
 * from zfs diff`s escape function:
 *
 * Prints a file name out a character at a time.  If the character is
 * not in the range of what we consider "printable" ASCII, display it
 * as an escaped 3-digit octal value.  ASCII values less than a space
 * are all control characters and we declare the upper end as the
 * DELete character.  This also is the last 7-bit ASCII character.
 * We choose to treat all 8-bit ASCII as not printable for this
 * application.
 */
func UnescapeFilepath(path string) (string, error) {
	buf := make([]byte, 0, len(path))
	llen := len(path)
	for i := 0; i < llen; {
		if path[i] == '\\' {
			if llen < i+4 {
				return "", fmt.Errorf("invalid octal code: too short")
			}
			octalCode := path[(i + 1):(i + 4)]
			val, err := strconv.ParseUint(octalCode, 8, 8)
			if err != nil {
				return "", fmt.Errorf("invalid octal code: %w", err)
			}
			buf = append(buf, byte(val))
			i += 4
		} else {
			buf = append(buf, path[i])
			i++
		}
	}
	return string(buf), nil
}
