package migrations

import (
	"strings"
	"testing"
)

func TestRemoveLegacyConfigurationMigration(t *testing.T) {
	content, err := FS.ReadFile("200_remove_legacy_configuration.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(content))
	for _, fragment := range []string{
		"active subscription cannot be mapped",
		"unused subscription redeem code cannot be mapped",
		"paid subscription order cannot be mapped",
		"update usage_logs",
		"rename column group_id to platform_id",
		"credentials - 'model_mapping' - 'model_whitelist' - 'openai_capabilities'",
		"drop column if exists legacy_group_id",
		"drop table if exists api_key_allowed_groups",
		"drop table if exists account_groups",
		"drop table if exists billing_profiles",
		"drop table if exists composite_model_routes",
		"drop table if exists channels",
		"drop table if exists groups",
		"'available_channels_enabled'",
		"'allow_ungrouped_key_scheduling'",
		"'{{group_name}}', '{{platform_name}}'",
		"'{{subscription_group}}', '{{subscription_plan}}'",
		"delete from ops_alert_rules",
		"'group_available_accounts'",
		"trg_api_key_platforms_auth_cache_invalidation",
		"trg_api_key_subscription_plans_auth_cache_invalidation",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
	if strings.Contains(sql, "drop table groups cascade") {
		t.Fatal("migration must explicitly remove dependencies before dropping groups")
	}
}
