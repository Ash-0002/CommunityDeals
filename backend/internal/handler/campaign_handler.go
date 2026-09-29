package handler

import (
	"errors"
	"strconv"

	"github.com/community-platform/backend/internal/domain"
	"github.com/community-platform/backend/internal/dto"
	"github.com/community-platform/backend/internal/middleware"
	"github.com/community-platform/backend/internal/repository"
	"github.com/community-platform/backend/internal/service"
	"github.com/community-platform/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// CampaignHandler handles all campaign-related HTTP endpoints.
type CampaignHandler struct {
	campaignSvc service.CampaignService
	userRepo    repository.UserRepository
}

// NewCampaignHandler creates a CampaignHandler.
func NewCampaignHandler(campaignSvc service.CampaignService, userRepo repository.UserRepository) *CampaignHandler {
	return &CampaignHandler{campaignSvc: campaignSvc, userRepo: userRepo}
}

// CreateCampaign godoc
// POST /campaigns
// Creates a new campaign with pricing tiers. Published immediately in MVP.
func (h *CampaignHandler) CreateCampaign(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.CreateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	campaign, err := h.campaignSvc.Create(c.Request.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommunityNotFound):
			response.NotFound(c, "Community not found")
		case errors.Is(err, service.ErrTiersNotValid):
			response.BadRequest(c, "INVALID_TIERS", err.Error(), nil)
		default:
			response.BadRequest(c, "CREATE_FAILED", err.Error(), nil)
		}
		return
	}

	response.Created(c, "Campaign created successfully", campaign)
}

// GetCampaign godoc
// GET /campaigns/:id
// Returns full campaign details including pricing tiers and join status.
func (h *CampaignHandler) GetCampaign(c *gin.Context) {
	campaignID := c.Param("id")
	userID := middleware.GetUserID(c)

	campaign, err := h.campaignSvc.GetByID(c.Request.Context(), campaignID, userID)
	if err != nil {
		if errors.Is(err, service.ErrCampaignNotFound) {
			response.NotFound(c, "Campaign not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "Campaign fetched", campaign)
}

// GetCampaignBySlug godoc
// GET /c/:slug  (public, no auth required — used for shareable links)
// Returns campaign details by the slug in the shareable URL.
func (h *CampaignHandler) GetCampaignBySlug(c *gin.Context) {
	slug := c.Param("slug")
	// requestingUserID may be empty for unauthenticated visitors
	userID := c.GetString(middleware.ContextUserID)

	campaign, err := h.campaignSvc.GetBySlug(c.Request.Context(), slug, userID)
	if err != nil {
		if errors.Is(err, service.ErrCampaignNotFound) {
			response.NotFound(c, "Campaign not found")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "Campaign fetched", campaign)
}

// ListCampaigns godoc
// GET /campaigns?community_id=xxx&status=PUBLISHED&page=1&limit=20
// Returns a paginated list of campaigns with optional filters.
func (h *CampaignHandler) ListCampaigns(c *gin.Context) {
	var query dto.ListCampaignsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "VALIDATION_ERROR", "Invalid query parameters", err.Error())
		return
	}

	result, err := h.campaignSvc.List(c.Request.Context(), query, middleware.GetUserID(c))
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "Campaigns fetched", result)
}

// JoinCampaign godoc
// POST /campaigns/:id/join
// Joins the campaign. Uses SELECT FOR UPDATE under the hood to prevent race conditions.
func (h *CampaignHandler) JoinCampaign(c *gin.Context) {
	campaignID := c.Param("id")
	userID := middleware.GetUserID(c)

	result, err := h.campaignSvc.Join(c.Request.Context(), campaignID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCampaignNotFound):
			response.NotFound(c, "Campaign not found")
		case errors.Is(err, service.ErrAlreadyJoined):
			response.Conflict(c, "ALREADY_JOINED", "You have already joined this campaign")
		case errors.Is(err, service.ErrCampaignFull):
			response.Conflict(c, "CAMPAIGN_FULL", "This campaign is full — no more spots available")
		case errors.Is(err, service.ErrCampaignNotJoinable):
			response.BadRequest(c, "NOT_JOINABLE", "This campaign is not open for joining", nil)
		default:
			response.BadRequest(c, "JOIN_FAILED", err.Error(), nil)
		}
		return
	}

	response.OK(c, result.Message, result)
}

// LeaveCampaign godoc
// POST /campaigns/:id/leave
// Removes the user from a campaign they previously joined.
func (h *CampaignHandler) LeaveCampaign(c *gin.Context) {
	campaignID := c.Param("id")
	userID := middleware.GetUserID(c)

	if err := h.campaignSvc.Leave(c.Request.Context(), campaignID, userID); err != nil {
		switch {
		case errors.Is(err, service.ErrCampaignNotFound):
			response.NotFound(c, "Campaign not found")
		case errors.Is(err, service.ErrNotJoined):
			response.BadRequest(c, "NOT_JOINED", "You have not joined this campaign", nil)
		default:
			response.BadRequest(c, "LEAVE_FAILED", err.Error(), nil)
		}
		return
	}

	response.OK(c, "You have left the campaign", nil)
}

// ListParticipants godoc
// GET /campaigns/:id/participants?page=1&limit=20
func (h *CampaignHandler) ListParticipants(c *gin.Context) {
	campaignID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	participants, total, err := h.campaignSvc.ListParticipants(c.Request.Context(), campaignID, page, limit)
	if err != nil {
		if errors.Is(err, service.ErrCampaignNotFound) {
			response.NotFound(c, "Campaign not found")
			return
		}
		response.InternalError(c)
		return
	}

	// Enrich each participant with their display name/avatar for the
	// "N people joined" avatar-stack UI. Best-effort — a lookup failure
	// just falls back to an empty name rather than failing the request.
	enriched := make([]dto.ParticipantResponse, 0, len(participants))
	for _, p := range participants {
		item := dto.ParticipantResponse{
			UserID:   p.UserID,
			JoinedAt: p.JoinedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if user, err := h.userRepo.FindByID(c.Request.Context(), p.UserID); err == nil {
			item.Name = user.Name
			item.AvatarURL = user.AvatarURL
		}
		enriched = append(enriched, item)
	}

	response.OK(c, "Participants fetched", gin.H{
		"participants": enriched,
		"total":        total,
		"page":         page,
		"limit":        limit,
	})
}

// UpdateCampaignStatus godoc
// PATCH /campaigns/:id/status
// Transitions a campaign to a new status. Enforces the state machine.
func (h *CampaignHandler) UpdateCampaignStatus(c *gin.Context) {
	campaignID := c.Param("id")
	userID := middleware.GetUserID(c)

	var req dto.UpdateCampaignStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	newStatus := domain.CampaignStatus(req.Status)
	if err := h.campaignSvc.UpdateStatus(c.Request.Context(), campaignID, userID, newStatus); err != nil {
		switch {
		case errors.Is(err, service.ErrCampaignNotFound):
			response.NotFound(c, "Campaign not found")
		case errors.Is(err, service.ErrInvalidStatusChange):
			response.BadRequest(c, "INVALID_TRANSITION", err.Error(), nil)
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, "Campaign status updated", gin.H{"status": req.Status})
}
