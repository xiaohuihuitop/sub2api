package migrations

import (
	"strings"
	"testing"
)

func TestBackfillPlatformCatalogMigratesGroupsModelsAndAccounts(t *testing.T) {
	content, err := FS.ReadFile("199_backfill_platform_catalog.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(content))
	for _, fragment := range []string{
		"insert into platforms",
		"legacy_group_id",
		"jsonb_array_elements_text",
		"insert into platform_model_rules",
		"update accounts",
		"row_number() over",
		"on conflict (code) do nothing",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration missing %q", fragment)
		}
	}
}
