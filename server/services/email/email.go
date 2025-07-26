package email

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/mattermost/focalboard/server/services/config"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

// Provider defines the interface for email providers
type Provider interface {
	SendEmail(to, from, subject, htmlBody, textBody string) error
}

// Service handles email operations
type Service struct {
	config    *config.Configuration
	provider  Provider
	logger    mlog.LoggerIFace
	templates *EmailTemplates
}

// New creates a new email service
func New(cfg *config.Configuration, logger mlog.LoggerIFace) (*Service, error) {
	service := &Service{
		config: cfg,
		logger: logger,
	}

	// Load email templates
	templates, err := LoadTemplates(cfg.EmailConfig.TemplatesPath)
	if err != nil {
		logger.Warn("Failed to load email templates, using defaults", mlog.Err(err))
		// Continue with default templates
		templates, _ = LoadTemplates("")
	}
	service.templates = templates

	// Determine which provider to use based on configuration
	var provider Provider
	var providerErr error

	if cfg.EmailConfig.PostmarkAPIToken != "" {
		provider, providerErr = NewPostmarkProvider(cfg.EmailConfig.PostmarkAPIToken, logger)
		if providerErr != nil {
			return nil, fmt.Errorf("failed to create Postmark provider: %w", providerErr)
		}
		logger.Info("Email service initialized with Postmark provider")
	} else if cfg.EmailConfig.SMTPServer != "" {
		provider, providerErr = NewSMTPProvider(cfg.EmailConfig, logger)
		if providerErr != nil {
			return nil, fmt.Errorf("failed to create SMTP provider: %w", providerErr)
		}
		logger.Info("Email service initialized with SMTP provider")
	} else {
		return nil, fmt.Errorf("no email provider configured. Please set either SMTP or Postmark configuration")
	}

	service.provider = provider
	return service, nil
}

// SendInvitation sends a board invitation email
func (s *Service) SendInvitation(toEmail, boardTitle, inviterName, inviteToken, serverRoot string) error {
	if s.provider == nil {
		return fmt.Errorf("email provider not configured")
	}

	// Generate invitation link
	inviteURL := fmt.Sprintf("%s/invite/%s", strings.TrimSuffix(serverRoot, "/"), inviteToken)

	// Prepare template data
	data := InvitationData{
		BoardTitle:  boardTitle,
		InviterName: inviterName,
		InviteURL:   inviteURL,
		FromName:    s.config.EmailConfig.FromName,
	}

	// Render templates
	subject := s.templates.RenderSubject(data)
	
	htmlBody, err := s.templates.RenderHTMLTemplate(data)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	textBody, err := s.templates.RenderTextTemplate(data)
	if err != nil {
		return fmt.Errorf("failed to render text template: %w", err)
	}

	// Format from address with name
	fromAddr := fmt.Sprintf("%s <%s>", s.config.EmailConfig.FromName, s.config.EmailConfig.FromEmail)
	
	return s.provider.SendEmail(toEmail, fromAddr, subject, htmlBody, textBody)
}

// GenerateInviteToken generates a secure random token for invitations
func (s *Service) GenerateInviteToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}


// IsConfigured returns true if email service is properly configured
func (s *Service) IsConfigured() bool {
	return s.provider != nil && 
		s.config.EmailConfig.FromEmail != "" &&
		(s.config.EmailConfig.SMTPServer != "" || s.config.EmailConfig.PostmarkAPIToken != "")
}