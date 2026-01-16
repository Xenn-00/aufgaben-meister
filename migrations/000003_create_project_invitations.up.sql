-- ENUM TYPE FOR INVITATION STATUS
CREATE TYPE invitation_status_enum AS ENUM ('Pending', 'Accepted', 'Revoked', 'Expired');

-- PROJECT INVITATIONS
CREATE TABLE project_invitations (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    invited_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    invited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    role user_role_enum NOT NULL DEFAULT 'Mitarbeiter',

    status invitation_status_enum NOT NULL DEFAULT 'Pending',
    token_hash TEXT NOT NULL,

    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    accepted_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),

    CONSTRAINT unique_active_invite UNIQUE (project_id, invited_user_id)
);