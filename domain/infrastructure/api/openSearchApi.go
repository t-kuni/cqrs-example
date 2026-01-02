//go:generate go tool mockgen -source=$GOFILE -destination=${GOFILE}_mock.go -package=$GOPACKAGE
package api

import (
	"context"
)

// IOpenSearchApi は OpenSearch に対する操作を提供するインターフェースです。
// RDB上のデータをOpenSearchに同期する際に使用します。
type IOpenSearchApi interface {
	// IndexDocument は OpenSearch にドキュメントを登録または更新します。
	// 既に同じドキュメントIDが存在する場合は更新されます。
	//
	// Parameters:
	//   - ctx: コンテキスト
	//   - indexName: インデックス名
	//   - documentID: ドキュメントID
	//   - document: JSON形式のドキュメント文字列
	//
	// Returns:
	//   - error: エラーが発生した場合
	IndexDocument(ctx context.Context, indexName string, documentID string, document string) error
}
