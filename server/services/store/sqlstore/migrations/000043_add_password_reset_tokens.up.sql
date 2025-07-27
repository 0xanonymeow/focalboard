CREATE TABLE IF NOT EXISTS {{.prefix}}password_reset_tokens (
    token VARCHAR(128) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    email VARCHAR(255) NOT NULL,
    create_at BIGINT NOT NULL,
    expires_at BIGINT NOT NULL,
    used BOOLEAN DEFAULT FALSE{{if .mysql}},
    INDEX idx_user_id (user_id),
    INDEX idx_email (email),
    INDEX idx_expires_at (expires_at){{end}}
) {{if .mysql}}DEFAULT CHARACTER SET utf8mb4{{end}};

{{if .postgres}}
CREATE INDEX IF NOT EXISTS idx_password_reset_user_id ON {{.prefix}}password_reset_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_email ON {{.prefix}}password_reset_tokens (email);
CREATE INDEX IF NOT EXISTS idx_password_reset_expires_at ON {{.prefix}}password_reset_tokens (expires_at);
{{end}}

-- Add last_password_reset_at to track cooldowns per user
ALTER TABLE {{.prefix}}users ADD COLUMN IF NOT EXISTS last_password_reset_at BIGINT DEFAULT 0;