package service

import (
	"fmt"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrPlatformModelNotFound  = infraerrors.NotFound("PLATFORM_MODEL_NOT_FOUND", "model is not assigned to an active platform")
	ErrPlatformModelAmbiguous = infraerrors.Conflict("PLATFORM_MODEL_AMBIGUOUS", "model matches multiple platform rules")
	ErrPlatformModelRule      = infraerrors.BadRequest("INVALID_PLATFORM_MODEL_RULE", "invalid platform model rule")
)

type platformModelResolver struct {
	rules []PlatformModelRule
}

func newPlatformModelResolver(rules []PlatformModelRule) *platformModelResolver {
	return &platformModelResolver{rules: append([]PlatformModelRule(nil), rules...)}
}

func validatePlatformModelRules(rules []PlatformModelRule) error {
	seen := make(map[string]struct{}, len(rules))
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		pattern, _, err := normalizePlatformModelPattern(rule.ModelPattern)
		if err != nil {
			return err
		}
		if rule.PlatformID <= 0 {
			return fmt.Errorf("%w: platform id is required", ErrPlatformModelRule)
		}
		key := fmt.Sprintf("%d:%s", rule.PlatformID, pattern)
		if _, exists := seen[key]; exists {
			return fmt.Errorf("%w: duplicate pattern %q on platform %d", ErrPlatformModelRule, rule.ModelPattern, rule.PlatformID)
		}
		seen[key] = struct{}{}
	}

	return validateCrossPlatformModelPatternOverlap(rules)
}

func validateCrossPlatformModelPatternOverlap(rules []PlatformModelRule) error {
	for left := range rules {
		if !rules[left].Enabled {
			continue
		}
		leftPattern, _, _ := normalizePlatformModelPattern(rules[left].ModelPattern)
		for right := left + 1; right < len(rules); right++ {
			if !rules[right].Enabled || rules[left].PlatformID == rules[right].PlatformID {
				continue
			}
			rightPattern, _, _ := normalizePlatformModelPattern(rules[right].ModelPattern)
			if platformModelPatternsOverlap(leftPattern, rightPattern) {
				return fmt.Errorf("%w: %q overlaps %q across platforms %d and %d", ErrPlatformModelRule, rules[left].ModelPattern, rules[right].ModelPattern, rules[left].PlatformID, rules[right].PlatformID)
			}
		}
	}
	return nil
}

func (r *platformModelResolver) Resolve(requestedModel string) (*ResolvedPlatformModel, error) {
	requested := strings.TrimSpace(requestedModel)
	if requested == "" {
		return nil, ErrPlatformModelNotFound
	}

	var best *PlatformModelRule
	bestExact := false
	bestPrefixLength := -1
	for index := range r.rules {
		rule := &r.rules[index]
		if !rule.Enabled {
			continue
		}
		pattern, wildcard, err := normalizePlatformModelPattern(rule.ModelPattern)
		if err != nil || !platformModelRuleMatches(strings.ToLower(requested), pattern, wildcard) {
			continue
		}
		if best == nil || shouldPreferPlatformModelRule(*rule, pattern, wildcard, *best, bestExact, bestPrefixLength) {
			best = rule
			bestExact = !wildcard
			bestPrefixLength = len(strings.TrimSuffix(pattern, "*"))
			continue
		}
		if samePlatformModelRulePriority(pattern, wildcard, *best, bestExact, bestPrefixLength) {
			return nil, ErrPlatformModelAmbiguous
		}
	}
	if best == nil {
		return nil, ErrPlatformModelNotFound
	}

	upstream := best.UpstreamModel
	if upstream == "" {
		upstream = requested
	}
	return &ResolvedPlatformModel{
		PlatformID:           best.PlatformID,
		PlatformCode:         best.PlatformCode,
		AccountPlatform:      best.AccountPlatform,
		RequestedModel:       requested,
		UpstreamModel:        upstream,
		EndpointCapabilities: append([]string(nil), best.EndpointCapabilities...),
		LegacyGroupID:        clonePlatformInt64Pointer(best.LegacyGroupID),
		RuleID:               best.ID,
	}, nil
}

func normalizePlatformModelPattern(raw string) (string, bool, error) {
	pattern := strings.ToLower(strings.TrimSpace(raw))
	if pattern == "" || strings.Count(pattern, "*") > 1 || (strings.Contains(pattern, "*") && !strings.HasSuffix(pattern, "*")) {
		return "", false, fmt.Errorf("%w: model pattern %q must be exact or use one suffix wildcard", ErrPlatformModelRule, raw)
	}
	return pattern, strings.HasSuffix(pattern, "*"), nil
}

func platformModelPatternsOverlap(left, right string) bool {
	leftPrefix, leftWildcard := strings.TrimSuffix(left, "*"), strings.HasSuffix(left, "*")
	rightPrefix, rightWildcard := strings.TrimSuffix(right, "*"), strings.HasSuffix(right, "*")
	switch {
	case !leftWildcard && !rightWildcard:
		return leftPrefix == rightPrefix
	case leftWildcard && rightWildcard:
		return strings.HasPrefix(leftPrefix, rightPrefix) || strings.HasPrefix(rightPrefix, leftPrefix)
	case leftWildcard:
		return strings.HasPrefix(rightPrefix, leftPrefix)
	default:
		return strings.HasPrefix(leftPrefix, rightPrefix)
	}
}

func platformModelRuleMatches(requested, pattern string, wildcard bool) bool {
	if wildcard {
		return strings.HasPrefix(requested, strings.TrimSuffix(pattern, "*"))
	}
	return requested == pattern
}

func shouldPreferPlatformModelRule(candidate PlatformModelRule, pattern string, wildcard bool, best PlatformModelRule, bestExact bool, bestPrefixLength int) bool {
	if !wildcard && !bestExact {
		return true
	}
	if wildcard && bestExact {
		return false
	}
	if wildcard {
		return len(strings.TrimSuffix(pattern, "*")) > bestPrefixLength
	}
	return false
}

func samePlatformModelRulePriority(pattern string, wildcard bool, best PlatformModelRule, bestExact bool, bestPrefixLength int) bool {
	if wildcard != !bestExact {
		return false
	}
	if wildcard {
		return len(strings.TrimSuffix(pattern, "*")) == bestPrefixLength
	}
	return true
}
