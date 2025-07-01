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
go build -o client main.go                              # ローカル用
GOOS=linux GOARCH=arm64 go build -o client-arm64 main.go  # ARM64用
```

## 使用方法

### 基本的な使用法
```bash
# reboot（デフォルト）
./client

# 特定のコマンドを実行
./client -cmd=<command>

# 別のサーバーを指定
./client -server=192.168.1.100:50051 -cmd=disk
```

### 利用可能なコマンド

#### システム制御
```bash
# 即座にreboot
./client -cmd=reboot

# 5秒遅延してreboot
./client -cmd=reboot -delay=5

# 即座にshutdown
./client -cmd=shutdown

# 10秒遅延してshutdown
./client -cmd=shutdown -delay=10
```

#### DRSサービス管理
```bash
# DRSサービスを停止
./client -cmd=drs-stop

# DRSサービスを再起動
./client -cmd=drs-restart

# DRSサービスの状態確認
./client -cmd=drs-status
```

#### Recorderサービス管理
```bash
# Recorderサービスを停止
./client -cmd=recorder-stop

# Recorderサービスを再起動
./client -cmd=recorder-restart

# Recorderサービスの状態確認
./client -cmd=recorder-status
```

#### ディスク使用量確認
```bash
# ルートディスクの使用量
./client -cmd=disk

# 特定のパスの使用量
./client -cmd=disk -path=/home
```

#### PTP時刻同期確認
```bash
# ローカルのPTP同期状態のみ確認
./client -cmd=ptp

# ローカルとリモートデバイスのPTP同期状態を確認
./client -cmd=ptp-all
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
./client-arm64 -server=192.168.20.1:50051 -cmd=disk

# Lite版サーバーでsystemdサービステスト（エラーになる）
./client -server=localhost:50053 -cmd=drs-status

# フル版サーバーで全機能テスト
./client -server=localhost:50051 -cmd=drs-status

# PTP同期状態確認
./client -cmd=ptp
./client -server=192.168.1.100:50051 -cmd=ptp-all
```

## 注意

- reboot/shutdownコマンドは実際にシステムを操作します
- テスト環境でのみ使用してください
- systemdサービス管理はLite版では利用できません
- PTP同期確認には`pmc`コマンドが必要です
- リモートデバイスのPTP確認にはARPテーブルエントリが必要です