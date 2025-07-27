// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/focalboard/server/model"
	"github.com/mattermost/focalboard/server/services/audit"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

func (a *API) registerPasswordResetRoutes(r *mux.Router) {
	// Password reset routes (no auth required)
	r.HandleFunc("/users/password/reset", a.handleRequestPasswordReset).Methods("POST")
	r.HandleFunc("/users/password/reset/{token}", a.handleResetPassword).Methods("POST")
}

func (a *API) handleRequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	// swagger:operation POST /api/v2/users/password/reset requestPasswordReset
	//
	// Request a password reset email
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: Body
	//   in: body
	//   description: Email to send reset link to
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/PasswordResetRequest"
	// responses:
	//   '200':
	//     description: success
	//   '400':
	//     description: invalid request
	//   '429':
	//     description: too many requests (cooldown active)
	//   '500':
	//     description: internal error
	//     schema:
	//       "$ref": "#/definitions/ErrorResponse"

	if !a.app.GetConfig().EnablePublicSharedBoards {
		a.errorResponse(w, r, model.NewErrUnauthorized("password reset is disabled"))
		return
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	var resetRequest model.PasswordResetRequest
	if err = json.Unmarshal(requestBody, &resetRequest); err != nil {
		a.errorResponse(w, r, model.NewErrBadRequest("invalid request body"))
		return
	}

	if err = resetRequest.IsValid(); err != nil {
		a.errorResponse(w, r, err)
		return
	}

	auditRec := a.makeAuditRecord(r, "requestPasswordReset", audit.Fail)
	defer a.audit.LogRecord(audit.LevelAuth, auditRec)
	auditRec.AddMeta("email", resetRequest.Email)

	// Get user by email
	user, err := a.app.GetUserByEmail(resetRequest.Email)
	if err != nil {
		// Don't reveal if email exists or not for security
		a.logger.Debug("Password reset requested for non-existent email",
			mlog.String("email", resetRequest.Email),
		)
		jsonStringResponse(w, http.StatusOK, "{}")
		auditRec.Success()
		return
	}

	// Check cooldown
	lastResetAt, err := a.app.GetUserLastPasswordResetAt(user.ID)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	timeSinceLastReset := model.GetMillis() - lastResetAt
	if timeSinceLastReset < int64(model.PasswordResetCooldown.Milliseconds()) {
		remainingTime := int64(model.PasswordResetCooldown.Milliseconds()) - timeSinceLastReset
		a.errorResponse(w, r, model.NewErrBadRequest(fmt.Sprintf("Please wait %d seconds before requesting another reset", remainingTime/1000)))
		return
	}

	// Generate token
	token := &model.PasswordResetToken{
		Token:     model.GeneratePasswordResetToken(),
		UserID:    user.ID,
		Email:     user.Email,
		CreateAt:  model.GetMillis(),
		ExpiresAt: model.GetMillis() + int64(model.PasswordResetTokenExpiry.Milliseconds()),
		Used:      false,
	}

	// Store token
	if err = a.app.CreatePasswordResetToken(token); err != nil {
		a.errorResponse(w, r, err)
		return
	}

	// Update last reset timestamp
	if err = a.app.UpdateUserLastPasswordResetAt(user.ID, model.GetMillis()); err != nil {
		a.logger.Error("Failed to update last password reset timestamp",
			mlog.String("userID", user.ID),
			mlog.Err(err),
		)
	}

	// Send email
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", a.app.GetConfig().ServerRoot, token.Token)

	if err = a.app.SendPasswordResetEmail(user.Email, resetURL); err != nil {
		a.logger.Error("Failed to send password reset email",
			mlog.String("email", user.Email),
			mlog.Err(err),
		)
		// Don't return error to user for security
	}

	jsonStringResponse(w, http.StatusOK, "{}")
	auditRec.Success()
}

func (a *API) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	// swagger:operation POST /api/v2/users/password/reset/{token} resetPassword
	//
	// Reset password using token
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: token
	//   in: path
	//   description: Password reset token
	//   required: true
	//   type: string
	// - name: Body
	//   in: body
	//   description: New password
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/ResetPasswordRequest"
	// responses:
	//   '200':
	//     description: success
	//   '400':
	//     description: invalid request
	//   '404':
	//     description: token not found or expired
	//   '500':
	//     description: internal error
	//     schema:
	//       "$ref": "#/definitions/ErrorResponse"

	vars := mux.Vars(r)
	token := vars["token"]

	if token == "" {
		a.errorResponse(w, r, model.NewErrBadRequest("token is required"))
		return
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	var resetRequest model.ResetPasswordRequest
	if err = json.Unmarshal(requestBody, &resetRequest); err != nil {
		a.errorResponse(w, r, model.NewErrBadRequest("invalid request body"))
		return
	}

	resetRequest.Token = token
	if err = resetRequest.IsValid(); err != nil {
		a.errorResponse(w, r, err)
		return
	}

	auditRec := a.makeAuditRecord(r, "resetPassword", audit.Fail)
	defer a.audit.LogRecord(audit.LevelAuth, auditRec)

	// Get token
	resetToken, err := a.app.GetPasswordResetToken(token)
	if err != nil {
		a.errorResponse(w, r, model.NewErrNotFound("invalid or expired token"))
		return
	}

	// Check if token is expired
	if resetToken.IsExpired() {
		a.errorResponse(w, r, model.NewErrNotFound("token has expired"))
		return
	}

	// Check if token is already used
	if resetToken.Used {
		a.errorResponse(w, r, model.NewErrNotFound("token has already been used"))
		return
	}

	// Update password
	if err = a.app.UpdateUserPasswordByID(resetToken.UserID, resetRequest.NewPassword); err != nil {
		a.errorResponse(w, r, err)
		return
	}

	// Mark token as used
	if err = a.app.UpdatePasswordResetTokenUsed(token); err != nil {
		a.logger.Error("Failed to mark password reset token as used",
			mlog.String("token", token),
			mlog.Err(err),
		)
	}

	// Clean up expired tokens
	go func() {
		if err := a.app.DeleteExpiredPasswordResetTokens(); err != nil {
			a.logger.Error("Failed to delete expired password reset tokens", mlog.Err(err))
		}
	}()

	auditRec.AddMeta("userID", resetToken.UserID)
	jsonStringResponse(w, http.StatusOK, "{}")
	auditRec.Success()
}