package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/applicationgateway"
	"github.com/Wei-Shaw/sub2api/internal/gatewayruntime"
	"github.com/Wei-Shaw/sub2api/internal/productcore"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

var ErrLegacyRuntimeExchangeUnavailable = errors.New("legacy runtime exchange is unavailable")

type legacyGinHandler func(*gin.Context)

type ginContextCarrier interface {
	GinContext() *gin.Context
}

func (e *ginHTTPExchange) GinContext() *gin.Context {
	if e == nil {
		return nil
	}
	return e.context
}

type legacyEndpointExecutor struct {
	handler legacyGinHandler
}

func (e legacyEndpointExecutor) Execute(
	ctx context.Context,
	request gatewayruntime.Request,
	sink gatewayruntime.UsageSink,
) (gatewayruntime.Result, error) {
	carrier, ok := request.Exchange.(ginContextCarrier)
	if !ok || carrier.GinContext() == nil || e.handler == nil {
		return gatewayruntime.Result{}, ErrLegacyRuntimeExchangeUnavailable
	}
	c := carrier.GinContext()
	originalRequest := c.Request
	if originalRequest != nil {
		baseContext := ctx
		if baseContext == nil {
			baseContext = originalRequest.Context()
		}
		c.Request = originalRequest.WithContext(gatewayruntime.WithUsageSink(baseContext, sink))
		defer func() { c.Request = originalRequest }()
	}
	e.handler(c)
	status := c.Writer.Status()
	event := gatewayruntime.UsageEvent{
		RequestID: request.RequestID,
		Success:   status >= http.StatusOK && status < http.StatusBadRequest,
		Facts: gatewayruntime.UsageFacts{
			TerminalStatus:           http.StatusText(status),
			RequestWasClientStream:   request.Stream,
			ResponseWasPartiallySent: c.Writer.Size() > 0,
		},
	}
	if !event.Success {
		event.Error = gatewayruntime.RuntimeErrorFromStatus(status, http.StatusText(status))
	}
	if sink != nil {
		if err := sink.RecordFinal(ctx, event); err != nil && !errors.Is(err, gatewayruntime.ErrTerminalAlreadyRecorded) {
			return gatewayruntime.Result{}, err
		}
	}
	return gatewayruntime.Result{StatusCode: status, Response: gatewayruntime.Response{Streamed: request.Stream}}, nil
}

type contextDecisionProvider struct{}

func (contextDecisionProvider) Resolve(ctx context.Context, _ productcore.AccessGrant, _ productcore.Request) (*productcore.Decision, error) {
	route, ok := service.GatewayPlatformAssetContextFromContext(ctx)
	if !ok || route == nil || route.Platform == nil {
		return nil, service.ErrAPIKeyPlatformForbidden
	}
	platform := route.Platform
	decision := &productcore.Decision{
		Platform: productcore.Platform{
			ID:                   platform.PlatformID,
			Code:                 platform.PlatformCode,
			AccountPlatform:      platform.AccountPlatform,
			RequestedModel:       platform.RequestedModel,
			UpstreamModel:        platform.UpstreamModel,
			EndpointCapabilities: append([]string(nil), platform.EndpointCapabilities...),
			MatchPriority:        platform.MatchPriority,
		},
	}
	if route.BillingAsset != nil {
		asset := route.BillingAsset
		decision.BillingAsset = &productcore.BillingAsset{
			Source:         asset.Source,
			SubscriptionID: cloneInt64Pointer(asset.SubscriptionID),
			PlanID:         cloneInt64Pointer(asset.PlanID),
			RateMultiplier: asset.RateMultiplier,
		}
	}
	return decision, nil
}

type noOpRuntimeUsageSink struct{}

func (noOpRuntimeUsageSink) RecordFinal(context.Context, gatewayruntime.UsageEvent) error { return nil }

type noOpRuntimeUsageSinkFactory struct{}

func (noOpRuntimeUsageSinkFactory) ForDecision(applicationgateway.DecisionSnapshot) gatewayruntime.UsageSink {
	return noOpRuntimeUsageSink{}
}

func (h *GatewayHandler) dispatchLegacyEndpoint(c *gin.Context, endpoint gatewayruntime.Endpoint, legacy legacyGinHandler) error {
	if h != nil && h.applicationGateway != nil {
		return dispatchLegacyEndpointWithGateway(c, endpoint, legacy, h.applicationGateway)
	}
	return dispatchLegacyEndpoint(c, endpoint, legacy)
}

func dispatchLegacyEndpoint(c *gin.Context, endpoint gatewayruntime.Endpoint, legacy legacyGinHandler) error {
	adapter := NewSub2APIRuntimeAdapter(map[gatewayruntime.Endpoint]Sub2APIEndpointExecutor{
		endpoint: legacyEndpointExecutor{handler: legacy},
	})
	gateway := applicationgateway.New(contextDecisionProvider{}, adapter, noOpRuntimeUsageSinkFactory{})
	return dispatchLegacyEndpointWithGateway(c, endpoint, legacy, gateway)
}

func dispatchLegacyEndpointWithGateway(c *gin.Context, endpoint gatewayruntime.Endpoint, legacy legacyGinHandler, gateway *applicationgateway.Gateway) error {
	if c == nil || c.Request == nil {
		return ErrLegacyRuntimeExchangeUnavailable
	}
	// Direct unit callers and malformed requests retain the existing auth error
	// envelope. Authenticated production requests must already carry a ProductCore route.
	if _, ok := middleware2.GetAPIKeyFromContext(c); !ok {
		if legacy == nil {
			return ErrLegacyRuntimeExchangeUnavailable
		}
		legacy(c)
		return nil
	}
	if _, ok := middleware2.GetAuthSubjectFromContext(c); !ok {
		if legacy == nil {
			return ErrLegacyRuntimeExchangeUnavailable
		}
		legacy(c)
		return nil
	}
	if gateway == nil {
		return ErrSub2APIRuntimeUnavailable
	}
	apiKey, _ := middleware2.GetAPIKeyFromContext(c)
	subject, _ := middleware2.GetAuthSubjectFromContext(c)
	grant := productcore.AccessGrant{UserID: subject.UserID}
	if apiKey != nil {
		grant.KeyID = apiKey.ID
		grant.PlatformIDs = append([]int64(nil), apiKey.AllowedPlatformIDs...)
		grant.SubscriptionPlanIDs = append([]int64(nil), apiKey.AllowedSubscriptionPlanIDs...)
		grant.AllowBalance = apiKey.AllowBalance
	}

	route, _ := service.GatewayPlatformAssetContextFromContext(c.Request.Context())
	model := ""
	if route != nil && route.Platform != nil {
		model = strings.TrimSpace(route.Platform.RequestedModel)
	}
	request := applicationgateway.DispatchRequest{
		Grant: grant,
		Product: productcore.Request{
			Model:              model,
			EndpointCapability: endpointCapabilityForRuntime(endpoint),
		},
		Runtime: gatewayruntime.Request{
			Endpoint:        endpoint,
			InboundEndpoint: c.Request.URL.Path,
			Stream:          requestLikelyStreams(c),
			Exchange:        NewGinHTTPExchange(c),
		},
	}
	returnError := func(err error) {
		if !c.Writer.Written() {
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"type": "server_error", "message": err.Error()}})
		}
	}
	_, err := gateway.Dispatch(c.Request.Context(), request)
	if err != nil {
		returnError(err)
	}
	return err
}

// NewSub2APIProductionApplicationGateway composes the existing protocol
// handlers behind one runtime adapter. The handlers remain protocol-specific;
// the application boundary is shared and is the only production dispatch path.
func NewSub2APIProductionApplicationGateway(gatewayHandler *GatewayHandler, openaiHandler *OpenAIGatewayHandler, apiKeys service.ProductUsageAPIKeyLoader) *applicationgateway.Gateway {
	executors := make(map[gatewayruntime.Endpoint]Sub2APIEndpointExecutor, 12)
	for _, endpoint := range []gatewayruntime.Endpoint{
		gatewayruntime.EndpointMessages,
		gatewayruntime.EndpointChatCompletions,
		gatewayruntime.EndpointResponses,
	} {
		executors[endpoint] = sub2APIMessagesExecutor{
			gatewayHandler: gatewayHandler,
			openaiHandler:  openaiHandler,
			endpoint:       endpoint,
		}
	}
	for _, endpoint := range []gatewayruntime.Endpoint{
		gatewayruntime.EndpointGeminiNative,
		gatewayruntime.EndpointEmbeddings,
		gatewayruntime.EndpointAlphaSearch,
		gatewayruntime.EndpointImages,
		gatewayruntime.EndpointVideos,
		gatewayruntime.EndpointCountTokens,
		gatewayruntime.EndpointLive,
	} {
		executors[endpoint] = sub2APIAuxiliaryExecutor{
			gatewayHandler: gatewayHandler,
			openAIHandler:  openaiHandler,
			endpoint:       endpoint,
		}
	}
	adapter := NewSub2APIRuntimeAdapter(executors)
	usageFactory := service.NewSub2APIProductUsageSinkFactory(
		gatewayServiceFromHandler(gatewayHandler),
		openAIGatewayServiceFromHandler(openaiHandler),
		apiKeys,
	)
	return applicationgateway.New(contextDecisionProvider{}, adapter, usageFactory)
}

func gatewayServiceFromHandler(h *GatewayHandler) *service.GatewayService {
	if h == nil {
		return nil
	}
	return h.gatewayService
}

func openAIGatewayServiceFromHandler(h *OpenAIGatewayHandler) *service.OpenAIGatewayService {
	if h == nil {
		return nil
	}
	return h.gatewayService
}

func (h *OpenAIGatewayHandler) dispatchLegacyEndpoint(c *gin.Context, endpoint gatewayruntime.Endpoint, legacy legacyGinHandler) error {
	if h != nil && h.applicationGateway != nil {
		return dispatchLegacyEndpointWithGateway(c, endpoint, legacy, h.applicationGateway)
	}
	return dispatchLegacyEndpoint(c, endpoint, legacy)
}

func endpointCapabilityForRuntime(endpoint gatewayruntime.Endpoint) string {
	switch endpoint {
	case gatewayruntime.EndpointChatCompletions:
		return "chat_completions"
	case gatewayruntime.EndpointResponses:
		return "responses"
	case gatewayruntime.EndpointMessages:
		return "messages"
	default:
		return string(endpoint)
	}
}

func requestLikelyStreams(c *gin.Context) bool {
	if c == nil || c.Request == nil {
		return false
	}
	return strings.EqualFold(c.GetHeader("Accept"), "text/event-stream")
}

func cloneInt64Pointer(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
