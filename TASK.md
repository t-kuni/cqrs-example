# タスク一覧

## 概要

`spec/データモデル.md` の仕様追加に伴い、`commands/seed-v2/main.go` にproductsテーブルのキー・インデックスの退避・復元機能を追加する。

**追加された仕様:**
- productsテーブルについて、各種キーとインデックスの情報を取得・退避しておき、キーとインデックスを削除する
- レコード登録完了時に復元する
- information_schema.STATISTICS や information_schema.KEY_COLUMN_USAGE などを活用する

**目的:**
大量データ（100万レコード）のLOAD DATA LOCAL INFILE実行時のパフォーマンス向上のため、productsテーブルのキー・インデックスを一時的に削除し、データ投入後に復元する。

**仕様書:** `spec/データモデル.md`

## タスク

### productsテーブルのキー・インデックス退避・復元機能の実装

- [ ] `commands/seed-v2/main.go` にキー・インデックスの退避・復元機能を追加
  - **処理フロー:**
    1. 各テーブルをTRUNCATE（既存の処理を維持）
    2. SET FOREIGN_KEY_CHECKS=0（既存の処理を維持）
    3. SET sql_log_bin=0（既存の処理を維持）
    4. **【新規】productsテーブルのキー・インデックス情報を取得・退避**
    5. **【新規】productsテーブルのキー・インデックスを削除**
    6. 各テーブル用のCSVファイルを生成（既存の処理を維持）
    7. Users用のトランザクション: BEGIN → LOAD DATA LOCAL INFILE → COMMIT（既存の処理を維持）
    8. Tenants用のトランザクション: BEGIN → LOAD DATA LOCAL INFILE → COMMIT（既存の処理を維持）
    9. Categories用のトランザクション: BEGIN → LOAD DATA LOCAL INFILE → COMMIT（既存の処理を維持）
    10. Products用のトランザクション: BEGIN → LOAD DATA LOCAL INFILE → COMMIT（既存の処理を維持）
    11. **【新規】productsテーブルのキー・インデックスを復元**
    12. 完了メッセージを表示（既存の処理を維持）

  - **実装詳細:**
    - キー・インデックス情報の取得
      - `information_schema.STATISTICS` テーブルからproductsテーブルのインデックス情報を取得
      - `information_schema.KEY_COLUMN_USAGE` テーブルから外部キー情報を取得
      - 取得した情報を構造体に格納して保持
    - キー・インデックスの削除
      - PRIMARY KEYは削除しない（削除すると復元が困難なため）
      - 外部キー制約を削除（`ALTER TABLE products DROP FOREIGN KEY <constraint_name>`）
      - インデックスを削除（`ALTER TABLE products DROP INDEX <index_name>`）
      - ※ PRIMARY KEY以外のインデックス・外部キーを削除対象とする
    - キー・インデックスの復元
      - 削除時に退避した情報を元に、外部キー制約を再作成（`ALTER TABLE products ADD CONSTRAINT <constraint_name> FOREIGN KEY ...`）
      - 削除時に退避した情報を元に、インデックスを再作成（`ALTER TABLE products ADD INDEX <index_name> ...`）
    - エラーハンドリング
      - キー・インデックスの取得・削除・復元で発生したエラーは適切にハンドリングしてpanicする
      - 復元に失敗した場合は、どのキー・インデックスの復元に失敗したかをログに出力する

  - **実装方針:**
    - キー・インデックス情報を保持する構造体を定義する
      ```go
      type IndexInfo struct {
          IndexName    string
          ColumnName   string
          NonUnique    bool
          IndexType    string
      }
      
      type ForeignKeyInfo struct {
          ConstraintName      string
          ColumnName          string
          ReferencedTableName string
          ReferencedColumnName string
      }
      ```
    - キー・インデックス情報を取得する関数を実装する
      ```go
      func getProductsIndexes(db *sql.DB) ([]IndexInfo, error)
      func getProductsForeignKeys(db *sql.DB) ([]ForeignKeyInfo, error)
      ```
    - キー・インデックスを削除する関数を実装する
      ```go
      func dropProductsIndexes(db *sql.DB, indexes []IndexInfo) error
      func dropProductsForeignKeys(db *sql.DB, fks []ForeignKeyInfo) error
      ```
    - キー・インデックスを復元する関数を実装する
      ```go
      func restoreProductsIndexes(db *sql.DB, indexes []IndexInfo) error
      func restoreProductsForeignKeys(db *sql.DB, fks []ForeignKeyInfo) error
      ```
    - main関数内の適切な箇所でこれらの関数を呼び出す

  - **注意事項:**
    - PRIMARY KEYは削除・復元の対象外とする
    - 外部キー制約の復元時は、参照先テーブル（tenants, categories）のデータが既に投入されている必要がある
    - インデックスの復元は、データ投入後に実行する必要がある
    - `SET FOREIGN_KEY_CHECKS=0` が設定されているため、外部キー制約の削除・復元は慎重に行う

### ビルド確認

- [ ] `make generate` を実行してビルドエラーがないことを確認

## 対象外のタスク

以下のタスクは不要です：

- **動作確認**: 実際にseed-v2を実行してのデータ投入確認は不要（ビルドの確認のみ）
- **テスト実行**: `make test` の実行は不要
- **swagger.yml や ent/schema の修正**: 今回の変更では不要
- **他のテーブル（users, tenants, categories）のキー・インデックス退避・復元**: 仕様書ではproductsテーブルのみが対象

## 指示者宛ての懸念事項（作業対象外）

特になし。以下の点は確認済み：

- PRIMARY KEYは削除対象外とする（仕様書に追記済み）
- 外部キー制約の復元タイミングは問題なし（products投入後に復元）

