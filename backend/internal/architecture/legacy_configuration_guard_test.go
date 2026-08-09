package architecture

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

var forbiddenLegacyConfiguration = []string{
	`json:"group_id"`,
	`json:"group_ids"`,
	"AllowedGroups",
	"AllowedGroupIDs",
	"ResolveBillingGroupForRequest",
	"PricingGroupID",
	"LegacyPricingGroupID",
	"legacy_group_id",
	"api_key_allowed_groups",
	"account_groups",
	"user_allowed_groups",
	"user_group_rate_multipliers",
	"platform_assets_v2_enabled",
	"available_channels_enabled",
	"allow_ungrouped_key_scheduling",
	"user_group_rate_cache_ttl_seconds",
	"models_list_cache_ttl_seconds",
	"skip_default_group_bind",
	"GroupID",
	"groupID",
	"SchedulingGroup",
	"OPS_GROUP_ID_INVALID",
	"UserGroupRPM",
	"userGroupRPM",
	"rpm:ug:",
	"ensureCompositeTargetPlatform",
	"compositeTargetPlatform",
	"compositeBillableModel",
	"BATCH_IMAGE_GROUP_DISABLED",
	"GROUP_NOT_SUBSCRIPTION_TYPE",
	"subscription_group",
	"resolveChannelPricing",
	"resolveOpenAIChannelPricing",
	"allowOpenAICompatibleMessagesDispatch",
	"liveEnabledForAPIKey",
	"all_groups",
	"group_count",
	"legacyAuthorization",
}

func TestNoLegacyConfigurationInActiveSource(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture test path")
	}
	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
	repositoryRoot := filepath.Dir(backendRoot)

	var violations []string
	for _, root := range []string{
		filepath.Join(backendRoot, "internal"),
		filepath.Join(backendRoot, "ent", "schema"),
	} {
		violations = append(violations, scanGoProductionSource(root)...)
	}
	violations = append(violations, scanFrontendProductionSource(filepath.Join(repositoryRoot, "frontend", "src"))...)
	if len(violations) > 0 {
		t.Fatalf("legacy configuration returned to active source:\n%s", strings.Join(violations, "\n"))
	}
}

func TestOpsContractsUsePlatformTerminology(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture test path")
	}
	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
	repositoryRoot := filepath.Dir(backendRoot)

	files := []string{
		filepath.Join(backendRoot, "internal", "service", "ops_models.go"),
		filepath.Join(backendRoot, "internal", "service", "ops_trend_models.go"),
		filepath.Join(backendRoot, "internal", "service", "ops_user_error.go"),
		filepath.Join(backendRoot, "internal", "repository", "ops_repo.go"),
		filepath.Join(backendRoot, "internal", "repository", "ops_repo_trends.go"),
		filepath.Join(repositoryRoot, "frontend", "src", "api", "admin", "ops.ts"),
		filepath.Join(repositoryRoot, "frontend", "src", "components", "user", "UserErrorRequestsTable.vue"),
	}
	files = append(files, productionFilesUnder(filepath.Join(repositoryRoot, "frontend", "src", "views", "admin", "ops"))...)

	forbidden := []string{"group_name", "GroupName", "top_groups", "TopGroups", "groupId", "group-id", "topGroups", "selectGroup"}
	var violations []string
	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			violations = append(violations, fmt.Sprintf("%s: %v", path, err))
			continue
		}
		for _, token := range forbidden {
			if strings.Contains(string(content), token) {
				violations = append(violations, fmt.Sprintf("%s: %s", path, token))
			}
		}
	}
	if len(violations) > 0 {
		t.Fatalf("ops contract still exposes legacy group terminology:\n%s", strings.Join(violations, "\n"))
	}
}

func productionFilesUnder(root string) []string {
	var files []string
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() {
			if entry.Name() == "__tests__" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".vue") || (strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".spec.ts")) {
			files = append(files, path)
		}
		return nil
	})
	return files
}

func scanGoProductionSource(root string) []string {
	fset := token.NewFileSet()
	var violations []string
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			violations = append(violations, fmt.Sprintf("%s: parse error: %v", path, parseErr))
			return nil
		}
		ast.Inspect(file, func(node ast.Node) bool {
			var value string
			switch typed := node.(type) {
			case *ast.Ident:
				value = typed.Name
			case *ast.BasicLit:
				value = typed.Value
				if unquoted, unquoteErr := strconv.Unquote(typed.Value); unquoteErr == nil {
					value = unquoted
				}
			default:
				return true
			}
			for _, forbidden := range forbiddenLegacyConfiguration {
				if strings.Contains(value, forbidden) {
					position := fset.Position(node.Pos())
					violations = append(violations, fmt.Sprintf("%s:%d: %s", path, position.Line, forbidden))
				}
			}
			return true
		})
		return nil
	})
	return violations
}

func scanFrontendProductionSource(root string) []string {
	var violations []string
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() {
			if entry.Name() == "__tests__" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".spec.ts") || (!strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".vue")) {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			violations = append(violations, fmt.Sprintf("%s: read error: %v", path, readErr))
			return nil
		}
		text := string(content)
		for _, forbidden := range forbiddenLegacyConfiguration {
			if strings.Contains(text, forbidden) {
				violations = append(violations, fmt.Sprintf("%s: %s", path, forbidden))
			}
		}
		return nil
	})
	return violations
}
