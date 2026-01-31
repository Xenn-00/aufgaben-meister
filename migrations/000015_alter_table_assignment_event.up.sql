ALTER TABLE aufgaben_assignment_events
    ADD COLUMN task_archived_at TIMESTAMPTZ NULL,
    ADD COLUMN task_archived_by UUID NULL;

ALTER TABLE aufgaben
    RENAME COLUMN deleted_at TO archived_at;

ALTER TABLE aufgaben
    ADD COLUMN archived_by UUID NULL;

ALTER TYPE action_events ADD VALUE 'Task_Archived';
ALTER TYPE action_events ADD VALUE 'Due_Date_Updated';