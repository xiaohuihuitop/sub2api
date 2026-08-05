package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidatePlatformModelRulesRejectsCrossPlatformOverlap(t *testing.T) {
	err := validatePlatformModelRules([]PlatformModelRule{
		{PlatformID: 1, ModelPattern: "gpt-*", Enabled: true},
		{PlatformID: 2, ModelPattern: "gpt-4o", Enabled: true},
	})

	require.ErrorContains(t, err, "overlaps")
}

func TestResolvePlatformModelUsesExactBeforeSuffixWildcard(t *testing.T) {
	resolver := newPlatformModelResolver([]PlatformModelRule{
		{PlatformID: 1, PlatformCode: PlatformOpenAI, ModelPattern: "gpt-*", Enabled: true},
		{PlatformID: 2, PlatformCode: "grok", ModelPattern: "gpt-4o", UpstreamModel: "gpt-4o-2024-08-06", Enabled: true},
	})

	got, err := resolver.Resolve("gpt-4o")

	require.NoError(t, err)
	require.Equal(t, int64(2), got.PlatformID)
	require.Equal(t, "grok", got.PlatformCode)
	require.Equal(t, "gpt-4o-2024-08-06", got.UpstreamModel)
}

func TestResolvePlatformModelRejectsAnUnmatchedModel(t *testing.T) {
	resolver := newPlatformModelResolver([]PlatformModelRule{
		{PlatformID: 1, PlatformCode: PlatformOpenAI, ModelPattern: "gpt-*", Enabled: true},
	})

	_, err := resolver.Resolve("claude-sonnet-4")

	require.ErrorIs(t, err, ErrPlatformModelNotFound)
}
