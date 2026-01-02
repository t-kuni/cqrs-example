//go:generate go tool mockgen -source=$GOFILE -destination=${GOFILE}_mock.go -package=$GOPACKAGE
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/t-kuni/cqrs-example/domain/infrastructure/api"
	"github.com/t-kuni/cqrs-example/domain/infrastructure/db"
	"github.com/t-kuni/cqrs-example/ent"
	"github.com/t-kuni/cqrs-example/ent/product"
)

// IProductTransferService は RDB上のproductをOpenSearchに同期するサービスのインターフェースです。
// データ移行やバッチ処理で使用されることを想定しています。
type IProductTransferService interface {
	// TransferAllProducts は RDB の全 product を OpenSearch に同期します。
	// エラーが発生した場合は処理を中断します。
	TransferAllProducts(ctx context.Context) error

	// TransferProduct は 指定された product を OpenSearch に同期します。
	// 既に同じproductIdが存在する場合は更新されます。
	TransferProduct(ctx context.Context, productID uuid.UUID) error
}

// ProductTransferService は IProductTransferService の実装です。
type ProductTransferService struct {
	DBConnector   db.IConnector
	OpenSearchApi api.IOpenSearchApi
}

// NewProductTransferService は ProductTransferService の新しいインスタンスを作成します。
func NewProductTransferService(conn db.IConnector, openSearchApi api.IOpenSearchApi) (IProductTransferService, error) {
	return &ProductTransferService{
		DBConnector:   conn,
		OpenSearchApi: openSearchApi,
	}, nil
}

// TransferAllProducts は RDB の全 product を OpenSearch に同期します。
func (s *ProductTransferService) TransferAllProducts(ctx context.Context) error {
	client := s.DBConnector.GetEnt()

	// 全productのIDを取得
	productIDs, err := client.Product.
		Query().
		IDs(ctx)
	if err != nil {
		return eris.Wrap(err, "")
	}

	fmt.Printf("Total products to transfer: %d\n", len(productIDs))

	// 各productIDに対してTransferProductを呼び出す
	for i, productID := range productIDs {
		err := s.TransferProduct(ctx, productID)
		if err != nil {
			return eris.Wrap(err, "")
		}

		// 進捗表示（10000件ごと）
		if (i+1)%10000 == 0 {
			fmt.Printf("Progress: %d/%d products transferred\n", i+1, len(productIDs))
		}
	}

	fmt.Printf("Completed: %d/%d products transferred\n", len(productIDs), len(productIDs))

	return nil
}

// TransferProduct は 指定された product を OpenSearch に同期します。
func (s *ProductTransferService) TransferProduct(ctx context.Context, productID uuid.UUID) error {
	client := s.DBConnector.GetEnt()

	// productを取得（関連エンティティも含む）
	p, err := client.Product.
		Query().
		Where(product.ID(productID)).
		WithTenant(func(tq *ent.TenantQuery) {
			tq.WithOwner()
		}).
		WithCategory().
		Only(ctx)
	if err != nil {
		return eris.Wrap(err, "")
	}

	// OpenSearchのドキュメント構造に変換
	doc := make(map[string]interface{})
	doc["id"] = p.ID.String()
	doc["name"] = p.Name
	doc["price"] = p.Price
	doc["listed_at"] = p.ListedAt

	// properties
	if p.Properties != nil {
		propertiesMap := make(map[string]interface{})
		if p.Properties.Size != nil {
			propertiesMap["size"] = *p.Properties.Size
		}
		if p.Properties.Latitude != nil {
			propertiesMap["latitude"] = *p.Properties.Latitude
		}
		if p.Properties.Longitude != nil {
			propertiesMap["longitude"] = *p.Properties.Longitude
		}
		if p.Properties.Color != nil {
			propertiesMap["color"] = *p.Properties.Color
		}
		doc["properties"] = propertiesMap

		// location フィールドの生成（latitude と longitude の両方が存在する場合のみ）
		if p.Properties.Latitude != nil && p.Properties.Longitude != nil {
			lat, err := strconv.ParseFloat(*p.Properties.Latitude, 64)
			if err == nil {
				lon, err := strconv.ParseFloat(*p.Properties.Longitude, 64)
				if err == nil {
					doc["location"] = map[string]interface{}{
						"lat": lat,
						"lon": lon,
					}
				}
			}
		}
	}

	// tenant
	if p.Edges.Tenant != nil {
		tenant := p.Edges.Tenant
		doc["tenant"] = map[string]interface{}{
			"id":   tenant.ID.String(),
			"name": tenant.Name,
		}

		// user (tenant.owner)
		if tenant.Edges.Owner != nil {
			user := tenant.Edges.Owner
			doc["user"] = map[string]interface{}{
				"id":   user.ID.String(),
				"name": user.Name,
			}
		}
	}

	// category
	if p.Edges.Category != nil {
		category := p.Edges.Category
		doc["category"] = map[string]interface{}{
			"id":   category.ID.String(),
			"name": category.Name,
		}
	}

	// JSON文字列に変換
	documentJSON, err := json.Marshal(doc)
	if err != nil {
		return eris.Wrap(err, "")
	}

	// OpenSearchに登録
	err = s.OpenSearchApi.IndexDocument(ctx, "products", p.ID.String(), string(documentJSON))
	if err != nil {
		return eris.Wrap(err, "")
	}

	return nil
}
