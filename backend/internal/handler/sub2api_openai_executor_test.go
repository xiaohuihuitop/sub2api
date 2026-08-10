//go:build unit

package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/gatewayruntime"
	"github.com/stretchr/testify/require"
)

type openAIExecutorSink struct {
	events []gatewayruntime.UsageEvent
}

func (s *openAIExecutorSink) RecordFinal(_ context.Context, event gatewayruntime.UsageEvent) error {
	s.events = append(s.events, event)
	return nil
}

func TestSub2APIOpenAIExecutorDispatchesRegisteredProtocol(t *testing.T) {
	for _, endpoint := range []gatewayruntime.Endpoint{
		gatewayruntime.EndpointMessages,
		gatewayruntime.EndpointChatCompletions,
		gatewayruntime.EndpointResponses,
	} {
		t.Run(string(endpoint), func(t *testing.T) {
			sink := &openAIExecutorSink{}
			executor := sub2APIOpenAIExecutor{
				execute: func(context.Context, gatewayruntime.Request, gatewayruntime.UsageSink) (gatewayruntime.Result, error) {
					return gatewayruntime.Result{StatusCode: http.StatusOK}, nil
				},
			}
			result, err := executor.Execute(context.Background(), gatewayruntime.Request{Endpoint: endpoint}, sink)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, result.StatusCode)
			require.Len(t, sink.events, 1)
			require.True(t, sink.events[0].Success)
		})
	}
}
