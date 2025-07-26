package model

import "time"

// BoardInvitation represents a board invitation
type BoardInvitation struct {
	ID        string    `json:"id" db:"id"`
	BoardID   string    `json:"boardId" db:"board_id"`
	Email     string    `json:"email" db:"email"`
	Token     string    `json:"token" db:"token"`
	Role      string    `json:"role" db:"role"`
	CreatedBy string    `json:"createdBy" db:"created_by"`
	CreatedAt int64     `json:"createdAt" db:"created_at"`
	ExpiresAt int64     `json:"expiresAt" db:"expires_at"`
	UsedAt    *int64    `json:"usedAt,omitempty" db:"used_at"`
	UsedBy    *string   `json:"usedBy,omitempty" db:"used_by"`
}

// BoardInviteRequest represents a request to invite someone to a board
type BoardInviteRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// IsExpired checks if the invitation has expired
func (bi *BoardInvitation) IsExpired() bool {
	return time.Now().Unix() > bi.ExpiresAt
}

// IsUsed checks if the invitation has been used
func (bi *BoardInvitation) IsUsed() bool {
	return bi.UsedAt != nil
}