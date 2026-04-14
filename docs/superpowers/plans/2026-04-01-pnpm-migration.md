# pnpm 移行 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `ui/dashboard` パッケージを npm から pnpm に移行し、CI・Docker も含めてすべて pnpm で動作させる。

**Architecture:** `pnpm import` で既存の `package-lock.json` を `pnpm-lock.yaml` に変換して依存ツリーを保持する。pnpm バージョンは `packageManager` フィールドと Corepack で固定する。Dependabot は変更不要（`package-ecosystem: "npm"` が `pnpm-lock.yaml` を自動検出する）。

**Tech Stack:** pnpm 10, Corepack (Node.js 組み込み), node:22-alpine (Docker)

---

## ファイル変更マップ

| ファイル | 操作 |
|---|---|
| `ui/dashboard/pnpm-lock.yaml` | 新規生成 |
| `ui/dashboard/package.json` | `packageManager` フィールド追加 |
| `ui/dashboard/package-lock.json` | 削除 |
| `ui/dashboard/Dockerfile` | pnpm 対応に更新 |
| `docker/dashboard/Dockerfile` | pnpm 対応に更新 |
| `ui/dashboard/pnpm-workspace.yaml` | 新規作成（`minimumReleaseAge` 設定） |
| `.github/workflows/dashboard.yml` | Corepack 有効化・コマンド更新 |

---

## Task 1: pnpm をセットアップしてロックファイルを生成する

**Files:**
- Create: `ui/dashboard/pnpm-lock.yaml`

- [ ] **Step 1: Corepack を有効化して pnpm の最新版を準備する**

```bash
corepack enable
corepack prepare pnpm@latest --activate
```

`sudo` が必要な場合は `sudo corepack enable`。

Expected output (例):
```
Preparing pnpm@10.12.4 for immediate activation...
```

- [ ] **Step 2: pnpm のバージョンを確認する**

```bash
pnpm --version
```

Expected output (例): `10.12.4`  
このバージョン番号を次の Task 2 で使用する。

- [ ] **Step 3: `pnpm import` で pnpm-lock.yaml を生成する**

`ui/dashboard/` ディレクトリで実行する。

```bash
cd ui/dashboard
pnpm import
```

Expected output:
```
 WARN  1 issue
Lockfile was created by pnpm from a npm lockfile. ...
```
警告は正常。`pnpm-lock.yaml` が生成されていることを確認:

```bash
ls pnpm-lock.yaml
```

Expected: `pnpm-lock.yaml`

- [ ] **Step 4: `pnpm install` を実行して node_modules を生成する**

```bash
pnpm install
```

Expected output (例):
```
Packages: +NNN
Progress: resolved NNN, reused NNN, downloaded 0, added NNN, done
```

- [ ] **Step 5: ビルドが通ることを確認する**

```bash
pnpm run build
```

Expected: エラーなくビルドが完了し `dist/` ディレクトリが生成される。

- [ ] **Step 6: pnpm-lock.yaml をコミットする**

```bash
git add ui/dashboard/pnpm-lock.yaml
git commit -m "chore: add pnpm-lock.yaml generated from package-lock.json"
```

---

## Task 2: package.json に packageManager フィールドを追加する

**Files:**
- Modify: `ui/dashboard/package.json`

- [ ] **Step 1: packageManager フィールドを追加する**

Task 1 Step 2 で確認したバージョンを使って以下のコマンドを実行する（`ui/dashboard/` で実行）。

```bash
pnpm pkg set packageManager="pnpm@$(pnpm --version)"
```

Expected: コマンドがエラーなく完了する。

- [ ] **Step 2: package.json の内容を確認する**

```bash
grep packageManager package.json
```

Expected output (例):
```
  "packageManager": "pnpm@10.12.4",
```

- [ ] **Step 3: コミットする**

```bash
git add ui/dashboard/package.json
git commit -m "chore: pin pnpm version via packageManager field"
```

---

## Task 3: package-lock.json を削除する

**Files:**
- Delete: `ui/dashboard/package-lock.json`

- [ ] **Step 1: package-lock.json を削除する**

```bash
# ui/dashboard/ で実行
rm package-lock.json
```

- [ ] **Step 2: pnpm install が引き続き動作することを確認する**

```bash
pnpm install --frozen-lockfile
```

Expected output: エラーなし。`frozen-lockfile` フラグで pnpm-lock.yaml が正しいことも検証できる。

- [ ] **Step 3: コミットする**

```bash
git add ui/dashboard/package-lock.json
git commit -m "chore: remove package-lock.json (replaced by pnpm-lock.yaml)"
```

---

## Task 4: ui/dashboard/Dockerfile を更新する

**Files:**
- Modify: `ui/dashboard/Dockerfile`

- [ ] **Step 1: Dockerfile をまるごと以下の内容に置き換える**

```dockerfile
# Multi-stage build
FROM node:22-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY package.json pnpm-lock.yaml ./

# Install dependencies
RUN corepack enable && pnpm install --frozen-lockfile

# Copy source code
COPY . .

# Accept build-time variables
ARG VITE_API_BASE_URL=/api/v1

# Build the application
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
RUN pnpm run build

# Production stage
FROM nginx:alpine

# Copy built assets from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Expose port 3000
EXPOSE 3000

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
```

- [ ] **Step 2: Docker ビルドが通ることを確認する**

```bash
# ui/dashboard/ で実行
docker build -t dashboard-test .
```

Expected: `Successfully built` で終わること。Docker が使えない環境ではスキップ可。

- [ ] **Step 3: コミットする**

```bash
git add ui/dashboard/Dockerfile
git commit -m "chore: migrate ui/dashboard/Dockerfile to pnpm"
```

---

## Task 5: docker/dashboard/Dockerfile を更新する

**Files:**
- Modify: `docker/dashboard/Dockerfile`

- [ ] **Step 1: Dockerfile をまるごと以下の内容に置き換える**

```dockerfile
# Multi-stage build
FROM node:22-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files from ui/dashboard
COPY ui/dashboard/package.json ui/dashboard/pnpm-lock.yaml ./

# Install dependencies
RUN corepack enable && pnpm install --frozen-lockfile

# Copy source code from ui/dashboard
COPY ui/dashboard/ ./

# Accept build-time variables
ARG VITE_API_BASE_URL=/api/v1

# Build the application
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
RUN pnpm run build

# Production stage
FROM nginx:alpine

# Copy built assets from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration
COPY docker/dashboard/nginx.conf /etc/nginx/nginx.conf

# Expose port 3000
EXPOSE 3000

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
```

- [ ] **Step 2: Docker ビルドが通ることを確認する**

```bash
# リポジトリルートで実行
docker build -t dashboard-test-root -f docker/dashboard/Dockerfile .
```

Expected: `Successfully built` で終わること。Docker が使えない環境ではスキップ可。

- [ ] **Step 3: コミットする**

```bash
git add docker/dashboard/Dockerfile
git commit -m "chore: migrate docker/dashboard/Dockerfile to pnpm"
```

---

## Task 6: GitHub Actions ワークフローを更新する

**Files:**
- Modify: `.github/workflows/dashboard.yml`

- [ ] **Step 1: dashboard.yml をまるごと以下の内容に置き換える**

```yaml
name: Dashboard CI

on:
  push:
    branches: [ main, develop ]
    paths:
      - 'ui/dashboard/**'
      - '.github/workflows/dashboard.yml'
  pull_request:
    branches: [ main, develop ]
    paths:
      - 'ui/dashboard/**'
      - '.github/workflows/dashboard.yml'

defaults:
  run:
    working-directory: ui/dashboard

jobs:
  lint:
    name: Lint Dashboard
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Node.js
        uses: actions/setup-node@v6
        with:
          node-version: '20'
          cache: 'pnpm'
          cache-dependency-path: ui/dashboard/pnpm-lock.yaml

      - name: Enable Corepack
        run: corepack enable

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      - name: Run ESLint
        run: pnpm run lint

      - name: Check code formatting
        run: pnpm run format:check

  type-check:
    name: TypeScript Type Check
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Node.js
        uses: actions/setup-node@v6
        with:
          node-version: '20'
          cache: 'pnpm'
          cache-dependency-path: ui/dashboard/pnpm-lock.yaml

      - name: Enable Corepack
        run: corepack enable

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      - name: Run TypeScript compiler
        run: pnpm exec tsc -b --noEmit

  build:
    name: Build Dashboard
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Node.js
        uses: actions/setup-node@v6
        with:
          node-version: '20'
          cache: 'pnpm'
          cache-dependency-path: ui/dashboard/pnpm-lock.yaml

      - name: Enable Corepack
        run: corepack enable

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      - name: Build production bundle
        run: pnpm run build
        env:
          VITE_ENABLE_MOCKS: false

      - name: Upload build artifacts
        uses: actions/upload-artifact@v4
        with:
          name: dashboard-dist
          path: ui/dashboard/dist/
          retention-days: 7

  test:
    name: Test Dashboard
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Node.js
        uses: actions/setup-node@v6
        with:
          node-version: '20'
          cache: 'pnpm'
          cache-dependency-path: ui/dashboard/pnpm-lock.yaml

      - name: Enable Corepack
        run: corepack enable

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      # Add test script when available
      # - name: Run tests
      #   run: pnpm test

      - name: Check for security vulnerabilities
        run: pnpm audit --audit-level=high
        continue-on-error: true
```

- [ ] **Step 2: YAML 構文エラーがないことを確認する**

```bash
# リポジトリルートで実行
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/dashboard.yml'))" && echo "YAML OK"
```

Expected: `YAML OK`

- [ ] **Step 3: コミットする**

```bash
git add .github/workflows/dashboard.yml
git commit -m "chore: migrate GitHub Actions CI to pnpm"
```

---

## Task 7: 最終確認

- [ ] **Step 1: ローカルで全コマンドが通ることを確認する**

```bash
# ui/dashboard/ で実行
pnpm install --frozen-lockfile
pnpm run lint
pnpm run format:check
pnpm exec tsc -b --noEmit
pnpm run build
```

Expected: すべてエラーなし。

- [ ] **Step 2: 変更ファイルを確認する**

```bash
git log --oneline -6
```

Expected (例):
```
xxxxxxx chore: migrate GitHub Actions CI to pnpm
xxxxxxx chore: migrate docker/dashboard/Dockerfile to pnpm
xxxxxxx chore: migrate ui/dashboard/Dockerfile to pnpm
xxxxxxx chore: remove package-lock.json (replaced by pnpm-lock.yaml)
xxxxxxx chore: pin pnpm version via packageManager field
xxxxxxx chore: add pnpm-lock.yaml generated from package-lock.json
```

- [ ] **Step 3: リポジトリに残存する npm 参照がないことを確認する**

```bash
# リポジトリルートで実行
grep -r "npm ci\|npm run\|npm install\|package-lock" \
  ui/dashboard/Dockerfile \
  docker/dashboard/Dockerfile \
  .github/workflows/dashboard.yml
```

Expected: 何も出力されないこと（マッチなし）。
