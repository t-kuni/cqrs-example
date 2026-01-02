# タスク一覧

## フェーズ1: OpenSearch連携基盤の実装

### 前提条件

- OpenSearchには既に `/products` インデックスが `spec/openSearchScheme/products.json` として登録されている
- 環境変数 `OPENSEARCH_ORIGIN` は各種envファイルに設定済み（値: `http://opensearch-node1:9200`）

### OpenSearch APIラッパーの作成

- [x] `domain/infrastructure/api/openSearchApi.go` を作成
  - インターフェース `IOpenSearchApi` を定義
  - 以下のメソッドを定義する
    - `IndexDocument(ctx context.Context, indexName string, documentID string, document string) error`
      - OpenSearchにドキュメントを登録または更新する
      - 既に同じドキュメントIDが存在する場合は更新する
      - `document` はJSON形式の文字列
  - go docコメントを記載する
    - 各メソッドの目的と使用場面を記載
  - モック生成用のコメントを記載
    - `//go:generate go tool mockgen -source=$GOFILE -destination=${GOFILE}_mock.go -package=$GOPACKAGE`

- [x] `infrastructure/api/openSearchApi.go` を作成
  - `domain/infrastructure/api/openSearchApi.go` のインターフェースを実装
  - `github.com/opensearch-project/opensearch-go` を使用する
  - 構造体 `OpenSearchApi` を定義
    - フィールド: `client *opensearch.Client`
  - コンストラクタ `NewOpenSearchApi() (api.IOpenSearchApi, error)` を実装
    - 環境変数 `OPENSEARCH_ORIGIN` から接続先を取得
    - OpenSearchクライアントを初期化
  - 各メソッドを実装
    - `IndexDocument`: `client.Index()` を使用
      - ドキュメントIDを指定することで、既存ドキュメントが存在する場合は更新される
  - エラーハンドリング: `eris.Wrap` で包んで返す

### Product同期サービスの作成

- [x] `domain/service/productTransferService.go` を作成
  - インターフェース `IProductTransferService` を定義
  - 以下のメソッドを定義する
    - `TransferAllProducts(ctx context.Context) error`
      - RDBの全productをOpenSearchに同期する
      - 内部で `TransferProduct` を全レコード分ループで呼び出す
      - エラーが発生した場合は処理を中断する
    - `TransferProduct(ctx context.Context, productID uuid.UUID) error`
      - 指定されたproductをOpenSearchに同期する
      - 既に同じproductIdが存在する場合は更新する
      - RDBからproduct、tenant、category、userを取得
      - OpenSearchのドキュメント構造に変換
        - `spec/openSearchScheme/products.json` を参照
        - `location` フィールドは `properties.latitude` と `properties.longitude` から生成
          - 形式: `{"lat": <latitude>, "lon": <longitude>}`
      - OpenSearchに登録
  - 構造体 `ProductTransferService` を定義
    - フィールド:
      - `DBConnector db.IConnector`
      - `OpenSearchApi api.IOpenSearchApi`
  - コンストラクタ `NewProductTransferService(conn db.IConnector, openSearchApi api.IOpenSearchApi) (IProductTransferService, error)` を実装
  - 各メソッドを実装
    - `TransferAllProducts`:
      - RDBから全productのIDを取得
      - 各productIDに対して `TransferProduct` を呼び出す
      - エラーが発生した場合は即座に処理を中断して返す
      - 進捗表示（10000件ごと）
    - `TransferProduct`:
      - entを使用してproductを取得
        - `WithTenant()`, `WithCategory()` でリレーションを取得
        - tenantから `WithOwner()` でuserを取得
      - OpenSearchのドキュメント構造に変換
        - JSON形式の文字列を生成
        - `location` フィールドについて
          - `properties.latitude` と `properties.longitude` の両方が存在する場合のみ生成
          - どちらか一方でもnullの場合は `location` フィールド自体を省略する
      - `OpenSearchApi.IndexDocument()` を呼び出す
        - インデックス名は `products` でハードコード
        - ドキュメントIDにはproductのIDを使用（既存ドキュメントがあれば更新される）
  - go docコメントを記載
  - モック生成用のコメントを記載

### コマンドの作成

- [x] `commands/transferProducts/main.go` を作成
  - 参考: `commands/seed-v2/main.go`
  - DIコンテナを使用してサービスを取得
  - `ProductTransferService.TransferAllProducts()` を呼び出す
  - エラーハンドリングとログ出力を実装
  - エラーが発生した場合は処理を中断してpanicする

### DIコンテナへの登録

- [x] `di/container.go` を編集
  - `fx.Provide` に以下を追加
    - `api.NewOpenSearchApi`
    - `service.NewProductTransferService`

### ビルド確認

- [x] `make generate` を実行してビルドが通ることを確認
  - エラーが出た場合は修正する

## 指示者宛ての懸念事項（作業対象外）

### パフォーマンスについて

- 100万件のproductを1件ずつ同期するため、処理時間が長くなる可能性がある
  - 将来的にパフォーマンスが問題になった場合は `BulkIndexDocuments` を使用したバッチ処理への変更を検討

## 事前修正提案

特になし

