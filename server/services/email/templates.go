package email

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// EmailTemplates holds parsed email templates
type EmailTemplates struct {
	InvitationHTML *template.Template
	InvitationText *template.Template
	SubjectText    string
}

// InvitationData contains data for invitation email templates
type InvitationData struct {
	BoardTitle  string
	InviterName string
	InviteURL   string
	FromName    string
}

// LoadTemplates loads email templates from the templates directory
func LoadTemplates(templatesPath string) (*EmailTemplates, error) {
	if templatesPath == "" {
		templatesPath = "./templates/email"
	}

	templates := &EmailTemplates{}
	
	// Load HTML template
	htmlPath := filepath.Join(templatesPath, "invitation.html")
	if _, err := os.Stat(htmlPath); err == nil {
		htmlTemplate, err := template.ParseFiles(htmlPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse HTML template: %w", err)
		}
		templates.InvitationHTML = htmlTemplate
	}
	
	// Load text template
	textPath := filepath.Join(templatesPath, "invitation.txt")
	if _, err := os.Stat(textPath); err == nil {
		textTemplate, err := template.ParseFiles(textPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse text template: %w", err)
		}
		templates.InvitationText = textTemplate
	}
	
	// Load subject template
	subjectPath := filepath.Join(templatesPath, "invitation_subject.txt")
	if _, err := os.Stat(subjectPath); err == nil {
		subjectBytes, err := os.ReadFile(subjectPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read subject template: %w", err)
		}
		templates.SubjectText = strings.TrimSpace(string(subjectBytes))
	}

	// Use defaults if templates don't exist
	if templates.InvitationHTML == nil {
		htmlTemplate, err := template.New("invitation_html").Parse(defaultHTMLTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse default HTML template: %w", err)
		}
		templates.InvitationHTML = htmlTemplate
	}
	
	if templates.InvitationText == nil {
		textTemplate, err := template.New("invitation_text").Parse(defaultTextTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse default text template: %w", err)
		}
		templates.InvitationText = textTemplate
	}
	
	if templates.SubjectText == "" {
		templates.SubjectText = defaultSubject
	}

	return templates, nil
}

// RenderHTMLTemplate renders the HTML invitation template
func (t *EmailTemplates) RenderHTMLTemplate(data InvitationData) (string, error) {
	var buf strings.Builder
	err := t.InvitationHTML.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to render HTML template: %w", err)
	}
	return buf.String(), nil
}

// RenderTextTemplate renders the text invitation template
func (t *EmailTemplates) RenderTextTemplate(data InvitationData) (string, error) {
	var buf strings.Builder
	err := t.InvitationText.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to render text template: %w", err)
	}
	return buf.String(), nil
}

// RenderSubject renders the subject line
func (t *EmailTemplates) RenderSubject(data InvitationData) string {
	// Simple string replacement for subject
	subject := t.SubjectText
	subject = strings.ReplaceAll(subject, "{{.BoardTitle}}", data.BoardTitle)
	subject = strings.ReplaceAll(subject, "{{.InviterName}}", data.InviterName)
	subject = strings.ReplaceAll(subject, "{{.FromName}}", data.FromName)
	return subject
}

// Default templates (used as fallbacks)
const defaultHTMLTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Board Invitation</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #f8f9fa; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .content { padding: 20px 0; }
        .button { 
            display: inline-block; 
            background-color: #007bff; 
            color: white; 
            padding: 12px 24px; 
            text-decoration: none; 
            border-radius: 5px; 
            margin: 20px 0; 
        }
        .footer { color: #666; font-size: 12px; margin-top: 30px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>You've been invited to join a board!</h1>
        </div>
        <div class="content">
            <p>Hi there!</p>
            <p><strong>{{.InviterName}}</strong> has invited you to collaborate on the board <strong>"{{.BoardTitle}}"</strong> in Focalboard.</p>
            <p>Click the button below to accept the invitation and start collaborating:</p>
            <a href="{{.InviteURL}}" class="button">Accept Invitation</a>
            <p>If the button doesn't work, copy and paste this link into your browser:</p>
            <p><a href="{{.InviteURL}}">{{.InviteURL}}</a></p>
        </div>
        <div class="footer">
            <p>This invitation was sent by {{.FromName}}. If you weren't expecting this invitation, you can safely ignore this email.</p>
            <p>Powered by Focalboard</p>
        </div>
    </div>
</body>
</html>`

const defaultTextTemplate = `You've been invited to join a board!

Hi there!

{{.InviterName}} has invited you to collaborate on the board "{{.BoardTitle}}" in Focalboard.

To accept the invitation and start collaborating, visit this link:
{{.InviteURL}}

If you weren't expecting this invitation, you can safely ignore this email.

This invitation was sent by {{.FromName}}.
Powered by Focalboard`

const defaultSubject = `You've been invited to join "{{.BoardTitle}}"`