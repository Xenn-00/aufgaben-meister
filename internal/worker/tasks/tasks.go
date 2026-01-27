package worker_task

import "time"

const TaskInvitationExpire = "low:invitation_expire"

const TaskSendProjectInvitationEmail = "email:send_project_invitation"

const TaskOverdueAufgabenReminders = "low:overdue_aufgaben_reminders"

const TaskSendProjectProgressReminder = "low:send_project_progress_reminder"

const TaskHandoverRequestNotifyMeister = "email:handover_request_notify_meister"

type SendInvitationEmailPayload struct {
	InvitationID string `json:"invitation_id"`
	RawToken     string `json:"raw_token"`
}

type SendProjectProgressReminder struct {
	AufgabeID string `json:"aufgabe_id"`
}

type HandoverRequestNotifyMeister struct {
	AufgabeID        string    `json:"aufgabe_id"`
	AufgabeTitle     string    `json:"aufgabe_title"`
	AufgabeStatus    string    `json:"aufgabe_status"`
	ProjectID        string    `json:"project_id"`
	ProjectName      string    `json:"project_name"`
	AssigneeID       string    `json:"assignee_id"`
	TargetAssigneeID string    `json:"target_assignee_id"`
	RequestedAt      time.Time `json:"requested_at"`
	DueDate          time.Time `json:"due_date"`
	Note             string    `json:"note"`
}
