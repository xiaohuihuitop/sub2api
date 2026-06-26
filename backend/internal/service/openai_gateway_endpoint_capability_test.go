package service

import (
	"context"
	"testing"
)

func TestOpenAISelectAccountWithSchedulerForCapability_FiltersEndpointCapability(t *testing.T) {
	groupID := int64(1)
	embeddingsOnly := Account{
		ID:          1,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Priority:    0,
		Credentials: map[string]any{
			"openai_capabilities": []any{"embeddings"},
		},
	}
	chatCapable := Account{
		ID:          2,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Priority:    1,
		Credentials: map[string]any{
			"openai_capabilities": []any{"chat_completions"},
		},
	}

	svc := &OpenAIGatewayService{
		accountRepo:        stubOpenAIAccountRepo{accounts: []Account{embeddingsOnly, chatCapable}},
		concurrencyService: NewConcurrencyService(stubConcurrencyCache{}),
	}

	selection, _, err := svc.SelectAccountWithSchedulerForCapability(
		context.Background(),
		&groupID,
		"",
		"",
		"gpt-4",
		nil,
		OpenAIUpstreamTransportAny,
		OpenAIEndpointCapabilityChatCompletions,
		false,
	)
	if err != nil {
		t.Fatalf("SelectAccountWithSchedulerForCapability error: %v", err)
	}
	if selection == nil || selection.Account == nil {
		t.Fatalf("expected selection with account")
	}
	if selection.Account.ID != chatCapable.ID {
		t.Fatalf("expected account %d, got %d", chatCapable.ID, selection.Account.ID)
	}
	if selection.ReleaseFunc != nil {
		selection.ReleaseFunc()
	}
}
