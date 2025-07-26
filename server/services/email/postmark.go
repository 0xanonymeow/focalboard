package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

// PostmarkProvider handles email sending via Postmark API
type PostmarkProvider struct {
	apiToken string
	logger   mlog.LoggerIFace
	client   *http.Client
}

// PostmarkMessage represents a Postmark email message
type PostmarkMessage struct {
	From     string `json:"From"`
	To       string `json:"To"`
	Subject  string `json:"Subject"`
	HTMLBody string `json:"HtmlBody"`
	TextBody string `json:"TextBody"`
	Tag      string `json:"Tag,omitempty"`
}

// PostmarkResponse represents Postmark API response
type PostmarkResponse struct {
	ErrorCode int    `json:"ErrorCode"`
	Message   string `json:"Message"`
	MessageID string `json:"MessageID"`
}

// NewPostmarkProvider creates a new Postmark email provider
func NewPostmarkProvider(apiToken string, logger mlog.LoggerIFace) (*PostmarkProvider, error) {
	if apiToken == "" {
		return nil, fmt.Errorf("Postmark API token not configured")
	}

	return &PostmarkProvider{
		apiToken: apiToken,
		logger:   logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// SendEmail sends an email via Postmark API
func (p *PostmarkProvider) SendEmail(to, from, subject, htmlBody, textBody string) error {
	// Prepare Postmark message
	message := PostmarkMessage{
		From:     from,
		To:       to,
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
		Tag:      "focalboard-invitation",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Postmark message: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.postmarkapp.com/email", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create Postmark request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", p.apiToken)
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Postmark request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var postmarkResp PostmarkResponse
	if err := json.NewDecoder(resp.Body).Decode(&postmarkResp); err != nil {
		return fmt.Errorf("failed to decode Postmark response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK || postmarkResp.ErrorCode != 0 {
		return fmt.Errorf("Postmark API error (code %d): %s", postmarkResp.ErrorCode, postmarkResp.Message)
	}

	p.logger.Debug("Email sent successfully via Postmark", 
		mlog.String("to", to), 
		mlog.String("messageId", postmarkResp.MessageID))

	return nil
}