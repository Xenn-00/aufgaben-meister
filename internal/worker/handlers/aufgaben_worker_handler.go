package worker_handler

import (
	"context"

	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	worker_task "github.com/Xenn-00/aufgaben-meister/internal/worker/tasks"
	"github.com/goccy/go-json"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

func (wh *WorkerHander) OverdueAufgaben() asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		// TODO
		// List all aufgaben that match condition now() >= a.due_date
		aufgaben, err := wh.ar.ListShouldRemindOverdue(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Worker handler: Error occured when list aufgaben")
			return err
		}
		// When there is no matches, do nothing
		if len(aufgaben) == 0 {
			return nil
		}
		// If there are matches, start transaction
		tx, txErr := wh.txManager.Begin(ctx)
		if txErr != nil {
			log.Error().Err(txErr).Msg("Worker handler: Failed to open db transaction")
			return txErr
		}
		defer tx.Rollback(ctx)
		// Send email to assignee from their related contact (email)
		aufgabenID := []string{}
		for _, aufgabe := range aufgaben {
			if err := wh.mailer.SendReminderAufgabenOverdue(&aufgabe); err != nil {
				log.Error().Err(err).Msg("Worker handler: Error occured when trying to send email.")
				continue
			}

			aufgabenID = append(aufgabenID, aufgabe.ID)
		}

		// Update value last_reminder_at (batch)
		if err := wh.ar.BatchUpdateAufgabenReminderOverdue(ctx, tx, aufgabenID); err != nil {
			log.Error().Err(err).Msg("Worker handler: Error occured when update aufgaben")

			return err
		}

		// Commit
		if err := tx.Commit(ctx); err != nil {
			log.Error().Err(err).Msg("Worker handler: Error when initiating commit transaction")
			return err
		}

		return nil
	}
}

func (wh *WorkerHander) ReminderAufgaben() asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		// TODO
		// List all aufgaben that match condition now() >= due_date - interval 1 hour
		var p worker_task.SendProjectProgressReminder
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}

		aufgabe, err := wh.ar.ShouldRemind(ctx, p.AufgabeID)
		if err != nil {
			log.Error().Err(err).Msg("Worker handler: Error occured when list aufgaben")
			return err
		}
		// Idempotency check
		if aufgabe.LastReminderAt != nil {
			return nil
		}

		// safety check
		if aufgabe.Status != entity.AufgabenInProgress {
			return nil
		}

		// If there are matches, start transaction
		tx, txErr := wh.txManager.Begin(ctx)
		if txErr != nil {
			log.Error().Err(txErr).Msg("Worker handler: Failed to open db transaction")
			return nil
		}
		defer tx.Rollback(ctx)

		// Send email to assignee from their related contact (email)
		if err := wh.mailer.SendReminderAufgabenProgress(aufgabe); err != nil {
			log.Error().Err(err).Msg("Worker handler: Error occured when trying to send email.")
			return nil
		}

		if err := wh.ar.UpdateAufgabeReminderBeforeDue(ctx, tx, aufgabe.ID); err != nil {
			log.Error().Err(err).Msg("Worker handler: Error occured when trying to update aufgabe.")
			return nil
		}

		// Commit
		if err := tx.Commit(ctx); err != nil {
			log.Error().Err(err).Msg("Worker handler: Error when initiating commit transaction")
			return nil
		}

		return nil
	}
}

func (wh *WorkerHander) HandoverRequestNotifyMeister() asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p worker_task.HandoverRequestNotifyMeister
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			log.Error().Err(err).Msg("Worker handler: Error occured when trying to unmarshal task payload.")
			return err
		}

		// Get Project
		project, err := wh.pr.GetProjectByID(ctx, p.ProjectID)
		if err != nil {
			log.Error().Err(err).Msg("Worker handler: error occured when fetch project info")
			return err
		}

		// Get meister info
		meister, err := wh.ur.FindByUserID(ctx, project.MasterID)
		if err != nil {
			log.Error().Err(err).Msg("Worker handler: error occured when fetch meister info")
		}

		// Get assignee info
		assignee, err := wh.ur.FindByUserID(ctx, p.AssigneeID)

		return wh.mailer.SendHandoverRequest(&p, meister.Email, assignee.Username)
	}
}
