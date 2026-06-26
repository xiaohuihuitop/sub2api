package service

import "testing"

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
