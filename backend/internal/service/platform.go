package service

import (
	"context"
	"time"
)

const PlatformStatusActive = "active"

// Platform owns one provider-specific account pool. LegacyGroupID only bridges
// the existing pricing context while the old group path remains available.
type Platform struct {
	ID                   int64
	Code                 string
	Name                 string
	AccountPlatform      string
	Status               string
	EndpointCapabilities map[string]bool
	SchedulingConfig     map[string]any
	LegacyGroupID        *int64
	ModelRules           []PlatformModelRule
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (p Platform) IsActive() bool {
	return p.Status == "" || p.Status == PlatformStatusActive
}

// PlatformModelRule maps one requested model name or a suffix wildcard to a
// platform. The persisted status is adapted to Enabled at the repository edge.
type PlatformModelRule struct {
	ID                   int64
	PlatformID           int64
	PlatformCode         string
	AccountPlatform      string
	LegacyGroupID        *int64
	ModelPattern         string
	UpstreamModel        string
	EndpointCapabilities []string
	Enabled              bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// PlatformModelResolver resolves a client model to its unique active platform.
// It keeps the request path dependent on a small capability rather than on the
// full administrator service implementation.
type PlatformModelResolver interface {
	ResolveModel(ctx context.Context, requestedModel string) (*ResolvedPlatformModel, error)
}

// ResolvedPlatformModel is the only valid platform target for a client model.
type ResolvedPlatformModel struct {
	PlatformID           int64
	PlatformCode         string
	AccountPlatform      string
	RequestedModel       string
	UpstreamModel        string
	EndpointCapabilities []string
	LegacyGroupID        *int64
	RuleID               int64
}
