package handler

import (
	"errors"
	"strconv"

	"github.com/community-platform/backend/internal/dto"
	"github.com/community-platform/backend/internal/middleware"
	"github.com/community-platform/backend/internal/service"
	"github.com/community-platform/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// CommunityHandler handles all community-related HTTP endpoints.
type CommunityHandler struct {
	communitySvc service.CommunityService
}

// NewCommunityHandler creates a CommunityHandler.
func NewCommunityHandler(communitySvc service.CommunityService) *CommunityHandler {
	return &CommunityHandler{communitySvc: communitySvc}
}

// CreateCommunity godoc
// POST /communities
// Creates a new community. The authenticated user becomes the first admin.
func (h *CommunityHandler) CreateCommunity(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.CreateCommunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	community, err := h.communitySvc.Create(c.Request.Context(), userID, req)
	if err != nil {
		response.BadRequest(c, "CREATE_FAILED", err.Error(), nil)
		return
	}

	response.Created(c, "Community created successfully", community)
}

// GetCommunity godoc
// GET /communities/:id
// Returns community details. Invite code only visible to admins.
func (h *CommunityHandler) GetCommunity(c *gin.Context) {
	communityID := c.Param("id")
	userID := middleware.GetUserID(c)

	community, err := h.communitySvc.GetByID(c.Request.Context(), communityID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommunityNotFound):
			response.NotFound(c, "Community not found")
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, "Community fetched", community)
}

// ListCommunities godoc
// GET /communities?type=SOCIETY&city=Mumbai&pin_code=400001&page=1&limit=20
// Returns a paginated list of active communities with optional filters.
func (h *CommunityHandler) ListCommunities(c *gin.Context) {
	var query dto.ListCommunitiesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "VALIDATION_ERROR", "Invalid query parameters", err.Error())
		return
	}

	result, err := h.communitySvc.List(c.Request.Context(), query)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "Communities fetched", result)
}

// JoinCommunity godoc
// POST /communities/:id/join
// Joins the authenticated user to the community.
// If the community requires approval, status will be PENDING.
func (h *CommunityHandler) JoinCommunity(c *gin.Context) {
	communityID := c.Param("id")
	userID := middleware.GetUserID(c)

	var req dto.JoinCommunityRequest
	// Body is optional (invite_code may or may not be sent)
	_ = c.ShouldBindJSON(&req)

	result, err := h.communitySvc.Join(c.Request.Context(), communityID, userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommunityNotFound):
			response.NotFound(c, "Community not found")
		case errors.Is(err, service.ErrAlreadyMember):
			response.Conflict(c, "ALREADY_MEMBER", "You are already a member of this community")
		case errors.Is(err, service.ErrInvalidInviteCode):
			response.BadRequest(c, "INVALID_INVITE_CODE", "The invite code is incorrect", nil)
		case errors.Is(err, service.ErrCommunityInactive):
			response.BadRequest(c, "COMMUNITY_INACTIVE", "This community is not currently active", nil)
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, result.Message, result)
}

// LeaveCommunity godoc
// POST /communities/:id/leave
// Removes the authenticated user from the community.
func (h *CommunityHandler) LeaveCommunity(c *gin.Context) {
	communityID := c.Param("id")
	userID := middleware.GetUserID(c)

	if err := h.communitySvc.Leave(c.Request.Context(), communityID, userID); err != nil {
		switch {
		case errors.Is(err, service.ErrNotCommunityMember):
			response.BadRequest(c, "NOT_MEMBER", "You are not a member of this community", nil)
		default:
			response.BadRequest(c, "LEAVE_FAILED", err.Error(), nil)
		}
		return
	}

	response.OK(c, "You have left the community", nil)
}

// ListCommunityMembers godoc
// GET /communities/:id/members?page=1&limit=20
// Lists approved members. Only accessible to community members.
func (h *CommunityHandler) ListCommunityMembers(c *gin.Context) {
	communityID := c.Param("id")
	userID := middleware.GetUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	members, total, err := h.communitySvc.ListMembers(c.Request.Context(), communityID, userID, page, limit)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotCommunityMember):
			response.Forbidden(c, "NOT_MEMBER", "You must be a member to view the member list")
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, "Members fetched", gin.H{
		"members": members,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// ApproveMember godoc
// POST /communities/:id/members/:user_id/approve
// Approves a pending join request. Only community admins can do this.
func (h *CommunityHandler) ApproveMember(c *gin.Context) {
	communityID := c.Param("id")
	targetUserID := c.Param("user_id")
	adminID := middleware.GetUserID(c)

	if err := h.communitySvc.ApproveMember(c.Request.Context(), communityID, adminID, targetUserID); err != nil {
		switch {
		case errors.Is(err, service.ErrNotCommunityAdmin):
			response.Forbidden(c, "NOT_ADMIN", "You do not have admin rights in this community")
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, "Member approved", nil)
}

// MyCommunities godoc
// GET /users/communities
// Returns all communities the authenticated user belongs to.
func (h *CommunityHandler) MyCommunities(c *gin.Context) {
	userID := middleware.GetUserID(c)

	communities, err := h.communitySvc.ListUserCommunities(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "Your communities fetched", gin.H{"communities": communities})
}
