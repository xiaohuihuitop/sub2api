//go:build unit

package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type groupBillingProfileAdminServiceStub struct {
	service.AdminService
	profile *service.BillingProfile
	input   *service.UpdateBillingProfileInput
}

func (s *groupBillingProfileAdminServiceStub) GetGroupBillingProfile(_ context.Context, _ int64) (*service.BillingProfile, error) {
	return s.profile, nil
}

func (s *groupBillingProfileAdminServiceStub) UpdateGroupBillingProfile(_ context.Context, _ int64, input *service.UpdateBillingProfileInput) (*service.BillingProfile, error) {
	s.input = input
	return s.profile, nil
}

func TestGroupBillingProfileHandlerUsesSeparateBillingPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &groupBillingProfileAdminServiceStub{profile: &service.BillingProfile{GroupID: 7, BalanceRateMultiplier: 1.25}}
	handler := NewGroupHandler(stub, nil, nil)
	router := gin.New()
	router.PUT("/groups/:id/billing-profile", handler.UpdateBillingProfile)

	req := httptest.NewRequest(http.MethodPut, "/groups/7/billing-profile", bytes.NewBufferString(`{
		"balance_rate_multiplier": 1.25,
		"peak_rate_enabled": true,
		"peak_start": "09:00",
		"peak_end": "18:00",
		"peak_rate_multiplier": 1.5,
		"image_rate_independent": true,
		"image_rate_multiplier": 2,
		"batch_image_discount_multiplier": 0.5,
		"batch_image_hold_multiplier": 0.6,
		"video_rate_independent": false,
		"video_rate_multiplier": 1
	}`))
	req.Header.Set("Content-Type", "application/json")
	writer := httptest.NewRecorder()

	router.ServeHTTP(writer, req)

	require.Equal(t, http.StatusOK, writer.Code)
	require.NotNil(t, stub.input)
	require.Equal(t, 1.25, stub.input.BalanceRateMultiplier)
	require.True(t, stub.input.PeakRateEnabled)
	require.Equal(t, "09:00", stub.input.PeakStart)
	require.Equal(t, 1.5, stub.input.PeakRateMultiplier)
}

type groupRPMOverridesAdminServiceStub struct {
	service.AdminService
	entries []service.UserGroupRateEntry
}

func (s *groupRPMOverridesAdminServiceStub) GetGroupRateMultipliers(_ context.Context, _ int64) ([]service.UserGroupRateEntry, error) {
	return s.entries, nil
}

func TestGroupRPMOverridesHandlerDoesNotExposeLegacyRateMultiplier(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rate := 0.5
	rpm := 42
	handler := NewGroupHandler(&groupRPMOverridesAdminServiceStub{entries: []service.UserGroupRateEntry{
		{UserID: 1, UserEmail: "rate-only@example.com", RateMultiplier: &rate},
		{UserID: 2, UserEmail: "rpm@example.com", RateMultiplier: &rate, RPMOverride: &rpm},
	}}, nil, nil)
	router := gin.New()
	router.GET("/groups/:id/rpm-overrides", handler.GetGroupRPMOverrides)

	req := httptest.NewRequest(http.MethodGet, "/groups/7/rpm-overrides", nil)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, req)

	require.Equal(t, http.StatusOK, writer.Code)
	var response struct {
		Data []map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(writer.Body.Bytes(), &response))
	require.Len(t, response.Data, 1)
	require.Equal(t, float64(2), response.Data[0]["user_id"])
	require.Equal(t, float64(42), response.Data[0]["rpm_override"])
	require.NotContains(t, response.Data[0], "rate_multiplier")
}
