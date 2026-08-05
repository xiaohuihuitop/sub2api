package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
)

// PlatformSchedulingScope identifies one V2 provider account pool. AccountPlatform
// is the internal protocol adapter, while PlatformID is the business-pool boundary.
type PlatformSchedulingScope struct {
	PlatformID      int64
	PlatformCode    string
	AccountPlatform string
}

// PlatformPoolAccountLister stays deliberately narrow so legacy repository test
// doubles do not need to implement V2-only scheduling methods.
type PlatformPoolAccountLister interface {
	ListSchedulableByPlatformPool(ctx context.Context, platformID int64, accountPlatform string) ([]Account, error)
}

type PlatformAccountBindingValidator interface {
	ValidatePlatformAccountBinding(ctx context.Context, platformID int64, accountPlatform string) error
}

func WithPlatformSchedulingScope(ctx context.Context, scope PlatformSchedulingScope) context.Context {
	normalized, ok := normalizePlatformSchedulingScope(scope)
	if ctx == nil || !ok {
		return ctx
	}
	return context.WithValue(ctx, ctxkey.PlatformSchedulingScope, normalized)
}

func PlatformSchedulingScopeFromContext(ctx context.Context) (PlatformSchedulingScope, bool) {
	if ctx == nil {
		return PlatformSchedulingScope{}, false
	}
	scope, ok := ctx.Value(ctxkey.PlatformSchedulingScope).(PlatformSchedulingScope)
	if !ok {
		return PlatformSchedulingScope{}, false
	}
	return normalizePlatformSchedulingScope(scope)
}

func listPlatformPoolSchedulableAccounts(
	ctx context.Context,
	lister PlatformPoolAccountLister,
	scope PlatformSchedulingScope,
) ([]Account, error) {
	normalized, ok := normalizePlatformSchedulingScope(scope)
	if !ok {
		return nil, fmt.Errorf("%w: invalid platform scheduling scope", ErrPlatformInvalid)
	}
	if lister == nil {
		return nil, fmt.Errorf("%w: platform account pool is not configured", ErrPlatformInvalid)
	}
	return lister.ListSchedulableByPlatformPool(ctx, normalized.PlatformID, normalized.AccountPlatform)
}

func normalizePlatformSchedulingScope(scope PlatformSchedulingScope) (PlatformSchedulingScope, bool) {
	scope.PlatformCode = strings.ToLower(strings.TrimSpace(scope.PlatformCode))
	scope.AccountPlatform = strings.ToLower(strings.TrimSpace(scope.AccountPlatform))
	return scope, scope.PlatformID > 0 && scope.AccountPlatform != ""
}

// platformSchedulingCacheGroupID keeps V2 sticky-session entries separate from
// legacy group entries. Legacy IDs are positive and -1 is reserved by account
// list filters, so -(platformID+1) is a stable non-colliding namespace.
func platformSchedulingCacheGroupID(scope PlatformSchedulingScope) *int64 {
	normalized, ok := normalizePlatformSchedulingScope(scope)
	if !ok {
		return nil
	}
	cacheID := -normalized.PlatformID - 1
	return &cacheID
}

func accountMatchesPlatformSchedulingScope(ctx context.Context, account *Account) bool {
	scope, scoped := PlatformSchedulingScopeFromContext(ctx)
	return !scoped || platformSchedulingScopeMatchesAccount(scope, account)
}

func platformSchedulingScopeMatchesAccount(scope PlatformSchedulingScope, account *Account) bool {
	normalized, ok := normalizePlatformSchedulingScope(scope)
	if !ok || account == nil || account.PlatformID == nil {
		return false
	}
	return *account.PlatformID == normalized.PlatformID &&
		strings.EqualFold(strings.TrimSpace(account.Platform), normalized.AccountPlatform)
}
