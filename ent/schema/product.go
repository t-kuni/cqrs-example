package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/t-kuni/go-web-api-template/domain/model"
)

// Product holds the schema definition for the Product entity.
type Product struct {
	ent.Schema
}

// Fields of the Product.
func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"),
		field.String("tenant_id"),
		field.String("category_id"),
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

