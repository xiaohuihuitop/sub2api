package handler

import (
	"context"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/gatewayruntime"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// sub2APIAuxiliaryExecutor keeps the remaining Sub2API protocol handlers
// behind the same runtime boundary. Billing-capable handlers publish their
// existing usage facts through the injected sink; capability-only endpoints
// receive the ApplicationGateway non-billing sink.
type sub2APIAuxiliaryExecutor struct {
	gatewayHandler *GatewayHandler
	openAIHandler  *OpenAIGatewayHandler
	endpoint       gatewayruntime.Endpoint
}

func (e sub2APIAuxiliaryExecutor) Execute(ctx context.Context, request gatewayruntime.Request, sink gatewayruntime.UsageSink) (gatewayruntime.Result, error) {
	route, _ := service.GatewayPlatformAssetContextFromContext(ctx)
	if route == nil || route.Platform == nil {
		if compatibilityRoute := runtimeCompatibilityRoute(request); compatibilityRoute != nil {
			ctx = service.WithGatewayPlatformAssetContext(ctx, compatibilityRoute)
			route = compatibilityRoute
		}
	}
	if route == nil || route.Platform == nil {
		return gatewayruntime.Result{}, service.ErrAPIKeyPlatformForbidden
	}
	handler := e.handlerFor(request, route.Platform.AccountPlatform)
	if handler == nil {
		return gatewayruntime.Result{}, ErrSub2APIRuntimeEndpointUnavailable
	}
	return (legacyEndpointExecutor{handler: handler}).Execute(ctx, request, sink)
}

func (e sub2APIAuxiliaryExecutor) handlerFor(request gatewayruntime.Request, adapter string) legacyGinHandler {
	adapter = strings.ToLower(strings.TrimSpace(adapter))
	switch e.endpoint {
	case gatewayruntime.EndpointGeminiNative:
		if e.gatewayHandler != nil {
			return e.gatewayHandler.legacyGeminiV1BetaModels
		}
	case gatewayruntime.EndpointCountTokens:
		if adapter == service.PlatformOpenAI && e.openAIHandler != nil {
			return e.openAIHandler.legacyCountTokens
		}
		if adapter == service.PlatformGrok && e.openAIHandler != nil {
			return e.openAIHandler.legacyGrokCountTokens
		}
		if e.gatewayHandler != nil {
			return e.gatewayHandler.legacyCountTokens
		}
	case gatewayruntime.EndpointEmbeddings:
		if e.openAIHandler != nil {
			return e.openAIHandler.legacyEmbeddings
		}
	case gatewayruntime.EndpointAlphaSearch:
		if e.openAIHandler != nil {
			return e.openAIHandler.legacyAlphaSearch
		}
	case gatewayruntime.EndpointImages:
		if e.openAIHandler != nil {
			if adapter == service.PlatformGrok {
				return e.openAIHandler.legacyGrokImages
			}
			return e.openAIHandler.legacyImages
		}
	case gatewayruntime.EndpointVideos:
		return e.videoHandler(runtimeRequestPath(request))
	case gatewayruntime.EndpointLive:
		if e.openAIHandler != nil {
			return e.openAIHandler.legacyLive
		}
	}
	return nil
}

func (e sub2APIAuxiliaryExecutor) videoHandler(path string) legacyGinHandler {
	if e.openAIHandler == nil {
		return nil
	}
	switch {
	case strings.Contains(path, "/videos/edits"):
		return e.openAIHandler.legacyGrokVideoEdit
	case strings.Contains(path, "/videos/extensions"):
		return e.openAIHandler.legacyGrokVideoExtension
	default:
		return e.openAIHandler.legacyGrokVideoGeneration
	}
}

func runtimeRequestPath(request gatewayruntime.Request) string {
	if request.Exchange != nil && request.Exchange.Request() != nil && request.Exchange.Request().URL != nil {
		return request.Exchange.Request().URL.Path
	}
	return request.InboundEndpoint
}

var _ Sub2APIEndpointExecutor = (*sub2APIAuxiliaryExecutor)(nil)
