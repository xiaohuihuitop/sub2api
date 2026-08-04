package migrate

import (
	"testing"

	"entgo.io/ent/dialect/sql/schema"
)

func TestBillingProfileGroupForeignKeyDeletesDependentProfile(t *testing.T) {
	foreignKeys := BillingProfilesTable.ForeignKeys
	if len(foreignKeys) != 1 {
		t.Fatalf("BillingProfilesTable.ForeignKeys length = %d, want 1", len(foreignKeys))
	}

	if got := foreignKeys[0].OnDelete; got != schema.Cascade {
		t.Errorf("BillingProfilesTable.ForeignKeys[0].OnDelete = %s, want %s", got, schema.Cascade)
	}
}
