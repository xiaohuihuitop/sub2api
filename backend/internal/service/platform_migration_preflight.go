package service

import (
	"sort"
	"strconv"
	"strings"
)

const (
	SettingKeyPlatformAssetsV2Enabled     = "platform_assets_v2_enabled"
	SettingKeyGlobalBalanceRateMultiplier = "global_balance_rate_multiplier"

	platformMigrationConflictModelPatternOverlap       = "model_pattern_overlap"
	platformMigrationConflictAccountPlatformAmbiguity  = "account_platform_ambiguity"
	platformMigrationConflictAPIKeyPlanAmbiguity       = "api_key_plan_ambiguity"
	platformMigrationConflictAPIKeyPlanUnmapped        = "api_key_plan_unmapped"
	platformMigrationConflictAPIKeyAssetTypeUnmapped   = "api_key_asset_type_unmapped"
	platformMigrationConflictAPIKeyPlatformUnmapped    = "api_key_platform_unmapped"
	platformMigrationConflictAPIKeyAuthorizationNeeded = "api_key_explicit_authorization_required"
)

type PlatformMigrationPreflight struct {
	Ready     bool                        `json:"ready"`
	Conflicts []PlatformMigrationConflict `json:"conflicts"`
}

type PlatformMigrationConflict struct {
	Kind    string `json:"kind"`
	Subject string `json:"subject"`
	Detail  string `json:"detail"`
}

type legacyPlatformMigrationSnapshot struct {
	GroupPlatforms         map[int64]string
	GroupPlanIDs           map[int64][]int64
	GroupSubscriptionTypes map[int64]string
	ModelRules             []legacyPlatformModelRule
	Accounts               []legacyAccountPlatformBinding
	APIKeys                []legacyAPIKeyGroupBinding
}

type legacyPlatformModelRule struct {
	GroupID  int64
	Platform string
	Pattern  string
}

type legacyAccountPlatformBinding struct {
	AccountID int64
	Platforms []string
}

type legacyAPIKeyGroupBinding struct {
	APIKeyID int64
	GroupIDs []int64
}

func buildPlatformMigrationPreflight(snapshot legacyPlatformMigrationSnapshot) PlatformMigrationPreflight {
	conflicts := collectPlatformMigrationConflicts(snapshot)
	return PlatformMigrationPreflight{
		Ready:     len(conflicts) == 0,
		Conflicts: conflicts,
	}
}

func collectPlatformMigrationConflicts(snapshot legacyPlatformMigrationSnapshot) []PlatformMigrationConflict {
	conflicts := make([]PlatformMigrationConflict, 0)
	conflicts = append(conflicts, collectModelPatternConflicts(snapshot.ModelRules)...)
	conflicts = append(conflicts, collectAccountPlatformConflicts(snapshot.Accounts)...)
	conflicts = append(conflicts, collectAPIKeyAssetConflicts(snapshot)...)
	sortMigrationConflicts(conflicts)
	return conflicts
}

func collectModelPatternConflicts(rules []legacyPlatformModelRule) []PlatformMigrationConflict {
	conflicts := make([]PlatformMigrationConflict, 0)
	for left := range rules {
		for right := left + 1; right < len(rules); right++ {
			if rules[left].Platform == rules[right].Platform || !modelPatternsOverlap(rules[left].Pattern, rules[right].Pattern) {
				continue
			}
			conflicts = append(conflicts, PlatformMigrationConflict{
				Kind:    platformMigrationConflictModelPatternOverlap,
				Subject: rules[left].Pattern + "|" + rules[right].Pattern,
				Detail:  rules[left].Platform + " conflicts with " + rules[right].Platform,
			})
		}
	}
	return conflicts
}

func collectAccountPlatformConflicts(accounts []legacyAccountPlatformBinding) []PlatformMigrationConflict {
	conflicts := make([]PlatformMigrationConflict, 0)
	for _, account := range accounts {
		if len(uniquePlatforms(account.Platforms)) <= 1 {
			continue
		}
		conflicts = append(conflicts, PlatformMigrationConflict{
			Kind:    platformMigrationConflictAccountPlatformAmbiguity,
			Subject: formatMigrationID("account", account.AccountID),
			Detail:  "legacy account is associated with multiple platforms",
		})
	}
	return conflicts
}

func collectAPIKeyAssetConflicts(snapshot legacyPlatformMigrationSnapshot) []PlatformMigrationConflict {
	conflicts := make([]PlatformMigrationConflict, 0)
	for _, key := range snapshot.APIKeys {
		if len(key.GroupIDs) == 0 {
			conflicts = append(conflicts, PlatformMigrationConflict{
				Kind:    platformMigrationConflictAPIKeyAuthorizationNeeded,
				Subject: formatMigrationID("api_key", key.APIKeyID),
				Detail:  "unbound legacy API key needs explicit platform and billing authorization",
			})
			continue
		}
		for _, groupID := range key.GroupIDs {
			conflicts = append(conflicts, collectAPIKeyGroupConflicts(snapshot, key.APIKeyID, groupID)...)
		}
	}
	return conflicts
}

func collectAPIKeyGroupConflicts(
	snapshot legacyPlatformMigrationSnapshot,
	apiKeyID, groupID int64,
) []PlatformMigrationConflict {
	conflicts := make([]PlatformMigrationConflict, 0, 2)
	if platformForLegacyGroup(snapshot, groupID) == "" {
		conflicts = append(conflicts, PlatformMigrationConflict{
			Kind:    platformMigrationConflictAPIKeyPlatformUnmapped,
			Subject: formatMigrationKeyGroup(apiKeyID, groupID),
			Detail:  "legacy group has no unique platform mapping",
		})
	}
	plans := uniquePositiveIDs(snapshot.GroupPlanIDs[groupID])
	if !isSubscriptionGroup(snapshot, groupID) {
		if isBalanceGroup(snapshot, groupID) {
			return conflicts
		}
		return append(conflicts, PlatformMigrationConflict{
			Kind:    platformMigrationConflictAPIKeyAssetTypeUnmapped,
			Subject: formatMigrationKeyGroup(apiKeyID, groupID),
			Detail:  "legacy group has no recognized billing asset type",
		})
	}
	if len(plans) == 1 {
		return conflicts
	}
	kind := platformMigrationConflictAPIKeyPlanAmbiguity
	detail := "legacy group maps to multiple subscription plans"
	if len(plans) == 0 {
		kind = platformMigrationConflictAPIKeyPlanUnmapped
		detail = "legacy group has no subscription plan mapping"
	}
	return append(conflicts, PlatformMigrationConflict{Kind: kind, Subject: formatMigrationKeyGroup(apiKeyID, groupID), Detail: detail})
}

func isSubscriptionGroup(snapshot legacyPlatformMigrationSnapshot, groupID int64) bool {
	return normalizeLegacySubscriptionType(snapshot.GroupSubscriptionTypes[groupID]) == SubscriptionTypeSubscription
}

func isBalanceGroup(snapshot legacyPlatformMigrationSnapshot, groupID int64) bool {
	return normalizeLegacySubscriptionType(snapshot.GroupSubscriptionTypes[groupID]) == SubscriptionTypeStandard
}

func normalizeLegacySubscriptionType(subscriptionType string) string {
	return strings.ToLower(strings.TrimSpace(subscriptionType))
}

func platformForLegacyGroup(snapshot legacyPlatformMigrationSnapshot, groupID int64) string {
	if platform := normalizePlatform(snapshot.GroupPlatforms[groupID]); platform != "" {
		return platform
	}
	platforms := make([]string, 0, 1)
	for _, rule := range snapshot.ModelRules {
		if rule.GroupID == groupID {
			platforms = append(platforms, rule.Platform)
		}
	}
	unique := uniquePlatforms(platforms)
	if len(unique) != 1 {
		return ""
	}
	return unique[0]
}

func modelPatternsOverlap(left, right string) bool {
	leftPrefix, leftWildcard := splitMigrationModelPattern(left)
	rightPrefix, rightWildcard := splitMigrationModelPattern(right)
	if (leftPrefix == "" && !leftWildcard) || (rightPrefix == "" && !rightWildcard) {
		return false
	}
	if (leftWildcard && leftPrefix == "") || (rightWildcard && rightPrefix == "") {
		return true
	}
	if !leftWildcard && !rightWildcard {
		return leftPrefix == rightPrefix
	}
	if leftWildcard && rightWildcard {
		return strings.HasPrefix(leftPrefix, rightPrefix) || strings.HasPrefix(rightPrefix, leftPrefix)
	}
	if leftWildcard {
		return strings.HasPrefix(rightPrefix, leftPrefix)
	}
	return strings.HasPrefix(leftPrefix, rightPrefix)
}

func splitMigrationModelPattern(pattern string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(pattern))
	if strings.HasSuffix(normalized, "*") {
		return strings.TrimSuffix(normalized, "*"), true
	}
	return normalized, false
}

func uniquePlatforms(platforms []string) []string {
	seen := make(map[string]struct{}, len(platforms))
	result := make([]string, 0, len(platforms))
	for _, platform := range platforms {
		normalized := normalizePlatform(platform)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func normalizePlatform(platform string) string {
	return strings.ToLower(strings.TrimSpace(platform))
}

func uniquePositiveIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, exists := seen[id]; !exists {
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result
}

func platformMigrationConflictCodes(conflicts []PlatformMigrationConflict) []string {
	codes := make([]string, len(conflicts))
	for index := range conflicts {
		codes[index] = conflicts[index].Kind
	}
	return codes
}

func formatMigrationID(kind string, id int64) string {
	return kind + ":" + strconv.FormatInt(id, 10)
}

func formatMigrationKeyGroup(apiKeyID, groupID int64) string {
	return formatMigrationID("api_key", apiKeyID) + ":group:" + strconv.FormatInt(groupID, 10)
}

func sortMigrationConflicts(conflicts []PlatformMigrationConflict) {
	sort.Slice(conflicts, func(left, right int) bool {
		if conflicts[left].Kind != conflicts[right].Kind {
			return conflicts[left].Kind < conflicts[right].Kind
		}
		return conflicts[left].Subject < conflicts[right].Subject
	})
}
