package routers

import (
	user_handlers "github.com/Xenn-00/aufgaben-meister/internal/handlers/user"
	"github.com/Xenn-00/aufgaben-meister/internal/i18n"
	"github.com/Xenn-00/aufgaben-meister/internal/middleware"
	"github.com/Xenn-00/aufgaben-meister/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func UserRouter(api fiber.Router, db *pgxpool.Pool, redis *redis.Client, i18n *i18n.I18nService, paseto *utils.PasetoMaker) {
	r := api.Group("/user", middleware.AuthMiddleware(paseto, redis))
	userHandler := user_handlers.NewUserHandler(db, redis, i18n)
	r.Get("/me", userHandler.FetchUserSelfProfile)
	r.Get("/:id", userHandler.FetchUserProfile)
	r.Patch("/me", userHandler.UpdateSelfProfile)
	r.Post("/me/deactivate", userHandler.DeactivateSelfUser)
}
