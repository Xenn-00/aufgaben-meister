package aufgaben_dto

import "time"

type CreateNewAufgabenResponse struct {
	AufgabenID  string     `json:"aufgaben_id"`
	ProjectID   string     `json:"project_id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	AssigneeID  *string    `json:"assignee_id,omitempty"`
	CreateAt    time.Time  `json:"created_at"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

type AufgabenItem struct {
	AufgabenID  string     `json:"aufgaben_id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	AssigneeID  *string    `json:"assignee_id,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

type AssignedAufgabenListItem struct {
	AufgabenID  string    `json:"aufgaben_id"`
	ProjectName string    `json:"project_name"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	DueDate     time.Time `json:"due_date"`
}

type AufgabenAssignResponse struct {
	AufgabenID string    `json:"aufgaben_id"`
	ProjectID  string    `json:"project_id"`
	Status     string    `json:"status"`
	Priority   string    `json:"priority"`
	AssigneeID string    `json:"assignee_id,omitempty"`
	CreatedBy  string    `json:"created_by"`
	DueDate    time.Time `json:"due_date"`
}

type AufgabenForwardProgressResponse struct {
	AufgabenID  string    `json:"aufgaben_id"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	CreatedBy   string    `json:"created_by"`
	CompletedAt time.Time `json:"completed_at"`
}

type UnassignAufgabenResponse struct {
	AufgabenID string `json:"aufgaben_id"`
	Status     string `json:"status"`
	Note       string `json:"note"`
	Action     string `json:"action"`
	Reason     string `json:"reason"`
}

type ReassignAufgabenResponse struct {
	AufgabenID    string  `json:"aufgaben_id"`
	NewAssigneeID *string `json:"new_assignee_id"`
	Status        string  `json:"status,omitempty"`
	Note          string  `json:"note"`
	Action        string  `json:"action"`
	Reason        *string `json:"reason,omitempty"`
}

type UpdateDueDateResponse struct {
	AufgabenID string    `json:"aufgaben_id"`
	DueDate    time.Time `json:"due_date"`
}

type AufgabenEventItem struct {
	EventID     string    `json:"event_id"`
	AufgabeID   string    `json:"aufgabe_id"`
	ActorID     string    `json:"actor_id"`
	EventAction string    `json:"event_action"`
	Note        *string   `json:"note,omitempty"`
	TargetID    *string   `json:"target_id,omitempty"`
	ReasonCode  *string   `json:"reason_code,omitempty"`
	ReasonText  *string   `json:"reason_text,omitempty"`
	EventTime   time.Time `json:"event_time"`
}
