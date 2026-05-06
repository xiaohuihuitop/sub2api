package service

import (
	"context"
	"fmt"
)

func normalizeAPIKeyGroupIDs(groupIDs []int64, groupID *int64) []int64 {
	ids := make([]int64, 0, len(groupIDs)+1)
	seen := make(map[int64]struct{}, len(groupIDs)+1)
	for _, id := range groupIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 && groupID != nil && *groupID > 0 {
		ids = append(ids, *groupID)
	}
	return ids
}

func (s *APIKeyService) validateAPIKeyAllowedGroups(ctx context.Context, user *User, groupIDs []int64) ([]Group, error) {
	groups := make([]Group, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		group, err := s.groupRepo.GetByID(ctx, groupID)
		if err != nil {
			return nil, fmt.Errorf("get group: %w", err)
		}
		if !s.canUserBindGroup(ctx, user, group) {
			return nil, ErrGroupNotAllowed
		}
		groups = append(groups, *group)
	}
	return groups, nil
}
