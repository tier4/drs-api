#ifndef ROS2_BRIDGE_SENSING_HANDLER_HPP_
#define ROS2_BRIDGE_SENSING_HANDLER_HPP_

#include "drs/ros2bridge/v1/sensing_service.grpc.pb.h"
#include "ros2_bridge/lazy_topic_cache.hpp"
#include "ros2_bridge/lidar_camera_projector.hpp"

#include <rclcpp/rclcpp.hpp>

#include <sensor_msgs/msg/camera_info.hpp>
#include <sensor_msgs/msg/compressed_image.hpp>
#include <sensor_msgs/msg/nav_sat_fix.hpp>
#include <sensor_msgs/msg/point_cloud2.hpp>

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

  grpc::Status GetLidarCameraProjectionPreview(
    grpc::ServerContext * context,
    const drs::ros2bridge::v1::GetLidarCameraProjectionPreviewRequest * request,
    drs::ros2bridge::v1::GetLidarCameraProjectionPreviewResponse * response) override;

private:
  // ROS2 callback for NavSatFix messages
  void navSatFixCallback(const sensor_msgs::msg::NavSatFix::SharedPtr msg);

  // Convert ROS2 NavSatFix to protobuf Position
  void convertNavSatFixToPosition(
    const sensor_msgs::msg::NavSatFix & nav_sat_fix, drs::ros2bridge::v1::Position * position);

  // Filter nodes based on filter string
  bool matchesFilter(const std::string & node_name, const std::string & filter);

  // Derives the decoded "_points" topic name from a requested "_packets"
  // topic name via suffix substitution (vendor-agnostic). Returns empty on
  // a topic_name that doesn't end with "_packets".
  static std::string derivePointsTopic(const std::string & packets_topic_name);

  // Extracts the LiDAR position segment (e.g. "front") from a topic name
  // shaped like "/sensing/lidar/{position}/{vendor}_packets". Returns empty
  // if topic_name doesn't match that shape.
  static std::string deriveLidarPosition(const std::string & packets_topic_name);

  // Periodic sweep that unsubscribes topics with no preview request in the
  // last kPreviewIdleTimeout, across all preview caches.
  void sweepIdlePreviewSubscriptions();

  rclcpp::Node::SharedPtr node_;
  rclcpp::Subscription<sensor_msgs::msg::NavSatFix>::SharedPtr nav_sat_fix_sub_;

  // Cached position data
  std::mutex position_mutex_;
  std::shared_ptr<sensor_msgs::msg::NavSatFix> cached_position_;
  std::chrono::steady_clock::time_point last_position_update_;

  // A topic with no preview request for kPreviewIdleTimeout is unsubscribed
  // by sweepIdlePreviewSubscriptions() so an ECU that relays another ECU's
  // sensor stream over the network doesn't keep paying for it forever after
  // the dashboard stops looking at it.
  static constexpr std::chrono::seconds kPreviewIdleTimeout{30};
  static constexpr std::chrono::seconds kPreviewIdleSweepInterval{5};

  LazyTopicCache<sensor_msgs::msg::CompressedImage> camera_cache_;

  // Server-side ceiling on points projected per GetLidarCameraProjectionPreview
  // call, regardless of what max_points the client requests.
  static constexpr int32_t kMaxPointCloudPreviewPoints = 5000;
  LazyTopicCache<sensor_msgs::msg::PointCloud2> point_cloud_cache_;

  // camera_info is small, low-rate, and typically latched by the driver, so
  // it shares the default reliable QoS used by camera_cache_ (verified
  // against a real vehicle's camera_info QoS profile; see design doc).
  LazyTopicCache<sensor_msgs::msg::CameraInfo> camera_info_cache_;

  LidarCameraProjector lidar_camera_projector_;

  rclcpp::TimerBase::SharedPtr preview_idle_timer_;
};

}  // namespace ros2_bridge

#endif  // ROS2_BRIDGE_SENSING_HANDLER_HPP_
