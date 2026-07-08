#ifndef ROS2_BRIDGE_SENSING_HANDLER_HPP_
#define ROS2_BRIDGE_SENSING_HANDLER_HPP_

#include "drs/ros2bridge/v1/sensing_service.grpc.pb.h"

#include <rclcpp/rclcpp.hpp>

#include <sensor_msgs/msg/compressed_image.hpp>
#include <sensor_msgs/msg/nav_sat_fix.hpp>

#include <chrono>
#include <memory>
#include <mutex>
#include <string>
#include <unordered_map>

namespace ros2_bridge
{

class SensingHandler : public drs::ros2bridge::v1::SensingService::Service
{
public:
  explicit SensingHandler(rclcpp::Node::SharedPtr node);
  ~SensingHandler() override = default;

  // gRPC service methods
  grpc::Status GetPosition(
    grpc::ServerContext * context, const drs::ros2bridge::v1::GetPositionRequest * request,
    drs::ros2bridge::v1::GetPositionResponse * response) override;

  grpc::Status ListNodes(
    grpc::ServerContext * context, const drs::ros2bridge::v1::ListNodesRequest * request,
    drs::ros2bridge::v1::ListNodesResponse * response) override;

  grpc::Status GetCameraPreview(
    grpc::ServerContext * context, const drs::ros2bridge::v1::GetCameraPreviewRequest * request,
    drs::ros2bridge::v1::GetCameraPreviewResponse * response) override;

private:
  // ROS2 callback for NavSatFix messages
  void navSatFixCallback(const sensor_msgs::msg::NavSatFix::SharedPtr msg);

  // Convert ROS2 NavSatFix to protobuf Position
  void convertNavSatFixToPosition(
    const sensor_msgs::msg::NavSatFix & nav_sat_fix, drs::ros2bridge::v1::Position * position);

  // Filter nodes based on filter string
  bool matchesFilter(const std::string & node_name, const std::string & filter);

  // ROS2 callback for camera CompressedImage messages, shared across all
  // lazily-subscribed camera topics
  void cameraImageCallback(
    const std::string & topic_name, const sensor_msgs::msg::CompressedImage::SharedPtr msg);

  // Periodic sweep that unsubscribes camera topics with no GetCameraPreview
  // request in the last kCameraIdleTimeout.
  void sweepIdleCameraSubscriptions();

  rclcpp::Node::SharedPtr node_;
  rclcpp::Subscription<sensor_msgs::msg::NavSatFix>::SharedPtr nav_sat_fix_sub_;

  // Cached position data
  std::mutex position_mutex_;
  std::shared_ptr<sensor_msgs::msg::NavSatFix> cached_position_;
  std::chrono::steady_clock::time_point last_position_update_;

  // Lazily-created per-topic camera subscriptions (created on first
  // GetCameraPreview request for that topic_name), the latest cached
  // compressed frame, and the last time each topic was requested. A topic
  // with no request for kCameraIdleTimeout is unsubscribed by
  // sweepIdleCameraSubscriptions() so an ECU that relays another ECU's
  // camera over the network doesn't keep paying for that stream forever
  // after the dashboard stops looking at it.
  static constexpr std::chrono::seconds kCameraIdleTimeout{30};
  static constexpr std::chrono::seconds kCameraIdleSweepInterval{5};

  std::mutex camera_mutex_;
  std::unordered_map<
    std::string, rclcpp::Subscription<sensor_msgs::msg::CompressedImage>::SharedPtr>
    camera_subs_;
  std::unordered_map<std::string, sensor_msgs::msg::CompressedImage::SharedPtr> cached_frames_;
  std::unordered_map<std::string, std::chrono::steady_clock::time_point> last_camera_request_;
  rclcpp::TimerBase::SharedPtr camera_idle_timer_;
};

}  // namespace ros2_bridge

#endif  // ROS2_BRIDGE_SENSING_HANDLER_HPP_
