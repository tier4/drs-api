# System Manager Client

reboot APIをテストするための最小限のクライアントです。

## ビルド

```bash
cd tools/client
export PATH=/usr/local/go/bin:$PATH

# ローカルアーキテクチャ用
make build

# ARM64用
make build-arm64

# 両方
make all

# または直接コマンド
mkdir -p bin && go build -o bin/client main.go                              # ローカル用
mkdir -p bin && GOOS=linux GOARCH=arm64 go build -o bin/client-arm64 main.go  # ARM64用
```

## 使用方法

### 基本的な使用法
```bash
# reboot（デフォルト）
./bin/client

# 特定のコマンドを実行
./bin/client -cmd=<command>

# 別のサーバーを指定
./bin/client -server=192.168.1.100:50051 -cmd=disk
```

### 利用可能なコマンド

#### システム制御
```bash
# 即座にreboot
./bin/client -cmd=reboot

# 5秒遅延してreboot
./bin/client -cmd=reboot -delay=5

# 即座にshutdown
./bin/client -cmd=shutdown

# 10秒遅延してshutdown
./bin/client -cmd=shutdown -delay=10
```

#### DRSサービス管理
```bash
# DRSサービスを停止
./bin/client -cmd=drs-stop

# DRSサービスを再起動
./bin/client -cmd=drs-restart

# DRSサービスの状態確認
./bin/client -cmd=drs-status
```

#### Recorderサービス管理
```bash
# Recorderサービスを停止
./bin/client -cmd=recorder-stop

# Recorderサービスを再起動
./bin/client -cmd=recorder-restart

# Recorderサービスの状態確認
./bin/client -cmd=recorder-status
```

#### ディスク使用量確認
```bash
# ルートディスクの使用量
./bin/client -cmd=disk

# 特定のパスの使用量
./bin/client -cmd=disk -path=/home
```

#### PTP時刻同期確認
```bash
# ローカルのPTP同期状態のみ確認
./bin/client -cmd=ptp

# ローカルとリモートデバイスのPTP同期状態を確認
./bin/client -cmd=ptp-all
```

## オプション

- `-server`: サーバーアドレス（デフォルト: localhost:50051）
- `-cmd`: 実行するコマンド（デフォルト: reboot）
  - システム制御: `reboot`, `shutdown`
  - DRSサービス: `drs-stop`, `drs-restart`, `drs-status`
  - Recorderサービス: `recorder-stop`, `recorder-restart`, `recorder-status`
  - ディスク: `disk`
  - PTP同期: `ptp` (ローカルのみ), `ptp-all` (ローカル＋リモート)
- `-delay`: reboot/shutdown前の遅延秒数（デフォルト: 0）
- `-path`: ディスク使用量確認のパス（デフォルト: /）

## 使用例

```bash
# ARM64環境のサーバーをテスト
./bin/client-arm64 -server=192.168.20.1:50051 -cmd=disk

# Lite版サーバーでsystemdサービステスト（エラーになる）
./bin/client -server=localhost:50053 -cmd=drs-status

# フル版サーバーで全機能テスト
./bin/client -server=localhost:50051 -cmd=drs-status

# PTP同期状態確認
./bin/client -cmd=ptp
./bin/client -server=192.168.1.100:50051 -cmd=ptp-all
```

## 注意

- reboot/shutdownコマンドは実際にシステムを操作します
- テスト環境でのみ使用してください
- systemdサービス管理はLite版では利用できません
- PTP同期確認には`pmc`コマンドが必要です
- リモートデバイスのPTP確認にはARPテーブルエントリが必要です