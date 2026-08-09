package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestForwardResponsesViaChatOverwritesActualEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"gpt-5.4","input":"hello","stream":false}`)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, openAIResponsesEndpoint, bytes.NewReader(body))
	SetActualOpenAIUpstreamEndpoint(c, openAIResponsesEndpoint)

	upstream := &httpUpstreamRecorder{resp: openAIEndpointMarkerErrorResponse()}
	svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}
	account := rawChatCompletionsTestAccount()
	account.Extra = map[string]any{"openai_responses_mode": "force_chat_completions"}

	_, err := svc.Forward(context.Background(), c, account, body)

	require.Error(t, err)
	require.NotNil(t, upstream.lastReq)
	require.Equal(t, grokChatRawEndpoint, upstream.lastReq.URL.Path)
	require.Equal(t, grokChatRawEndpoint, GetActualOpenAIUpstreamEndpoint(c))
}

func TestForwardResponsesOverwritesPreviousChatActualEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"gpt-5.4","input":"hello","stream":false}`)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, openAIResponsesEndpoint, bytes.NewReader(body))
	SetActualOpenAIUpstreamEndpoint(c, grokChatRawEndpoint)

	upstream := &httpUpstreamRecorder{resp: openAIEndpointMarkerErrorResponse()}
	svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}
	account := rawChatCompletionsTestAccount()
	account.Extra = map[string]any{"openai_responses_mode": "force_responses"}

	_, err := svc.Forward(context.Background(), c, account, body)

	require.Error(t, err)
	require.NotNil(t, upstream.lastReq)
	require.Equal(t, openAIResponsesEndpoint, upstream.lastReq.URL.Path)
	require.Equal(t, openAIResponsesEndpoint, GetActualOpenAIUpstreamEndpoint(c))
}

func TestForwardResponsesCompactMarksFullActualEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"gpt-5.4","input":"hello"}`)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, openAIResponsesEndpoint+"/compact", bytes.NewReader(body))

	upstream := &httpUpstreamRecorder{resp: openAIEndpointMarkerErrorResponse()}
	svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}
	account := rawChatCompletionsTestAccount()
	account.Extra = map[string]any{"openai_responses_mode": "force_responses"}

	_, err := svc.Forward(context.Background(), c, account, body)

	require.Error(t, err)
	require.NotNil(t, upstream.lastReq)
	require.Equal(t, openAIResponsesEndpoint+"/compact", upstream.lastReq.URL.Path)
	require.Equal(t, openAIResponsesEndpoint+"/compact", GetActualOpenAIUpstreamEndpoint(c))
}

func openAIEndpointMarkerErrorResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusBadRequest,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"error":{"type":"invalid_request_error","message":"marker test"}}`)),
	}
}
