package service

import (
	"context"
	"fmt"
	"math"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/billingprofile"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// BillingProfile contains balance pricing for one routing group. It never
// contains subscription limits or subscription multipliers.
type BillingProfile struct {
	GroupID                      int64
	BalanceRateMultiplier        float64
	PeakRateEnabled              bool
	PeakStart                    string
	PeakEnd                      string
	PeakRateMultiplier           float64
	ImageRateIndependent         bool
	ImageRateMultiplier          float64
	ImagePrice1K                 *float64
	ImagePrice2K                 *float64
	ImagePrice4K                 *float64
	BatchImageDiscountMultiplier float64
	BatchImageHoldMultiplier     float64
	VideoRateIndependent         bool
	VideoRateMultiplier          float64
	VideoPrice480P               *float64
	VideoPrice720P               *float64
	VideoPrice1080P              *float64
	WebSearchPricePerCall        *float64
}

// UpdateBillingProfileInput replaces the balance billing terms for one group.
// A nil media price means the upstream default price should be used.
type UpdateBillingProfileInput struct {
	BalanceRateMultiplier        float64
	PeakRateEnabled              bool
	PeakStart                    string
	PeakEnd                      string
	PeakRateMultiplier           float64
	ImageRateIndependent         bool
	ImageRateMultiplier          float64
	ImagePrice1K                 *float64
	ImagePrice2K                 *float64
	ImagePrice4K                 *float64
	BatchImageDiscountMultiplier float64
	BatchImageHoldMultiplier     float64
	VideoRateIndependent         bool
	VideoRateMultiplier          float64
	VideoPrice480P               *float64
	VideoPrice720P               *float64
	VideoPrice1080P              *float64
	WebSearchPricePerCall        *float64
}

func (s *adminServiceImpl) GetGroupBillingProfile(ctx context.Context, groupID int64) (*BillingProfile, error) {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if s.entClient == nil {
		return billingProfileFromLegacyGroup(group), nil
	}

	entity, err := s.entClient.BillingProfile.Query().
		Where(billingprofile.GroupIDEQ(groupID)).
		Only(ctx)
	if dbent.IsNotFound(err) {
		return billingProfileFromLegacyGroup(group), nil
	}
	if err != nil {
		return nil, fmt.Errorf("get billing profile: %w", err)
	}
	return billingProfileFromEntity(entity), nil
}

func (s *adminServiceImpl) UpdateGroupBillingProfile(ctx context.Context, groupID int64, input *UpdateBillingProfileInput) (*BillingProfile, error) {
	if input == nil {
		return nil, infraerrors.BadRequest("BILLING_PROFILE_INVALID", "billing profile is required")
	}
	if _, err := s.groupRepo.GetByID(ctx, groupID); err != nil {
		return nil, err
	}
	if s.entClient == nil {
		return nil, infraerrors.InternalServer("BILLING_PROFILE_UNAVAILABLE", "billing profile storage is unavailable")
	}

	profile, err := billingProfileFromInput(groupID, input)
	if err != nil {
		return nil, err
	}
	entity, err := saveBillingProfile(ctx, s.entClient, profile)
	if err != nil {
		return nil, err
	}
	if s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByGroupID(ctx, groupID)
	}
	return billingProfileFromEntity(entity), nil
}

func billingProfileFromLegacyGroup(group *Group) *BillingProfile {
	if group == nil {
		return nil
	}
	return &BillingProfile{
		GroupID:                      group.ID,
		BalanceRateMultiplier:        group.RateMultiplier,
		PeakRateEnabled:              group.PeakRateEnabled,
		PeakStart:                    group.PeakStart,
		PeakEnd:                      group.PeakEnd,
		PeakRateMultiplier:           group.PeakRateMultiplier,
		ImageRateIndependent:         group.ImageRateIndependent,
		ImageRateMultiplier:          group.ImageRateMultiplier,
		ImagePrice1K:                 copyFloat64(group.ImagePrice1K),
		ImagePrice2K:                 copyFloat64(group.ImagePrice2K),
		ImagePrice4K:                 copyFloat64(group.ImagePrice4K),
		BatchImageDiscountMultiplier: group.BatchImageDiscountMultiplier,
		BatchImageHoldMultiplier:     group.BatchImageHoldMultiplier,
		VideoRateIndependent:         group.VideoRateIndependent,
		VideoRateMultiplier:          group.VideoRateMultiplier,
		VideoPrice480P:               copyFloat64(group.VideoPrice480P),
		VideoPrice720P:               copyFloat64(group.VideoPrice720P),
		VideoPrice1080P:              copyFloat64(group.VideoPrice1080P),
		WebSearchPricePerCall:        copyFloat64(group.WebSearchPricePerCall),
	}
}

func billingProfileFromEntity(entity *dbent.BillingProfile) *BillingProfile {
	if entity == nil {
		return nil
	}
	return &BillingProfile{
		GroupID:                      entity.GroupID,
		BalanceRateMultiplier:        entity.BalanceRateMultiplier,
		PeakRateEnabled:              entity.PeakRateEnabled,
		PeakStart:                    entity.PeakStart,
		PeakEnd:                      entity.PeakEnd,
		PeakRateMultiplier:           entity.PeakRateMultiplier,
		ImageRateIndependent:         entity.ImageRateIndependent,
		ImageRateMultiplier:          entity.ImageRateMultiplier,
		ImagePrice1K:                 copyFloat64(entity.ImagePrice1k),
		ImagePrice2K:                 copyFloat64(entity.ImagePrice2k),
		ImagePrice4K:                 copyFloat64(entity.ImagePrice4k),
		BatchImageDiscountMultiplier: entity.BatchImageDiscountMultiplier,
		BatchImageHoldMultiplier:     entity.BatchImageHoldMultiplier,
		VideoRateIndependent:         entity.VideoRateIndependent,
		VideoRateMultiplier:          entity.VideoRateMultiplier,
		VideoPrice480P:               copyFloat64(entity.VideoPrice480p),
		VideoPrice720P:               copyFloat64(entity.VideoPrice720p),
		VideoPrice1080P:              copyFloat64(entity.VideoPrice1080p),
		WebSearchPricePerCall:        copyFloat64(entity.WebSearchPricePerCall),
	}
}

func billingProfileFromInput(groupID int64, input *UpdateBillingProfileInput) (*BillingProfile, error) {
	profile := &BillingProfile{
		GroupID:                      groupID,
		BalanceRateMultiplier:        input.BalanceRateMultiplier,
		PeakRateEnabled:              input.PeakRateEnabled,
		PeakStart:                    input.PeakStart,
		PeakEnd:                      input.PeakEnd,
		PeakRateMultiplier:           input.PeakRateMultiplier,
		ImageRateIndependent:         input.ImageRateIndependent,
		ImageRateMultiplier:          input.ImageRateMultiplier,
		ImagePrice1K:                 copyFloat64(input.ImagePrice1K),
		ImagePrice2K:                 copyFloat64(input.ImagePrice2K),
		ImagePrice4K:                 copyFloat64(input.ImagePrice4K),
		BatchImageDiscountMultiplier: input.BatchImageDiscountMultiplier,
		BatchImageHoldMultiplier:     input.BatchImageHoldMultiplier,
		VideoRateIndependent:         input.VideoRateIndependent,
		VideoRateMultiplier:          input.VideoRateMultiplier,
		VideoPrice480P:               copyFloat64(input.VideoPrice480P),
		VideoPrice720P:               copyFloat64(input.VideoPrice720P),
		VideoPrice1080P:              copyFloat64(input.VideoPrice1080P),
		WebSearchPricePerCall:        copyFloat64(input.WebSearchPricePerCall),
	}
	if err := validateBillingProfile(profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func validateBillingProfile(profile *BillingProfile) error {
	if profile == nil || profile.GroupID <= 0 {
		return infraerrors.BadRequest("BILLING_PROFILE_INVALID", "group_id must be positive")
	}
	for _, value := range []struct {
		name  string
		value float64
	}{
		{"balance_rate_multiplier", profile.BalanceRateMultiplier},
		{"peak_rate_multiplier", profile.PeakRateMultiplier},
		{"image_rate_multiplier", profile.ImageRateMultiplier},
		{"batch_image_discount_multiplier", profile.BatchImageDiscountMultiplier},
		{"batch_image_hold_multiplier", profile.BatchImageHoldMultiplier},
		{"video_rate_multiplier", profile.VideoRateMultiplier},
	} {
		if !isNonNegativeFinite(value.value) {
			return infraerrors.BadRequest("BILLING_PROFILE_INVALID", value.name+" must be a finite number >= 0")
		}
	}
	if profile.BatchImageHoldMultiplier < profile.BatchImageDiscountMultiplier {
		return infraerrors.BadRequest("BILLING_PROFILE_INVALID", "batch_image_hold_multiplier must be >= batch_image_discount_multiplier")
	}
	for _, price := range []struct {
		name  string
		value *float64
	}{
		{"image_price_1k", profile.ImagePrice1K},
		{"image_price_2k", profile.ImagePrice2K},
		{"image_price_4k", profile.ImagePrice4K},
		{"video_price_480p", profile.VideoPrice480P},
		{"video_price_720p", profile.VideoPrice720P},
		{"video_price_1080p", profile.VideoPrice1080P},
		{"web_search_price_per_call", profile.WebSearchPricePerCall},
	} {
		if price.value != nil && !isNonNegativeFinite(*price.value) {
			return infraerrors.BadRequest("BILLING_PROFILE_INVALID", price.name+" must be a finite number >= 0")
		}
	}
	if profile.PeakRateEnabled {
		if err := validateBillingProfilePeakRate(profile.PeakStart, profile.PeakEnd, profile.PeakRateMultiplier); err != nil {
			return err
		}
		return nil
	}
	profile.PeakStart = ""
	profile.PeakEnd = ""
	profile.PeakRateMultiplier = 1
	return nil
}

func validateBillingProfilePeakRate(start, end string, multiplier float64) error {
	if start == "" || end == "" {
		return infraerrors.BadRequest("BILLING_PROFILE_INVALID", "peak_start and peak_end are required when peak pricing is enabled")
	}
	startMinutes, validStart := parseMinutes(start)
	endMinutes, validEnd := parseMinutes(end)
	if !validStart || !validEnd || startMinutes >= endMinutes {
		return infraerrors.BadRequest("BILLING_PROFILE_INVALID", "peak time range must be a valid same-day HH:MM interval")
	}
	if !isNonNegativeFinite(multiplier) {
		return infraerrors.BadRequest("BILLING_PROFILE_INVALID", "peak_rate_multiplier must be a finite number >= 0")
	}
	return nil
}

func isNonNegativeFinite(value float64) bool {
	return value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func saveBillingProfile(ctx context.Context, client *dbent.Client, profile *BillingProfile) (*dbent.BillingProfile, error) {
	existing, err := client.BillingProfile.Query().
		Where(billingprofile.GroupIDEQ(profile.GroupID)).
		Only(ctx)
	if err != nil && !dbent.IsNotFound(err) {
		return nil, fmt.Errorf("load billing profile: %w", err)
	}
	if dbent.IsNotFound(err) {
		return createBillingProfile(ctx, client, profile)
	}
	return updateBillingProfile(ctx, client, existing.ID, profile)
}

func createBillingProfile(ctx context.Context, client *dbent.Client, profile *BillingProfile) (*dbent.BillingProfile, error) {
	return client.BillingProfile.Create().
		SetGroupID(profile.GroupID).
		SetBalanceRateMultiplier(profile.BalanceRateMultiplier).
		SetPeakRateEnabled(profile.PeakRateEnabled).
		SetPeakStart(profile.PeakStart).
		SetPeakEnd(profile.PeakEnd).
		SetPeakRateMultiplier(profile.PeakRateMultiplier).
		SetImageRateIndependent(profile.ImageRateIndependent).
		SetImageRateMultiplier(profile.ImageRateMultiplier).
		SetNillableImagePrice1k(profile.ImagePrice1K).
		SetNillableImagePrice2k(profile.ImagePrice2K).
		SetNillableImagePrice4k(profile.ImagePrice4K).
		SetBatchImageDiscountMultiplier(profile.BatchImageDiscountMultiplier).
		SetBatchImageHoldMultiplier(profile.BatchImageHoldMultiplier).
		SetVideoRateIndependent(profile.VideoRateIndependent).
		SetVideoRateMultiplier(profile.VideoRateMultiplier).
		SetNillableVideoPrice480p(profile.VideoPrice480P).
		SetNillableVideoPrice720p(profile.VideoPrice720P).
		SetNillableVideoPrice1080p(profile.VideoPrice1080P).
		SetNillableWebSearchPricePerCall(profile.WebSearchPricePerCall).
		Save(ctx)
}

func updateBillingProfile(ctx context.Context, client *dbent.Client, id int64, profile *BillingProfile) (*dbent.BillingProfile, error) {
	builder := client.BillingProfile.UpdateOneID(id).
		SetBalanceRateMultiplier(profile.BalanceRateMultiplier).
		SetPeakRateEnabled(profile.PeakRateEnabled).
		SetPeakStart(profile.PeakStart).
		SetPeakEnd(profile.PeakEnd).
		SetPeakRateMultiplier(profile.PeakRateMultiplier).
		SetImageRateIndependent(profile.ImageRateIndependent).
		SetImageRateMultiplier(profile.ImageRateMultiplier).
		SetBatchImageDiscountMultiplier(profile.BatchImageDiscountMultiplier).
		SetBatchImageHoldMultiplier(profile.BatchImageHoldMultiplier).
		SetVideoRateIndependent(profile.VideoRateIndependent).
		SetVideoRateMultiplier(profile.VideoRateMultiplier)
	setBillingProfileNullablePrices(builder, profile)
	return builder.Save(ctx)
}

func setBillingProfileNullablePrices(builder *dbent.BillingProfileUpdateOne, profile *BillingProfile) {
	if profile.ImagePrice1K == nil {
		builder.ClearImagePrice1k()
	} else {
		builder.SetImagePrice1k(*profile.ImagePrice1K)
	}
	if profile.ImagePrice2K == nil {
		builder.ClearImagePrice2k()
	} else {
		builder.SetImagePrice2k(*profile.ImagePrice2K)
	}
	if profile.ImagePrice4K == nil {
		builder.ClearImagePrice4k()
	} else {
		builder.SetImagePrice4k(*profile.ImagePrice4K)
	}
	if profile.VideoPrice480P == nil {
		builder.ClearVideoPrice480p()
	} else {
		builder.SetVideoPrice480p(*profile.VideoPrice480P)
	}
	if profile.VideoPrice720P == nil {
		builder.ClearVideoPrice720p()
	} else {
		builder.SetVideoPrice720p(*profile.VideoPrice720P)
	}
	if profile.VideoPrice1080P == nil {
		builder.ClearVideoPrice1080p()
	} else {
		builder.SetVideoPrice1080p(*profile.VideoPrice1080P)
	}
	if profile.WebSearchPricePerCall == nil {
		builder.ClearWebSearchPricePerCall()
	} else {
		builder.SetWebSearchPricePerCall(*profile.WebSearchPricePerCall)
	}
}
