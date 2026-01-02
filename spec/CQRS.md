# CQRS

* OpenSearchはCQRSを実現するための検索用のストレージとして利用する
* RDBのデータが更新されたらOpenSearchに同期する
* OpenSearchでは検索に最適化した構造で保持する（非正規化など）
* データの同期は RDB から OpenSearchの単方向
    * 逆方向の処理はない

## products の同期処理

* RDBのデータ構造は以下を参照
    * spec/models/products_properties.yaml
    * spec/データモデル.md
* OpenSearchのデータ構造は以下を参照
    * spec/openSearchScheme/products.json
* 実装について
    * commands/transferProducts/main.go として実装する
    * RDBの全レコードをOpenSearchに同期する
    * 既に同じproductIdが存在する場合は更新する
    * 主なロジックは domain/service に実装する
    * 1productを同期する関数を作成し、それを全レコード分ループで処理する
        * 将来的に別サービスに切り出す可能性あり
    * OpenSearchとの通信は github.com/opensearch-project/opensearch-go を利用する
        * ラッパーを infrastructure/api/openSearch.go として作成する（ここにはロジックを含めない）
    * locationフィールドについて
        * properties.latitude と properties.longitude の両方が存在する場合のみ生成する
        * どちらか一方でもnullの場合は locationフィールド自体を省略する
    * インデックス名は `products` でハードコードする
    