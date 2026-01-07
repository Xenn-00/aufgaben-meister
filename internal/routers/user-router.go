package routers

import (
	"github.com/Xenn-00/aufgaben-meister/internal/handlers"
	user_handler "github.com/Xenn-00/aufgaben-meister/internal/handlers/user"
	"github.com/Xenn-00/aufgaben-meister/internal/middleware"
	"github.com/Xenn-00/aufgaben-meister/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func UserRouter(api fiber.Router, db *pgxpool.Pool, redis *redis.Client, paseto *utils.PasetoMaker) {
	r := api.Group("/user", middleware.AuthMiddleware(paseto, redis))
	userHandler := user_handler.NewUserHandler(db, redis)
	r.Get("/me", handlers.Wrap(userHandler.FetchUserSelfProfile))
	r.Get("/:id", handlers.Wrap(userHandler.FetchUserProfile))
	r.Patch("/me", handlers.Wrap(userHandler.UpdateSelfProfile))
	r.Post("/me/deactivate", handlers.Wrap(userHandler.DeactivateSelfUser))
}
