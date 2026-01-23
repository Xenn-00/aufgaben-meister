package routers

import (
	"fmt"
	"strings"
	"time"

	project_handlers "github.com/Xenn-00/aufgaben-meister/internal/handlers/project"
	"github.com/Xenn-00/aufgaben-meister/internal/i18n"
	"github.com/Xenn-00/aufgaben-meister/internal/middleware"
	"github.com/Xenn-00/aufgaben-meister/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	redis_fiber "github.com/gofiber/storage/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func ProjectRouter(api fiber.Router, db *pgxpool.Pool, redis *redis.Client, i18n *i18n.I18nService, paseto *utils.PasetoMaker, cfgStorage CfgRedisStorage) {
	r := api.Group("/project", middleware.AuthMiddleware(paseto, redis))
	projectHandler := project_handlers.NewProjectHandler(db, redis, i18n)

	// prepare redis storage for rate limiter fiber
	redisAddr := strings.Split(redis.Options().Addr, ":") // seperate host and port
	redisStore := redis_fiber.New(redis_fiber.Config{
		Host:     redisAddr[0],
		Password: redis.Options().Password,
		Port:     6379,
		Database: 1,
	})
	r.Post("/create", projectHandler.CreateNewProject)
	r.Get("/me", projectHandler.GetSelfProject)
	r.Get("/:project_id/detail", projectHandler.GetProjectDetail)
	r.Post("/invite/accept", projectHandler.AcceptProjectMember)
	r.Post("/invite/:project_id", limiter.New(limiter.Config{
		Max:        5,
		Expiration: 30 * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			userID := c.Locals("user_id")
			projectID := c.Params("project_id")
			if userID == nil {
				return "invite:ip:" + c.IP() // fallback to ip
			}
			return fmt.Sprintf("invite:%v:%s", userID, projectID)
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"status": "error",
				"error":  "too_many_request",
			})
		},
		Storage: redisStore,
	}), projectHandler.InviteProjectMember)
	r.Get("/invitations", projectHandler.ListSelfPendingInvitations)
	r.Get("/:project_id/invitations", projectHandler.GetInvitationsInProject)
	r.Post("/invite/:invitation_id/reject", projectHandler.RejectSelfPendingInvitation)
	r.Post("/invite/:invitation_id/resend", limiter.New(limiter.Config{
		Max:        10,
		Expiration: 24 * time.Hour,
		KeyGenerator: func(c *fiber.Ctx) string {
			userID := c.Locals("user_id")
			projectID := c.Params("project_id")
			if userID == nil {
				return "invite:ip:" + c.IP() // fallback to ip
			}
			return fmt.Sprintf("invite:%v:%s", userID, projectID)
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"status": "error",
				"error":  "too_many_request",
			})
		},
		Storage: redisStore,
	}), projectHandler.ResendProjectInvitations)
	r.Post("/invite/:project_id/revoke", limiter.New(limiter.Config{
		Max:        5,
		Expiration: 30 * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			userID := c.Locals("user_id")
			projectID := c.Params("project_id")
			if userID == nil {
				return "invite:ip:" + c.IP() // fallback to ip
			}
			return fmt.Sprintf("invite:%v:%s", userID, projectID)
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"status": "error",
				"error":  "too_many_request",
			})
		},
		Storage: redisStore,
	}), projectHandler.RevokeProjectInvitations)

}
