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

type AufgabenListItem struct {
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
	AufgabenID    string `json:"aufgaben_id"`
	NewAssigneeID string `json:"new_assignee_id"`
	Status        string `json:"status,omitempty"`
	Note          string `json:"note"`
	Action        string `json:"action"`
	Reason        string `json:"reason,omitempty"`
}
