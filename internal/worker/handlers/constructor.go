package worker_handler

import (
	"github.com/Xenn-00/aufgaben-meister/internal/mail"
	aufgaben_repo "github.com/Xenn-00/aufgaben-meister/internal/repo/aufgaben-repo"
	project_repo "github.com/Xenn-00/aufgaben-meister/internal/repo/project-repo"
	user_repo "github.com/Xenn-00/aufgaben-meister/internal/repo/user-repo"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkerHander struct {
	db     *pgxpool.Pool
	pr     project_repo.ProjectRepoContract
	ar     aufgaben_repo.AufgabenRepoContract
	ur     user_repo.UserRepoContract
	mailer mail.Mailer
}

func NewWorkerHandler(db *pgxpool.Pool, mailer mail.Mailer) *WorkerHander {
	return &WorkerHander{
		db:     db,
		pr:     project_repo.NewUserRepo(db),
		ar:     aufgaben_repo.NewAufgabenRepo(db),
		ur:     user_repo.NewUserRepo(db),
		mailer: mailer,
	}
}
