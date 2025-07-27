// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"time"
)

const (
	PasswordResetTokenSize   = 32
	PasswordResetTokenExpiry = 30 * time.Minute
	PasswordResetCooldown    = 60 * time.Second
)

// PasswordResetToken represents a password reset request
type PasswordResetToken struct {
	Token      string `json:"token"`
	UserID     string `json:"user_id"`
	Email      string `json:"email"`
	CreateAt   int64  `json:"create_at"`
	ExpiresAt  int64  `json:"expires_at"`
	Used       bool   `json:"used"`
}

// PasswordResetRequest represents a request to reset password
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest represents the actual password reset
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"password"`
}

// IsValid validates a password reset request
func (r *PasswordResetRequest) IsValid() error {
	if r.Email == "" {
		return NewErrBadRequest("email is required")
	}
	if !isValidEmail(r.Email) {
		return NewErrBadRequest("invalid email format")
	}
	return nil
}

// isValidEmail checks if email format is valid
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValid validates a reset password request
func (r *ResetPasswordRequest) IsValid() error {
	if r.Token == "" {
		return NewErrBadRequest("token is required")
	}
	if r.NewPassword == "" {
		return NewErrBadRequest("password is required")
	}
	if len(r.NewPassword) < 8 {
		return NewErrBadRequest("password must be at least 8 characters")
	}
	return nil
}

// GeneratePasswordResetToken creates a new random token
func GeneratePasswordResetToken() string {
	b := make([]byte, PasswordResetTokenSize)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// IsExpired checks if the token has expired
func (t *PasswordResetToken) IsExpired() bool {
	return GetMillis() > t.ExpiresAt
}