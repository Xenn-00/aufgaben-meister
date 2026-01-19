package project_handlers

import (
	"strings"

	project_dto "github.com/Xenn-00/aufgaben-meister/internal/dtos/project-dto"
	app_errors "github.com/Xenn-00/aufgaben-meister/internal/errors"
	"github.com/Xenn-00/aufgaben-meister/internal/handlers"
	internal_i18n "github.com/Xenn-00/aufgaben-meister/internal/i18n"
	project_case "github.com/Xenn-00/aufgaben-meister/internal/use-cases/project-case"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type ProjectHander struct {
	validator *validator.Validate
	service   project_case.ProjectServiceContract
	i18n      *internal_i18n.I18nService
}

func NewProjectHandler(db *pgxpool.Pool, redis *redis.Client, i18n *internal_i18n.I18nService) *ProjectHander {
	validate := validator.New()
	validate.RegisterValidation("typeProject", project_dto.IsValidTypeProject)
	validate.RegisterValidation("visibility", project_dto.IsValidVisibility)
	validate.RegisterValidation("invitationStatus", project_dto.IsValidInvitationStatus)
	return &ProjectHander{
		validator: validate,
		service:   project_case.NewProjectService(db, redis),
		i18n:      i18n,
	}
}

func (h *ProjectHander) CreateNewProject(c *fiber.Ctx) error {
	var req project_dto.CreateNewProjectRequest

	// Req Body geparst werden
	if err := c.BodyParser(&req); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidBody, "request.invalid_body", err)
	}

	// Req validiert werden
	if err := h.validator.Struct(req); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	// Abruf der BenutzerID von c.Locals
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	// Service abrufen
	resp, err := h.service.CreateNewProject(c.Context(), req, userID)
	if err != nil {
		return err
	}

	// Anwort zurückgeben
	reqID := handlers.GetRequestID(c)
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_create_project", nil), resp, reqID)
	if err := c.Status(fiber.StatusCreated).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}
	return nil
}

func (h *ProjectHander) GetSelfProject(c *fiber.Ctx) error {
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.GetSelfProject(c.Context(), userID)
	if err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_fetch_self_project", nil), resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}

	return nil
}

func (h *ProjectHander) GetProjectDetail(c *fiber.Ctx) error {
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	var req *project_dto.ParamGetProjectByID
	if err := c.ParamsParser(&req); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidParam, "request.invalid_param", err)
	}

	if err := h.validator.Struct(req); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	resp, err := h.service.GetProjectDetail(c.Context(), req.ID, userID)
	if err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_fetch_project_detail", nil), resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}

	return nil
}

func (h *ProjectHander) InviteProjectMember(c *fiber.Ctx) error {
	// Need sessioned userID for check, whether user role is meister oder nicht
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	// get param: project_id
	var params project_dto.ParamGetProjectByID
	if err := c.ParamsParser(&params); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidParam, "request.invalid_param", err)
	}

	// get req body
	var req *project_dto.InviteProjectMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidBody, "request.invalid_body", err)
	}

	if err := h.validator.Struct(req); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	if len(req.UserIDs) == 0 {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidBody, "request.invalid_body", nil)
	}

	if len(req.UserIDs) > 50 {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidBody, "request.too_many_invites", nil)
	}

	// call service
	resp, err := h.service.InviteProjectMember(c.Context(), params.ID, userID, req)
	if err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_invite_member", nil), resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}
	return nil
}

func (h *ProjectHander) AcceptProjectMember(c *fiber.Ctx) error {
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}
	var req project_dto.InvitationQueryRequest
	if err := c.QueryParser(&req); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidQuery, "request.invalid_query", nil)
	}

	if err := h.validator.Struct(req); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	// call service
	resp, err := h.service.AcceptInvitationProject(c.Context(), &req, userID)
	if err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_accept_invitation", nil), resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}
	return nil
}

func (h *ProjectHander) ListSelfPendingInvitations(c *fiber.Ctx) error {
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	// call service
	resp, err := h.service.GetSelfInvitationPending(c.Context(), userID)
	if err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_list_invitation", nil), resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}
	return nil
}

func (h *ProjectHander) RejectSelfPendingInvitation(c *fiber.Ctx) error {
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	// get invitation_id from param
	var param project_dto.InvitationParamRequest
	if err := c.ParamsParser(&param); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "request.invalid_param", err)
	}

	if err := h.validator.Struct(param); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	// call service
	resp, err := h.service.RejectProjectInvitation(c.Context(), param.InvitationID, userID)
	if err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_reject_invitation", nil), resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}

	return nil
}

func (h *ProjectHander) RevokeProjectInvitations(c *fiber.Ctx) error {
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	var params project_dto.ParamGetProjectByID
	if err := c.ParamsParser(&params); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidParam, "request.invalid_param", err)
	}

	if err := h.validator.Struct(params); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	var req *project_dto.RevokeProjectMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidBody, "request.invalid_body", err)
	}

	if err := h.validator.Struct(req); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	// call service
	resp, err := h.service.RevokeProjectInvitations(c.Context(), params.ID, userID, req)
	if err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_revoke_user", nil), resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}

	return nil
}

func (h *ProjectHander) ResendProjectInvitations(c *fiber.Ctx) error {
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	var params project_dto.InvitationParamRequest
	if err := c.ParamsParser(&params); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidParam, "request.invalid_param", err)
	}

	if err := h.validator.Struct(params); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	if err := h.service.ResendProjectInvitations(c.Context(), params.InvitationID, userID); err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_resend_invitation", nil), "OK", reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}

	return nil
}

func (h *ProjectHander) GetInvitationsInProject(c *fiber.Ctx) error {
	userID, err := handlers.GetUserID(c)
	if err != nil {
		return err
	}

	// get project_id from param
	var params project_dto.ParamGetProjectByID
	if err := c.ParamsParser(&params); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidParam, "request.invalid_param", err)
	}

	if err := h.validator.Struct(params); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	// get filter queries
	var filters project_dto.FilterProjectInvitation
	if err := c.QueryParser(&filters); err != nil {
		return app_errors.NewAppError(fiber.StatusBadRequest, app_errors.ErrInvalidQuery, "request.invalid_query", err)
	}

	if filters.Status != nil {
		s := strings.ToTitle(strings.TrimSpace(*filters.Status))
		filters.Status = &s
	}

	if err := h.validator.Struct(filters); err != nil {
		return app_errors.NewValidationError(app_errors.ParseValidationError(err))
	}

	// call service
	resp, err := h.service.GetInvitationsInProject(c.Context(), params.ID, userID, filters)
	if err != nil {
		return err
	}

	reqID := handlers.GetRequestID(c)
	lang, _ := c.Locals("lang").(string)
	webResp := handlers.CreateResponse(h.i18n.T(lang, "response.success_fetch_invitation_in_project", nil), resp, reqID)
	if err := c.Status(fiber.StatusOK).JSON(webResp); err != nil {
		return app_errors.NewAppError(fiber.StatusInternalServerError, app_errors.ErrInternal, "response.write_failed", err)
	}

	return nil
}
