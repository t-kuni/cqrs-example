package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Tenant holds the schema definition for the Tenant entity.
type Tenant struct {
	ent.Schema
}

// Fields of the Tenant.
func (Tenant) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"),
		field.String("owner_id"),
		field.String("name"),
	}
}

// Edges of the Tenant.
func (Tenant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("tenants").
			Unique().
			Required().
			Field("owner_id"),
		edge.To("products", Product.Type),
	}
}
