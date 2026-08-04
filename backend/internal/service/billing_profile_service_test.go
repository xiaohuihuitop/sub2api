//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type billingProfileGroupRepoStub struct {
	GroupRepository
	group *Group
}

func (s billingProfileGroupRepoStub) GetByID(context.Context, int64) (*Group, error) {
	return s.group, nil
}

type billingProfileAuthCacheInvalidator struct {
	groupIDs []int64
}

func (s *billingProfileAuthCacheInvalidator) InvalidateAuthCacheByKey(context.Context, string) {}
func (s *billingProfileAuthCacheInvalidator) InvalidateAuthCacheByUserID(context.Context, int64) {
}
func (s *billingProfileAuthCacheInvalidator) InvalidateAuthCacheByGroupID(_ context.Context, groupID int64) {
	s.groupIDs = append(s.groupIDs, groupID)
}

func TestAdminServiceUpdateGroupBillingProfilePersistsBalancePricing(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	groupEntity, err := client.Group.Create().SetName("billing-route").Save(ctx)
	require.NoError(t, err)

	imagePrice := 0.12
	videoPrice := 0.34
	invalidator := &billingProfileAuthCacheInvalidator{}
	svc := &adminServiceImpl{
		entClient: client,
		groupRepo: billingProfileGroupRepoStub{group: &Group{
			ID:               groupEntity.ID,
			RateMultiplier:   9,
			SubscriptionType: SubscriptionTypeStandard,
		}},
		authCacheInvalidator: invalidator,
	}

	updated, err := svc.UpdateGroupBillingProfile(ctx, groupEntity.ID, &UpdateBillingProfileInput{
		BalanceRateMultiplier:        1.25,
		PeakRateEnabled:              true,
		PeakStart:                    "09:00",
		PeakEnd:                      "18:00",
		PeakRateMultiplier:           1.5,
		ImageRateIndependent:         true,
		ImageRateMultiplier:          2,
		BatchImageDiscountMultiplier: 0.5,
		BatchImageHoldMultiplier:     0.6,
		VideoRateIndependent:         true,
		VideoRateMultiplier:          3,
		ImagePrice1K:                 &imagePrice,
		VideoPrice1080P:              &videoPrice,
	})

	require.NoError(t, err)
	require.Equal(t, groupEntity.ID, updated.GroupID)
	require.Equal(t, 1.25, updated.BalanceRateMultiplier)
	require.True(t, updated.PeakRateEnabled)
	require.Equal(t, "09:00", updated.PeakStart)
	require.Equal(t, 1.5, updated.PeakRateMultiplier)
	require.Equal(t, &imagePrice, updated.ImagePrice1K)
	require.Equal(t, &videoPrice, updated.VideoPrice1080P)
	require.Equal(t, []int64{groupEntity.ID}, invalidator.groupIDs)

	loaded, err := svc.GetGroupBillingProfile(ctx, groupEntity.ID)
	require.NoError(t, err)
	require.Equal(t, updated, loaded)
}
