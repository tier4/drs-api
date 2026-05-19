#ifndef ROS2_BRIDGE_SENSING_HANDLER_HPP_
#define ROS2_BRIDGE_SENSING_HANDLER_HPP_

#include "drs/ros2bridge/v1/sensing_service.grpc.pb.h"

#include <rclcpp/rclcpp.hpp>

#include <sensor_msgs/msg/nav_sat_fix.hpp>

#include <memory>
#include <mutex>
#include <string>

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

private:
  // ROS2 callback for NavSatFix messages
  void navSatFixCallback(const sensor_msgs::msg::NavSatFix::SharedPtr msg);

  // Convert ROS2 NavSatFix to protobuf Position
  void convertNavSatFixToPosition(
    const sensor_msgs::msg::NavSatFix & nav_sat_fix, drs::ros2bridge::v1::Position * position);

  // Filter nodes based on filter string
  bool matchesFilter(const std::string & node_name, const std::string & filter);

  rclcpp::Node::SharedPtr node_;
  rclcpp::Subscription<sensor_msgs::msg::NavSatFix>::SharedPtr nav_sat_fix_sub_;

  // Cached position data
  std::mutex position_mutex_;
  std::shared_ptr<sensor_msgs::msg::NavSatFix> cached_position_;
  std::chrono::steady_clock::time_point last_position_update_;
};

}  // namespace ros2_bridge

#endif  // ROS2_BRIDGE_SENSING_HANDLER_HPP_
