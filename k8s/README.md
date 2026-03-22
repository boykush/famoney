# Kubernetes Manifests

このディレクトリには、Famoneyアプリケーションのkubernetesマニフェストが含まれています。
Kustomizeを使用した階層構造で管理しており、環境ごとのオーバーレイで設定を上書きできます。

## ディレクトリ構成

```
k8s/
├── base/                      # ベースとなるマニフェスト
│   ├── platform/              # プラットフォーム層（ArgoCD, Istio等）
│   │   ├── argocd/
│   │   ├── istio-base/
│   │   ├── istio-ingressgateway/
│   │   └── istiod/
│   └── workloads/             # アプリケーション層
│       ├── bff/
│       ├── expense/
│       ├── web/
│       ├── keycloak/          # 認証基盤（OIDC IdP）
│       ├── postgres/
│       └── istio/             # Istio設定（Gateway, VirtualService等）
└── overlays/                  # 環境別のオーバーレイ
    └── local/                 # ローカル開発環境
        ├── platform/
        └── workloads/
```

## 構成の考え方

### Platform vs Workloads

- **platform/**: インフラストラクチャ層
  - ArgoCD: GitOpsデプロイメントツール
  - Istio: サービスメッシュ（istio-base, istiod, ingressgateway）

- **workloads/**: アプリケーション層
  - マイクロサービス（bff, expense, web）
  - 認証基盤（keycloak）
  - データベース（postgres - CloudNativePG）
  - Istio設定（Gateway, VirtualService, AuthorizationPolicy等）

この分離により、プラットフォーム層とアプリケーション層を独立して管理できます。

## 認証アーキテクチャ

Keycloakをセルフホスト型のOIDC Identity Providerとして使用しています。

```
ブラウザ → Istio IngressGateway → BFF (go-oidc トークン検証) → expense サービス
                                    ↕
                                Keycloak (OIDC IdP)
                                    ↕
                                PostgreSQL (keycloak_db)
```

### 構成要素

- **Keycloak**: OIDC IdP。`famoney` realmと`famoney-bff`クライアントはConfigMap経由で起動時に自動インポート
- **BFF**: [go-oidc](https://github.com/coreos/go-oidc)ミドルウェアでBearerトークンを検証。Keycloak未起動時はログ出力のみで起動し、認証エンドポイントは503を返す
- **ユーザー管理**: Keycloak管理コンソールで手動登録

### エンドポイントの認証

| パス | 認証 |
|------|------|
| `*/health` | 不要 |
| その他 | Bearerトークン必須（401/503） |

### ローカルでのKeycloak管理コンソール

```bash
# ポートフォワード
kubectl port-forward svc/keycloak 8180:8080

# http://localhost:8180 でアクセス
# ユーザー: admin / パスワード: keycloak-admin-secret参照
```

### Base vs Overlays

- **base/**: 環境に依存しない共通のマニフェスト
- **overlays/**: 環境ごとの設定（現在はlocalのみ）

## CD パイプライン

mainブランチへのpush時に、変更があったサービスのDockerイメージをGHCR（GitHub Container Registry）にビルド・プッシュし、kustomize imagesでbase kustomization.yamlのイメージタグを自動更新します。ArgoCD自動syncでデプロイまで完結します。

## ローカル開発

### 一括セットアップ
```bash
# クラスタ作成、プラットフォームデプロイ、ポートフォワードまで一括実行
mise run k8s:local:start
```

### 個別操作
```bash
# クラスタの作成
mise run k8s:local:cluster:create

# プラットフォームリソースのデプロイ（ArgoCD, Istio等）
mise run k8s:local:deploy-platform

# ポートフォワード
mise run k8s:local:forward

# クラスタの削除
mise run k8s:local:cluster:delete
```

workloadsのデプロイはArgoCD自動syncで行われるため、手動デプロイは不要です。

## 利用可能なコマンド
```bash
mise tasks | grep k8s
```
