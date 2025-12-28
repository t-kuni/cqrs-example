package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/t-kuni/cqrs-example/domain/model"
)

// Product holds the schema definition for the Product entity.
type Product struct {
	ent.Schema
}

// Fields of the Product.
func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("tenant_id", uuid.UUID{}),
		field.UUID("category_id", uuid.UUID{}),
		field.String("name"),
		field.Int64("price"),
		field.JSON("properties", &model.ProductProperties{}),
	}
}

// Edges of the Product.
func (Product) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("products").
			Unique().
			Required().
			Field("tenant_id"),
		edge.From("category", Category.Type).
			Ref("products").
			Unique().
			Required().
			Field("category_id"),
	}
}

