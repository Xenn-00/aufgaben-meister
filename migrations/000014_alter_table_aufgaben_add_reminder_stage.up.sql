CREATE TYPE reminder_stage AS ENUM ('None', 'Before_Due', 'Overdue');

ALTER TABLE aufgaben
    ADD COLUMN reminder_stage reminder_stage NOT NULL DEFAULT 'None',
    ADD CONSTRAINT check_reminder_stage_valid
CHECK (
    (status IN ('Done', 'Archived') AND reminder_stage = 'None')
    OR status NOT IN ('Done', 'Archived')
);