CREATE TABLE IF NOT EXISTS {{.prefix}}board_invitations (
    id VARCHAR(36) PRIMARY KEY,
    board_id VARCHAR(36) NOT NULL,
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    role VARCHAR(32) NOT NULL DEFAULT 'viewer',
    created_by VARCHAR(36) NOT NULL,
    created_at BIGINT NOT NULL,
    expires_at BIGINT NOT NULL,
    used_at BIGINT,
    used_by VARCHAR(36)
);

CREATE INDEX IF NOT EXISTS {{.prefix}}board_invitations_board_id_idx ON {{.prefix}}board_invitations(board_id);
CREATE INDEX IF NOT EXISTS {{.prefix}}board_invitations_token_idx ON {{.prefix}}board_invitations(token);
CREATE INDEX IF NOT EXISTS {{.prefix}}board_invitations_email_idx ON {{.prefix}}board_invitations(email);
CREATE INDEX IF NOT EXISTS {{.prefix}}board_invitations_expires_at_idx ON {{.prefix}}board_invitations(expires_at);