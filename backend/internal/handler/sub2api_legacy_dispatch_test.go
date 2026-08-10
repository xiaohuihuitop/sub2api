//go:build unit

package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/gatewayruntime"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestDispatchLegacyEndpointUsesApplicationGatewayRoute(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{ID: 11, UserID: 12, User: &service.User{ID: 12}})
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 12, Concurrency: 1})
	c.Request = c.Request.WithContext(service.WithGatewayPlatformAssetContext(c.Request.Context(), &service.GatewayPlatformAssetContext{
		Platform: &service.ResolvedPlatformModel{
			PlatformID:      42,
			PlatformCode:    "openai-main",
			AccountPlatform: service.PlatformOpenAI,
			RequestedModel:  "gpt-5.6",
			UpstreamModel:   "gpt-5.6",
		},
		SchedulingScope: service.PlatformSchedulingScope{
			PlatformID:      42,
			PlatformCode:    "openai-main",
			AccountPlatform: service.PlatformOpenAI,
		},
	}))

	called := false
	err := (&GatewayHandler{}).dispatchLegacyEndpoint(c, gatewayruntime.EndpointResponses, func(ctx *gin.Context) {
		called = true
		ctx.Status(http.StatusOK)
	})

	require.NoError(t, err)
	require.True(t, called)
	require.Equal(t, http.StatusOK, recorder.Code)
}
