package routers

import (
	"github.com/Xenn-00/aufgaben-meister/internal/handlers"
	auth_handlers "github.com/Xenn-00/aufgaben-meister/internal/handlers/auth"
	"github.com/Xenn-00/aufgaben-meister/internal/middleware"
	"github.com/Xenn-00/aufgaben-meister/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// AuthRouter richtet die Authentifizierungsrouten ein.
func AuthRouter(api fiber.Router, db *pgxpool.Pool, redis *redis.Client, paseto *utils.PasetoMaker) {
	r := api.Group("/auth")
	authHandler := auth_handlers.NewAuthHandler(db, redis, paseto)
	r.Post("/registieren", handlers.Wrap(authHandler.RegisterUser))
	r.Post("/anmelden", handlers.Wrap(authHandler.LoginUser))
	r.Delete("/abmelden", middleware.AuthMiddleware(paseto, redis), handlers.Wrap(authHandler.LogoutUser))
	r.Delete("/alle-device-abgemeldet", middleware.AuthMiddleware(paseto, redis), handlers.Wrap(authHandler.LogoutAllDevices))
	r.Get("/device-auflisten", middleware.AuthMiddleware(paseto, redis), handlers.Wrap(authHandler.ListAllUserDevices))
}
