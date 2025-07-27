DROP TABLE IF EXISTS {{.prefix}}password_reset_tokens;

ALTER TABLE {{.prefix}}users DROP COLUMN IF EXISTS last_password_reset_at;