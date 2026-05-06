//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type updateGroupsAPIKeyRepoStub struct {
	apiKey              *APIKey
	updated             *APIKey
	replacedAllowedKeys []int64
	replacedAllowedSets [][]int64
}

func (s *updateGroupsAPIKeyRepoStub) Create(context.Context, *APIKey) error {
	panic("unexpected Create call")
}

func (s *updateGroupsAPIKeyRepoStub) GetByID(context.Context, int64) (*APIKey, error) {
	if s.apiKey == nil {
		return nil, ErrAPIKeyNotFound
	}
	clone := *s.apiKey
	if s.apiKey.GroupID != nil {
		groupID := *s.apiKey.GroupID
		clone.GroupID = &groupID
	}
	clone.AllowedGroupIDs = append([]int64(nil), s.apiKey.AllowedGroupIDs...)
	clone.AllowedGroups = append([]Group(nil), s.apiKey.AllowedGroups...)
	return &clone, nil
}

func (s *updateGroupsAPIKeyRepoStub) GetKeyAndOwnerID(context.Context, int64) (string, int64, error) {
	panic("unexpected GetKeyAndOwnerID call")
}

func (s *updateGroupsAPIKeyRepoStub) GetByKey(context.Context, string) (*APIKey, error) {
	panic("unexpected GetByKey call")
}

func (s *updateGroupsAPIKeyRepoStub) GetByKeyForAuth(context.Context, string) (*APIKey, error) {
	panic("unexpected GetByKeyForAuth call")
}

func (s *updateGroupsAPIKeyRepoStub) Update(_ context.Context, key *APIKey) error {
	clone := *key
	if key.GroupID != nil {
		groupID := *key.GroupID
		clone.GroupID = &groupID
	}
	clone.AllowedGroupIDs = append([]int64(nil), key.AllowedGroupIDs...)
	clone.AllowedGroups = append([]Group(nil), key.AllowedGroups...)
	s.updated = &clone
	return nil
}

func (s *updateGroupsAPIKeyRepoStub) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}

func (s *updateGroupsAPIKeyRepoStub) ListByUserID(context.Context, int64, pagination.PaginationParams, APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error) {
	panic("unexpected ListByUserID call")
}

func (s *updateGroupsAPIKeyRepoStub) VerifyOwnership(context.Context, int64, []int64) ([]int64, error) {
	panic("unexpected VerifyOwnership call")
}

func (s *updateGroupsAPIKeyRepoStub) CountByUserID(context.Context, int64) (int64, error) {
	panic("unexpected CountByUserID call")
}

func (s *updateGroupsAPIKeyRepoStub) ExistsByKey(context.Context, string) (bool, error) {
	panic("unexpected ExistsByKey call")
}

func (s *updateGroupsAPIKeyRepoStub) ListByGroupID(context.Context, int64, pagination.PaginationParams) ([]APIKey, *pagination.PaginationResult, error) {
	panic("unexpected ListByGroupID call")
}

func (s *updateGroupsAPIKeyRepoStub) SearchAPIKeys(context.Context, int64, string, int) ([]APIKey, error) {
	panic("unexpected SearchAPIKeys call")
}

func (s *updateGroupsAPIKeyRepoStub) ClearGroupIDByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected ClearGroupIDByGroupID call")
}

func (s *updateGroupsAPIKeyRepoStub) UpdateGroupIDByUserAndGroup(context.Context, int64, int64, int64) (int64, error) {
	panic("unexpected UpdateGroupIDByUserAndGroup call")
}

func (s *updateGroupsAPIKeyRepoStub) CountByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected CountByGroupID call")
}

func (s *updateGroupsAPIKeyRepoStub) ListKeysByUserID(context.Context, int64) ([]string, error) {
	panic("unexpected ListKeysByUserID call")
}

func (s *updateGroupsAPIKeyRepoStub) ListKeysByGroupID(context.Context, int64) ([]string, error) {
	panic("unexpected ListKeysByGroupID call")
}

func (s *updateGroupsAPIKeyRepoStub) ReplaceAllowedGroups(_ context.Context, keyID int64, groupIDs []int64) error {
	s.replacedAllowedKeys = append(s.replacedAllowedKeys, keyID)
	s.replacedAllowedSets = append(s.replacedAllowedSets, append([]int64(nil), groupIDs...))
	return nil
}

func (s *updateGroupsAPIKeyRepoStub) ListAllowedGroups(context.Context, int64) ([]Group, error) {
	panic("unexpected ListAllowedGroups call")
}

func (s *updateGroupsAPIKeyRepoStub) IncrementQuotaUsed(context.Context, int64, float64) (float64, error) {
	panic("unexpected IncrementQuotaUsed call")
}

func (s *updateGroupsAPIKeyRepoStub) UpdateLastUsed(context.Context, int64, time.Time) error {
	panic("unexpected UpdateLastUsed call")
}

func (s *updateGroupsAPIKeyRepoStub) IncrementRateLimitUsage(context.Context, int64, float64) error {
	panic("unexpected IncrementRateLimitUsage call")
}

func (s *updateGroupsAPIKeyRepoStub) ResetRateLimitWindows(context.Context, int64) error {
	panic("unexpected ResetRateLimitWindows call")
}

func (s *updateGroupsAPIKeyRepoStub) GetRateLimitData(context.Context, int64) (*APIKeyRateLimitData, error) {
	panic("unexpected GetRateLimitData call")
}

type updateGroupsUserRepoStub struct {
	user *User
}

func (s *updateGroupsUserRepoStub) Create(context.Context, *User) error {
	panic("unexpected Create call")
}

func (s *updateGroupsUserRepoStub) GetByID(context.Context, int64) (*User, error) {
	if s.user == nil {
		return nil, ErrUserNotFound
	}
	return s.user, nil
}

func (s *updateGroupsUserRepoStub) GetByEmail(context.Context, string) (*User, error) {
	panic("unexpected GetByEmail call")
}

func (s *updateGroupsUserRepoStub) GetFirstAdmin(context.Context) (*User, error) {
	panic("unexpected GetFirstAdmin call")
}

func (s *updateGroupsUserRepoStub) Update(context.Context, *User) error {
	panic("unexpected Update call")
}

func (s *updateGroupsUserRepoStub) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}

func (s *updateGroupsUserRepoStub) GetUserAvatar(context.Context, int64) (*UserAvatar, error) {
	panic("unexpected GetUserAvatar call")
}

func (s *updateGroupsUserRepoStub) UpsertUserAvatar(context.Context, int64, UpsertUserAvatarInput) (*UserAvatar, error) {
	panic("unexpected UpsertUserAvatar call")
}

func (s *updateGroupsUserRepoStub) DeleteUserAvatar(context.Context, int64) error {
	panic("unexpected DeleteUserAvatar call")
}

func (s *updateGroupsUserRepoStub) List(context.Context, pagination.PaginationParams) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (s *updateGroupsUserRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, UserListFilters) ([]User, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (s *updateGroupsUserRepoStub) GetLatestUsedAtByUserIDs(context.Context, []int64) (map[int64]*time.Time, error) {
	panic("unexpected GetLatestUsedAtByUserIDs call")
}

func (s *updateGroupsUserRepoStub) GetLatestUsedAtByUserID(context.Context, int64) (*time.Time, error) {
	panic("unexpected GetLatestUsedAtByUserID call")
}

func (s *updateGroupsUserRepoStub) UpdateUserLastActiveAt(context.Context, int64, time.Time) error {
	panic("unexpected UpdateUserLastActiveAt call")
}

func (s *updateGroupsUserRepoStub) UpdateBalance(context.Context, int64, float64) error {
	panic("unexpected UpdateBalance call")
}

func (s *updateGroupsUserRepoStub) DeductBalance(context.Context, int64, float64) error {
	panic("unexpected DeductBalance call")
}

func (s *updateGroupsUserRepoStub) UpdateConcurrency(context.Context, int64, int) error {
	panic("unexpected UpdateConcurrency call")
}

func (s *updateGroupsUserRepoStub) ExistsByEmail(context.Context, string) (bool, error) {
	panic("unexpected ExistsByEmail call")
}

func (s *updateGroupsUserRepoStub) RemoveGroupFromAllowedGroups(context.Context, int64) (int64, error) {
	panic("unexpected RemoveGroupFromAllowedGroups call")
}

func (s *updateGroupsUserRepoStub) RemoveGroupFromUserAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected RemoveGroupFromUserAllowedGroups call")
}

func (s *updateGroupsUserRepoStub) AddGroupToAllowedGroups(context.Context, int64, int64) error {
	panic("unexpected AddGroupToAllowedGroups call")
}

func (s *updateGroupsUserRepoStub) ListUserAuthIdentities(context.Context, int64) ([]UserAuthIdentityRecord, error) {
	panic("unexpected ListUserAuthIdentities call")
}

func (s *updateGroupsUserRepoStub) UnbindUserAuthProvider(context.Context, int64, string) error {
	panic("unexpected UnbindUserAuthProvider call")
}

func (s *updateGroupsUserRepoStub) UpdateTotpSecret(context.Context, int64, *string) error {
	panic("unexpected UpdateTotpSecret call")
}

func (s *updateGroupsUserRepoStub) EnableTotp(context.Context, int64) error {
	panic("unexpected EnableTotp call")
}

func (s *updateGroupsUserRepoStub) DisableTotp(context.Context, int64) error {
	panic("unexpected DisableTotp call")
}

func TestAPIKeyServiceUpdateClearsAllowedGroupsWhenGroupIDsEmpty(t *testing.T) {
	groupID := int64(101)
	repo := &updateGroupsAPIKeyRepoStub{
		apiKey: &APIKey{
			ID:              1,
			UserID:          7,
			Key:             "k1",
			Name:            "old",
			GroupID:         &groupID,
			AllowedGroupIDs: []int64{groupID},
			AllowedGroups:   []Group{{ID: groupID, Name: "sub"}},
			Status:          StatusActive,
		},
	}
	groupIDs := []int64{}
	svc := &APIKeyService{
		apiKeyRepo: repo,
		userRepo:   &updateGroupsUserRepoStub{user: &User{ID: 7}},
	}

	updated, err := svc.Update(context.Background(), 1, 7, UpdateAPIKeyRequest{
		GroupIDs: &groupIDs,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Nil(t, updated.GroupID)
	require.Nil(t, updated.Group)
	require.Empty(t, updated.AllowedGroupIDs)
	require.Empty(t, updated.AllowedGroups)
	require.Len(t, repo.replacedAllowedKeys, 1)
	require.Equal(t, int64(1), repo.replacedAllowedKeys[0])
	require.Equal(t, [][]int64{{}}, repo.replacedAllowedSets)
	require.NotNil(t, repo.updated)
	require.Nil(t, repo.updated.GroupID)
	require.Empty(t, repo.updated.AllowedGroupIDs)
}

func TestAPIKeyServiceUpdateKeepsAllowedGroupsWhenGroupIDsOmitted(t *testing.T) {
	groupID := int64(101)
	repo := &updateGroupsAPIKeyRepoStub{
		apiKey: &APIKey{
			ID:              2,
			UserID:          7,
			Key:             "k2",
			Name:            "old",
			GroupID:         &groupID,
			AllowedGroupIDs: []int64{groupID},
			AllowedGroups:   []Group{{ID: groupID, Name: "sub"}},
			Status:          StatusActive,
		},
	}
	name := "new-name"
	svc := &APIKeyService{
		apiKeyRepo: repo,
		userRepo:   &updateGroupsUserRepoStub{user: &User{ID: 7}},
	}

	updated, err := svc.Update(context.Background(), 2, 7, UpdateAPIKeyRequest{
		Name: &name,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Equal(t, name, updated.Name)
	require.NotNil(t, updated.GroupID)
	require.Equal(t, groupID, *updated.GroupID)
	require.Equal(t, []int64{groupID}, updated.AllowedGroupIDs)
	require.Len(t, repo.replacedAllowedKeys, 0)
	require.NotNil(t, repo.updated)
	require.NotNil(t, repo.updated.GroupID)
	require.Equal(t, groupID, *repo.updated.GroupID)
	require.Equal(t, []int64{groupID}, repo.updated.AllowedGroupIDs)
}
