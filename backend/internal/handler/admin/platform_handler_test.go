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

type platformHandlerServiceStub struct {
	platforms []service.Platform
	created   service.CreatePlatformInput
	updated   service.UpdatePlatformInput
	deletedID int64
	deleteErr error
}

func (s *platformHandlerServiceStub) List(context.Context) ([]service.Platform, error) {
	return append([]service.Platform(nil), s.platforms...), nil
}

func (s *platformHandlerServiceStub) GetByID(_ context.Context, id int64) (*service.Platform, error) {
	for index := range s.platforms {
		if s.platforms[index].ID == id {
			platform := s.platforms[index]
			return &platform, nil
		}
	}
	return nil, service.ErrPlatformNotFound
}

func (s *platformHandlerServiceStub) Create(_ context.Context, input service.CreatePlatformInput) (*service.Platform, error) {
	s.created = input
	return &service.Platform{ID: 7, Code: input.Code, Name: input.Name, AccountPlatform: input.AccountPlatform, Status: service.PlatformStatusActive, EndpointCapabilities: input.EndpointCapabilities, ModelRules: input.ModelRules}, nil
}

func (s *platformHandlerServiceStub) Update(_ context.Context, id int64, input service.UpdatePlatformInput) (*service.Platform, error) {
	s.updated = input
	return &service.Platform{ID: id, Code: "gpt", Name: "GPT", AccountPlatform: service.PlatformOpenAI, Status: service.PlatformStatusActive}, nil
}

func (s *platformHandlerServiceStub) Delete(_ context.Context, id int64) error {
	s.deletedID = id
	return s.deleteErr
}

func setupPlatformHandlerRouter(svc platformManagementService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewPlatformHandler(svc)
	router.GET("/api/v1/admin/platforms", handler.List)
	router.GET("/api/v1/admin/platforms/:id", handler.GetByID)
	router.POST("/api/v1/admin/platforms", handler.Create)
	router.PUT("/api/v1/admin/platforms/:id", handler.Update)
	router.DELETE("/api/v1/admin/platforms/:id", handler.Delete)
	return router
}

func TestPlatformHandlerDeletesUnusedPlatform(t *testing.T) {
	stub := &platformHandlerServiceStub{}
	router := setupPlatformHandlerRouter(stub)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodDelete, "/api/v1/admin/platforms/7", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, int64(7), stub.deletedID)
	require.JSONEq(t, `{"code":0,"message":"success"}`, recorder.Body.String())
}

func TestPlatformHandlerReturnsConflictForReferencedPlatform(t *testing.T) {
	stub := &platformHandlerServiceStub{
		deleteErr: service.ErrPlatformInUse.WithMetadata(map[string]string{"accounts": "1"}),
	}
	router := setupPlatformHandlerRouter(stub)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodDelete, "/api/v1/admin/platforms/7", nil))

	require.Equal(t, http.StatusConflict, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"reason":"PLATFORM_IN_USE"`)
	require.Contains(t, recorder.Body.String(), `"accounts":"1"`)
}

func TestPlatformHandlerRejectsInvalidDeleteID(t *testing.T) {
	router := setupPlatformHandlerRouter(&platformHandlerServiceStub{})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodDelete, "/api/v1/admin/platforms/0", nil))

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestPlatformHandlerReturnsNotFoundWhenDeletingMissingPlatform(t *testing.T) {
	stub := &platformHandlerServiceStub{deleteErr: service.ErrPlatformNotFound}
	router := setupPlatformHandlerRouter(stub)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodDelete, "/api/v1/admin/platforms/7", nil))

	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"reason":"PLATFORM_NOT_FOUND"`)
}

func TestPlatformHandlerListsAndCreatesPlatformPools(t *testing.T) {
	stub := &platformHandlerServiceStub{platforms: []service.Platform{{
		ID:              7,
		Code:            "gpt",
		Name:            "GPT",
		AccountPlatform: service.PlatformOpenAI,
		Status:          service.PlatformStatusActive,
	}}}
	router := setupPlatformHandlerRouter(stub)

	listRecorder := httptest.NewRecorder()
	router.ServeHTTP(listRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/admin/platforms", nil))
	require.Equal(t, http.StatusOK, listRecorder.Code)
	require.JSONEq(t, `{"code":0,"message":"success","data":[{"id":7,"code":"gpt","name":"GPT","account_platform":"openai","status":"active","endpoint_capabilities":[],"model_rules":[]}]}`, listRecorder.Body.String())

	body, err := json.Marshal(map[string]any{
		"code":                  "glm",
		"name":                  "GLM",
		"account_platform":      "openai",
		"endpoint_capabilities": []string{"chat_completions", "responses"},
		"model_rules": []map[string]any{{
			"model_pattern":  "glm-4-*",
			"upstream_model": "glm-4-plus",
		}},
	})
	require.NoError(t, err)
	createRecorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/platforms", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(createRecorder, request)

	require.Equal(t, http.StatusOK, createRecorder.Code)
	require.Equal(t, "glm", stub.created.Code)
	require.Len(t, stub.created.ModelRules, 1)
	require.True(t, stub.created.ModelRules[0].Enabled)
	require.Equal(t, []string{"chat_completions", "responses"}, stub.created.EndpointCapabilities)
}
