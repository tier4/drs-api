# System Manager Service

システム制御のためのgRPCサービスです。設定ファイルによってカスタマイズ可能です。

## 設定ファイル

### 設定ファイルの場所
以下の順序で設定ファイルを検索します：
1. `./config.yaml`
2. `./config.yml`
3. `実行ファイルと同じディレクトリ/config.yaml`
4. `実行ファイルと同じディレクトリ/config.yml`
5. `/etc/system-manager/config.yaml`
6. `/etc/system-manager/config.yml`

### 設定例（config.yaml）
```yaml
server:
  port: 50051
  mode: "full"  # "full" or "lite"

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
./bin/system-manager
```

### 設定ファイルを指定して実行
```bash
./bin/system-manager -config=custom-config.yaml
```

### ポートを指定して実行（設定ファイルより優先）
```bash
./bin/system-manager -port=50052
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