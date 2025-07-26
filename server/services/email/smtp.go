package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/mattermost/focalboard/server/services/config"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

// SMTPProvider handles email sending via SMTP
type SMTPProvider struct {
	config config.EmailConfig
	logger mlog.LoggerIFace
}

// NewSMTPProvider creates a new SMTP email provider
func NewSMTPProvider(cfg config.EmailConfig, logger mlog.LoggerIFace) (*SMTPProvider, error) {
	if cfg.SMTPServer == "" {
		return nil, fmt.Errorf("SMTP server not configured")
	}
	if cfg.FromEmail == "" {
		return nil, fmt.Errorf("from email not configured")
	}

	return &SMTPProvider{
		config: cfg,
		logger: logger,
	}, nil
}

// SendEmail sends an email via SMTP
func (p *SMTPProvider) SendEmail(to, from, subject, htmlBody, textBody string) error {
	// Prepare message
	message := p.buildMessage(to, from, subject, htmlBody, textBody)

	// SMTP server configuration
	host := p.config.SMTPServer
	port := p.config.SMTPPort
	addr := fmt.Sprintf("%s:%d", host, port)

	// Authentication
	var auth smtp.Auth
	if p.config.SMTPUsername != "" && p.config.SMTPPassword != "" {
		auth = smtp.PlainAuth("", p.config.SMTPUsername, p.config.SMTPPassword, host)
	}

	// Send email
	if p.config.SMTPTls {
		return p.sendTLS(addr, auth, to, message)
	}
	return p.sendPlain(addr, auth, to, message)
}

// sendTLS sends email with TLS encryption
func (p *SMTPProvider) sendTLS(addr string, auth smtp.Auth, to string, message []byte) error {
	// Connect to server
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: p.config.SMTPServer,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server with TLS: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, p.config.SMTPServer)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Authenticate if credentials provided
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	// Set sender
	if err = client.Mail(p.config.FromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send message
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to send message data: %w", err)
	}
	defer wc.Close()

	if _, err = wc.Write(message); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	p.logger.Debug("Email sent successfully via SMTP TLS", 
		mlog.String("to", to), 
		mlog.String("server", p.config.SMTPServer))
	
	return nil
}

// sendPlain sends email without TLS (for development/testing)
func (p *SMTPProvider) sendPlain(addr string, auth smtp.Auth, to string, message []byte) error {
	err := smtp.SendMail(addr, auth, p.config.FromEmail, []string{to}, message)
	if err != nil {
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	p.logger.Debug("Email sent successfully via SMTP", 
		mlog.String("to", to), 
		mlog.String("server", p.config.SMTPServer))
	
	return nil
}

// buildMessage constructs the email message with proper headers
func (p *SMTPProvider) buildMessage(to, from, subject, htmlBody, textBody string) []byte {
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "multipart/alternative; boundary=\"boundary123\""

	var message strings.Builder

	// Add headers
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")

	// Add multipart content
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(textBody)
	message.WriteString("\r\n\r\n")

	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n\r\n")

	message.WriteString("--boundary123--\r\n")

	return []byte(message.String())
}