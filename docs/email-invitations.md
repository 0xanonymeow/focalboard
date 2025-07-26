# Email Invitations Configuration

Focalboard supports sending email invitations to invite users to collaborate on boards. This feature supports both SMTP and Postmark email providers.

## Prerequisites

- Focalboard server with `enablePublicSharedBoards` set to `true`
- Email provider configuration (SMTP or Postmark)
- Valid email configuration settings

## Configuration

### Using SMTP

To configure SMTP email sending, set the following environment variables:

```bash
# SMTP server configuration
FOCALBOARD_EMAIL_SMTP_SERVER=smtp.gmail.com
FOCALBOARD_EMAIL_SMTP_PORT=587
FOCALBOARD_EMAIL_SMTP_USERNAME=your-email@gmail.com
FOCALBOARD_EMAIL_SMTP_PASSWORD=your-app-password
FOCALBOARD_EMAIL_SMTP_TLS=true

# Required email settings
FOCALBOARD_EMAIL_FROM_EMAIL=noreply@your-domain.com
FOCALBOARD_EMAIL_FROM_NAME=Your Organization
```

### Using Postmark

To configure Postmark email sending, set the following environment variables:

```bash
# Postmark configuration
FOCALBOARD_EMAIL_POSTMARK_API_TOKEN=your-postmark-server-token

# Required email settings
FOCALBOARD_EMAIL_FROM_EMAIL=noreply@your-domain.com
FOCALBOARD_EMAIL_FROM_NAME=Your Organization
```

### Configuration File

Alternatively, you can configure email settings in your `config.json` file:

```json
{
  "emailConfig": {
    "smtpServer": "smtp.gmail.com",
    "smtpPort": 587,
    "smtpUsername": "your-email@gmail.com",
    "smtpPassword": "your-app-password",
    "smtpTls": true,
    "postmarkApiToken": "",
    "fromEmail": "noreply@your-domain.com",
    "fromName": "Focalboard",
    "inviteSubject": "You've been invited to join a board"
  }
}
```

## Email Provider Setup

### Gmail SMTP

1. Enable 2-factor authentication on your Gmail account
2. Generate an app password at https://myaccount.google.com/apppasswords
3. Use the app password in the `FOCALBOARD_EMAIL_SMTP_PASSWORD` setting

### Postmark

1. Sign up for a Postmark account at https://postmarkapp.com
2. Create a server and get your Server API Token
3. Set up your sender signature or domain authentication
4. Use the Server API Token in the `FOCALBOARD_EMAIL_POSTMARK_API_TOKEN` setting

### Other SMTP Providers

For other SMTP providers, configure the appropriate server settings:

- **Office 365**: `smtp.office365.com:587` with TLS
- **Yahoo Mail**: `smtp.mail.yahoo.com:587` with TLS
- **Outlook.com**: `smtp-mail.outlook.com:587` with TLS

## Using Email Invitations

Once configured, board administrators can:

1. Open the board sharing dialog
2. Use the "Invite by email" section
3. Enter an email address and select a role (Viewer, Commenter, Editor, Admin)
4. Click "Send Invitation"

The invited user will receive an email with a link to join the board. When they click the link and log in (or create an account), they will be automatically added to the board with the specified role.

## API Usage

You can also send invitations programmatically using the REST API:

```bash
curl -X POST \
  http://your-focalboard-server/api/v2/boards/{boardId}/invite \
  -H 'Authorization: Bearer your-auth-token' \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "user@example.com",
    "role": "viewer"
  }'
```

## Troubleshooting

### Email not sending

1. Check that email configuration is correct
2. Verify firewall/network settings allow SMTP traffic
3. Check server logs for email sending errors
4. Test email configuration with a simple test message

### Invalid invitation links

1. Ensure `serverRoot` is configured correctly in your Focalboard configuration
2. Check that the invitation hasn't expired (default: 7 days)
3. Verify the user's email matches the invitation email exactly

### Permission issues

1. Ensure the sender has `ManageBoardRoles` permission on the board
2. Check that `enablePublicSharedBoards` is set to `true`
3. Verify the board allows the specified role for invitations

## Email Templates

Email invitations use customizable templates located in `./templates/email/`:

- `invitation.html` - HTML email template  
- `invitation.txt` - Plain text email template
- `invitation_subject.txt` - Email subject template

Templates support the following variables:
- `{{.InviterName}}` - Name of the person sending invitation
- `{{.BoardTitle}}` - Name of the board  
- `{{.InviteURL}}` - Secure invitation acceptance link
- `{{.FromName}}` - Configured sender name

If templates are not found, the system uses built-in defaults.

## Security Considerations

- Use TLS/SSL encryption for SMTP connections
- Store email credentials securely (use environment variables, not config files in production)
- Set appropriate invitation expiry times
- Monitor invitation usage to prevent abuse
- Use strong authentication for email providers (app passwords, API tokens)

## Invitation Lifecycle

1. **Creation**: Invitation created with 7-day expiry
2. **Email Sent**: Invitation email sent to recipient
3. **Acceptance**: User clicks link and logs in/registers
4. **Board Access**: User automatically added to board with specified role
5. **Cleanup**: Expired invitations are automatically cleaned up