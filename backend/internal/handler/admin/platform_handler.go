package admin

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// platformManagementService keeps the handler independent from the concrete
// service, making the administrator contract directly testable.
type platformManagementService interface {
	List(ctx context.Context) ([]service.Platform, error)
	GetByID(ctx context.Context, id int64) (*service.Platform, error)
	Create(ctx context.Context, input service.CreatePlatformInput) (*service.Platform, error)
	Update(ctx context.Context, id int64, input service.UpdatePlatformInput) (*service.Platform, error)
}

// PlatformHandler manages business platform account pools. Deletion is not
// exposed: disabling a platform preserves existing account ownership.
type PlatformHandler struct {
	platforms platformManagementService
}

func NewPlatformHandler(platforms platformManagementService) *PlatformHandler {
	return &PlatformHandler{platforms: platforms}
}

type platformModelRuleRequest struct {
	ModelPattern         string   `json:"model_pattern" binding:"required"`
	UpstreamModel        string   `json:"upstream_model"`
	EndpointCapabilities []string `json:"endpoint_capabilities"`
	Enabled              *bool    `json:"enabled"`
}

func (r platformModelRuleRequest) toServiceRule() service.PlatformModelRule {
	enabled := true
	if r.Enabled != nil {
		enabled = *r.Enabled
	}
	return service.PlatformModelRule{
		ModelPattern:         r.ModelPattern,
		UpstreamModel:        r.UpstreamModel,
		EndpointCapabilities: append([]string(nil), r.EndpointCapabilities...),
		Enabled:              enabled,
	}
}

type createPlatformRequest struct {
	Code            string                     `json:"code" binding:"required"`
	Name            string                     `json:"name" binding:"required"`
	AccountPlatform string                     `json:"account_platform" binding:"required"`
	Status          string                     `json:"status"`
	LegacyGroupID   *int64                     `json:"legacy_group_id"`
	ModelRules      []platformModelRuleRequest `json:"model_rules"`
}

func (r createPlatformRequest) toServiceInput() service.CreatePlatformInput {
	rules := make([]service.PlatformModelRule, len(r.ModelRules))
	for index := range r.ModelRules {
		rules[index] = r.ModelRules[index].toServiceRule()
	}
	return service.CreatePlatformInput{
		Code:            r.Code,
		Name:            r.Name,
		AccountPlatform: r.AccountPlatform,
		Status:          r.Status,
		LegacyGroupID:   r.LegacyGroupID,
		ModelRules:      rules,
	}
}

type updatePlatformRequest struct {
	Code             *string                     `json:"code"`
	Name             *string                     `json:"name"`
	AccountPlatform  *string                     `json:"account_platform"`
	Status           *string                     `json:"status"`
	LegacyGroupID    *int64                      `json:"legacy_group_id"`
	ClearLegacyGroup bool                        `json:"clear_legacy_group"`
	ModelRules       *[]platformModelRuleRequest `json:"model_rules"`
}

func (r updatePlatformRequest) toServiceInput() service.UpdatePlatformInput {
	result := service.UpdatePlatformInput{
		Code:             r.Code,
		Name:             r.Name,
		AccountPlatform:  r.AccountPlatform,
		Status:           r.Status,
		LegacyGroupID:    r.LegacyGroupID,
		ClearLegacyGroup: r.ClearLegacyGroup,
	}
	if r.ModelRules == nil {
		return result
	}
	rules := make([]service.PlatformModelRule, len(*r.ModelRules))
	for index := range *r.ModelRules {
		rules[index] = (*r.ModelRules)[index].toServiceRule()
	}
	result.ModelRules = &rules
	return result
}

type platformModelRuleResponse struct {
	ID                   int64    `json:"id"`
	ModelPattern         string   `json:"model_pattern"`
	UpstreamModel        string   `json:"upstream_model"`
	EndpointCapabilities []string `json:"endpoint_capabilities"`
	Enabled              bool     `json:"enabled"`
}

type platformResponse struct {
	ID              int64                       `json:"id"`
	Code            string                      `json:"code"`
	Name            string                      `json:"name"`
	AccountPlatform string                      `json:"account_platform"`
	Status          string                      `json:"status"`
	LegacyGroupID   *int64                      `json:"legacy_group_id,omitempty"`
	ModelRules      []platformModelRuleResponse `json:"model_rules"`
}

func platformResponseFromService(platform *service.Platform) platformResponse {
	if platform == nil {
		return platformResponse{ModelRules: []platformModelRuleResponse{}}
	}
	rules := make([]platformModelRuleResponse, len(platform.ModelRules))
	for index := range platform.ModelRules {
		rules[index] = platformModelRuleResponse{
			ID:                   platform.ModelRules[index].ID,
			ModelPattern:         platform.ModelRules[index].ModelPattern,
			UpstreamModel:        platform.ModelRules[index].UpstreamModel,
			EndpointCapabilities: append([]string(nil), platform.ModelRules[index].EndpointCapabilities...),
			Enabled:              platform.ModelRules[index].Enabled,
		}
		if rules[index].EndpointCapabilities == nil {
			rules[index].EndpointCapabilities = []string{}
		}
	}
	return platformResponse{
		ID:              platform.ID,
		Code:            platform.Code,
		Name:            platform.Name,
		AccountPlatform: platform.AccountPlatform,
		Status:          platform.Status,
		LegacyGroupID:   platform.LegacyGroupID,
		ModelRules:      rules,
	}
}

// List returns all platform account pools, including disabled configurations.
func (h *PlatformHandler) List(c *gin.Context) {
	platforms, err := h.platforms.List(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	result := make([]platformResponse, len(platforms))
	for index := range platforms {
		result[index] = platformResponseFromService(&platforms[index])
	}
	response.Success(c, result)
}

// GetByID returns one platform pool configuration.
func (h *PlatformHandler) GetByID(c *gin.Context) {
	id, ok := parsePositiveIDParam(c, "id")
	if !ok {
		return
	}
	platform, err := h.platforms.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, platformResponseFromService(platform))
}

// Create creates a platform account pool and its model rules atomically.
func (h *PlatformHandler) Create(c *gin.Context) {
	var req createPlatformRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	platform, err := h.platforms.Create(c.Request.Context(), req.toServiceInput())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, platformResponseFromService(platform))
}

// Update edits a platform pool. Administrators disable pools instead of
// deleting them, so existing account ownership remains intact.
func (h *PlatformHandler) Update(c *gin.Context) {
	id, ok := parsePositiveIDParam(c, "id")
	if !ok {
		return
	}
	var req updatePlatformRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	platform, err := h.platforms.Update(c.Request.Context(), id, req.toServiceInput())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, platformResponseFromService(platform))
}
