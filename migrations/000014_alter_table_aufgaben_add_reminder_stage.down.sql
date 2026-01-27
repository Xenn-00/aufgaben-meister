ALTER TABLE aufgaben
    DROP CONSTRAINT check_reminder_stage_valid;
    
ALTER TABLE aufgaben
    DROP COLUMN reminder_stage;

DROP TYPE IF EXISTS reminder_stage;