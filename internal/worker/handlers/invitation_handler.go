package worker_handler

import (
	"context"
	"fmt"

	"github.com/Xenn-00/aufgaben-meister/internal/entity"
	"github.com/Xenn-00/aufgaben-meister/internal/mail"
	project_repo "github.com/Xenn-00/aufgaben-meister/internal/repo/project-repo"
	worker_task "github.com/Xenn-00/aufgaben-meister/internal/worker/tasks"
	"github.com/goccy/go-json"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type WorkerHander struct {
	db     *pgxpool.Pool
	pr     project_repo.ProjectRepoContract
	mailer mail.Mailer
}

func NewWorkerHandler(db *pgxpool.Pool, mailer mail.Mailer) *WorkerHander {
	return &WorkerHander{
		db:     db,
		pr:     project_repo.NewUserRepo(db),
		mailer: mailer,
	}
}

func (wh *WorkerHander) WorkerInvitationEmailHandler() asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		log.Info().Msg("Worker handler: Worker invitation email hit.")
		var p worker_task.SendInvitationEmailPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			log.Error().Err(err).Msg("Worker handler: Error occured when trying to unmarshal task payload.")
			return err
		}

		inv, err := wh.pr.GetInvitationInfo(ctx, p.InvitationID)
		if err != nil {
			log.Error().Err(err).Msg("Worker handler: Error occured when trying to call repo -> GetInvitationInfo.")
			return err
		}

		if inv.InvitationStatus != string(entity.PENDING) {
			return nil // idempotent
		}

		link := fmt.Sprintf("http://localhost:8080/api/v1/project/invite/accept?invitation_id=%s&token=%s", p.InvitationID, p.RawToken) // we can change it to proper link

		log.Info().Msg("Worker handler: Preparing to hit SendInfitationEmail service.")
		return wh.mailer.SendInvitationEmail(inv.UserEmail, inv.ProjectName, link)

	}
}

func (wh *WorkerHander) WorkerInvitationExpireHandler() asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		log.Info().Msg("Worker handler: Worker invitation expire hit.")
		// begin transaction
		tx, txErr := wh.db.BeginTx(ctx, pgx.TxOptions{})
		if txErr != nil {
			log.Error().Err(txErr).Msg("Worker handler: Failed to open db transaction")
			return nil
		}
		defer tx.Rollback(ctx)
		// list invitations that expired
		invIDs, err := wh.pr.ListInvitationsExpire(ctx, tx)
		if err != nil {
			log.Error().Err(err).Msg("Worker handler: Error occured when list invitations")
			return nil
		}
		if len(invIDs) == 0 {
			return nil
		}
		// update invitations status
		if err := wh.pr.UpdateInvitationsExpire(ctx, tx, invIDs); err != nil {
			log.Error().Err(err).Msg("Worker handler: Failed to update invitations status")
			return nil
		}

		// commit
		if err := tx.Commit(ctx); err != nil {
			log.Error().Err(err).Msg("Worker handler: Error when initiating commit transaction")
			return nil
		}

		return nil
	}
}
