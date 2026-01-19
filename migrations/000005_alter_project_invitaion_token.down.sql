ALTER TABLE project_invitations
ALTER COLUMN token_hash SET NOT NULL,
ALTER COLUMN token_hash DROP DEFAULT;