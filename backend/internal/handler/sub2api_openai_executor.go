package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/gatewayruntime"
)

// sub2APIOpenAIExecutor owns the OpenAI-compatible protocol family at the
// runtime boundary. The existing handler methods remain the protocol
// implementation until they can be extracted without changing forwarding,
// OAuth, streaming, or failover behavior.
type sub2APIOpenAIExecutor struct {
	handler  *OpenAIGatewayHandler
	endpoint gatewayruntime.Endpoint
	execute  func(context.Context, gatewayruntime.Request, gatewayruntime.UsageSink) (gatewayruntime.Result, error)
}

func (e sub2APIOpenAIExecutor) Execute(ctx context.Context, request gatewayruntime.Request, sink gatewayruntime.UsageSink) (gatewayruntime.Result, error) {
	terminal := &openAIExecutorTerminalSink{sink: sink}
	if e.execute != nil {
		result, err := e.execute(ctx, request, terminal)
		return e.ensureTerminal(ctx, request, result, err, terminal)
	}
	if e.handler == nil {
		return gatewayruntime.Result{}, ErrSub2APIRuntimeUnavailable
	}
	legacy := e.legacyHandler()
	if legacy == nil {
		return gatewayruntime.Result{}, ErrSub2APIRuntimeEndpointUnavailable
	}
	result, err := (legacyEndpointExecutor{handler: legacy}).Execute(ctx, request, terminal)
	return e.ensureTerminal(ctx, request, result, err, terminal)
}

func (e sub2APIOpenAIExecutor) ensureTerminal(
	ctx context.Context,
	request gatewayruntime.Request,
	result gatewayruntime.Result,
	err error,
	terminal *openAIExecutorTerminalSink,
) (gatewayruntime.Result, error) {
	if terminal.recorded() {
		return result, err
	}
	status := result.StatusCode
	if status == 0 {
		status = http.StatusBadGateway
	}
	event := gatewayruntime.UsageEvent{
		RequestID: request.RequestID,
		Success:   err == nil && status >= http.StatusOK && status < http.StatusBadRequest,
		Facts: gatewayruntime.UsageFacts{
			Adapter:                  request.Adapter,
			RequestedModel:           request.RequestedModel,
			UpstreamModel:            request.UpstreamModel,
			InboundEndpoint:          request.InboundEndpoint,
			RequestWasClientStream:   request.Stream,
			ResponseWasPartiallySent: result.Response.Streamed,
		},
	}
	if !event.Success {
		if err != nil {
			event.Error = gatewayruntime.RuntimeErrorFromContext(err)
		}
		if event.Error == nil {
			event.Error = gatewayruntime.RuntimeErrorFromStatus(status, http.StatusText(status))
		}
	}
	if recordErr := terminal.RecordFinal(ctx, event); recordErr != nil {
		if err != nil {
			return result, errors.Join(err, recordErr)
		}
		return result, recordErr
	}
	return result, err
}

func (e sub2APIOpenAIExecutor) legacyHandler() legacyGinHandler {
	if e.handler == nil {
		return nil
	}
	switch e.endpoint {
	case gatewayruntime.EndpointMessages:
		return e.handler.legacyMessages
	case gatewayruntime.EndpointChatCompletions:
		return e.handler.legacyChatCompletions
	case gatewayruntime.EndpointResponses:
		return e.handler.legacyResponses
	default:
		return nil
	}
}

type openAIExecutorTerminalSink struct {
	sink gatewayruntime.UsageSink
	seen bool
}

func (s *openAIExecutorTerminalSink) RecordFinal(ctx context.Context, event gatewayruntime.UsageEvent) error {
	if s == nil || s.sink == nil {
		return gatewayruntime.ErrUsageSinkUnavailable
	}
	if s.seen {
		return gatewayruntime.ErrTerminalAlreadyRecorded
	}
	s.seen = true
	return s.sink.RecordFinal(ctx, event)
}

func (s *openAIExecutorTerminalSink) recorded() bool {
	return s != nil && s.seen
}

var _ Sub2APIEndpointExecutor = (*sub2APIOpenAIExecutor)(nil)
