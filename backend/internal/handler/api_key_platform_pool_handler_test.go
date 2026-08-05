//go:build unit

package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type apiKeyPlatformPoolListerStub struct {
	platforms []service.Platform
}

func (s apiKeyPlatformPoolListerStub) List(context.Context) ([]service.Platform, error) {
	return append([]service.Platform(nil), s.platforms...), nil
}

func TestAPIKeyHandlerAvailablePlatformsReturnsOnlyActivePoolMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &APIKeyHandler{
		platformPools: apiKeyPlatformPoolListerStub{platforms: []service.Platform{
			{ID: 11, Code: "openai-primary", Name: "OpenAI Primary", AccountPlatform: service.PlatformOpenAI, Status: service.PlatformStatusActive},
			{ID: 12, Code: "grok-paused", Name: "Grok Paused", AccountPlatform: service.PlatformGrok, Status: service.StatusDisabled},
		}},
	}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7})
		c.Next()
	})
	router.GET("/api/v1/platforms/available", handler.GetAvailablePlatforms)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/platforms/available", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"code":0,"message":"success","data":[{"id":11,"code":"openai-primary","name":"OpenAI Primary","account_platform":"openai"}]}`, recorder.Body.String())
	require.NotContains(t, recorder.Body.String(), "model_rules")
	require.NotContains(t, recorder.Body.String(), "legacy_group_id")
}
