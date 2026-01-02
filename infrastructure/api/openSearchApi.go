package api

import (
	"context"
	"os"
	"strings"

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/rotisserie/eris"
	"github.com/t-kuni/cqrs-example/domain/infrastructure/api"
)

// OpenSearchApi は IOpenSearchApi インターフェースの実装です。
type OpenSearchApi struct {
	client *opensearch.Client
}

// NewOpenSearchApi は OpenSearchApi の新しいインスタンスを作成します。
// 環境変数 OPENSEARCH_ORIGIN から接続先を取得します。
func NewOpenSearchApi() (api.IOpenSearchApi, error) {
	origin := os.Getenv("OPENSEARCH_ORIGIN")
	if origin == "" {
		return nil, eris.New("OPENSEARCH_ORIGIN environment variable is not set")
	}

	client, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{origin},
	})
	if err != nil {
		return nil, eris.Wrap(err, "")
	}

	return &OpenSearchApi{
		client: client,
	}, nil
}

// IndexDocument は OpenSearch にドキュメントを登録または更新します。
func (o *OpenSearchApi) IndexDocument(ctx context.Context, indexName string, documentID string, document string) error {
	res, err := o.client.Index(
		indexName,
		strings.NewReader(document),
		o.client.Index.WithDocumentID(documentID),
		o.client.Index.WithContext(ctx),
	)
	if err != nil {
		return eris.Wrap(err, "")
	}
	defer res.Body.Close()

	if res.IsError() {
		return eris.Errorf("failed to index document: %s", res.Status())
	}

	return nil
}
