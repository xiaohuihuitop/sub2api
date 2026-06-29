package repository

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestOpenAIAccountEndpointCapabilities_DefaultsToResponsesOnly(t *testing.T) {
	account := &service.Account{
		Platform: service.PlatformOpenAI,
		Type:     service.AccountTypeAPIKey,
	}

	require.Equal(
		t,
		[]service.OpenAIEndpointCapability{service.OpenAIEndpointCapabilityResponses},
		openAIAccountEndpointCapabilities(account),
	)
}

func TestOpenAIAccountEndpointCapabilities_EmbeddingsOnlyDoesNotEnableChatOrResponses(t *testing.T) {
	account := &service.Account{
		Platform: service.PlatformOpenAI,
		Type:     service.AccountTypeAPIKey,
		Credentials: map[string]any{
			"openai_capabilities": []any{"embeddings"},
		},
	}

	require.Empty(t, openAIAccountEndpointCapabilities(account))
}
