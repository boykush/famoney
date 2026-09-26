# Famoney

家族の家計データ基盤。マネーフォワード ME の入出金明細を DuckDB でデータレイクに取り込み、整えた明細を
認証付きの MCP サーバーで配る。主な利用先は [boykush/life](https://github.com/boykush/life) のエージェント。

```
マネーフォワード ME ──(ingest)──▶ raw 層 ──(transform)──▶ product 層 ──(mcp)──▶ life のエージェント
                     CSV(Shift_JIS)    CSV(UTF-8)          Parquet          Bearer 認証
                                  └──── Cloudflare R2（DuckDB httpfs で読み書き）────┘
```

API サーバーとフロントエンドは、必要になった時点で product 層の上に足す。

## 構成

1つの Go バイナリ（= 1つのイメージ）をサブコマンドで使い分ける。

| サブコマンド | 役割 | 動かし方（iac 側） |
| --- | --- | --- |
| `famoney ingest --month <月>` | マネーフォワード ME から対象月の CSV を取得し、raw 層へ置く | CronJob / Job |
| `famoney transform --month <月>` | raw 層の CSV を product の明細（Parquet）に変換する | CronJob / Job |
| `famoney mcp --addr <host:port>` | product の明細を MCP（Streamable HTTP、`/mcp`）で配る | Deployment |

`<月>` は `YYYY-MM`・`current`（既定）・`previous`。`current` / `previous` は JST で数える。同じ月を
流し直すと上書きするので、月の途中で何度流してもよい。

```
cmd/famoney/            エントリポイント（サブコマンド）
internal/moneyforward/  マネーフォワード ME の CSV ダウンロード（Shift_JIS → UTF-8、列の検証）
internal/lake/          DuckDB の接続とデータレイクのパス
internal/ingest/        raw 層への取り込み
internal/transform/     raw → product の変換（transactions.sql）
internal/mcpserver/     MCP サーバー（tool と Bearer 認証）
```

### データレイク

ルート（`FAMONEY_LAKE_ROOT`）は `s3://<bucket>[/prefix]`（S3 互換。本番は Cloudflare R2）かローカルのディレクトリ。
どちらも同じ配置になる。

```
raw/moneyforward/month=YYYY-MM/transactions.csv     マネーフォワードの CSV（UTF-8、全列文字列のまま）
product/transactions/month=YYYY-MM/data.parquet     明細
```

product の明細の列:

| 列 | 型 | 中身 |
| --- | --- | --- |
| `id` | VARCHAR | マネーフォワードの ID。重複は1件に寄せる |
| `date` | DATE | 日付 |
| `description` | VARCHAR | 内容 |
| `amount` | BIGINT | 金額（円）。収入が正、支出が負 |
| `account` / `category` / `subcategory` / `memo` | VARCHAR | 保有金融機関 / 大項目 / 中項目 / メモ |
| `is_target` / `is_transfer` | BOOLEAN | 計算対象 / 振替 |
| `kind` | VARCHAR | `income` / `expense` / `transfer`（振替）/ `excluded`（計算対象外）。収支に入るのは前の2つ |
| `month` | VARCHAR | パーティション（`YYYY-MM`）。読むときに hive partitioning で付く |

### MCP の tool

| tool | 返すもの |
| --- | --- |
| `list_months` | 取り込み済みの月と、月ごとの収入・支出・収支 |
| `get_monthly_summary` | 対象月の収支と大項目ごとの内訳 |
| `list_transactions` | 対象月の明細。区分・大項目・キーワードで絞り込める |
| `get_category_trend` | 大項目（任意で中項目）の月ごとの推移 |

`/mcp` は `Authorization: Bearer <token>` が `FAMONEY_MCP_TOKENS` のどれかと一致しないと 401 を返す。
家計は非公開の情報なので、トークンが空だとサーバーは起動しない。`/healthz` だけは無認証（probe 用）。

## 環境変数

| 変数 | 使うサブコマンド | 中身 |
| --- | --- | --- |
| `FAMONEY_LAKE_ROOT` | すべて | データレイクのルート（`s3://famoney` など） |
| `FAMONEY_S3_ENDPOINT` | すべて（s3 のとき） | R2 なら `<account_id>.r2.cloudflarestorage.com` |
| `FAMONEY_S3_REGION` | すべて（s3 のとき） | R2 なら `auto` |
| `FAMONEY_S3_URL_STYLE` | すべて（s3 のとき） | `vhost`（既定）か `path`。R2 は `path` |
| `FAMONEY_S3_ACCESS_KEY_ID` / `FAMONEY_S3_SECRET_ACCESS_KEY` | すべて（s3 のとき） | R2 の API トークンのアクセスキー（secret） |
| `MONEYFORWARD_COOKIE` | `ingest` | ログイン済みブラウザの Cookie ヘッダ（secret） |
| `FAMONEY_MCP_TOKENS` | `mcp` | 受け付ける Bearer トークン。カンマ区切りで複数（入れ替え用）（secret） |
| `FAMONEY_DUCKDB_EXTENSION_DIRECTORY` | すべて | DuckDB 拡張の置き場。イメージが設定済みなので普段は触らない |

### マネーフォワード ME の Cookie

マネーフォワード ME には公開 API が無いので、ブラウザと同じ CSV ダウンロード（`/cf/csv`）をログイン済みの
Cookie で叩いている。CSV のダウンロードはプレミアム会員の機能。

Cookie はブラウザの開発者ツールで `moneyforward.com` へのリクエストの `Cookie` ヘッダを写す。セッションが
切れると `ingest` は `moneyforward session expired` で失敗するので、そのときに入れ替える。

## デプロイ

この repo が出すのは `ghcr.io/boykush/famoney` のイメージまで（`.github/workflows/image.yml`。main で
`<commit 7桁>` と `main` の2つのタグを push）。k8s のマニフェスト・Secret・公開ホスト名・digest の追従は
[boykush/infrastructure-as-code](https://github.com/boykush/infrastructure-as-code) が持つ。

iac 側で要るもの:

- `ingest` → `transform` を順に流す CronJob（例: 毎日、`--month current`。月初に `--month previous` も）
- `mcp` の Deployment / Service（port 8080、probe は `/healthz`）と、Cloudflare Tunnel のホスト名
- Secret: R2 のアクセスキー、`MONEYFORWARD_COOKIE`、`FAMONEY_MCP_TOKENS`

利用側（life）は `.mcp.json` でヘッダにトークンを渡す。値は環境変数から展開させ、repo には書かない。

```json
{
  "mcpServers": {
    "famoney": {
      "type": "http",
      "url": "https://famoney-mcp.boykush.com/mcp",
      "headers": { "Authorization": "Bearer ${FAMONEY_MCP_TOKEN}" }
    }
  }
}
```

## 開発

[mise](https://mise.jdx.dev/) でツールとタスクを揃える。DuckDB を cgo でリンクするので C コンパイラが要る。

```bash
mise install
mise run check                          # fmt / vet / lint / tidy / test（CI と同じ）

# 手元のパイプライン（データレイクは .data/）
mise run ingest -- --month 2026-09      # .mise.local.toml の [env] に MONEYFORWARD_COOKIE を置く
mise run transform -- --month 2026-09
mise run mcp                            # 127.0.0.1:8080、トークンは local
```
