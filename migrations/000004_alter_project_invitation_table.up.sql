ALTER TABLE project_invitations
ALTER COLUMN accepted_at DROP NOT NULL,
ALTER COLUMN accepted_at SET DEFAULT NULL;
