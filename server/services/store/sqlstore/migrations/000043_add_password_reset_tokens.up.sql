CREATE TABLE IF NOT EXISTS {{.prefix}}password_reset_tokens (
    token VARCHAR(128) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    email VARCHAR(255) NOT NULL,
    create_at BIGINT NOT NULL,
    expires_at BIGINT NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    INDEX idx_user_id (user_id),
    INDEX idx_email (email),
    INDEX idx_expires_at (expires_at)
) {{if .mysql}}DEFAULT CHARACTER SET utf8mb4{{end}};

-- Add last_password_reset_at to track cooldowns per user
ALTER TABLE {{.prefix}}users ADD COLUMN IF NOT EXISTS last_password_reset_at BIGINT DEFAULT 0;