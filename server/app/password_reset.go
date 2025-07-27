// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"fmt"

	"github.com/mattermost/focalboard/server/model"
)

// GetUserByEmail gets a user by email
func (a *App) GetUserByEmail(email string) (*model.User, error) {
	return a.store.GetUserByEmail(email)
}

// CreatePasswordResetToken creates a new password reset token
func (a *App) CreatePasswordResetToken(token *model.PasswordResetToken) error {
	return a.store.CreatePasswordResetToken(token)
}

// GetPasswordResetToken gets a password reset token
func (a *App) GetPasswordResetToken(token string) (*model.PasswordResetToken, error) {
	return a.store.GetPasswordResetToken(token)
}

// UpdatePasswordResetTokenUsed marks a token as used
func (a *App) UpdatePasswordResetTokenUsed(token string) error {
	return a.store.UpdatePasswordResetTokenUsed(token)
}

// DeleteExpiredPasswordResetTokens removes expired tokens
func (a *App) DeleteExpiredPasswordResetTokens() error {
	return a.store.DeleteExpiredPasswordResetTokens()
}

// GetUserLastPasswordResetAt gets the last password reset timestamp
func (a *App) GetUserLastPasswordResetAt(userID string) (int64, error) {
	return a.store.GetUserLastPasswordResetAt(userID)
}

// UpdateUserLastPasswordResetAt updates the last password reset timestamp
func (a *App) UpdateUserLastPasswordResetAt(userID string, timestamp int64) error {
	return a.store.UpdateUserLastPasswordResetAt(userID, timestamp)
}

// UpdateUserPasswordByID updates a user's password by ID
func (a *App) UpdateUserPasswordByID(userID, password string) error {
	return a.store.UpdateUserPasswordByID(userID, password)
}

// SendPasswordResetEmail sends a password reset email
func (a *App) SendPasswordResetEmail(email, resetURL string) error {
	// Note: Email subject would be "Reset your Focalboard password"
	// but SendMessage doesn't support subject parameter
	body := fmt.Sprintf(`Hello,

You requested to reset your password for Focalboard.

Click the link below to reset your password:
%s

This link will expire in 30 minutes.

If you did not request this reset, please ignore this email.

Best regards,
The Focalboard Team`, resetURL)

	return a.store.SendMessage(body, "", []string{email})
}