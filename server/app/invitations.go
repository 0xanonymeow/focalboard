package app

import (
	"fmt"

	"github.com/mattermost/focalboard/server/model"
	"github.com/mattermost/focalboard/server/utils"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

// IsEmailConfigured returns true if email service is configured
func (a *App) IsEmailConfigured() bool {
	return a.email != nil && a.email.IsConfigured()
}

// GenerateInviteToken generates a secure token for invitations
func (a *App) GenerateInviteToken() (string, error) {
	if a.email == nil {
		return "", fmt.Errorf("email service not configured")
	}
	return a.email.GenerateInviteToken()
}

// SendInvitationEmail sends an invitation email
func (a *App) SendInvitationEmail(toEmail, boardTitle, inviterName, token string) error {
	if a.email == nil {
		return fmt.Errorf("email service not configured")
	}
	
	serverRoot := a.config.ServerRoot
	if serverRoot == "" {
		serverRoot = "http://localhost:8000"
	}
	
	return a.email.SendInvitation(toEmail, boardTitle, inviterName, token, serverRoot)
}

// CreateBoardInvitation creates a new board invitation
func (a *App) CreateBoardInvitation(invitation *model.BoardInvitation) error {
	if invitation.ID == "" {
		invitation.ID = utils.NewID(utils.IDTypeNone)
	}
	
	return a.store.CreateBoardInvitation(invitation)
}

// GetBoardInvitation retrieves a board invitation by token
func (a *App) GetBoardInvitation(token string) (*model.BoardInvitation, error) {
	return a.store.GetBoardInvitationByToken(token)
}

// UpdateBoardInvitation updates a board invitation
func (a *App) UpdateBoardInvitation(invitation *model.BoardInvitation) error {
	return a.store.UpdateBoardInvitation(invitation)
}

// GetBoardInvitationsForBoard retrieves all invitations for a board
func (a *App) GetBoardInvitationsForBoard(boardID string) ([]*model.BoardInvitation, error) {
	return a.store.GetBoardInvitationsForBoard(boardID)
}

// DeleteBoardInvitation deletes a board invitation
func (a *App) DeleteBoardInvitation(invitationID string) error {
	return a.store.DeleteBoardInvitation(invitationID)
}

// CleanupExpiredInvitations removes expired invitations
func (a *App) CleanupExpiredInvitations() error {
	invitations, err := a.store.GetExpiredBoardInvitations()
	if err != nil {
		return err
	}
	
	for _, invitation := range invitations {
		if err := a.store.DeleteBoardInvitation(invitation.ID); err != nil {
			a.logger.Error("Failed to delete expired invitation",
				mlog.String("invitationID", invitation.ID),
				mlog.Err(err))
		}
	}
	
	return nil
}