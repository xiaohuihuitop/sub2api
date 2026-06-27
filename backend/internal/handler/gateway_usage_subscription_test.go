//go:build unit

package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	middleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUsageUnrestricted_UsesSubscriptionGroupForLimits(t *testing.T) {
	gin.SetMode(gin.TestMode)

	keyDailyLimit := 30.0
	subDailyLimit := 60.0
	keyGroupID := int64(10)
	subGroupID := int64(20)

	keyGroup := &service.Group{
		ID:               keyGroupID,
		Name:             "openai",
		Status:           service.StatusActive,
		SubscriptionType: service.SubscriptionTypeSubscription,
		DailyLimitUSD:    &keyDailyLimit,
	}
	subGroup := service.Group{
		ID:               subGroupID,
		Name:             "glm",
		Status:           service.StatusActive,
		SubscriptionType: service.SubscriptionTypeSubscription,
		DailyLimitUSD:    &subDailyLimit,
	}
	apiKey := &service.APIKey{
		ID:            1,
		UserID:        7,
		Status:        service.StatusActive,
		GroupID:       &keyGroupID,
		Group:         keyGroup,
		AllowedGroups: []service.Group{*keyGroup, subGroup},
	}
	subscription := &service.UserSubscription{
		ID:            100,
		UserID:        apiKey.UserID,
		GroupID:       subGroupID,
		DailyUsageUSD: 12.5,
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/usage", nil)
	c.Set(string(middleware.ContextKeySubscription), subscription)

	h := &GatewayHandler{}
	h.usageUnrestricted(c, context.Background(), apiKey, middleware.AuthSubject{UserID: apiKey.UserID}, nil, nil)

	require.Equal(t, http.StatusOK, recorder.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))
	require.Equal(t, "glm", resp["planName"])
	require.Equal(t, subDailyLimit-subscription.DailyUsageUSD, resp["remaining"])

	subResp, ok := resp["subscription"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, subDailyLimit, subResp["daily_limit_usd"])
	require.Equal(t, subscription.DailyUsageUSD, subResp["daily_usage_usd"])
}
