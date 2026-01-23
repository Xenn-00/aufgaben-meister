package routers

import (
	aufgaben_handlers "github.com/Xenn-00/aufgaben-meister/internal/handlers/aufgaben"
	"github.com/Xenn-00/aufgaben-meister/internal/i18n"
	"github.com/Xenn-00/aufgaben-meister/internal/middleware"
	"github.com/Xenn-00/aufgaben-meister/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func AufgabenRouter(api fiber.Router, db *pgxpool.Pool, redis *redis.Client, i18n *i18n.I18nService, paseto *utils.PasetoMaker) {
	r := api.Group("/project/:project_id/aufgaben", middleware.AuthMiddleware(paseto, redis))
	aufgabenHandler := aufgaben_handlers.NewAufgabenHandler(db, redis, i18n)

	r.Post("/create", aufgabenHandler.CreateNewAufgaben)
	r.Get("/list", aufgabenHandler.ListTasks)
	r.Post("/:task_id/assign", aufgabenHandler.AssignTask)
	r.Patch("/:task_id/forward-progress", aufgabenHandler.ForwardProgress)
	r.Patch("/:task_id/unassign", aufgabenHandler.UnassignTask)
	r.Patch("/:task_id/force-unassign", aufgabenHandler.ForceUnassignTask)
	r.Patch("/:task_id/reassign", aufgabenHandler.ReassignTask)
}
