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

## 認証基盤（Keycloak）

認証アーキテクチャの概要は[README.md](../README.md#認証アーキテクチャ)を参照してください。

### K8sリソース構成

- `keycloak/deployment.yaml`: Keycloak本体。`--import-realm`で起動時にrealm設定を自動インポート
- `keycloak/realm-configmap.yaml`: `famoney` realmと`famoney-bff`クライアントのJSON定義
- `keycloak/secret.yaml`: 管理コンソールのクレデンシャル
- `postgres/cluster.yaml`: `keycloak_db`を`postInitSQL`で作成

### ローカルでのKeycloak管理コンソール

```bash
mise run k8s:local:forward-keycloak
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
