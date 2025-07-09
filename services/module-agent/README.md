# Module Agent Service

DRSの各モジュール（Sensing Module、Storage Moduleなど）で動作するシステム制御用gRPCサービスです。
設定ファイルでAPIを個別に有効/無効化でき、各モジュールの要件に応じた構成が可能です。

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

disk:
  monitored_paths:
    - path: "/"
      name: "root"
      description: "Root filesystem"
    - path: "/home"
      name: "home"
      description: "Home directory"
    - path: "/var/log"
      name: "logs"
      description: "Log directory"
  default_path: "/"

services:
  enable_systemd_manage: true
  allowed_services:
    - "drs.service"
    - "recorder.service"

system:
  max_delay_seconds: 300
  allow_reboot: true
  allow_shutdown: true
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

## ディスク使用量API

設定ファイルで定義されたパスの監視が可能です：

### 名前による指定
```bash
# "home"という名前で設定されたパス（/home）の使用量
./client -cmd=disk -path=home

# "logs"という名前で設定されたパス（/var/log）の使用量  
./client -cmd=disk -path=logs
```

### 直接パス指定
```bash
# 直接パスを指定
./client -cmd=disk -path=/tmp
```

### デフォルトパス
```bash
# パス未指定時はdefault_pathを使用
./client -cmd=disk
```

## モジュール別設定例

### Sensing Module用設定
```yaml
apis:
  enable_reboot: true
  enable_shutdown: true
  enable_service_management: true  # レコーディングサービス管理用
  enable_disk_usage: true          # ストレージ監視用
  enable_ptp_check: true           # 時刻同期確認用
```

### Storage Module用設定
```yaml
apis:
  enable_reboot: true
  enable_shutdown: true
  enable_service_management: false # ストレージモジュールでは不要
  enable_disk_usage: true          # メイン機能
  enable_ptp_check: false          # ストレージモジュールでは不要
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