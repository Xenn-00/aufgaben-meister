package routers

import (
	"github.com/Xenn-00/aufgaben-meister/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// SetupRoutes richtet die API-Routen ein.
func SetupRoutes(app *fiber.App, db *pgxpool.Pool, redis *redis.Client, paseto *utils.PasetoMaker) {
	api := app.Group("/api/v1")

	AuthRouter(api, db, redis, paseto)
	HealthRouter(api, db)
}
