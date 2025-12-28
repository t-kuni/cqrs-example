# タスク一覧

## 概要

仕様書（spec/データモデル.md）の更新に伴い、各テーブルのIDカラムの型を `string` から `uuid` に変更する対応を行います。

**変更対象テーブル：**
- USERS
- TENANTS
- PRODUCTS
- CATEGORIES

**参照仕様書：** `spec/データモデル.md`

## フェーズ1: スキーマ定義の更新とコード生成

### 作業状況

**実施済み：**
- `ent/schema/user.go` のidフィールドをUUID型に変更（完了）
- `ent/schema/tenant.go` のidフィールドをUUID型に変更（完了）
- `ent/schema/product.go` のidフィールドをUUID型に変更（完了）
- `ent/schema/category.go` のidフィールドをUUID型に変更（完了）

**エラー発生：**
`make generate`実行時に以下のエラーが発生しました：
```
entc/gen: set "Category" foreign-keys: mismatch field type between edge field "category_id" and id of type "Category" (string != [16]byte)
```

**原因：**
各テーブルのIDフィールドをUUID型に変更しましたが、外部キーフィールド（FKフィールド）も同様にUUID型に変更する必要があります。

**追加で必要な作業（TASK.mdに記載されていなかった作業）：**
- `ent/schema/tenant.go` の `owner_id` フィールドをUUID型に変更（UserテーブルへのFK）
- `ent/schema/product.go` の `tenant_id` フィールドをUUID型に変更（TenantテーブルへのFK）
- `ent/schema/product.go` の `category_id` フィールドをUUID型に変更（CategoryテーブルへのFK）

**次のステップ：**
上記の外部キーフィールドもUUID型に変更する必要があります。これらの変更をTASK.mdに追加し、作業を継続するかどうかご指示ください。

### ent/schemaの修正

- [ ] `ent/schema/user.go` のidフィールドをUUID型に変更
  - 変更箇所： `Fields()` メソッド内の `field.String("id")` を UUID型に変更
  - 変更例： `field.UUID("id", uuid.UUID{}).Default(uuid.New)` または適切なUUID設定を使用
  - 注意： entのUUID型の使い方については、entのドキュメントを参照すること

- [ ] `ent/schema/tenant.go` のidフィールドをUUID型に変更
  - 変更箇所： `Fields()` メソッド内の `field.String("id")` を UUID型に変更
  - 変更例： `field.UUID("id", uuid.UUID{}).Default(uuid.New)` または適切なUUID設定を使用

- [ ] `ent/schema/product.go` のidフィールドをUUID型に変更
  - 変更箇所： `Fields()` メソッド内の `field.String("id")` を UUID型に変更
  - 変更例： `field.UUID("id", uuid.UUID{}).Default(uuid.New)` または適切なUUID設定を使用

- [ ] `ent/schema/category.go` のidフィールドをUUID型に変更
  - 変更箇所： `Fields()` メソッド内の `field.String("id")` を UUID型に変更
  - 変更例： `field.UUID("id", uuid.UUID{}).Default(uuid.New)` または適切なUUID設定を使用

