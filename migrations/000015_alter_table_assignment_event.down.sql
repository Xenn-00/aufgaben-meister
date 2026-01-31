ALTER TABLE aufgaben_assignment_events
    DROP COLUMN task_archived_at,
    DROP COLUMN task_archived_by;

ALTER TABLE aufgaben
    RENAME COLUMN archived_at TO deleted_at,
    DROP COLUMN archived_by;