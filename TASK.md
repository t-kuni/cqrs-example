# タスク一覧

仕様書の更新内容に基づいて、以下のタスクを実施します。

参照仕様書: `spec/データモデル.md`

## フェーズ1: データモデルの修正とサンプルデータ登録プログラムの作成

### productsテーブルへのlisted_atカラム追加

- [ ] `ent/schema/product.go`を編集
  - `Fields()`メソッドに`field.Time("listed_at")`を追加
  - 仕様書のER図に従い、datetime型のlisted_atカラムを追加する

### サンプルデータ登録プログラムの作成

- [ ] `commands/seed-v2/main.go`を新規作成
  - 参考実装: `commands/seed/main.go`
  - 以下のサンプルデータを登録するプログラムを作成
    - **users**: 200レコード
      - name: `ユーザX` (Xは1からの連番)
    - **tenants**: 1000レコード
      - name: `テナントX` (Xは1からの連番)
      - owner_id: usersを満遍なく紐付ける（200ユーザに対して1000テナントなので、各ユーザに5テナントずつ割り当て）
    - **categories**: 50レコード
      - name: `カテゴリX` (Xは1からの連番)
    - **products**: 100万レコード
      - tenant_id: tenantsを満遍なく紐付ける（1000テナントに対して100万商品なので、各テナントに1000商品ずつ割り当て）
      - category_id: categoriesを満遍なく紐付ける（50カテゴリに対して100万商品なので、各カテゴリに20000商品ずつ割り当て）
      - name: `商品X` (Xは1からの連番)
      - price: 100〜10000の乱数(整数)
      - properties: `domain/model/productProperties.go`の構造に従う
        - size: `spec/models/products_properties.yaml`のenumの値（S, M, L）をランダムに設定
        - latitude: 20.43〜45.55の乱数(実数、文字列型として保存)
        - longitude: 122.93〜153.99の乱数(実数、文字列型として保存)
        - color: `spec/models/products_properties.yaml`のenumの値（red, green, blue）をランダムに設定
      - listed_at: 現在時刻から過去1年間のランダムな日時
  - BulkInsertを活用して効率よくレコードを作成する
  - 実装方針:
    - entのBulkCreateを使用してバッチ挿入を行う
    - メモリ使用量を抑えるため、productsは適切なバッチサイズ（例: 1000件ずつ）に分割して挿入
    - 外部キー制約を考慮し、users → tenants → categories → products の順に登録
    - 乱数生成には`math/rand`パッケージを使用
    - 日時生成には`time`パッケージを使用

## 指示者宛ての懸念事項（作業対象外）

- 100万レコードのproductsデータを登録するため、実行時間が長くなる可能性があります。バッチサイズやトランザクション管理の調整が必要になるかもしれません。

