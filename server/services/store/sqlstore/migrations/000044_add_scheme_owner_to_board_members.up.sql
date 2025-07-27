{{if .mysql}}
ALTER TABLE {{.prefix}}board_members ADD COLUMN scheme_owner BOOLEAN DEFAULT FALSE;
{{end}}

{{if .postgres}}
ALTER TABLE {{.prefix}}board_members ADD COLUMN scheme_owner BOOLEAN DEFAULT FALSE;
{{end}}

{{if .sqlite}}
ALTER TABLE {{.prefix}}board_members ADD COLUMN scheme_owner BOOLEAN DEFAULT FALSE;
{{end}}

-- Ensure we have at least one owner per board
-- For existing boards, make the first admin the owner if no owner exists
{{if .mysql}}
UPDATE {{.prefix}}board_members bm1
SET scheme_owner = TRUE
WHERE scheme_admin = TRUE
AND NOT EXISTS (
    SELECT 1 FROM (SELECT * FROM {{.prefix}}board_members) bm2
    WHERE bm2.board_id = bm1.board_id
    AND bm2.scheme_owner = TRUE
)
AND bm1.user_id = (
    SELECT user_id FROM (SELECT * FROM {{.prefix}}board_members) bm3
    WHERE bm3.board_id = bm1.board_id
    AND bm3.scheme_admin = TRUE
    ORDER BY bm3.user_id
    LIMIT 1
);
{{end}}

{{if .postgres}}
UPDATE {{.prefix}}board_members bm1
SET scheme_owner = TRUE
WHERE scheme_admin = TRUE
AND NOT EXISTS (
    SELECT 1 FROM {{.prefix}}board_members bm2
    WHERE bm2.board_id = bm1.board_id
    AND bm2.scheme_owner = TRUE
)
AND bm1.user_id = (
    SELECT user_id FROM {{.prefix}}board_members bm3
    WHERE bm3.board_id = bm1.board_id
    AND bm3.scheme_admin = TRUE
    ORDER BY bm3.user_id
    LIMIT 1
);
{{end}}

{{if .sqlite}}
UPDATE {{.prefix}}board_members
SET scheme_owner = TRUE
WHERE scheme_admin = TRUE
AND NOT EXISTS (
    SELECT 1 FROM {{.prefix}}board_members bm2
    WHERE bm2.board_id = {{.prefix}}board_members.board_id
    AND bm2.scheme_owner = TRUE
)
AND user_id = (
    SELECT user_id FROM {{.prefix}}board_members bm3
    WHERE bm3.board_id = {{.prefix}}board_members.board_id
    AND bm3.scheme_admin = TRUE
    ORDER BY bm3.user_id
    LIMIT 1
);
{{end}}