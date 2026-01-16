package routers

import (
	auth_handlers "github.com/Xenn-00/aufgaben-meister/internal/handlers/auth"
	"github.com/Xenn-00/aufgaben-meister/internal/i18n"
	"github.com/Xenn-00/aufgaben-meister/internal/middleware"
	"github.com/Xenn-00/aufgaben-meister/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// AuthRouter richtet die Authentifizierungsrouten ein.
func AuthRouter(api fiber.Router, db *pgxpool.Pool, redis *redis.Client, i18n *i18n.I18nService, paseto *utils.PasetoMaker) {
	r := api.Group("/auth")
	authHandler := auth_handlers.NewAuthHandler(db, redis, i18n, paseto)
	r.Post("/registieren", authHandler.RegisterUser)
	r.Post("/anmelden", authHandler.LoginUser)
	r.Delete("/abmelden", middleware.AuthMiddleware(paseto, redis), authHandler.LogoutUser)
	r.Delete("/abmelden/alle", middleware.AuthMiddleware(paseto, redis), authHandler.LogoutAllDevices)
	r.Get("/device-auflisten", middleware.AuthMiddleware(paseto, redis), authHandler.ListAllUserDevices)
}
