# Server

Famoneyアプリケーションのバックエンドサービス群です。Go + gRPC で構築されています。

## アーキテクチャ概要

```
    ┌────────┐
    │  BFF   │ ← エントリポイント (grpc-gateway)
    └───┬────┘
        │ gRPC
        ▼
   ┌──────────────────────┐
   │  Backend Services    │
   ├──────────────────────┤
   │ • expense            │
   └──────────┬───────────┘
              │
              ▼
        ┌──────────┐
        │ PostgreSQL│
        └──────────┘
```

## サービス一覧

### [bff/](bff/)
**Backend for Frontend** - クライアントアプリケーションのAPIゲートウェイ

- 役割: REST APIエンドポイントを提供し、内部のgRPCマイクロサービスへリクエストをルーティング
- 技術: grpc-gateway を使用してgRPCをHTTP/RESTに変換
- ポート: 8080

### [expense/](expense/)
**支出管理サービス** - 支出データの管理API

- 役割: 支出の登録・取得
- プロトコル: gRPC
- ポート: 50052

## 開発

### 利用可能なコマンド
```bash
mise tasks | grep go
```

### Protocol Buffers コード生成
```bash
mise run proto:generate
```

### APIテスト
[hurl](https://hurl.dev/) を使用して、BFF grpc-gatewayが公開するRESTエンドポイントに対する統合テストを実行できます。全サービスのビルド・起動からテスト実行・クリーンアップまで一括で行います。

```bash
mise run server:test
```

テストファイルは [test/](test/) ディレクトリに配置されています。
