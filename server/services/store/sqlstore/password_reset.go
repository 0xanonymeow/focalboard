// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"database/sql"

	"github.com/mattermost/focalboard/server/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"

	sq "github.com/Masterminds/squirrel"
)

// CreatePasswordResetToken creates a new password reset token
func (s *SQLStore) CreatePasswordResetToken(token *model.PasswordResetToken) error {
	query := s.getQueryBuilder(s.db).
		Insert(s.tablePrefix+"password_reset_tokens").
		Columns("token", "user_id", "email", "create_at", "expires_at", "used").
		Values(token.Token, token.UserID, token.Email, token.CreateAt, token.ExpiresAt, false)

	if _, err := query.Exec(); err != nil {
		s.logger.Error("Failed to create password reset token",
			mlog.String("email", token.Email),
			mlog.Err(err),
		)
		return err
	}

	return nil
}

// GetPasswordResetToken retrieves a password reset token
func (s *SQLStore) GetPasswordResetToken(token string) (*model.PasswordResetToken, error) {
	query := s.getQueryBuilder(s.db).
		Select("token", "user_id", "email", "create_at", "expires_at", "used").
		From(s.tablePrefix + "password_reset_tokens").
		Where(sq.Eq{"token": token})

	row := query.QueryRow()

	var resetToken model.PasswordResetToken
	err := row.Scan(
		&resetToken.Token,
		&resetToken.UserID,
		&resetToken.Email,
		&resetToken.CreateAt,
		&resetToken.ExpiresAt,
		&resetToken.Used,
	)

	if err == sql.ErrNoRows {
		return nil, model.NewErrNotFound("password reset token not found")
	}

	if err != nil {
		s.logger.Error("Failed to get password reset token", mlog.Err(err))
		return nil, err
	}

	return &resetToken, nil
}

// UpdatePasswordResetTokenUsed marks a token as used
func (s *SQLStore) UpdatePasswordResetTokenUsed(token string) error {
	query := s.getQueryBuilder(s.db).
		Update(s.tablePrefix+"password_reset_tokens").
		Set("used", true).
		Where(sq.Eq{"token": token})

	result, err := query.Exec()
	if err != nil {
		s.logger.Error("Failed to update password reset token", mlog.Err(err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return model.NewErrNotFound("password reset token not found")
	}

	return nil
}

// DeleteExpiredPasswordResetTokens removes expired tokens
func (s *SQLStore) DeleteExpiredPasswordResetTokens() error {
	query := s.getQueryBuilder(s.db).
		Delete(s.tablePrefix+"password_reset_tokens").
		Where(sq.Or{
			sq.Lt{"expires_at": model.GetMillis()},
			sq.Eq{"used": true},
		})

	if _, err := query.Exec(); err != nil {
		s.logger.Error("Failed to delete expired password reset tokens", mlog.Err(err))
		return err
	}

	return nil
}

// GetUserLastPasswordResetAt gets the timestamp of the last password reset request
func (s *SQLStore) GetUserLastPasswordResetAt(userID string) (int64, error) {
	query := s.getQueryBuilder(s.db).
		Select("COALESCE(last_password_reset_at, 0)").
		From(s.tablePrefix + "users").
		Where(sq.Eq{"id": userID})

	row := query.QueryRow()

	var timestamp int64
	err := row.Scan(&timestamp)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		s.logger.Error("Failed to get user last password reset timestamp", 
			mlog.String("userID", userID),
			mlog.Err(err),
		)
		return 0, err
	}

	return timestamp, nil
}

// UpdateUserLastPasswordResetAt updates the timestamp of the last password reset request
func (s *SQLStore) UpdateUserLastPasswordResetAt(userID string, timestamp int64) error {
	query := s.getQueryBuilder(s.db).
		Update(s.tablePrefix+"users").
		Set("last_password_reset_at", timestamp).
		Where(sq.Eq{"id": userID})

	result, err := query.Exec()
	if err != nil {
		s.logger.Error("Failed to update user last password reset timestamp",
			mlog.String("userID", userID),
			mlog.Err(err),
		)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return model.NewErrNotFound("user not found")
	}

	return nil
}