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

### ent/schemaの修正（指示者が実施）

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

