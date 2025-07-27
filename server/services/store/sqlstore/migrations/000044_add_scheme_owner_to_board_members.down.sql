{{if .mysql}}
ALTER TABLE {{.prefix}}board_members DROP COLUMN scheme_owner;
{{end}}

{{if .postgres}}
ALTER TABLE {{.prefix}}board_members DROP COLUMN scheme_owner;
{{end}}

{{if .sqlite}}
-- SQLite doesn't support DROP COLUMN directly in older versions
-- We need to recreate the table without the column
CREATE TABLE {{.prefix}}board_members_temp (
    board_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    roles TEXT,
    minimum_role TEXT,
    scheme_admin BOOLEAN,
    scheme_editor BOOLEAN,
    scheme_commenter BOOLEAN,
    scheme_viewer BOOLEAN,
    synthetic BOOLEAN,
    PRIMARY KEY (board_id, user_id)
);

INSERT INTO {{.prefix}}board_members_temp (board_id, user_id, roles, minimum_role, scheme_admin, scheme_editor, scheme_commenter, scheme_viewer, synthetic)
SELECT board_id, user_id, roles, minimum_role, scheme_admin, scheme_editor, scheme_commenter, scheme_viewer, synthetic
FROM {{.prefix}}board_members;

DROP TABLE {{.prefix}}board_members;

ALTER TABLE {{.prefix}}board_members_temp RENAME TO {{.prefix}}board_members;
{{end}}