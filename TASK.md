# タスク一覧

## 概要

`spec/データモデル.md` のトランザクション管理方針の変更に伴い、`commands/seed-v2/main.go` のトランザクション処理を修正する。

**変更内容:**
- `SET AUTOCOMMIT=0` の設定を削除
- 各 LOAD DATA LOCAL INFILE 実行毎に個別のトランザクションで処理するように変更
- 最後の COMMIT を削除（各テーブルのロード後に個別にコミットするため）

**仕様書:** `spec/データモデル.md`

## タスク

### seed-v2プログラムの修正

- [x] `commands/seed-v2/main.go` のトランザクション処理を修正
  - 40-43行目の `SET AUTOCOMMIT=0` の設定を削除
    - 現在: `_, err = db.Exec("SET AUTOCOMMIT=0")` とそのエラーハンドリング
    - 変更後: この処理を削除
  - 131-135行目の最後の `COMMIT` を削除
    - 現在: `_, err = db.Exec("COMMIT")` とそのエラーハンドリング
    - 変更後: この処理を削除
  - 各 LOAD DATA LOCAL INFILE の実行を個別のトランザクションで実行するように変更
    - Users のロード処理（99-105行目）を以下のように変更:
      1. `BEGIN` を実行
      2. LOAD DATA LOCAL INFILE を実行（既存の処理）
      3. `COMMIT` を実行
      4. エラーハンドリングを適切に実装
    - Tenants のロード処理（107-113行目）を同様に変更
    - Categories のロード処理（115-121行目）を同様に変更
    - Products のロード処理（123-129行目）を同様に変更
  - 処理フロー（変更後）:
    1. 各テーブルをTRUNCATE（既存の処理を維持）
    2. SET FOREIGN_KEY_CHECKS=0（既存の処理を維持）
    3. SET sql_log_bin=0（既存の処理を維持）
    4. 各テーブル用のCSVファイルを生成（既存の処理を維持）
    5. Users用のトランザクション: BEGIN → LOAD DATA LOCAL INFILE → COMMIT
    6. Tenants用のトランザクション: BEGIN → LOAD DATA LOCAL INFILE → COMMIT
    7. Categories用のトランザクション: BEGIN → LOAD DATA LOCAL INFILE → COMMIT
    8. Products用のトランザクション: BEGIN → LOAD DATA LOCAL INFILE → COMMIT
    9. 完了メッセージを表示（既存の処理を維持）

### ビルド確認

- [x] `make generate` を実行してビルドエラーがないことを確認

## 対象外のタスク

以下のタスクは不要です：

- **動作確認**: 実際にseed-v2を実行してのデータ投入確認は不要
- **テスト実行**: `make test` の実行は不要
- **swagger.yml や ent/schema の修正**: 今回の変更では不要

## 指示者宛ての懸念事項（作業対象外）

特になし。今回の変更は既存の実装に対する軽微な修正であり、トランザクション管理方針の変更のみです。

