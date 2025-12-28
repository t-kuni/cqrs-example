# タスク一覧

## フェーズ1: seed-v2プログラムの修正

### 仕様書

`spec/データモデル.md` の「サンプルデータ」セクションを参照

### 実装タスク

- [ ] `commands/seed-v2/main.go` の修正
  - データ投入前に以下のSQL設定を追加する
    - `SET FOREIGN_KEY_CHECKS=0` を実行（外部キー制約を一時的に無効化）
    - `SET AUTOCOMMIT=0` を実行（自動コミットを無効化）
    - `SET sql_log_bin=0` を実行（バイナリログを無効化）
  - 各テーブルをTRUNCATEする処理を追加
    - 対象テーブル: `users`, `tenants`, `categories`, `products`
    - TRUNCATEは外部キー制約無効化の後、データ投入の前に実行する
  - 処理終了時に `SET FOREIGN_KEY_CHECKS=1` を実行して外部キー制約を再有効化
  - 参考実装: `commands/seed/main.go` の46-57行目に類似の実装がある
  - 実装方法:
    - `conn.GetDB()` で `*sql.DB` を取得
    - `db.Exec()` メソッドでSQL文を実行
    - エラーハンドリングを適切に行う
    - defer文を使って確実に `FOREIGN_KEY_CHECKS=1` に戻す

### ビルド確認

- [ ] `make generate` を実行してビルドエラーがないことを確認

## 備考

- `commands/seed-v2/main.go` は現在 `conn.GetEnt()` のみを使用しているが、今回の修正で `conn.GetDB()` も使用することになる
- `SET AUTOCOMMIT=0` と `SET sql_log_bin=0` の設定は、大量データ投入時のパフォーマンス向上のために追加される
- 既存の `commands/seed/main.go` には `SET FOREIGN_KEY_CHECKS` と `TRUNCATE TABLE` の実装があるため、これを参考にすること

