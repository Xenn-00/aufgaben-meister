-- ENUM TYPE FOR AUFGABEN STATUS
CREATE TYPE aufgaben_status AS ENUM ('Todo', 'In_Progress', 'Done', 'Archived');

-- ENUM TYPE FOR AUFGABEN PRIORITY
CREATE TYPE aufgaben_priority AS ENUM ('Low', 'Medium', 'High', 'Urgent');

-- AUFGABEN
CREATE TABLE aufgaben (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    
    title VARCHAR(255) NOT NULL,
    description TEXT NULL,
    
    status aufgaben_status NOT NULL DEFAULT 'Todo',
    priority aufgaben_priority NOT NULL DEFAULT 'Medium',

    assignee_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    
    due_date TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ NULL,
    deleted_at TIMESTAMPTZ NULL
);

-- TRIGGER
CREATE TRIGGER set_timestamp_aufgaben
BEFORE UPDATE ON aufgaben
FOR EACH ROW
EXECUTE FUNCTION set_timestamp();

-- INDEX
CREATE INDEX idx_aufgaben_project ON aufgaben(project_id);
CREATE INDEX idx_aufgaben_project_status ON aufgaben(project_id, status);
CREATE INDEX idx_aufgaben_assignee ON aufgaben(assignee_id);
CREATE INDEX idx_aufgaben_due_date ON aufgaben(due_date);
CREATE INDEX idx_aufgaben_not_deleted ON aufgaben(id) WHERE deleted_at IS NULL;