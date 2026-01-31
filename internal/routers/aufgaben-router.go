package routers

import (
	"fmt"
	"strings"
	"time"

	aufgaben_handlers "github.com/Xenn-00/aufgaben-meister/internal/handlers/aufgaben"
	"github.com/Xenn-00/aufgaben-meister/internal/i18n"
	"github.com/Xenn-00/aufgaben-meister/internal/middleware"
	"github.com/Xenn-00/aufgaben-meister/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	redis_fiber "github.com/gofiber/storage/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func AufgabenRouter(api fiber.Router, db *pgxpool.Pool, redis *redis.Client, i18n *i18n.I18nService, paseto *utils.PasetoMaker, cfgStorage CfgRedisStorage) {
	r := api.Group("/project/:project_id/aufgaben", middleware.AuthMiddleware(paseto, redis))
	aufgabenHandler := aufgaben_handlers.NewAufgabenHandler(db, redis, i18n)

	// prepare redis storage for rate limiter fiber
	redisAddr := strings.Split(redis.Options().Addr, ":") // seperate host and port
	redisStore := redis_fiber.New(redis_fiber.Config{
		Host:     redisAddr[0],
		Password: redis.Options().Password,
		Port:     6379,
		Database: 1,
	})

	r.Post("/create", aufgabenHandler.CreateNewAufgaben)
	r.Get("/list", aufgabenHandler.ListTasks)
	r.Get("/:task_id", aufgabenHandler.GetAufgabeDetails)
	r.Post("/:task_id/assign", aufgabenHandler.AssignTask)
	r.Post("/:task_id/forward-progress", aufgabenHandler.ForwardProgress)
	r.Post("/:task_id/unassign", aufgabenHandler.UnassignTask)
	r.Post("/:task_id/force-unassign", aufgabenHandler.ForceUnassignTask)
	r.Post("/:task_id/reassign", limiter.New(limiter.Config{
		Max:        5,
		Expiration: 30 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			userID := c.Locals("user_id")
			projectID := c.Params("project_id")
			taskID := c.Params("task_id")
			if userID == nil {
				return "reassign:ip:" + c.IP() // fallback to ip
			}
			return fmt.Sprintf("reassign:%v:%s:%s", userID, projectID, taskID)
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"status": "error",
				"error":  "too_many_request",
			})
		},
		Storage: redisStore,
	}), aufgabenHandler.ReassignTask)
	r.Get("/assigned-task", aufgabenHandler.ListAssignedTasks)
	r.Post("/:task_id/archive", aufgabenHandler.ArchiveTask)
	r.Patch("/:task_id/update-due-date", aufgabenHandler.UpdateDueDate)
	r.Get("/:task_id/events", aufgabenHandler.FetchEventsForTask)
	r.Post("/:task_id/force-handover", aufgabenHandler.ForceAufgabeHandover)
}
