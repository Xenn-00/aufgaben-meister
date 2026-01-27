package entity

import "time"

type AufgabenEntity struct {
	ID          string           `json:"id"`
	ProjectID   string           `json:"project_id"`
	ProjectName *string          `json:"project_name,omitempty"`
	Title       string           `json:"title"`
	Description *string          `json:"Description,omitempty"`
	Status      AufgabenStatus   `json:"status"`
	Priority    AufgabenPriority `json:"priority"`
	AssigneeID  *string          `json:"assignee_id,omitempty"`
	CreatedBy   string           `json:"created_by"`
	DueDate     *time.Time       `json:"due_date,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   *time.Time       `json:"updated_at,omitempty"`
	DeletedAt   *time.Time       `json:"deleted_at,omitempty"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
}

type AssignedAufgaben struct {
	ID          string           `json:"id"`
	ProjectName string           `json:"project_name"`
	Title       string           `json:"title"`
	Description *string          `json:"description,omitempty"`
	Status      AufgabenStatus   `json:"status"`
	Priority    AufgabenPriority `json:"priority"`
	DueDate     time.Time        `json:"due_date"`
}

type ReminderAufgaben struct {
	ID             string           `json:"id"`
	ProjectID      string           `json:"project_id"`
	ProjectName    string           `json:"project_name"`
	Title          string           `json:"title"`
	Status         AufgabenStatus   `json:"status"`
	Priority       AufgabenPriority `json:"priority"`
	AssigneeID     string           `json:"assignee_id"`
	EmailAssignee  string           `json:"assignee_email"`
	DueDate        time.Time        `json:"due_date"`
	LastReminderAt *time.Time       `json:"last_reminder_at,omitempty"`
}

type AssignmentEventEntity struct {
	ID               string           `json:"id"`
	AufgabenID       string           `json:"aufgaben_id"`
	ActorID          string           `json:"actor_id"`
	TargetAssigneeID *string          `json:"target_assignee_id,omitempty"`
	Action           ActionEvent      `json:"action"`
	Note             *string          `json:"note,omitempty"`
	ReasonCode       *ReasonCodeEvent `json:"reason_code,omitempty"`
	ReasonText       *string          `json:"reason_text,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
}

type AssignTaskEntity struct {
	ID         string           `json:"id"`
	Status     AufgabenStatus   `json:"status"`
	Priority   AufgabenPriority `json:"priority"`
	AssigneeID string           `json:"assignee_id,omitempty"`
	CreatedBy  string           `json:"created_by"`
	DueDate    time.Time        `json:"due_date"`
}

type UnassignTaskEntity struct {
	ID         string `json:"id"`
	AssigneeID string `json:"assignee_id"`
}

type CompleteTaskEntity struct {
	ID          string           `json:"id"`
	Status      AufgabenStatus   `json:"status"`
	Priority    AufgabenPriority `json:"priority"`
	AssigneeID  string           `json:"assignee_id,omitempty"`
	CreatedBy   string           `json:"created_by"`
	CompletedAt time.Time        `json:"completed_at"`
}

type AddAssignment struct {
	ID               string          `json:"id"`
	AufgabenID       string          `json:"aufgaben_id"`
	ActorID          string          `json:"actor_id"`
	TargetAssigneeID *string         `json:"target_assignee_id,omitempty"`
	Action           ActionEvent     `json:"action"`
	Note             *string         `json:"note,omitempty"`
	ReasonText       *string         `json:"reason_text,omitempty"`
	ReasonCode       ReasonCodeEvent `json:"reason_code,omitempty"`
}

type ActionEvent string

const (
	ActionAssign          ActionEvent = "Assign"
	ActionUnassign        ActionEvent = "Unassign"
	ActionProgress        ActionEvent = "Progress"
	ActionComplete        ActionEvent = "Complete"
	ActionHandoverRequest ActionEvent = "Handover_Request"
	ActionHandoverExecute ActionEvent = "Handover_Execute"
)

type ReasonCodeEvent string

const (
	ReasonOverload ReasonCodeEvent = "Overload"
	ReasonBlocked  ReasonCodeEvent = "Blocked"
	ReasonSick     ReasonCodeEvent = "Sick"
	ReasonOther    ReasonCodeEvent = "Other"
)

type AufgabenStatus string

const (
	AufgabenTodo       AufgabenStatus = "Todo"
	AufgabenInProgress AufgabenStatus = "In_Progress"
	AufgabenDone       AufgabenStatus = "Done"
	AufgabenArchived   AufgabenStatus = "Archived"
)

type AufgabenPriority string

const (
	PriorityLow    AufgabenPriority = "Low"
	PriorityMedium AufgabenPriority = "Medium"
	PriorityHigh   AufgabenPriority = "High"
	PriorityUrgent AufgabenPriority = "Urgent"
)
