//go:build unit

package gatewayruntime

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/productcore"
	"github.com/stretchr/testify/require"
)

func TestDispatchIntentContextRoundTripsIndependentCopies(t *testing.T) {
	intent := &DispatchIntent{
		Platform:     productcore.Platform{ID: 3, EndpointCapabilities: []string{"responses"}},
		BillingAsset: &productcore.BillingAsset{Source: "balance", RateMultiplier: 1.25},
	}
	ctx := WithDispatchIntent(context.Background(), intent)
	intent.Platform.EndpointCapabilities[0] = "mutated-before-read"

	first, ok := DispatchIntentFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, "responses", first.Platform.EndpointCapabilities[0])

	first.Platform.EndpointCapabilities[0] = "mutated-after-read"
	second, ok := DispatchIntentFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, "responses", second.Platform.EndpointCapabilities[0])
}

func TestDispatchIntentAllowsNilBillingAsset(t *testing.T) {
	ctx := WithDispatchIntent(context.Background(), &DispatchIntent{
		Platform: productcore.Platform{ID: 3},
	})
	got, ok := DispatchIntentFromContext(ctx)
	require.True(t, ok)
	require.Nil(t, got.BillingAsset)
}
