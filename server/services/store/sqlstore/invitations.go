package sqlstore

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/mattermost/focalboard/server/model"
	"github.com/mattermost/focalboard/server/utils"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

// CreateBoardInvitation creates a new board invitation
func (s *SQLStore) CreateBoardInvitation(invitation *model.BoardInvitation) error {
	if invitation.ID == "" {
		invitation.ID = utils.NewID(utils.IDTypeNone)
	}

	query := s.getQueryBuilder(s.db).
		Insert(s.tablePrefix+"board_invitations").
		Columns(
			"id",
			"board_id",
			"email",
			"token",
			"role",
			"created_by",
			"created_at",
			"expires_at",
		).
		Values(
			invitation.ID,
			invitation.BoardID,
			invitation.Email,
			invitation.Token,
			invitation.Role,
			invitation.CreatedBy,
			invitation.CreatedAt,
			invitation.ExpiresAt,
		)

	if _, err := query.Exec(); err != nil {
		s.logger.Error("CreateBoardInvitation error",
			mlog.String("invitationID", invitation.ID),
			mlog.String("boardID", invitation.BoardID),
			mlog.Err(err))
		return err
	}

	return nil
}

// GetBoardInvitationByID retrieves a board invitation by ID
func (s *SQLStore) GetBoardInvitationByID(invitationID string) (*model.BoardInvitation, error) {
	query := s.getQueryBuilder(s.db).
		Select(
			"id",
			"board_id",
			"email",
			"token",
			"role",
			"created_by",
			"created_at",
			"expires_at",
			"used_at",
			"used_by",
			"last_sent_at",
		).
		From(s.tablePrefix + "board_invitations").
		Where(sq.Eq{"id": invitationID})

	row := query.QueryRow()

	invitation := &model.BoardInvitation{}
	var usedAt sql.NullInt64
	var usedBy sql.NullString
	var lastSentAt sql.NullInt64

	err := row.Scan(
		&invitation.ID,
		&invitation.BoardID,
		&invitation.Email,
		&invitation.Token,
		&invitation.Role,
		&invitation.CreatedBy,
		&invitation.CreatedAt,
		&invitation.ExpiresAt,
		&usedAt,
		&usedBy,
		&lastSentAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewErrNotFound("board invitation")
		}
		s.logger.Error("GetBoardInvitationByID error", mlog.Err(err))
		return nil, err
	}

	if usedAt.Valid {
		invitation.UsedAt = &usedAt.Int64
	}
	if usedBy.Valid {
		invitation.UsedBy = &usedBy.String
	}
	if lastSentAt.Valid {
		invitation.LastSentAt = &lastSentAt.Int64
	}

	return invitation, nil
}

// GetBoardInvitationByToken retrieves a board invitation by token
func (s *SQLStore) GetBoardInvitationByToken(token string) (*model.BoardInvitation, error) {
	query := s.getQueryBuilder(s.db).
		Select(
			"id",
			"board_id",
			"email",
			"token",
			"role",
			"created_by",
			"created_at",
			"expires_at",
			"used_at",
			"used_by",
			"last_sent_at",
		).
		From(s.tablePrefix + "board_invitations").
		Where(sq.Eq{"token": token})

	row := query.QueryRow()

	invitation := &model.BoardInvitation{}
	var usedAt sql.NullInt64
	var usedBy sql.NullString
	var lastSentAt sql.NullInt64

	err := row.Scan(
		&invitation.ID,
		&invitation.BoardID,
		&invitation.Email,
		&invitation.Token,
		&invitation.Role,
		&invitation.CreatedBy,
		&invitation.CreatedAt,
		&invitation.ExpiresAt,
		&usedAt,
		&usedBy,
		&lastSentAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewErrNotFound("board invitation")
		}
		s.logger.Error("GetBoardInvitationByToken error", mlog.Err(err))
		return nil, err
	}

	if usedAt.Valid {
		invitation.UsedAt = &usedAt.Int64
	}
	if usedBy.Valid {
		invitation.UsedBy = &usedBy.String
	}
	if lastSentAt.Valid {
		invitation.LastSentAt = &lastSentAt.Int64
	}

	return invitation, nil
}

// GetBoardInvitationsForBoard retrieves all invitations for a board
func (s *SQLStore) GetBoardInvitationsForBoard(boardID string) ([]*model.BoardInvitation, error) {
	query := s.getQueryBuilder(s.db).
		Select(
			"id",
			"board_id",
			"email",
			"token",
			"role",
			"created_by",
			"created_at",
			"expires_at",
			"used_at",
			"used_by",
			"last_sent_at",
		).
		From(s.tablePrefix + "board_invitations").
		Where(sq.Eq{"board_id": boardID}).
		OrderBy("created_at DESC")

	rows, err := query.Query()
	if err != nil {
		s.logger.Error("GetBoardInvitationsForBoard error", mlog.String("boardID", boardID), mlog.Err(err))
		return nil, err
	}
	defer rows.Close()

	invitations := []*model.BoardInvitation{}

	for rows.Next() {
		invitation := &model.BoardInvitation{}
		var usedAt sql.NullInt64
		var usedBy sql.NullString
		var lastSentAt sql.NullInt64

		err := rows.Scan(
			&invitation.ID,
			&invitation.BoardID,
			&invitation.Email,
			&invitation.Token,
			&invitation.Role,
			&invitation.CreatedBy,
			&invitation.CreatedAt,
			&invitation.ExpiresAt,
			&usedAt,
			&usedBy,
			&lastSentAt,
		)
		if err != nil {
			s.logger.Error("GetBoardInvitationsForBoard scan error", mlog.Err(err))
			return nil, err
		}

		if usedAt.Valid {
			invitation.UsedAt = &usedAt.Int64
		}
		if usedBy.Valid {
			invitation.UsedBy = &usedBy.String
		}
		if lastSentAt.Valid {
			invitation.LastSentAt = &lastSentAt.Int64
		}

		invitations = append(invitations, invitation)
	}

	return invitations, nil
}

// UpdateBoardInvitation updates a board invitation
func (s *SQLStore) UpdateBoardInvitation(invitation *model.BoardInvitation) error {
	query := s.getQueryBuilder(s.db).
		Update(s.tablePrefix+"board_invitations").
		Set("email", invitation.Email).
		Set("role", invitation.Role).
		Set("expires_at", invitation.ExpiresAt).
		Where(sq.Eq{"id": invitation.ID})

	if invitation.UsedAt != nil {
		query = query.Set("used_at", *invitation.UsedAt)
	}
	if invitation.UsedBy != nil {
		query = query.Set("used_by", *invitation.UsedBy)
	}
	if invitation.LastSentAt != nil {
		query = query.Set("last_sent_at", *invitation.LastSentAt)
	}

	if _, err := query.Exec(); err != nil {
		s.logger.Error("UpdateBoardInvitation error",
			mlog.String("invitationID", invitation.ID),
			mlog.Err(err))
		return err
	}

	return nil
}

// DeleteBoardInvitation deletes a board invitation
func (s *SQLStore) DeleteBoardInvitation(invitationID string) error {
	query := s.getQueryBuilder(s.db).
		Delete(s.tablePrefix+"board_invitations").
		Where(sq.Eq{"id": invitationID})

	if _, err := query.Exec(); err != nil {
		s.logger.Error("DeleteBoardInvitation error",
			mlog.String("invitationID", invitationID),
			mlog.Err(err))
		return err
	}

	return nil
}

// GetExpiredBoardInvitations retrieves all expired invitations
func (s *SQLStore) GetExpiredBoardInvitations() ([]*model.BoardInvitation, error) {
	now := time.Now().Unix()

	query := s.getQueryBuilder(s.db).
		Select(
			"id",
			"board_id",
			"email",
			"token",
			"role",
			"created_by",
			"created_at",
			"expires_at",
			"used_at",
			"used_by",
			"last_sent_at",
		).
		From(s.tablePrefix + "board_invitations").
		Where(sq.Lt{"expires_at": now}).
		Where(sq.Eq{"used_at": nil})

	rows, err := query.Query()
	if err != nil {
		s.logger.Error("GetExpiredBoardInvitations error", mlog.Err(err))
		return nil, err
	}
	defer rows.Close()

	invitations := []*model.BoardInvitation{}

	for rows.Next() {
		invitation := &model.BoardInvitation{}
		var usedAt sql.NullInt64
		var usedBy sql.NullString

		err := rows.Scan(
			&invitation.ID,
			&invitation.BoardID,
			&invitation.Email,
			&invitation.Token,
			&invitation.Role,
			&invitation.CreatedBy,
			&invitation.CreatedAt,
			&invitation.ExpiresAt,
			&usedAt,
			&usedBy,
		)
		if err != nil {
			s.logger.Error("GetExpiredBoardInvitations scan error", mlog.Err(err))
			return nil, err
		}

		if usedAt.Valid {
			invitation.UsedAt = &usedAt.Int64
		}
		if usedBy.Valid {
			invitation.UsedBy = &usedBy.String
		}

		invitations = append(invitations, invitation)
	}

	return invitations, nil
}