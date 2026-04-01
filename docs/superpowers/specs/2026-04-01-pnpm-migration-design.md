# pnpm 移行設計

**Date:** 2026-04-01
**Scope:** `ui/dashboard` パッケージの npm → pnpm 移行

---

## 概要

`ui/dashboard` の単一パッケージを npm から pnpm に移行する。
ロックファイル変換には `pnpm import` を使用し既存の依存ツリーを保持する。
pnpm バージョンは `package.json` の `packageManager` フィールドと Corepack で固定する。

---

## 変更対象ファイル

| ファイル | 変更内容 |
|---|---|
| `ui/dashboard/package.json` | `"packageManager"` フィールド追加 |
| `ui/dashboard/package-lock.json` | 削除 |
| `ui/dashboard/pnpm-lock.yaml` | 新規生成（`pnpm import` → `pnpm install`） |
| `ui/dashboard/Dockerfile` | pnpm 対応（Corepack 有効化、コマンド更新） |
| `docker/dashboard/Dockerfile` | 同上 |
| `.github/workflows/dashboard.yml` | Corepack 有効化、cache・コマンド更新 |
| `.github/dependabot.yml` | 変更不要 |

---

## 各ファイルの変更詳細

### `ui/dashboard/package.json`

`"packageManager"` フィールドを追加する。バージョンは実行時に `pnpm --version` で確認した最新の pnpm 10 安定版を使用する。

```json
{
  "packageManager": "pnpm@10.x.x"
}
```

### ロックファイル

1. `pnpm import` を実行し `package-lock.json` から `pnpm-lock.yaml` を生成する
2. `pnpm install` を実行してロックファイルを確定させる
3. `package-lock.json` を削除する

### Dockerfile（`ui/dashboard/Dockerfile` / `docker/dashboard/Dockerfile`）

ビルドステージのみ変更。本番ステージ（nginx）は無変更。

```diff
- COPY package*.json ./
- RUN npm ci
+ COPY package.json pnpm-lock.yaml ./
+ RUN corepack enable && pnpm install --frozen-lockfile

- RUN npm run build
+ RUN pnpm run build
```

`docker/dashboard/Dockerfile` の COPY パスは `ui/dashboard/` プレフィックス付き:

```diff
- COPY ui/dashboard/package*.json ./
+ COPY ui/dashboard/package.json ui/dashboard/pnpm-lock.yaml ./
```

### `.github/workflows/dashboard.yml`

各 job（lint / type-check / build / test）に同じパターンで変更を適用する。

```diff
  - name: Set up Node.js
    uses: actions/setup-node@v6
    with:
      node-version: '20'
-     cache: 'npm'
-     cache-dependency-path: ui/dashboard/package-lock.json
+     cache: 'pnpm'
+     cache-dependency-path: ui/dashboard/pnpm-lock.yaml

+ - name: Enable Corepack
+   run: corepack enable

  - name: Install dependencies
-   run: npm ci
+   run: pnpm install --frozen-lockfile
```

コマンドの置き換え:

| 変更前 | 変更後 |
|---|---|
| `npm run lint` | `pnpm run lint` |
| `npm run format:check` | `pnpm run format:check` |
| `npx tsc -b --noEmit` | `pnpm exec tsc -b --noEmit` |
| `npm run build` | `pnpm run build` |
| `npm audit --audit-level=high` | `pnpm audit --audit-level=high` |

### `.github/dependabot.yml`

変更不要。`package-ecosystem: "npm"` は `pnpm-lock.yaml` を自動検出する。

---

## 移行手順

1. pnpm をインストール: `corepack enable && corepack prepare pnpm@latest --activate`
2. `ui/dashboard/` に移動
3. `pnpm import` を実行（`pnpm-lock.yaml` 生成）
4. `pnpm install` を実行（ロックファイル確定）
5. ローカルで `pnpm run build` が通ることを確認
6. `package-lock.json` を削除
7. 各ファイルを上記の差分通りに更新
8. `pnpm run lint` / `pnpm run format:check` が通ることを確認
9. Docker ビルドが通ることを確認（任意）

---

## 非対応事項

- pnpm workspaces は使用しない（現在 JS パッケージは1つのみ）
- `.npmrc` は追加しない（互換性設定が不要なため）
