# タスク一覧

## 概要

`commands/seed-v2/main.go` のサンプルデータ投入処理を BulkInsert 方式から CSV + LOAD DATA LOCAL INFILE 方式に変更する。

仕様書：`spec/データモデル.md`

## タスク

### MySQL設定の変更

- [ ] `local-env/db/mysql.cnf` に LOAD DATA LOCAL INFILE を有効化する設定を追加
  - `[mysqld]` セクションに `local_infile=1` を追加
  - LOAD DATA LOCAL INFILE を使用するために必要な設定

### DB接続設定の変更

- [ ] infrastructure層のDB接続部分で LOAD DATA LOCAL INFILE を有効化
  - DB接続のDSN（Data Source Name）に `allowAllFiles=true` パラメータを追加
  - 該当ファイル：infrastructure層のDB接続を初期化しているファイル
  - 補足：既存のDSN設定を確認し、パラメータに追加する

### seed-v2プログラムの修正

- [ ] `commands/seed-v2/main.go` を修正
  - 変更概要：
    - BulkInsert（CreateBulk）を使用したデータ投入をやめる
    - 代わりに、CSV ファイルを生成してから LOAD DATA LOCAL INFILE でデータを投入する方式に変更
  - 処理フロー：
    1. 各テーブルをTRUNCATE（既存の処理を維持）
    2. SET FOREIGN_KEY_CHECKS=0（既存の処理を維持）
    3. SET AUTOCOMMIT=0（既存の処理を維持）
    4. SET sql_log_bin=0（既存の処理を維持）
    5. 各テーブル用のCSVファイルを生成
       - users.csv：200レコード、カラム：id（UUID）、name
       - tenants.csv：1000レコード、カラム：id（UUID）、owner_id（UUID）、name
       - categories.csv：50レコード、カラム：id（UUID）、name
       - products.csv：100万レコード、カラム：id（UUID）、tenant_id（UUID）、category_id（UUID）、name、price、properties（JSON文字列）、listed_at（datetime形式）
       - CSV出力先：`/tmp/` ディレクトリ
       - CSVフォーマット：カンマ区切り、ヘッダーなし、文字列はダブルクォートで囲む
       - JSONのpropertiesフィールドは文字列としてエスケープして出力
       - UUIDはgoogle/uuidパッケージで生成
    6. LOAD DATA LOCAL INFILE を実行してCSVファイルからデータを投入
       - 各テーブルごとに実行
       - SQL例：`LOAD DATA LOCAL INFILE '/tmp/users.csv' INTO TABLE users FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n' (id, name);`
       - products テーブルについては properties カラムが JSON 型なので適切にマッピング
    7. COMMIT を実行
    8. 完了メッセージを表示
  - データ生成ロジック：
    - 既存の仕様書通り（spec/データモデル.md参照）
    - users：200レコード、name は `ユーザX`（X は 1 からの連番）
    - tenants：1000レコード、name は `テナントX`、userを満遍なく紐付ける
    - categories：50レコード、name は `カテゴリX`
    - products：100万レコード、name は `商品X`、price は 100〜10000 の乱数、properties は spec/models/products_properties.yaml の仕様通り、listed_at は過去1年以内のランダムな日時
  - 注意点：
    - CSV生成時に大量のメモリを消費しないよう、バッファリングを活用すること
    - 特に products.csv（100万レコード）は逐次書き込みを行うこと

### ビルド確認

- [ ] `make generate` を実行してビルドエラーがないことを確認

## 対象外のタスク

以下のタスクは不要です：

- **動作確認**: 実際にseed-v2を実行してのデータ投入確認は不要
- **テスト実行**: `make test` の実行は不要

## 指示者宛ての懸念事項（作業対象外）

- LOAD DATA LOCAL INFILE はセキュリティリスクがあるため、本番環境では使用しないことを推奨
- 本機能は開発環境専用のサンプルデータ投入ツールという位置づけであることを確認
- MySQLのバージョンによっては LOAD DATA LOCAL INFILE の動作が異なる可能性がある（MySQL 8.0以降は制限が厳しくなっている）
- Docker環境でのファイルパスの扱いに注意が必要（コンテナ内のパスとホストのパスの違い）

