#include "ros2_bridge/sensing_handler.hpp"

#include <opencv2/imgcodecs.hpp>
#include <opencv2/imgproc.hpp>
#include <rclcpp/rclcpp.hpp>

#include <algorithm>
#include <chrono>
#include <cmath>
#include <regex>
#include <string>
#include <vector>

namespace ros2_bridge
{

SensingHandler::SensingHandler(rclcpp::Node::SharedPtr node) : node_(node)
{
  // Subscribe to NavSatFix topic
  nav_sat_fix_sub_ = node_->create_subscription<sensor_msgs::msg::NavSatFix>(
    "/sensing/ins/oxts/nav_sat_fix", 10,
    std::bind(&SensingHandler::navSatFixCallback, this, std::placeholders::_1));

  camera_idle_timer_ = node_->create_wall_timer(
    kCameraIdleSweepInterval, std::bind(&SensingHandler::sweepIdleCameraSubscriptions, this));

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

  response->set_has_data(true);
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

grpc::Status SensingHandler::GetCameraPreview(
  grpc::ServerContext * /* context */, const drs::ros2bridge::v1::GetCameraPreviewRequest * request,
  drs::ros2bridge::v1::GetCameraPreviewResponse * response)
{
  const std::string & topic_name = request->topic_name();
  if (topic_name.empty()) {
    return grpc::Status(grpc::StatusCode::INVALID_ARGUMENT, "topic_name is required");
  }

  sensor_msgs::msg::CompressedImage::SharedPtr frame;
  {
    std::lock_guard<std::mutex> lock(camera_mutex_);

    // Lazily subscribe on first request for this topic. try_emplace avoids a
    // second map lookup on the (common) already-subscribed path, and only
    // constructs the subscription when insertion actually happens.
    auto [sub_it, inserted] = camera_subs_.try_emplace(topic_name, nullptr);
    if (inserted) {
      sub_it->second = node_->create_subscription<sensor_msgs::msg::CompressedImage>(
        topic_name, 10, [this, topic_name](const sensor_msgs::msg::CompressedImage::SharedPtr msg) {
          cameraImageCallback(topic_name, msg);
        });
      RCLCPP_INFO(node_->get_logger(), "Lazily subscribed to camera topic: %s", topic_name.c_str());
    }

    // Mark this topic as actively viewed so the idle sweep doesn't drop it.
    last_camera_request_[topic_name] = std::chrono::steady_clock::now();

    auto it = cached_frames_.find(topic_name);
    if (it != cached_frames_.end()) {
      frame = it->second;
    }
  }

  if (!frame) {
    response->set_has_data(false);
    return grpc::Status::OK;
  }

  // Decode the already-compressed frame, resize to 720p (never upscale), re-encode as JPEG.
  cv::Mat encoded(
    1, static_cast<int>(frame->data.size()), CV_8UC1, const_cast<uint8_t *>(frame->data.data()));
  cv::Mat decoded = cv::imdecode(encoded, cv::IMREAD_COLOR);
  if (decoded.empty()) {
    RCLCPP_WARN(node_->get_logger(), "Failed to decode frame from topic: %s", topic_name.c_str());
    response->set_has_data(false);
    return grpc::Status::OK;
  }

  cv::Mat resized;
  if (decoded.rows > 720) {
    int new_width =
      std::max(1, static_cast<int>(std::lround(decoded.cols * (720.0 / decoded.rows))));
    cv::resize(decoded, resized, cv::Size(new_width, 720));
  } else {
    resized = decoded;
  }

  std::vector<uchar> jpeg_bytes;
  if (!cv::imencode(".jpg", resized, jpeg_bytes)) {
    RCLCPP_WARN(node_->get_logger(), "Failed to encode frame from topic: %s", topic_name.c_str());
    response->set_has_data(false);
    return grpc::Status::OK;
  }

  response->set_has_data(true);
  response->set_content_type("image/jpeg");
  response->set_image_data(jpeg_bytes.data(), jpeg_bytes.size());
  return grpc::Status::OK;
}

void SensingHandler::cameraImageCallback(
  const std::string & topic_name, const sensor_msgs::msg::CompressedImage::SharedPtr msg)
{
  std::lock_guard<std::mutex> lock(camera_mutex_);
  cached_frames_[topic_name] = msg;
}

void SensingHandler::sweepIdleCameraSubscriptions()
{
  std::lock_guard<std::mutex> lock(camera_mutex_);
  const auto now = std::chrono::steady_clock::now();

  for (auto it = last_camera_request_.begin(); it != last_camera_request_.end();) {
    const std::string & topic_name = it->first;
    if (now - it->second > kCameraIdleTimeout) {
      RCLCPP_INFO(node_->get_logger(), "Unsubscribing idle camera topic: %s", topic_name.c_str());
      camera_subs_.erase(topic_name);
      cached_frames_.erase(topic_name);
      it = last_camera_request_.erase(it);
    } else {
      ++it;
    }
  }
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
