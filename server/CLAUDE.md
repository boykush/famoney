# Server Claude Code Guide

## APIテスト

serverのAPI（BFF grpc-gatewayが公開するRESTエンドポイント）を修正した場合は、`server/test/` 配下のhurlテストファイルも合わせて更新してください。

### テストファイル構成

```
server/test/
├── expense_health.hurl          # Expense HealthCheck
└── expense_operations.hurl      # Expense Operations
```

### テスト実行

```bash
mise run server:test
```
