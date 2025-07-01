# Proto API - gRPC Microservices

このプロジェクトは、gRPCを使用してマシンを操作するマイクロサービスのAPIを提供します。

## サービス構成

### 1. System Manager (基本サービス)
- **言語**: Go
- **ポート**: 50051
- **機能**:
  - rebootとshutdownの実行
  - drs.serviceの管理（stop/restart/status）
  - recorder.serviceの管理（stop/restart/status）
  - ディスク容量の取得

### 2. System Manager Lite (制限版)
- **言語**: Go
- **ポート**: 50053
- **機能**:
  - rebootとshutdownの実行
  - ディスク容量の取得

### 3. ROS2 Bridge Service
- **言語**: C++
- **ポート**: 50052
- **機能**:
  - ROS2 topicのsubscribeとgRPC経由での値の取得
  - gRPC経由でのROS2 service呼び出し

## ディレクトリ構成

```
proto_api/
├── proto/                     # gRPC定義ファイル
├── services/                  # マイクロサービス群
│   ├── system-manager/        # 基本サービス（フル版・Lite版共用）
│   └── ros2-bridge/          # ROS2ブリッジ
├── scripts/                   # ビルド・実行スクリプト
└── docs/                      # ドキュメント
```

## セットアップ

### 1. 必要なツールのインストール

```bash
# protobufコンパイラのインストール
sudo apt-get install -y protobuf-compiler

# gRPC C++プラグインのインストール（ROS2ブリッジ用）
sudo apt-get install -y libgrpc++-dev protobuf-compiler-grpc

# Go用プラグインのインストール
cd scripts
./install-protoc-plugins.sh
```

### 注意事項
- System Managerサービスは**ネイティブ実行**を前提としています（shutdown/reboot/systemd操作のため）
- ARM64環境でのデプロイにはクロスコンパイルしたバイナリを使用してください
- 古いglibc環境では静的リンク版（`*-static`）を使用してください

### 2. Proto定義からコード生成

```bash
cd scripts
./generate-proto.sh
```

## ビルドと実行

### System Manager (Go)
静的リンク版は古いglibc環境でも動作します。
```bash
cd services/system-manager

# フル版のビルド（ローカルアーキテクチャ）
make build

# Lite版のビルド（ローカルアーキテクチャ）
make build-lite

# ARM64版のビルド
make build-arm64
make build-lite-arm64

# 全バージョンビルド
make all

# 実行（フル版）
./bin/system-manager -port=50051

# 実行（Lite版）
./bin/system-manager-lite -port=50053

# ARM64環境での実行
./bin/system-manager-arm64 -port=50051
./bin/system-manager-lite-arm64 -port=50053
```

#### ROS2 Bridge (C++)
```bash
cd services/ros2-bridge
colcon build
source install/setup.bash
ros2 run ros2_bridge ros2_bridge_node
```

## API仕様

各サービスのAPI仕様は`proto/`ディレクトリ内の`.proto`ファイルを参照してください。

## 開発者向け情報

### ビルドタグによる機能切り替え

System Managerは、ビルドタグまたは環境変数で機能を切り替えられます：

- ビルドタグ: `go build -tags lite`でLite版をビルド
- 環境変数: `SERVICE_MODE=lite`でLite版として動作