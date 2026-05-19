#include "ros2_bridge/sensing_handler.hpp"

#include <rclcpp/rclcpp.hpp>

#include <chrono>
#include <regex>
#include <string>

namespace ros2_bridge
{

SensingHandler::SensingHandler(rclcpp::Node::SharedPtr node) : node_(node)
{
  // Subscribe to NavSatFix topic
  nav_sat_fix_sub_ = node_->create_subscription<sensor_msgs::msg::NavSatFix>(
    "/sensing/ins/oxts/nav_sat_fix", 10,
    std::bind(&SensingHandler::navSatFixCallback, this, std::placeholders::_1));

  RCLCPP_INFO(node_->get_logger(), "SensingHandler initialized");
}

grpc::Status SensingHandler::GetPosition(
  grpc::ServerContext * /* context */,
  const drs::ros2bridge::v1::GetPositionRequest * /* request */,
  drs::ros2bridge::v1::GetPositionResponse * response)
{
  std::lock_guard<std::mutex> lock(position_mutex_);

  if (!cached_position_) {
    return grpc::Status(grpc::StatusCode::NOT_FOUND, "Position data not available");
  }

  // Check if data is too old (more than 5 seconds)
  auto now = std::chrono::steady_clock::now();
  auto age = std::chrono::duration_cast<std::chrono::seconds>(now - last_position_update_);
  if (age.count() > 5) {
    return grpc::Status(grpc::StatusCode::UNAVAILABLE, "Position data is stale");
  }

  convertNavSatFixToPosition(*cached_position_, response->mutable_position());
  return grpc::Status::OK;
}

grpc::Status SensingHandler::ListNodes(
  grpc::ServerContext * /* context */, const drs::ros2bridge::v1::ListNodesRequest * request,
  drs::ros2bridge::v1::ListNodesResponse * response)
{
  // Get all node names (ROS2 Humble API)
  auto node_names = node_->get_node_names();

  for (const auto & name : node_names) {
    // Apply filter if provided
    if (!request->filter().empty() && !matchesFilter(name, request->filter())) {
      continue;
    }

    auto * node_proto = response->add_nodes();
    node_proto->set_name(name);

    // Extract namespace from node name
    size_t last_slash = name.find_last_of('/');
    if (last_slash != std::string::npos) {
      node_proto->set_namespace_(name.substr(0, last_slash + 1));
    } else {
      node_proto->set_namespace_("/");
    }
  }

  return grpc::Status::OK;
}

void SensingHandler::navSatFixCallback(const sensor_msgs::msg::NavSatFix::SharedPtr msg)
{
  std::lock_guard<std::mutex> lock(position_mutex_);
  cached_position_ = msg;
  last_position_update_ = std::chrono::steady_clock::now();

  RCLCPP_DEBUG(
    node_->get_logger(), "Updated position cache: lat=%.6f, lon=%.6f, alt=%.3f", msg->latitude,
    msg->longitude, msg->altitude);
}

void SensingHandler::convertNavSatFixToPosition(
  const sensor_msgs::msg::NavSatFix & nav_sat_fix, drs::ros2bridge::v1::Position * position)
{
  // Convert header
  auto * header = position->mutable_header();
  header->mutable_stamp()->set_seconds(nav_sat_fix.header.stamp.sec);
  header->mutable_stamp()->set_nanos(nav_sat_fix.header.stamp.nanosec);
  header->set_frame_id(nav_sat_fix.header.frame_id);

  // Convert navigation satellite status
  auto * nav_sat_status = position->mutable_nav_sat_status();
  nav_sat_status->set_status(nav_sat_fix.status.status);
  nav_sat_status->set_service(nav_sat_fix.status.service);

  // Convert position data
  position->set_latitude(nav_sat_fix.latitude);
  position->set_longitude(nav_sat_fix.longitude);
  position->set_altitude(nav_sat_fix.altitude);

  // Convert position covariance (3x3 matrix, row-major)
  for (double covariance : nav_sat_fix.position_covariance) {
    position->add_position_covariance(covariance);
  }
  position->set_position_covariance_type(nav_sat_fix.position_covariance_type);
}

bool SensingHandler::matchesFilter(const std::string & node_name, const std::string & filter)
{
  // Simple filter format: "namespace:=/sensing"
  if (filter.find("namespace:=") == 0) {
    std::string target_namespace = filter.substr(11);  // Remove "namespace:="
    return node_name.find(target_namespace) == 0;
  }

  // Default: no filter matches all
  return true;
}

}  // namespace ros2_bridge
