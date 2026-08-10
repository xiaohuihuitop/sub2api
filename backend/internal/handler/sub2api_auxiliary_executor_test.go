//go:build unit

package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/gatewayruntime"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSub2APIAuxiliaryExecutorRoutesPlatformCapabilities(t *testing.T) {
	gateway := &GatewayHandler{}
	openai := &OpenAIGatewayHandler{}
	tests := []struct {
		endpoint gatewayruntime.Endpoint
		adapter  string
		method   string
		path     string
	}{
		{gatewayruntime.EndpointGeminiNative, service.PlatformGemini, http.MethodPost, "/v1beta/models/gemini:generateContent"},
		{gatewayruntime.EndpointCountTokens, service.PlatformAnthropic, http.MethodPost, "/v1/messages/count_tokens"},
		{gatewayruntime.EndpointCountTokens, service.PlatformOpenAI, http.MethodPost, "/v1/messages/count_tokens"},
		{gatewayruntime.EndpointImages, service.PlatformOpenAI, http.MethodPost, "/v1/images/generations"},
		{gatewayruntime.EndpointImages, service.PlatformGrok, http.MethodPost, "/v1/images/generations"},
		{gatewayruntime.EndpointVideos, service.PlatformGrok, http.MethodPost, "/v1/videos/edits"},
		{gatewayruntime.EndpointLive, service.PlatformOpenAI, http.MethodPost, "/v1/live"},
	}
	for _, test := range tests {
		t.Run(string(test.endpoint)+"_"+test.adapter+"_"+test.method, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(test.method, test.path, nil)
			executor := sub2APIAuxiliaryExecutor{
				gatewayHandler: gateway,
				openAIHandler:  openai,
				endpoint:       test.endpoint,
			}
			require.NotNil(t, executor.handlerFor(gatewayruntime.Request{Exchange: NewGinHTTPExchange(c)}, test.adapter))
		})
	}
}
