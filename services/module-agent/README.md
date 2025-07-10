# Module Agent Service

DRSの各モジュール（Sensing Module、Storage Moduleなど）で動作するシステム制御用gRPCサービスです。
設定ファイルでAPIを個別に有効/無効化でき、各モジュールの要件に応じた構成が可能です。

## 機能

### システム制御API
- **Reboot**: システムの再起動（遅延設定可能）
- **Shutdown**: システムのシャットダウン（遅延設定可能）

### サービス管理API
- **GetService**: 特定サービスの情報取得
- **ListServices**: 設定済みサービスの一覧取得
- **StartService**: サービスの起動
- **StopService**: サービスの停止
- **RestartService**: サービスの再起動
- **EnableService**: サービスの自動起動有効化
- **DisableService**: サービスの自動起動無効化

### リソース監視API
- **GetDiskUsage**: ディスク使用状況の取得（単一パス）
- **CheckPTPSync**: PTP（Precision Time Protocol）同期状態の確認

## 設定ファイル

### 設定ファイルの場所
以下の順序で設定ファイルを検索します：
1. `./config.yaml`
2. `./config.yml`
3. `実行ファイルと同じディレクトリ/config.yaml`
4. `実行ファイルと同じディレクトリ/config.yml`
5. `/etc/module-agent/config.yaml`
6. `/etc/module-agent/config.yml`

### 設定例（config.yaml）
```yaml
server:
  port: 50051

# APIごとの有効/無効設定
apis:
  enable_reboot: true           # 再起動API
  enable_shutdown: true         # シャットダウンAPI
  enable_service_management: true  # サービス管理API
  enable_disk_usage: true       # ディスク使用率API
  enable_ptp_check: true        # PTP同期確認API

# ディスク監視設定（ECUごとに単一パス）
disk:
  monitor_path: "/"             # 監視対象パス

# サービス管理設定
services:
  enable_systemd_manage: true
  services:
    drs_sensor:
      systemd_name: "drs-sensor.service"
      description: "DRS Sensor Service"
    drs_recorder:
      systemd_name: "drs-recorder.service"
      description: "DRS Recorder Service"

# システム設定
system:
  max_delay_seconds: 300
  allow_reboot: true
  allow_shutdown: true

# PTP同期設定
ptp:
  sync_threshold_ns: 1000000    # 同期閾値（ナノ秒）
  remote_devices:               # リモートデバイスのPTP同期確認
    - name: "sensor1"
      address: "192.168.1.101:50051"
    - name: "sensor2"
      address: "192.168.1.102:50051"
```

## 実行方法

### デフォルト設定で実行
```bash
./bin/module-agent
```

### 設定ファイルを指定して実行
```bash
./bin/module-agent -config=custom-config.yaml
```

### ポートを指定して実行（設定ファイルより優先）
```bash
./bin/module-agent -port=50052
```

## API使用例

### ディスク使用量取得
```bash
# 設定ファイルで定義されたパスの使用量を取得
./tools/client -cmd=disk
```

### サービス管理
```bash
# サービス一覧の取得
./tools/client -cmd=list-services

# サービス情報の取得（リソース名形式）
./tools/client -cmd=service -name=services/drs_sensor -action=status

# サービスの起動/停止/再起動
./tools/client -cmd=service -name=services/drs_recorder -action=start
./tools/client -cmd=service -name=services/drs_recorder -action=stop
./tools/client -cmd=service -name=services/drs_recorder -action=restart

# サービスの自動起動設定
./tools/client -cmd=service -name=services/drs_sensor -action=enable
./tools/client -cmd=service -name=services/drs_sensor -action=disable
```

### PTP同期確認
```bash
# ローカルのPTP同期状態確認
./tools/client -cmd=ptp

# すべてのデバイス（ローカル＋リモート）のPTP同期状態確認
./tools/client -cmd=ptp-all
```

### システム制御
```bash
# 即座に再起動
./tools/client -cmd=reboot

# 60秒後に再起動
./tools/client -cmd=reboot -delay=60

# 即座にシャットダウン
./tools/client -cmd=shutdown

# 30秒後にシャットダウン
./tools/client -cmd=shutdown -delay=30
```

## モジュール別設定例

### Sensing Module用設定
```yaml
server:
  port: 50051

apis:
  enable_reboot: true
  enable_shutdown: true
  enable_service_management: true  # レコーディングサービス管理用
  enable_disk_usage: true          # ストレージ監視用
  enable_ptp_check: true           # 時刻同期確認用

disk:
  monitor_path: "/data"            # センサーデータ保存領域

services:
  enable_systemd_manage: true
  services:
    drs_sensor:
      systemd_name: "drs-sensor.service"
      description: "DRS Sensor Service"
    drs_recorder:
      systemd_name: "drs-recorder.service"
      description: "DRS Recorder Service"

ptp:
  sync_threshold_ns: 1000000       # 1ms以内の同期を要求
```

### Storage Module用設定
```yaml
server:
  port: 50051

apis:
  enable_reboot: true
  enable_shutdown: true
  enable_service_management: false # ストレージモジュールでは不要
  enable_disk_usage: true          # メイン機能
  enable_ptp_check: false          # ストレージモジュールでは不要

disk:
  monitor_path: "/storage"         # データストレージ領域

system:
  max_delay_seconds: 300
  allow_reboot: true
  allow_shutdown: true
```

### Control Module用設定
```yaml
server:
  port: 50051

apis:
  enable_reboot: true
  enable_shutdown: true
  enable_service_management: true  # 全体システム管理用
  enable_disk_usage: true
  enable_ptp_check: true           # センサーモジュールとの同期確認

services:
  enable_systemd_manage: true
  services:
    drs_control:
      systemd_name: "drs-control.service"
      description: "DRS Control Service"

ptp:
  sync_threshold_ns: 1000000
  remote_devices:                  # センサーモジュールの同期状態監視
    - name: "sensor1"
      address: "192.168.1.101:50051"
    - name: "sensor2"
      address: "192.168.1.102:50051"
```

## ビルド

```bash
# 通常版
make build

# 静的リンク版（古いglibc環境用）
make build-static

# ARM64版
make build-arm64

# ARM64静的リンク版
make build-arm64-static
```