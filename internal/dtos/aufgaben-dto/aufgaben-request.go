package aufgaben_dto

import (
	"time"

	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	"github.com/go-playground/validator/v10"
)

type CreateNewAufgabenRequest struct {
	Title       string     `json:"title" validate:"required"`
	Description *string    `json:"description,omitempty"`
	Priority    *string    `json:"priority,omitempty" validate:"omitempty,aufgabenPriority"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	AssigneeID  *string    `json:"assignee_id,omitempty" validate:"omitempty,uuid"`
}

type AufgabenListFilter struct {
	Status     *string `query:"status,omitempty" validate:"omitempty,aufgabenStatus"`
	AssigneeID *string `query:"assigned_id,omitempty" validate:"omitempty,uuid"`
	Limit      int     `query:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Page       int     `query:"page,omitempty" validate:"omitempty,min=1,max=100"`
}

type AssignedAufgabenFilter struct {
	Status    *string `query:"status,omitempty" validate:"omitempty,aufgabenStatus"`
	Priority  *string `query:"priority,omitempty" validate:"omitempty,aufgabenPriority"`
	ProjectID *string `query:"project_id,omitempty" validate:"omitempty,uuid"`
	Limit     int     `query:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Cursor    *string `query:"cursor,omitempty" validate:"omitempty,uuid"`
}

type ParamTaskID struct {
	ID string `params:"task_id" validate:"required,uuid"`
}

type AufgabenAssignRequest struct {
	DueDate time.Time `json:"due_date" validate:"required"`
}

type AufgabenUpdateProgressRequest struct {
	Status string `json:"status" validate:"required,aufgabenStatus"`
}

type UnassignAufgabenRequest struct {
	Note       *string `json:"note,omitempty" validate:"omitempty,min=3"`
	ReasonCode string  `json:"reason_code" validate:"required,reasonCode"`
	Reason     string  `json:"reason" validate:"required"`
}

type ForceUnassignAufgabenRequest struct {
	TargetID   string  `json:"target_id" validate:"required,uuid"`
	Note       *string `json:"note,omitempty" validate:"omitempty,min=3"`
	ReasonCode string  `json:"reason_code" validate:"required,reasonCode"`
	Reason     string  `json:"reason" validate:"required"`
}

type ReassignAufgabenRequest struct {
	TargetID   string  `json:"target_id" validate:"required,uuid"`
	Note       string  `json:"note" validate:"required"`
	Reason     *string `json:"reason,omitempty" validate:"omitempty,min=3"`
	ReasonCode *string `json:"reason_code,omitempty" validate:"omitempty,reasonCode"`
}

func IsValidReasonCode(fl validator.FieldLevel) bool {
	v := fl.Field().String()
	switch entity.ReasonCodeEvent(v) {
	case entity.ReasonOverload, entity.ReasonBlocked, entity.ReasonSick, entity.ReasonOther:
		return true
	default:
		return false
	}
}

func IsValidAufgabenStatus(fl validator.FieldLevel) bool {
	v := fl.Field().String()
	switch entity.AufgabenStatus(v) {
	case entity.AufgabenTodo, entity.AufgabenInProgress, entity.AufgabenDone, entity.AufgabenArchived:
		return true
	default:
		return false
	}
}

func IsValidAufgabenPriority(fl validator.FieldLevel) bool {
	v := fl.Field().String()
	switch entity.AufgabenPriority(v) {
	case entity.PriorityLow, entity.PriorityMedium, entity.PriorityHigh, entity.PriorityUrgent:
		return true
	default:
		return false
	}
}
