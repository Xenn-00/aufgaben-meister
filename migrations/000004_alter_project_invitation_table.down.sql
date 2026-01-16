ALTER TABLE project_invitations
ALTER COLUMN accepted_at SET NOT NULL,
ALTER COLUMN accepted_at DROP DEFAULT;
