package middleware

import (
	"errors"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/googleapi"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// APIKeyAuthGoogle is a Google-style error wrapper for API key auth.
func APIKeyAuthGoogle(apiKeyService *service.APIKeyService, cfg *config.Config) gin.HandlerFunc {
	return APIKeyAuthWithSubscriptionGoogle(apiKeyService, nil, cfg)
}

// APIKeyAuthWithSubscriptionGoogle behaves like API key auth for Gemini-compatible endpoints.
func APIKeyAuthWithSubscriptionGoogle(apiKeyService *service.APIKeyService, subscriptionService *service.SubscriptionService, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if v := strings.TrimSpace(c.Query("api_key")); v != "" {
			abortWithGoogleError(c, 400, "Query parameter api_key is deprecated. Use Authorization header or key instead.")
			return
		}

		apiKeyString := extractAPIKeyForGoogle(c)
		if apiKeyString == "" {
			abortWithGoogleError(c, 401, "API key is required")
			return
		}

		apiKey, err := apiKeyService.GetByKey(c.Request.Context(), apiKeyString)
		if err != nil {
			if errors.Is(err, service.ErrAPIKeyNotFound) {
				abortWithGoogleError(c, 401, "Invalid API key")
				return
			}
			abortWithGoogleError(c, 500, "Failed to validate API key")
			return
		}

		if !apiKey.IsActive() &&
			apiKey.Status != service.StatusAPIKeyExpired &&
			apiKey.Status != service.StatusAPIKeyQuotaExhausted {
			abortWithGoogleError(c, 401, "API key is disabled")
			return
		}
		if apiKey.User == nil {
			abortWithGoogleError(c, 401, "User associated with API key not found")
			return
		}
		if !apiKey.User.IsActive() {
			abortWithGoogleError(c, 401, "User account is not active")
			return
		}

		targetPlatform, _ := c.Request.Context().Value(ctxkey.ForcePlatform).(string)
		if targetPlatform == "" {
			targetPlatform = service.PlatformGemini
		}

		subscription, err := apiKeyService.ResolveBillingGroupForRequest(
			c.Request.Context(),
			apiKey,
			subscriptionService,
			false,
			targetPlatform,
		)
		if err != nil {
			handleGoogleBillingResolutionError(c, err)
			return
		}

		if cfg.RunMode == config.RunModeSimple {
			setAuthContext(c, apiKey, subscription)
			_ = apiKeyService.TouchLastUsed(c.Request.Context(), apiKey.ID)
			c.Next()
			return
		}

		switch apiKey.Status {
		case service.StatusAPIKeyQuotaExhausted:
			abortWithGoogleError(c, 429, "API key quota exhausted")
			return
		case service.StatusAPIKeyExpired:
			abortWithGoogleError(c, 403, "API key expired")
			return
		}
		if apiKey.IsExpired() {
			abortWithGoogleError(c, 403, "API key expired")
			return
		}
		if apiKey.IsQuotaExhausted() {
			abortWithGoogleError(c, 429, "API key quota exhausted")
			return
		}

		setAuthContext(c, apiKey, subscription)
		_ = apiKeyService.TouchLastUsed(c.Request.Context(), apiKey.ID)
		c.Next()
	}
}

// extractAPIKeyForGoogle extracts API key for Google/Gemini endpoints.
// Priority: x-goog-api-key > Authorization: Bearer > x-api-key > query key.
func extractAPIKeyForGoogle(c *gin.Context) string {
	if k := strings.TrimSpace(c.GetHeader("x-goog-api-key")); k != "" {
		return k
	}

	auth := strings.TrimSpace(c.GetHeader("Authorization"))
	if auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			if k := strings.TrimSpace(parts[1]); k != "" {
				return k
			}
		}
	}

	if k := strings.TrimSpace(c.GetHeader("x-api-key")); k != "" {
		return k
	}

	if allowGoogleQueryKey(c.Request.URL.Path) {
		if v := strings.TrimSpace(c.Query("key")); v != "" {
			return v
		}
	}

	return ""
}

func allowGoogleQueryKey(path string) bool {
	return strings.HasPrefix(path, "/v1beta") || strings.HasPrefix(path, "/antigravity/v1beta")
}

func handleGoogleBillingResolutionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInsufficientBalance):
		abortWithGoogleError(c, 403, "Insufficient account balance")
	case errors.Is(err, service.ErrDailyLimitExceeded),
		errors.Is(err, service.ErrWeeklyLimitExceeded),
		errors.Is(err, service.ErrMonthlyLimitExceeded):
		abortWithGoogleError(c, 429, err.Error())
	case errors.Is(err, service.ErrSubscriptionNotFound):
		abortWithGoogleError(c, 403, "No active subscription found for this group")
	case errors.Is(err, service.ErrNoUsableBillingGroup):
		abortWithGoogleError(c, 403, "No usable billing group is available")
	default:
		abortWithGoogleError(c, 403, err.Error())
	}
}

func abortWithGoogleError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    status,
			"message": message,
			"status":  googleapi.HTTPStatusToGoogleStatus(status),
		},
	})
	c.Abort()
}
