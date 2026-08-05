package middleware

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

type platformAssetRequestModelReader func(*gin.Context) (string, error)
type platformAssetErrorWriter func(*gin.Context, int, string, string)

// NewPlatformAssetAuthorizationMiddleware activates only for API keys with
// explicit platform permissions. Legacy keys continue through the group-based
// request path unchanged.
func NewPlatformAssetAuthorizationMiddleware(
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	resolver service.PlatformModelResolver,
	cfg *config.Config,
) gin.HandlerFunc {
	return newPlatformAssetAuthorizationMiddleware(
		apiKeyService,
		subscriptionService,
		resolver,
		cfg,
		platformAssetJSONRequestModel,
		abortPlatformAssetRequestError,
	)
}

// NewPlatformAssetAuthorizationGoogleMiddleware applies the same V2 routing
// and billing policy to Gemini-native endpoints while preserving Google's error
// response envelope.
func NewPlatformAssetAuthorizationGoogleMiddleware(
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	resolver service.PlatformModelResolver,
	cfg *config.Config,
) gin.HandlerFunc {
	return newPlatformAssetAuthorizationMiddleware(
		apiKeyService,
		subscriptionService,
		resolver,
		cfg,
		platformAssetGoogleRequestModel,
		abortPlatformAssetGoogleRequestError,
	)
}

func newPlatformAssetAuthorizationMiddleware(
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	resolver service.PlatformModelResolver,
	cfg *config.Config,
	readModel platformAssetRequestModelReader,
	writeError platformAssetErrorWriter,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey, ok := GetAPIKeyFromContext(c)
		if !ok || !service.UsesPlatformAssetPermissions(apiKey) {
			c.Next()
			return
		}
		model, err := readModel(c)
		if err != nil {
			writeError(c, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
			return
		}
		if model == "" {
			c.Next()
			return
		}

		skipBilling := cfg != nil && cfg.RunMode == config.RunModeSimple
		route, err := apiKeyService.ResolvePlatformAssetRequest(
			c.Request.Context(),
			apiKey,
			resolver,
			subscriptionService,
			model,
			apiKeyBillingRequestEndpoint(c),
			skipBilling,
		)
		if err != nil {
			abortPlatformAssetResolutionError(c, err, writeError)
			return
		}

		c.Request = c.Request.WithContext(service.WithGatewayPlatformAssetContext(c.Request.Context(), route))
		if route.BillingAsset != nil && route.BillingAsset.Subscription != nil {
			c.Set(string(ContextKeySubscription), route.BillingAsset.Subscription)
		}
		c.Next()
	}
}

func platformAssetJSONRequestModel(c *gin.Context) (string, error) {
	if !platformAssetRequestCarriesModel(c) {
		return "", nil
	}
	body, err := pkghttputil.ReadRequestBodyWithPrealloc(c.Request)
	if err != nil {
		return "", err
	}
	resetPlatformAssetRequestBody(c, body)
	return strings.TrimSpace(gjson.GetBytes(body, "model").String()), nil
}

func platformAssetGoogleRequestModel(c *gin.Context) (string, error) {
	if c == nil || c.Request == nil || c.Request.Method != http.MethodPost {
		return "", nil
	}
	modelAction := strings.TrimPrefix(strings.TrimSpace(c.Param("modelAction")), "/")
	if modelAction == "" {
		modelAction = strings.TrimSpace(c.Param("model"))
	}
	if index := strings.LastIndex(modelAction, ":"); index >= 0 {
		modelAction = modelAction[:index]
	}
	return strings.TrimSpace(modelAction), nil
}

func platformAssetRequestCarriesModel(c *gin.Context) bool {
	if c == nil || c.Request == nil || c.Request.Method != http.MethodPost {
		return false
	}
	path := strings.ToLower(c.Request.URL.Path)
	return strings.Contains(path, "/messages") ||
		strings.Contains(path, "/responses") ||
		strings.Contains(path, "/chat/completions") ||
		strings.Contains(path, "/embeddings") ||
		strings.Contains(path, "/alpha/search") ||
		strings.Contains(path, "/images/") ||
		strings.Contains(path, "/videos/") ||
		strings.Contains(path, "/realtime/calls")
}

func resetPlatformAssetRequestBody(c *gin.Context, body []byte) {
	if c == nil || c.Request == nil {
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	c.Request.ContentLength = int64(len(body))
	c.Request.Header.Set("Content-Length", strconv.Itoa(len(body)))
}

func abortPlatformAssetResolutionError(c *gin.Context, err error, writeError platformAssetErrorWriter) {
	switch {
	case errors.Is(err, service.ErrPlatformModelNotFound):
		writeError(c, http.StatusBadRequest, "PLATFORM_MODEL_NOT_FOUND", "Model is not available for this API key")
	case errors.Is(err, service.ErrAPIKeyPlatformForbidden):
		writeError(c, http.StatusForbidden, "API_KEY_PLATFORM_FORBIDDEN", "API key is not authorized for this model platform")
	case errors.Is(err, service.ErrPlatformEndpointUnsupported):
		writeError(c, http.StatusForbidden, "PLATFORM_ENDPOINT_UNSUPPORTED", "The model platform does not support this endpoint")
	case errors.Is(err, service.ErrInsufficientBalance):
		writeError(c, http.StatusForbidden, "INSUFFICIENT_BALANCE", "Insufficient account balance")
	case errors.Is(err, service.ErrNoUsableBillingSource):
		writeError(c, http.StatusForbidden, "NO_USABLE_BILLING_SOURCE", "No authorized subscription or balance source is available")
	case errors.Is(err, service.ErrDailyLimitExceeded),
		errors.Is(err, service.ErrWeeklyLimitExceeded),
		errors.Is(err, service.ErrMonthlyLimitExceeded):
		writeError(c, http.StatusTooManyRequests, "USAGE_LIMIT_EXCEEDED", err.Error())
	default:
		writeError(c, http.StatusInternalServerError, "PLATFORM_ASSET_RESOLUTION_FAILED", "Failed to resolve platform request")
	}
}

func abortPlatformAssetRequestError(c *gin.Context, status int, code, message string) {
	AbortWithError(c, status, code, message)
}

func abortPlatformAssetGoogleRequestError(c *gin.Context, status int, _, message string) {
	abortWithGoogleError(c, status, message)
}
