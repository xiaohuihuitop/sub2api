package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// BillingProfile stores balance pricing for a routing group. Subscription terms
// are copied into UserSubscription, so editing this profile cannot alter them.
type BillingProfile struct {
	ent.Schema
}

func (BillingProfile) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "billing_profiles"}}
}

func (BillingProfile) Mixin() []ent.Mixin {
	return []ent.Mixin{mixins.TimeMixin{}}
}

func (BillingProfile) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("group_id").Unique(),
		field.Float("balance_rate_multiplier").SchemaType(map[string]string{dialect.Postgres: "decimal(10,4)"}).Default(1.0),
		field.Bool("peak_rate_enabled").Default(false),
		field.String("peak_start").MaxLen(5).Default(""),
		field.String("peak_end").MaxLen(5).Default(""),
		field.Float("peak_rate_multiplier").SchemaType(map[string]string{dialect.Postgres: "decimal(10,4)"}).Default(1.0),
		field.Bool("image_rate_independent").Default(false),
		field.Float("image_rate_multiplier").SchemaType(map[string]string{dialect.Postgres: "decimal(10,4)"}).Default(1.0),
		field.Float("image_price_1k").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Float("image_price_2k").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Float("image_price_4k").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Float("batch_image_discount_multiplier").SchemaType(map[string]string{dialect.Postgres: "decimal(10,4)"}).Default(0.5),
		field.Float("batch_image_hold_multiplier").SchemaType(map[string]string{dialect.Postgres: "decimal(10,4)"}).Default(0.6),
		field.Bool("video_rate_independent").Default(false),
		field.Float("video_rate_multiplier").SchemaType(map[string]string{dialect.Postgres: "decimal(10,4)"}).Default(1.0),
		field.Float("video_price_480p").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Float("video_price_720p").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Float("video_price_1080p").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Float("web_search_price_per_call").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
	}
}

func (BillingProfile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("group", Group.Type).
			Ref("billing_profile").
			Field("group_id").
			Required().
			Unique(),
	}
}
