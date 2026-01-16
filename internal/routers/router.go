package routers

import (
	"github.com/Xenn-00/aufgaben-meister/internal/i18n"
	"github.com/Xenn-00/aufgaben-meister/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type CfgRedisStorage struct {
	Host     string
	Password string
}

// SetupRoutes richtet die API-Routen ein.
func SetupRoutes(app *fiber.App, db *pgxpool.Pool, redis *redis.Client, i18n *i18n.I18nService, paseto *utils.PasetoMaker, cfgStorage CfgRedisStorage) {
	api := app.Group("/api/v1")

	AuthRouter(api, db, redis, i18n, paseto)
	UserRouter(api, db, redis, i18n, paseto)
	ProjectRouter(api, db, redis, i18n, paseto, cfgStorage)
	HealthRouter(api, db, redis)
}
