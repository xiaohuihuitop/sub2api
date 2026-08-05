package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildAccountForCreatePreservesPlatformPool(t *testing.T) {
	platformID := int64(42)
	account, err := buildAccountForCreate(&CreateAccountInput{
		Name:       "gpt-account",
		Platform:   PlatformOpenAI,
		PlatformID: &platformID,
		Type:       AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key": "test-key",
		},
		Concurrency: 1,
	}, map[string]any{})

	require.NoError(t, err)
	require.Equal(t, &platformID, account.PlatformID)
}
