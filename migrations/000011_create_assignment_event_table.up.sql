-- ENUM TYPE FOR ACTION ASSIGNMENT EVENTS
CREATE TYPE action_events AS ENUM ('Assign', 'Unassign', 'Progress', 'Complete');
-- ENUM TYPE FOR REASON CODE ASSIGNMENT EVENTS
CREATE TYPE reason_code_event AS ENUM ('Overload', 'Blocked', 'Sick', 'Other');

-- AUFGABEN ASSIGNMENT EVENTS
CREATE TABLE aufgaben_assignment_events (
    id UUID PRIMARY KEY,
    aufgaben_id UUID NOT NULL REFERENCES aufgaben(id) ON DELETE CASCADE,

    actor_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_assignee_id UUID NULL REFERENCES users(id),

    action action_events NOT NULL DEFAULT 'Assign',
    note TEXT NULL, 

    reason_code reason_code_event NULL,
    reason_text TEXT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);