package service

import "sort"

func normalizeAPIKeyGroupIDs(groupIDs []int64, legacyGroupID *int64) []int64 {
	ids := make([]int64, 0, len(groupIDs)+1)
	seen := make(map[int64]struct{}, len(groupIDs)+1)
	for _, id := range groupIDs {
		if id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 && len(groupIDs) == 0 && legacyGroupID != nil && *legacyGroupID > 0 {
		ids = append(ids, *legacyGroupID)
	}
	return ids
}

func sortAPIKeyAllowedGroups(groups []Group) []Group {
	sorted := append([]Group(nil), groups...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].SortOrder != sorted[j].SortOrder {
			return sorted[i].SortOrder < sorted[j].SortOrder
		}
		return sorted[i].ID < sorted[j].ID
	})
	return sorted
}

func apiKeyGroupIDs(groups []Group) []int64 {
	ids := make([]int64, 0, len(groups))
	for i := range groups {
		ids = append(ids, groups[i].ID)
	}
	return ids
}
