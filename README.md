# Famoney

支出を管理できる個人向けWebアプリケーション。マイクロサービスアーキテクチャで構築されています。

## プロジェクト構成

```
famoney/
├── server/            # バックエンドサービス群 (Go, gRPC)
│   ├── bff/          # Backend for Frontend (gRPC Gateway)
│   └── expense/      # 支出管理サービス
├── web/              # フロントエンド (Next.js)
├── k8s/              # Kubernetesマニフェスト
│   ├── base/         # ベースマニフェスト
│   └── overlays/     # 環境別設定（local）
└── .mise-tasks/      # タスク定義（mise用）
```

詳細は各ディレクトリのREADMEを参照してください：
- [server/README.md](server/README.md) - バックエンドサービスのアーキテクチャと開発ガイド
- [web/README.md](web/README.md) - フロントエンドの開発ガイド
- [k8s/README.md](k8s/README.md) - Kubernetesマニフェストの構成と管理

## 開発環境のセットアップ

### mise のインストール

このプロジェクトでは、[mise](https://mise.jdx.dev/)を使用してツールのバージョン管理とタスク実行を行っています。

```bash
# mise のインストール（macOS）
brew install mise

# シェル設定の追加（例: zsh）
echo 'eval "$(mise activate zsh)"' >> ~/.zshrc
source ~/.zshrc
```

### 依存ツールのインストール

プロジェクトルートで以下を実行すると、[.mise.toml](.mise.toml)で定義された必要なツールが自動的にインストールされます：

```bash
mise install

# インストール済みツールの確認
mise ls
```

## タスク管理

利用可能なタスクは[.mise-tasks/](.mise-tasks/)ディレクトリに定義されています。

```bash
# タスク一覧の確認
mise tasks
```

## クイックスタート

```bash
# 1. 依存ツールのインストール
mise install

# 2. ローカル環境の起動（クラスタ作成〜ポートフォワードまで一括実行）
mise run k8s:local:start
```
