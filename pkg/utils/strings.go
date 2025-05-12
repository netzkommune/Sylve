// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const Base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func FNVHash(s string) uint64 {
	hasher := fnv.New64a()
	hasher.Write([]byte(s))
	return hasher.Sum64()
}

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

	if count <= 0 {
		return input
	}

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

func GenerateDeterministicUUID(input string) string {
	hasher := sha256.New()
	hasher.Write([]byte(input))
	hash := hasher.Sum(nil)
	return uuid.NewSHA1(uuid.NameSpaceURL, hash).String()
}

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}
	return string(result)
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
	re := regexp.MustCompile(`(?m)^\n`)
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

func HumanFormatToSize(size string) uint64 {
	size = strings.TrimSpace(size)
	re := regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?)\s*([kmgtp]?b?)$`)

	matches := re.FindStringSubmatch(size)

	if len(matches) != 3 {
		reScientific := regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?(?:e[+-]?\d+)?)\s*([kmgtp]?b?)$`)
		matches = reScientific.FindStringSubmatch(size)
		if len(matches) != 3 {
			return 0
		}
	}

	num, err := strconv.ParseFloat(matches[1], 64)
	if err != nil || num < 0 {
		return 0
	}

	unit := strings.ToUpper(matches[2])
	if unit == "" {
		unit = "B"
	} else if !strings.HasSuffix(unit, "B") {
		unit += "B"
	}

	var multiplier float64
	switch unit {
	case "B":
		multiplier = 1
	case "KB":
		multiplier = 1 << 10
	case "MB":
		multiplier = 1 << 20
	case "GB":
		multiplier = 1 << 30
	case "TB":
		multiplier = 1 << 40
	case "PB":
		multiplier = 1 << 50
	default:
		return 0
	}

	maxVal := float64(^uint64(0))
	result := num * multiplier

	if num > maxVal/multiplier {
		return ^uint64(0)
	}

	if result >= maxVal {
		return ^uint64(0)
	}

	return uint64(result)
}

func IsIndented(line string) bool {
	return len(line) > 0 && unicode.IsSpace(rune(line[0]))
}

func Contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

func EncodeBase62(num uint64, length int) string {
	res := make([]byte, length)
	for i := length - 1; i >= 0; i-- {
		res[i] = Base62Chars[num%62]
		num /= 62
	}
	return string(res)
}

func ShortHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	num := binary.BigEndian.Uint64(hash[:8]) >> 16
	return EncodeBase62(num, 8)
}
