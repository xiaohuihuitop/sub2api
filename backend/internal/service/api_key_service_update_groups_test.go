//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type allowedGroupsAPIKeyRepoStub struct {
	*apiKeyRepoStub
	replaced [][]int64
}

func (s *allowedGroupsAPIKeyRepoStub) ReplaceAllowedGroups(_ context.Context, _ int64, groupIDs []int64) error {
	s.replaced = append(s.replaced, append([]int64{}, groupIDs...))
	return nil
}

type apiKeyGroupsRepoStub struct {
	groupRepoNoop
	groups map[int64]*Group
}

func (s *apiKeyGroupsRepoStub) GetByID(_ context.Context, id int64) (*Group, error) {
	group, ok := s.groups[id]
	if !ok {
		return nil, ErrGroupNotFound
	}
	copy := *group
	return &copy, nil
}

func TestAPIKeyServiceUpdateSortsAndPersistsAllowedGroups(t *testing.T) {
	legacyGroupID := int64(30)
	baseRepo := &apiKeyRepoStub{apiKey: &APIKey{
		ID: 1, UserID: 7, Key: "sk-groups", Name: "groups", Status: StatusActive,
		GroupID: &legacyGroupID,
	}}
	repo := &allowedGroupsAPIKeyRepoStub{apiKeyRepoStub: baseRepo}
	groupIDs := []int64{30, 20, 10, 20}
	svc := &APIKeyService{
		apiKeyRepo: repo,
		userRepo:   &userRepoStub{user: &User{ID: 7}},
		groupRepo: &apiKeyGroupsRepoStub{groups: map[int64]*Group{
			10: {ID: 10, SortOrder: 1, Status: StatusActive},
			20: {ID: 20, SortOrder: 1, Status: StatusActive},
			30: {ID: 30, SortOrder: 2, Status: StatusActive},
		}},
	}

	updated, err := svc.Update(context.Background(), 1, 7, UpdateAPIKeyRequest{GroupIDs: &groupIDs})

	require.NoError(t, err)
	require.Equal(t, []int64{10, 20, 30}, updated.AllowedGroupIDs)
	require.Equal(t, int64(10), *updated.GroupID)
	require.Equal(t, int64(10), updated.Group.ID)
	require.Equal(t, [][]int64{{10, 20, 30}}, repo.replaced)
}

func TestAPIKeyServiceUpdateClearsAllowedGroups(t *testing.T) {
	groupID := int64(10)
	baseRepo := &apiKeyRepoStub{apiKey: &APIKey{
		ID: 2, UserID: 7, Key: "sk-clear", Name: "clear", Status: StatusActive,
		GroupID: &groupID, AllowedGroupIDs: []int64{groupID}, AllowedGroups: []Group{{ID: groupID}},
	}}
	repo := &allowedGroupsAPIKeyRepoStub{apiKeyRepoStub: baseRepo}
	groupIDs := []int64{}
	svc := &APIKeyService{
		apiKeyRepo: repo,
		userRepo:   &userRepoStub{user: &User{ID: 7}},
		groupRepo:  &apiKeyGroupsRepoStub{},
	}

	updated, err := svc.Update(context.Background(), 2, 7, UpdateAPIKeyRequest{GroupIDs: &groupIDs})

	require.NoError(t, err)
	require.Nil(t, updated.GroupID)
	require.Nil(t, updated.Group)
	require.Empty(t, updated.AllowedGroupIDs)
	require.Empty(t, updated.AllowedGroups)
	require.Equal(t, [][]int64{{}}, repo.replaced)
}

func TestAPIKeyServiceUpdateKeepsExistingExpiredSubscriptionGroup(t *testing.T) {
	groupID := int64(10)
	baseRepo := &apiKeyRepoStub{apiKey: &APIKey{
		ID: 3, UserID: 7, Key: "sk-expired", Name: "expired", Status: StatusActive,
		GroupID: &groupID, AllowedGroupIDs: []int64{groupID},
	}}
	repo := &allowedGroupsAPIKeyRepoStub{apiKeyRepoStub: baseRepo}
	groupIDs := []int64{groupID}
	subscriptions := &userSubRepoStubForGroupUpdate{getActiveErr: ErrSubscriptionNotFound}
	svc := &APIKeyService{
		apiKeyRepo: repo,
		userRepo:   &userRepoStub{user: &User{ID: 7}},
		groupRepo: &apiKeyGroupsRepoStub{groups: map[int64]*Group{
			groupID: {ID: groupID, Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription},
		}},
		userSubRepo: subscriptions,
	}

	updated, err := svc.Update(context.Background(), 3, 7, UpdateAPIKeyRequest{GroupIDs: &groupIDs})

	require.NoError(t, err)
	require.Equal(t, []int64{groupID}, updated.AllowedGroupIDs)
	require.Equal(t, [][]int64{{groupID}}, repo.replaced)
}

func TestAPIKeyServiceUpdateRejectsNewExpiredSubscriptionGroup(t *testing.T) {
	baseRepo := &apiKeyRepoStub{apiKey: &APIKey{
		ID: 4, UserID: 7, Key: "sk-new-expired", Name: "new expired", Status: StatusActive,
	}}
	repo := &allowedGroupsAPIKeyRepoStub{apiKeyRepoStub: baseRepo}
	groupID := int64(10)
	groupIDs := []int64{groupID}
	svc := &APIKeyService{
		apiKeyRepo: repo,
		userRepo:   &userRepoStub{user: &User{ID: 7}},
		groupRepo: &apiKeyGroupsRepoStub{groups: map[int64]*Group{
			groupID: {ID: groupID, Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription},
		}},
		userSubRepo: &userSubRepoStubForGroupUpdate{getActiveErr: ErrSubscriptionNotFound},
	}

	_, err := svc.Update(context.Background(), 4, 7, UpdateAPIKeyRequest{GroupIDs: &groupIDs})

	require.ErrorIs(t, err, ErrGroupNotAllowed)
	require.Empty(t, repo.replaced)
}
