-- ENUM TYPE FOR USER ROLES, TEMPORARY
CREATE TYPE user_role_enum AS ENUM ('Meister', 'Mitarbeiter');
-- ENUM TYPE FOR PROJECT TYPES, TEMPORARY
CREATE TYPE project_type_enum AS ENUM ('Personal', 'Community', 'Corporate');
-- ENUM TYPE FOR PROJECT VISIBILITY, TOMPORARY
CREATE TYPE project_visibility_enum AS ENUM ('Public', 'Private');

-- USERS 
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    name VARCHAR(255) NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    role user_role_enum NOT NULL DEFAULT 'Mitarbeiter',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- PROJECTS (domain-level space)
CREATE TABLE projects (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type project_type_enum NOT NULL,
    visibility project_visibility_enum NOT NULL,
    master_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- USER-PROJECT RELATIONSHIP
CREATE TABLE project_members (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role user_role_enum NOT NULL DEFAULT 'Mitarbeiter',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT now(),

    CONSTRAINT unique_project_user UNIQUE (project_id, user_id) -- Ensure a user can't join the same project multiple times
);

-- TRIGGERS TO UPDATE UPDATED_AT FIELDS
CREATE OR REPLACE FUNCTION set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_timestamp_users
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION set_timestamp();

CREATE TRIGGER set_timestamp_projects
BEFORE UPDATE ON projects
FOR EACH ROW EXECUTE FUNCTION set_timestamp();


-- INDEXES
CREATE INDEX idx_project_master_id ON projects(master_id);
CREATE INDEX idx_project_members_pid_uid ON project_members(project_id, user_id);