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
	"hash/fnv"
	"regexp"
	"strconv"
	"strings"

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
	r, _ := strconv.ParseUint(s, 10, 64)
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
