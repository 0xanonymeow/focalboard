package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/focalboard/server/model"
	"github.com/mattermost/focalboard/server/services/audit"
	"github.com/mattermost/focalboard/server/services/auth"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

func (a *API) registerInvitationRoutes(r *mux.Router) {
	// Invitation APIs
	r.HandleFunc("/boards/{boardID}/invite", a.sessionRequired(a.handleSendInvitation)).Methods("POST")
	r.HandleFunc("/boards/{boardID}/invitations", a.sessionRequired(a.handleGetBoardInvitations)).Methods("GET")
	r.HandleFunc("/invitations/{invitationID}/resend", a.sessionRequired(a.handleResendInvitation)).Methods("POST")
	r.HandleFunc("/invitations/{invitationID}", a.sessionRequired(a.handleDeleteInvitation)).Methods("DELETE")
	r.HandleFunc("/invite/{token}", a.handleAcceptInvitation).Methods("GET")
	r.HandleFunc("/invite/{token}/accept", a.sessionRequired(a.handleAcceptInvitationPost)).Methods("POST")
}

func (a *API) handleSendInvitation(w http.ResponseWriter, r *http.Request) {
	// swagger:operation POST /boards/{boardID}/invite sendInvitation
	//
	// Send an email invitation to join a board
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: boardID
	//   in: path
	//   description: Board ID
	//   required: true
	//   type: string
	// - name: Body
	//   in: body
	//   description: Invitation request
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/BoardInviteRequest"
	// security:
	// - BearerAuth: []
	// responses:
	//   '200':
	//     description: success
	//   '400':
	//     description: invalid request
	//   '403':
	//     description: access denied
	//   default:
	//     description: internal error
	//     schema:
	//       "$ref": "#/definitions/ErrorResponse"

	boardID := mux.Vars(r)["boardID"]
	userID := getUserID(r)

	// Check if user has permission to invite others to this board
	if !a.permissions.HasPermissionToBoard(userID, boardID, model.PermissionManageBoardRoles) {
		a.errorResponse(w, r, model.NewErrPermission("access denied to invite members"))
		return
	}

	// Parse request body
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	var inviteReq model.BoardInviteRequest
	if err = json.Unmarshal(requestBody, &inviteReq); err != nil {
		a.errorResponse(w, r, model.NewErrBadRequest(err.Error()))
		return
	}

	// Validate email
	if !auth.IsEmailValid(inviteReq.Email) {
		a.errorResponse(w, r, model.NewErrBadRequest("invalid email address"))
		return
	}

	// Validate role
	if inviteReq.Role == "" {
		inviteReq.Role = "viewer" // Default role
	}

	auditRec := a.makeAuditRecord(r, "sendInvitation", audit.Fail)
	defer a.audit.LogRecord(audit.LevelModify, auditRec)
	auditRec.AddMeta("boardID", boardID)
	auditRec.AddMeta("email", inviteReq.Email)

	// Get board info
	board, err := a.app.GetBoard(boardID)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	// Get inviter info
	inviter, err := a.app.GetUser(userID)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	// Check if email service is configured
	if !a.app.IsEmailConfigured() {
		a.errorResponse(w, r, model.NewErrBadRequest("email service not configured"))
		return
	}

	// Generate invitation token
	token, err := a.app.GenerateInviteToken()
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	// Create invitation record
	invitation := &model.BoardInvitation{
		BoardID:   boardID,
		Email:     inviteReq.Email,
		Token:     token,
		Role:      inviteReq.Role,
		CreatedBy: userID,
		CreatedAt: time.Now().Unix(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days expiry
	}

	// Save invitation
	if err = a.app.CreateBoardInvitation(invitation); err != nil {
		a.errorResponse(w, r, err)
		return
	}

	// Set last sent timestamp
	now := time.Now().Unix()
	invitation.LastSentAt = &now

	// Send invitation email
	inviterName := inviter.Username
	if inviter.FirstName != "" || inviter.LastName != "" {
		inviterName = strings.TrimSpace(inviter.FirstName + " " + inviter.LastName)
	}

	err = a.app.SendInvitationEmail(inviteReq.Email, board.Title, inviterName, token)
	if err != nil {
		a.logger.Error("Failed to send invitation email",
			mlog.String("email", inviteReq.Email),
			mlog.String("boardID", boardID),
			mlog.Err(err))
		
		// Delete the invitation record since email failed
		deleteErr := a.app.DeleteBoardInvitation(invitation.ID)
		if deleteErr != nil {
			a.logger.Error("Failed to delete invitation after email failure",
				mlog.String("invitationID", invitation.ID),
				mlog.Err(deleteErr))
		}
		
		a.errorResponse(w, r, model.NewErrBadRequest("Failed to send invitation email: "+err.Error()))
		return
	}

	// Update invitation with last sent timestamp
	updateErr := a.app.UpdateBoardInvitation(invitation)
	if updateErr != nil {
		a.logger.Error("Failed to update invitation timestamp",
			mlog.String("invitationID", invitation.ID),
			mlog.Err(updateErr))
		// Don't fail the request, email was sent successfully
	}

	a.logger.Debug("Invitation sent successfully",
		mlog.String("boardID", boardID),
		mlog.String("email", inviteReq.Email),
		mlog.String("invitedBy", userID))

	jsonStringResponse(w, http.StatusOK, "{}")
	auditRec.Success()
}

func (a *API) handleAcceptInvitation(w http.ResponseWriter, r *http.Request) {
	// swagger:operation GET /invite/{token} getInvitation
	//
	// Get invitation details for display
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: token
	//   in: path
	//   description: Invitation Token
	//   required: true
	//   type: string
	// responses:
	//   '200':
	//     description: success
	//     schema:
	//       "$ref": "#/definitions/BoardInvitation"
	//   '404':
	//     description: invitation not found
	//   default:
	//     description: internal error
	//     schema:
	//       "$ref": "#/definitions/ErrorResponse"

	token := mux.Vars(r)["token"]

	auditRec := a.makeAuditRecord(r, "getInvitation", audit.Fail)
	defer a.audit.LogRecord(audit.LevelRead, auditRec)
	auditRec.AddMeta("token", token[:8]+"...") // Only log first 8 chars for security

	// Get invitation
	invitation, err := a.app.GetBoardInvitation(token)
	if err != nil {
		a.errorResponse(w, r, model.NewErrNotFound("invitation not found"))
		return
	}

	// Check if invitation is expired or used
	if invitation.IsExpired() {
		a.errorResponse(w, r, model.NewErrBadRequest("invitation has expired"))
		return
	}

	if invitation.IsUsed() {
		a.errorResponse(w, r, model.NewErrBadRequest("invitation has already been used"))
		return
	}

	// Get board info for display
	board, err := a.app.GetBoard(invitation.BoardID)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	// Return invitation info with board details
	response := map[string]interface{}{
		"boardTitle": board.Title,
		"email":      invitation.Email,
		"role":       invitation.Role,
		"boardId":    invitation.BoardID,
		"valid":      true,
	}

	data, err := json.Marshal(response)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	jsonBytesResponse(w, http.StatusOK, data)
	auditRec.Success()
}

func (a *API) handleAcceptInvitationPost(w http.ResponseWriter, r *http.Request) {
	// swagger:operation POST /invite/{token}/accept acceptInvitation
	//
	// Accept a board invitation
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: token
	//   in: path
	//   description: Invitation Token
	//   required: true
	//   type: string
	// security:
	// - BearerAuth: []
	// responses:
	//   '200':
	//     description: success
	//   '400':
	//     description: invalid invitation
	//   '404':
	//     description: invitation not found
	//   default:
	//     description: internal error
	//     schema:
	//       "$ref": "#/definitions/ErrorResponse"

	token := mux.Vars(r)["token"]
	userID := getUserID(r)

	auditRec := a.makeAuditRecord(r, "acceptInvitation", audit.Fail)
	defer a.audit.LogRecord(audit.LevelModify, auditRec)
	auditRec.AddMeta("token", token[:8]+"...")
	auditRec.AddMeta("userID", userID)

	// Get invitation
	invitation, err := a.app.GetBoardInvitation(token)
	if err != nil {
		a.errorResponse(w, r, model.NewErrNotFound("invitation not found"))
		return
	}

	// Check if invitation is expired or used
	if invitation.IsExpired() {
		a.errorResponse(w, r, model.NewErrBadRequest("invitation has expired"))
		return
	}

	if invitation.IsUsed() {
		a.errorResponse(w, r, model.NewErrBadRequest("invitation has already been used"))
		return
	}

	// Get user info to validate email matches
	user, err := a.app.GetUser(userID)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	// Check if user's email matches invitation email
	if user.Email != invitation.Email {
		a.errorResponse(w, r, model.NewErrBadRequest("invitation email does not match your account"))
		return
	}

	// Add user to board
	newBoardMember := &model.BoardMember{
		UserID:          userID,
		BoardID:         invitation.BoardID,
		SchemeAdmin:     invitation.Role == "admin",
		SchemeEditor:    invitation.Role == "editor",
		SchemeCommenter: invitation.Role == "commenter",
		SchemeViewer:    invitation.Role == "viewer" || invitation.Role == "",
	}

	_, err = a.app.AddMemberToBoard(newBoardMember)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	// Mark invitation as used
	now := time.Now().Unix()
	invitation.UsedAt = &now
	invitation.UsedBy = &userID

	err = a.app.UpdateBoardInvitation(invitation)
	if err != nil {
		a.logger.Error("Failed to mark invitation as used",
			mlog.String("token", token),
			mlog.String("userID", userID),
			mlog.Err(err))
		// Don't fail the request, user is already added to board
	}

	a.logger.Debug("Invitation accepted",
		mlog.String("boardID", invitation.BoardID),
		mlog.String("userID", userID),
		mlog.String("email", invitation.Email))

	jsonStringResponse(w, http.StatusOK, "{}")
	auditRec.Success()
}

func (a *API) handleGetBoardInvitations(w http.ResponseWriter, r *http.Request) {
	// swagger:operation GET /boards/{boardID}/invitations getBoardInvitations
	//
	// Get all pending invitations for a board
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: boardID
	//   in: path
	//   description: Board ID
	//   required: true
	//   type: string
	// security:
	// - BearerAuth: []
	// responses:
	//   '200':
	//     description: success
	//     schema:
	//       type: array
	//       items:
	//         "$ref": "#/definitions/BoardInvitation"
	//   '403':
	//     description: access denied
	//   default:
	//     description: internal error
	//     schema:
	//       "$ref": "#/definitions/ErrorResponse"

	boardID := mux.Vars(r)["boardID"]
	userID := getUserID(r)

	// Check if user has permission to manage board roles
	if !a.permissions.HasPermissionToBoard(userID, boardID, model.PermissionManageBoardRoles) {
		a.errorResponse(w, r, model.NewErrPermission("access denied to view invitations"))
		return
	}

	auditRec := a.makeAuditRecord(r, "getBoardInvitations", audit.Fail)
	defer a.audit.LogRecord(audit.LevelRead, auditRec)
	auditRec.AddMeta("boardID", boardID)

	// Get pending invitations
	invitations, err := a.app.GetBoardInvitations(boardID)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	// Calculate cooldown for each invitation
	for i := range invitations {
		invitations[i].ResendCooldownSeconds = invitations[i].CalculateResendCooldown()
	}

	data, err := json.Marshal(invitations)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	jsonBytesResponse(w, http.StatusOK, data)
	auditRec.Success()
}

func (a *API) handleResendInvitation(w http.ResponseWriter, r *http.Request) {
	// swagger:operation POST /invitations/{invitationID}/resend resendInvitation
	//
	// Resend a board invitation email
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: invitationID
	//   in: path
	//   description: Invitation ID
	//   required: true
	//   type: string
	// security:
	// - BearerAuth: []
	// responses:
	//   '200':
	//     description: success
	//   '403':
	//     description: access denied
	//   '404':
	//     description: invitation not found
	//   default:
	//     description: internal error
	//     schema:
	//       "$ref": "#/definitions/ErrorResponse"

	invitationID := mux.Vars(r)["invitationID"]
	userID := getUserID(r)

	auditRec := a.makeAuditRecord(r, "resendInvitation", audit.Fail)
	defer a.audit.LogRecord(audit.LevelModify, auditRec)
	auditRec.AddMeta("invitationID", invitationID)

	// Get invitation
	invitation, err := a.app.GetBoardInvitationByID(invitationID)
	if err != nil {
		a.errorResponse(w, r, model.NewErrNotFound("invitation not found"))
		return
	}

	// Check if user has permission to manage board roles
	if !a.permissions.HasPermissionToBoard(userID, invitation.BoardID, model.PermissionManageBoardRoles) {
		a.errorResponse(w, r, model.NewErrPermission("access denied to resend invitation"))
		return
	}

	// Check if invitation is already used or expired
	if invitation.IsUsed() {
		a.errorResponse(w, r, model.NewErrBadRequest("invitation has already been used"))
		return
	}

	if invitation.IsExpired() {
		a.errorResponse(w, r, model.NewErrBadRequest("invitation has expired"))
		return
	}

	// Check if invitation is in cooldown
	if !invitation.CanResend() {
		cooldownSeconds := invitation.CalculateResendCooldown()
		a.errorResponse(w, r, model.NewErrBadRequest(fmt.Sprintf("please wait %d seconds before resending", cooldownSeconds)))
		return
	}

	// Check if email service is configured
	if !a.app.IsEmailConfigured() {
		a.errorResponse(w, r, model.NewErrBadRequest("email service not configured"))
		return
	}

	// Get board info
	board, err := a.app.GetBoard(invitation.BoardID)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	// Get inviter info
	inviter, err := a.app.GetUser(invitation.CreatedBy)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	// Send invitation email
	inviterName := inviter.Username
	if inviter.FirstName != "" || inviter.LastName != "" {
		inviterName = strings.TrimSpace(inviter.FirstName + " " + inviter.LastName)
	}

	err = a.app.SendInvitationEmail(invitation.Email, board.Title, inviterName, invitation.Token)
	if err != nil {
		a.logger.Error("Failed to resend invitation email",
			mlog.String("email", invitation.Email),
			mlog.String("boardID", invitation.BoardID),
			mlog.Err(err))
		
		a.errorResponse(w, r, model.NewErrBadRequest("Failed to resend invitation email: "+err.Error()))
		return
	}

	// Update invitation with new last sent timestamp
	now := time.Now().Unix()
	invitation.LastSentAt = &now
	updateErr := a.app.UpdateBoardInvitation(invitation)
	if updateErr != nil {
		a.logger.Error("Failed to update invitation timestamp",
			mlog.String("invitationID", invitation.ID),
			mlog.Err(updateErr))
		// Don't fail the request, email was sent successfully
	}

	a.logger.Debug("Invitation resent successfully",
		mlog.String("boardID", invitation.BoardID),
		mlog.String("email", invitation.Email),
		mlog.String("resentBy", userID))

	jsonStringResponse(w, http.StatusOK, "{}")
	auditRec.Success()
}

func (a *API) handleDeleteInvitation(w http.ResponseWriter, r *http.Request) {
	// swagger:operation DELETE /invitations/{invitationID} deleteInvitation
	//
	// Delete a board invitation
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: invitationID
	//   in: path
	//   description: Invitation ID
	//   required: true
	//   type: string
	// security:
	// - BearerAuth: []
	// responses:
	//   '200':
	//     description: success
	//   '403':
	//     description: access denied
	//   '404':
	//     description: invitation not found
	//   default:
	//     description: internal error
	//     schema:
	//       "$ref": "#/definitions/ErrorResponse"

	invitationID := mux.Vars(r)["invitationID"]
	userID := getUserID(r)

	auditRec := a.makeAuditRecord(r, "deleteInvitation", audit.Fail)
	defer a.audit.LogRecord(audit.LevelModify, auditRec)
	auditRec.AddMeta("invitationID", invitationID)

	// Get invitation
	invitation, err := a.app.GetBoardInvitationByID(invitationID)
	if err != nil {
		a.errorResponse(w, r, model.NewErrNotFound("invitation not found"))
		return
	}

	// Check if user has permission to manage board roles
	if !a.permissions.HasPermissionToBoard(userID, invitation.BoardID, model.PermissionManageBoardRoles) {
		a.errorResponse(w, r, model.NewErrPermission("access denied to delete invitation"))
		return
	}

	// Delete invitation
	err = a.app.DeleteBoardInvitation(invitationID)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	a.logger.Debug("Invitation deleted successfully",
		mlog.String("boardID", invitation.BoardID),
		mlog.String("email", invitation.Email),
		mlog.String("deletedBy", userID))

	jsonStringResponse(w, http.StatusOK, "{}")
	auditRec.Success()
}