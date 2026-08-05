package productcore

import "context"

type PlatformCatalog interface {
	ResolveModel(context.Context, string) (*Platform, error)
}

type AssetSelector interface {
	Select(context.Context, AccessGrant, bool) (*BillingAsset, error)
}
