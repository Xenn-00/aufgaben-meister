package worker_task

const TaskSendProjectInvitationEmail = "email:send_project_invitation"

type SendInvitationEmailPayload struct {
	InvitationID string `json:"invitation_id"`
	RawToken     string `json:"raw_token"`
}
