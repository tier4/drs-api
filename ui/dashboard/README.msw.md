# MSW (Mock Service Worker) セットアップガイド

このプロジェクトでは、MSWを使用してDRS APIのモックサーバーを実装しています。

## 🚀 使い方

### モックAPIを使って開発サーバーを起動

```bash
npm run dev:mock
```

または、`.env.development`で`VITE_ENABLE_MOCKS=true`に設定してから：

```bash
npm run dev
```

### 実際のAPIを使って開発サーバーを起動

```bash
npm run dev:real
```

または、`.env.development`で`VITE_ENABLE_MOCKS=false`に設定してから：

```bash
npm run dev
```

## 📁 ファイル構成

```
src/mocks/
├── browser.ts      # MSWのブラウザ統合設定
├── handlers.ts     # APIエンドポイントのモックハンドラー
└── data.ts         # モックデータとステート管理
```

## 🔧 モックデータのカスタマイズ

### モジュールデータの変更

[src/mocks/data.ts](src/mocks/data.ts) の `mockModules` を編集：

```typescript
export const mockModules: ModuleStatus[] = [
  {
    hostname: 'ecu0',
    address: '192.168.20.1:50051',
    status: 'OK',  // 'OK' | 'WARN' | 'ERROR'
    // ... その他のプロパティ
  },
  // 新しいモジュールを追加可能
]
```

### APIレスポンスの遅延調整

[src/mocks/handlers.ts](src/mocks/handlers.ts) で `delay()` を変更：

```typescript
http.get(`${API_BASE}/modules`, async () => {
  await delay(300) // ミリ秒単位で調整
  return HttpResponse.json({ modules: mockModules })
})
```

## 🎯 実装済みエンドポイント

### モジュール管理
- `GET /api/v1/modules` - 全モジュール取得
- `GET /api/v1/modules/:hostname` - 単一モジュール取得

### レコーディング制御
- `GET /api/v1/recording/status` - レコーディングステータス取得
- `POST /api/v1/recording/start` - レコーディング開始
- `POST /api/v1/recording/stop` - レコーディング停止
- `POST /api/v1/recording/pause` - レコーディング一時停止
- `POST /api/v1/recording/resume` - レコーディング再開

### PTP時刻同期
- `GET /api/v1/ptp/status` - PTPステータス取得

### ROS2トピック監視
- `GET /api/v1/modules/:hostname/topics/status` - トピックステータス取得

### サービス管理
- `GET /api/v1/modules/:hostname/services` - サービス一覧取得
- `POST /api/v1/modules/:hostname/services/:service_name/start` - サービス開始
- `POST /api/v1/modules/:hostname/services/:service_name/stop` - サービス停止
- `POST /api/v1/modules/:hostname/services/:service_name/restart` - サービス再起動
- `POST /api/v1/modules/:hostname/services/restart` - センサーサービス再起動

### システム制御
- `POST /api/v1/system/restart` - システム全体再起動
- `POST /api/v1/system/shutdown` - システム全体シャットダウン
- `POST /api/v1/modules/:hostname/restart` - モジュール再起動
- `POST /api/v1/modules/:hostname/shutdown` - モジュールシャットダウン

### ヘルスチェック
- `GET /health` - ヘルスチェック

## 🎨 モックの特徴

### 1. リアルタイムステート管理
レコーディングの開始/停止など、状態が変更されるとモックデータも自動更新されます。

### 2. ネットワーク遅延のシミュレーション
実際のネットワークリクエストのような遅延を再現しています。

### 3. エラーケースの再現
存在しないモジュールへのアクセスなど、エラーケースもモックしています。

### 4. コンソールログ
重要な操作（システム再起動など）はコンソールにログ出力されます。

## 🔄 本番APIへの切り替え

MSWは開発環境でのみ動作し、本番ビルドには影響しません。

### 一時的に無効化する場合：
ブラウザのコンソールで：
```javascript
// MSWを停止
await window.msw.worker.stop()

// MSWを再開
await window.msw.worker.start()
```

### 環境変数で制御：
```bash
# .env.development
VITE_ENABLE_MOCKS=false
```

## 📝 新しいエンドポイントの追加

1. [src/mocks/handlers.ts](src/mocks/handlers.ts) に新しいハンドラーを追加：

```typescript
http.get('/api/v1/new-endpoint', async () => {
  await delay(200)
  return HttpResponse.json({ data: 'response' })
})
```

2. 必要に応じて [src/mocks/data.ts](src/mocks/data.ts) にモックデータを追加

## 🐛 トラブルシューティング

### MSWが起動しない
1. `public/mockServiceWorker.js` が存在するか確認
2. 存在しない場合：`npx msw init public/ --save`

### APIリクエストがモックされない
1. ブラウザコンソールで `[MSW] Mocking enabled` メッセージを確認
2. `.env.development` で `VITE_ENABLE_MOCKS=true` になっているか確認
3. 開発サーバーを再起動

### TypeScriptエラー
モックデータの型が [src/services/api.ts](src/services/api.ts) のインターフェースと一致しているか確認

## 📚 参考資料

- [MSW公式ドキュメント](https://mswjs.io/)
- [MSWクイックスタート](https://mswjs.io/docs/quick-start)
- [DRS API仕様](../../services/api-gateway/README.md)
