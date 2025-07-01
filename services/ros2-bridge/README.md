# ROS2 Bridge Service

ROS2ノードとgRPCサーバーを組み合わせたブリッジサービスです。

## 前提条件

### ROS2環境
```bash
# ROS2 Humbleがインストールされていること
source /opt/ros/humble/setup.bash
```

### gRPC C++ライブラリ
```bash
sudo apt-get install -y libgrpc++-dev protobuf-compiler-grpc
```

## ビルド手順

### 1. Proto定義からC++コードを生成
```bash
cd ../../scripts
./generate-proto.sh  # GoとC++の両方を生成
```

### 2. ROS2パッケージとしてビルド
```bash
cd ../services/ros2-bridge

# ROS2環境をセットアップ
source /opt/ros/humble/setup.bash

# colconでビルド
colcon build --packages-select ros2_bridge

# セットアップスクリプトをsource
source install/setup.bash
```

### 3. 実行
```bash
# ROS2ノードとして実行
ros2 run ros2_bridge ros2_bridge_node

# または直接実行
./install/ros2_bridge/lib/ros2_bridge/ros2_bridge_node
```

## 設定

### gRPCポート変更
```bash
ros2 run ros2_bridge ros2_bridge_node --ros-args -p grpc_port:=50052
```

## API仕様

gRPCサービスは`proto/ros2bridge/v1/bridge.proto`で定義されています。

### 主要機能
- ROS2 topicの購読とgRPC経由での値取得
- gRPC経由でのROS2 service呼び出し
- リアルタイムでのtopic streaming

## トラブルシューティング

### 1. gRPCライブラリが見つからない
```bash
sudo apt-get install -y libgrpc++-dev
```

### 2. protoファイルが生成されていない
```bash
cd ../../scripts
./generate-proto-go.sh
```

### 3. ROS2環境が見つからない
```bash
source /opt/ros/humble/setup.bash
```