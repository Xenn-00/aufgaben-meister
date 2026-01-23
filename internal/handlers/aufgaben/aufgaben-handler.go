package aufgaben_handlers

import (
	"strings"

	aufgaben_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/aufgaben-dto"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/Xenn-00/aufgaben-meister/internal/handlers"
	internal_i18n "github.com/Xenn-00/aufgaben-meister/internal/i18n"
	aufgaben_case "github.com/Xenn-00/aufgaben-meister/internal/use-cases/aufgaben-case"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type AufgabenHandler struct {
	validator *validator.Validate
	service   aufgaben_case.AufgabenServiceContract
	i18n      *internal_i18n.I18nService
}

func NewAufgabenHandler(db *pgxpool.Pool, redis *redis.Client, i18n *internal_i18n.I18nService) *AufgabenHandler {
	validate := validator.New()
	validate.RegisterValidation("aufgabenPriority", aufgaben_dto.IsValidAufgabenPriority)
	validate.RegisterValidation("aufgabenStatus", aufgaben_dto.IsValidAufgabenStatus)
	validate.RegisterValidation("reasonCode", aufgaben_dto.IsValidReasonCode)
	return &AufgabenHandler{
		validator: validate,
		service:   aufgaben_case.NewAufgabenService(db, redis),
		i18n:      i18n,
	}
}

func (h *AufgabenHandler) CreateNewAufgaben(c *fiber.Ctx) error {
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	// get project id from param
	projectID, err := handlers.GetParamProjectID(c, h.validator)
	if err != nil {
		return err
	}

	// get req body
	var req *aufgaben_dto.CreateNewAufgabenRequest
	if err := c.BodyParser(&req); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidBody, "request.invalid_body", err)
	}

	if req.Priority != nil {
		s := strings.Title(strings.TrimSpace(*req.Priority))
		req.Priority = &s
	}

	if err := h.validator.Struct(req); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	// call service
	resp, err := h.service.CreateNewAufgaben(c.Context(), userID, projectID, req)
	if err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_insert_new_aufgaben", nil), resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}

	return nil
}

func (h *AufgabenHandler) ListTasks(c *fiber.Ctx) error {
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	// get project ID from param
	projectID, err := handlers.GetParamProjectID(c, h.validator)
	if err != nil {
		return err
	}

	// get query filter
	var filters aufgaben_dto.AufgabenListFilter
	if err := c.QueryParser(&filters); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidQuery, "request.invalid_query", err)
	}

	if filters.Status != nil {
		s := handlers.NormalizeStatusCase(*filters.Status)
		filters.Status = &s
	}
	if filters.Limit == 0 {
		filters.Limit = 20
	} else if filters.Limit > 100 {
		filters.Limit = 100
	}

	if filters.Page == 0 {
		filters.Page = 1
	} else if filters.Page > 100 {
		filters.Page = 100
	}

	if err := h.validator.Struct(filters); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	// call service
	resp, paging, err := h.service.ListTasksProject(c.Context(), userID, projectID, filters)
	if err != nil {
		return err
	}

	// set http cache behavior
	c.Set("Cache-Control", "private, max-age=10")
	reqID := handlers.GetRequestID(c)
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_list_aufgaben", nil), resp, reqID, paging)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}

	return nil
}

func (h *AufgabenHandler) AssignTask(c *fiber.Ctx) error {
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	// get project id param
	projectID, err := handlers.GetParamProjectID(c, h.validator)
	if err != nil {
		return err
	}

	// get task id param
	taskID, err := handlers.GetParamTaskID(c, h.validator)
	if err != nil {
		return err
	}

	// get req
	var req *aufgaben_dto.AufgabenAssignRequest
	if err := c.BodyParser(&req); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidBody, "request.invalid_body", err)
	}

	if err := h.validator.Struct(req); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	// call service
	resp, err := h.service.AssignTask(c.Context(), userID, projectID, taskID, req)
	if err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_assign_task", nil), resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}

	return nil
}

func (h *AufgabenHandler) ForwardProgress(c *fiber.Ctx) error {
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	// get project id param
	projectID, err := handlers.GetParamProjectID(c, h.validator)
	if err != nil {
		return err
	}

	// get task id param
	taskID, err := handlers.GetParamTaskID(c, h.validator)
	if err != nil {
		return err
	}

	// call service
	resp, err := h.service.ForwardProgressTask(c.Context(), userID, projectID, taskID)

	if err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_forward_progress", nil), resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}

	return nil
}

func (h *AufgabenHandler) UnassignTask(c *fiber.Ctx) error {
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	// get project id param
	projectID, err := handlers.GetParamProjectID(c, h.validator)
	if err != nil {
		return err
	}

	// get task id param
	taskID, err := handlers.GetParamTaskID(c, h.validator)
	if err != nil {
		return err
	}

	// get req body
	var req *aufgaben_dto.UnassignAufgabenRequest
	if err := c.BodyParser(&req); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidBody, "request.invalid_body", err)
	}

	if req.ReasonCode != "" {
		s := strings.Title(strings.TrimSpace(req.ReasonCode))
		req.ReasonCode = s
	}

	if err := h.validator.Struct(req); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	// call service
	resp, err := h.service.UnassignTask(c.Context(), userID, projectID, taskID, req)
	if err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_unassign_task", nil), resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}

	return nil
}

func (h *AufgabenHandler) ForceUnassignTask(c *fiber.Ctx) error {
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	// get project id param
	projectID, err := handlers.GetParamProjectID(c, h.validator)
	if err != nil {
		return err
	}

	// get task id param
	taskID, err := handlers.GetParamTaskID(c, h.validator)
	if err != nil {
		return err
	}

	// get req body
	var req *aufgaben_dto.ForceUnassignAufgabenRequest
	if err := c.BodyParser(&req); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidBody, "request.invalid_body", err)
	}

	if req.ReasonCode != "" {
		s := strings.Title(strings.TrimSpace(req.ReasonCode))
		req.ReasonCode = s
	}

	if err := h.validator.Struct(req); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	// call service
	resp, err := h.service.ForceUnassignTask(c.Context(), userID, projectID, taskID, req)
	if err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_unassign_task", nil), resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}

	return nil
}
func (h *AufgabenHandler) ReassignTask(c *fiber.Ctx) error {
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	// get project id param
	projectID, err := handlers.GetParamProjectID(c, h.validator)
	if err != nil {
		return err
	}

	// get task id param
	taskID, err := handlers.GetParamTaskID(c, h.validator)
	if err != nil {
		return err
	}

	// get req body
	var req *aufgaben_dto.ReassignAufgabenRequest
	if err := c.BodyParser(&req); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidBody, "request.invalid_body", err)
	}

	if req.ReasonCode != nil {
		s := strings.Title(strings.TrimSpace(*req.ReasonCode))
		req.ReasonCode = &s
	}

	if err := h.validator.Struct(req); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	// call service
	resp, err := h.service.ReassignTask(c.Context(), userID, projectID, taskID, req)
	if err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_reassign_task", nil), resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}

	return nil
}
