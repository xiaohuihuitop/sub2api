package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
)

func TestAccountSupportsOpenAIEndpointCapability(t *testing.T) {
	tests := []struct {
		name       string
		account    *Account
		capability OpenAIEndpointCapability
		want       bool
	}{
		{
			name:       "missing capability requirement is allowed",
			account:    &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey},
			capability: "",
			want:       true,
		},
		{
			name:       "openai account without explicit capabilities preserves old behavior",
			account:    &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey},
			capability: OpenAIEndpointCapabilityChatCompletions,
			want:       true,
		},
		{
			name:       "openai account without explicit capabilities supports responses",
			account:    &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth},
			capability: OpenAIEndpointCapabilityResponses,
			want:       true,
		},
		{
			name: "array capability allows chat completions",
			account: &Account{
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
				Credentials: map[string]any{
					"openai_capabilities": []any{"chat_completions"},
				},
			},
			capability: OpenAIEndpointCapabilityChatCompletions,
			want:       true,
		},
		{
			name: "chat completions only blocks responses",
			account: &Account{
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
				Credentials: map[string]any{
					"openai_capabilities": []any{"chat_completions"},
				},
			},
			capability: OpenAIEndpointCapabilityResponses,
			want:       false,
		},
		{
			name: "responses probe failure blocks responses capability",
			account: &Account{
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
				Extra: map[string]any{
					openai_compat.ExtraKeyResponsesSupported: false,
				},
			},
			capability: OpenAIEndpointCapabilityResponses,
			want:       false,
		},
		{
			name: "force chat completions blocks responses capability",
			account: &Account{
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
				Extra: map[string]any{
					openai_compat.ExtraKeyResponsesMode: string(openai_compat.ResponsesSupportModeForceChatCompletions),
				},
			},
			capability: OpenAIEndpointCapabilityResponses,
			want:       false,
		},
		{
			name: "map capability blocks chat completions when disabled",
			account: &Account{
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
				Credentials: map[string]any{
					"openai_capabilities": map[string]any{
						"chat_completions": false,
						"embeddings":       true,
					},
				},
			},
			capability: OpenAIEndpointCapabilityChatCompletions,
			want:       false,
		},
		{
			name: "embeddings require API key account",
			account: &Account{
				Platform: PlatformOpenAI,
				Type:     AccountTypeOAuth,
				Credentials: map[string]any{
					"openai_capabilities": []string{"embeddings"},
				},
			},
			capability: OpenAIEndpointCapabilityEmbeddings,
			want:       false,
		},
		{
			name:       "non openai account is rejected",
			account:    &Account{Platform: PlatformAnthropic, Type: AccountTypeAPIKey},
			capability: OpenAIEndpointCapabilityChatCompletions,
			want:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.account.SupportsOpenAIEndpointCapability(tc.capability); got != tc.want {
				t.Fatalf("SupportsOpenAIEndpointCapability() = %v, want %v", got, tc.want)
			}
		})
	}
}
