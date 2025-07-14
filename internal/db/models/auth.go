// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package models

import (
	"time"
)

type User struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	Username      string    `gorm:"unique" json:"username"`
	Email         string    `gorm:"unique" json:"email"`
	Password      string    `json:"-"`
	Notes         string    `json:"notes"`
	TOTP          string    `json:"totp"`
	Admin         bool      `json:"admin"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
	LastLoginTime time.Time `json:"lastLoginTime"`
}

type Token struct {
	ID        uint      `gorm:"primarykey" json:"id,omitempty"`
	UserID    uint      `json:"userId,omitempty"`
	Token     string    `gorm:"index:,unique" json:"token,omitempty"`
	AuthType  string    `json:"authType,omitempty"`
	Expiry    time.Time `json:"expiry,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt,omitempty"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt,omitempty"`
}

type SystemSecrets struct {
	ID        uint   `gorm:"primarykey"`
	Name      string `gorm:"primarykey,unique"`
	Data      string
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt,omitempty"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt,omitempty"`
}
