#include "ros2_bridge/sensing_handler.hpp"

#include <opencv2/imgcodecs.hpp>
#include <opencv2/imgproc.hpp>
#include <rclcpp/rclcpp.hpp>

#include <sensor_msgs/point_cloud2_iterator.hpp>

#include <algorithm>
#include <chrono>
#include <cmath>
#include <cstring>
#include <memory>
#include <regex>
#include <string>
#include <unordered_map>
#include <vector>

namespace ros2_bridge
{

SensingHandler::SensingHandler(rclcpp::Node::SharedPtr node)
: node_(node),
  camera_cache_(node_, kPreviewIdleTimeout),
  // LiDAR decoders publish PointCloud2 with best-effort reliability (large,
  // high-rate sensor data); a reliable subscriber silently fails to connect
  // to a best-effort publisher at the DDS level (no error, just no
  // messages), so this must be at least as permissive as SensorDataQoS.
  point_cloud_cache_(node_, kPreviewIdleTimeout, rclcpp::SensorDataQoS()),
  camera_info_cache_(node_, kPreviewIdleTimeout),
  lidar_camera_projector_(node_)
{
  // Subscribe to NavSatFix topic
  nav_sat_fix_sub_ = node_->create_subscription<sensor_msgs::msg::NavSatFix>(
    "/sensing/ins/oxts/nav_sat_fix", 10,
    std::bind(&SensingHandler::navSatFixCallback, this, std::placeholders::_1));

  preview_idle_timer_ = node_->create_wall_timer(
    kPreviewIdleSweepInterval, std::bind(&SensingHandler::sweepIdlePreviewSubscriptions, this));

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

  sensor_msgs::msg::CompressedImage::SharedPtr frame = camera_cache_.getOrSubscribe(topic_name);

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

namespace
{

constexpr char kPacketsSuffix[] = "_packets";
constexpr char kPointsSuffix[] = "_points";

// Locates a PointField by name, or returns nullptr if absent.
const sensor_msgs::msg::PointField * findField(
  const sensor_msgs::msg::PointCloud2 & msg, const std::string & name)
{
  for (const auto & field : msg.fields) {
    if (field.name == name) {
      return &field;
    }
  }
  return nullptr;
}

// Per-field bounds check: the field must be FLOAT32 (PointCloud2Iterator<float>
// throws if the field's actual datatype doesn't match the requested type, so
// this also doubles as a crash guard against a decoder publishing a
// same-named field with an unexpected datatype) and offset + sizeof(float)
// must fit within point_step.
bool fieldFitsInStep(const sensor_msgs::msg::PointField & field, uint32_t point_step)
{
  if (field.datatype != sensor_msgs::msg::PointField::FLOAT32) {
    return false;
  }
  return static_cast<uint64_t>(field.offset) + sizeof(float) <= point_step;
}

// Fixed vehicle-hardware mapping from LiDAR position to the camera it's
// projected onto. Deliberately hardcoded rather than derived from /tf_static
// topology: each lidar_<position> frame has multiple camera children in the
// real TF tree (verified against a production rosbag), so topology alone
// can't disambiguate which camera is "the" intended one. TF is still used
// afterwards to validate that this pairing actually resolves.
const std::unordered_map<std::string, int> kPositionToCameraNumber = {
  {"front", 1},
  {"right", 2},
  {"rear", 5},
  {"left", 7},
};

// Decimated extraction of x/y/z/intensity from a PointCloud2, applying the
// same malformed-message guards as the removed GetPointCloudPreview did.
// Returns false (leaving points/intensities untouched) if the cloud is
// undersized or missing/invalid x/y/z fields. max_points <= 0 means
// unlimited: every point in the cloud is extracted.
bool extractPointsAndIntensities(
  const sensor_msgs::msg::PointCloud2 & cloud, int32_t max_points,
  std::vector<cv::Point3f> & points, std::vector<float> & intensities)
{
  const uint64_t expected_size =
    static_cast<uint64_t>(cloud.row_step) * static_cast<uint64_t>(cloud.height);
  if (cloud.data.size() < expected_size) {
    return false;
  }

  const auto * x_field = findField(cloud, "x");
  const auto * y_field = findField(cloud, "y");
  const auto * z_field = findField(cloud, "z");
  if (
    !x_field || !y_field || !z_field || !fieldFitsInStep(*x_field, cloud.point_step) ||
    !fieldFitsInStep(*y_field, cloud.point_step) || !fieldFitsInStep(*z_field, cloud.point_step)) {
    return false;
  }

  const auto * intensity_field = findField(cloud, "intensity");
  if (!intensity_field) {
    intensity_field = findField(cloud, "reflectivity");
  }
  const bool has_intensity = intensity_field && fieldFitsInStep(*intensity_field, cloud.point_step);

  const uint32_t total_points = cloud.width * cloud.height;
  const bool unlimited = max_points <= 0;
  const uint32_t max_points_u = unlimited ? total_points : static_cast<uint32_t>(max_points);
  const uint32_t stride = !unlimited && total_points > max_points_u
                            ? (total_points + max_points_u - 1) / max_points_u
                            : 1;

  sensor_msgs::PointCloud2ConstIterator<float> x_it(cloud, "x");
  sensor_msgs::PointCloud2ConstIterator<float> y_it(cloud, "y");
  sensor_msgs::PointCloud2ConstIterator<float> z_it(cloud, "z");
  std::unique_ptr<sensor_msgs::PointCloud2ConstIterator<float>> intensity_it;
  if (has_intensity) {
    intensity_it =
      std::make_unique<sensor_msgs::PointCloud2ConstIterator<float>>(cloud, intensity_field->name);
  }

  points.reserve(std::min(total_points, max_points_u));
  intensities.reserve(std::min(total_points, max_points_u));

  for (uint32_t i = 0; i < total_points && points.size() < static_cast<size_t>(max_points_u); ++i) {
    if (i % stride == 0) {
      points.emplace_back(*x_it, *y_it, *z_it);
      intensities.push_back(has_intensity ? **intensity_it : 0.0f);
    }
    ++x_it;
    ++y_it;
    ++z_it;
    if (has_intensity) {
      ++(*intensity_it);
    }
  }
  return true;
}

}  // namespace

std::string SensingHandler::derivePointsTopic(const std::string & packets_topic_name)
{
  const size_t suffix_len = std::strlen(kPacketsSuffix);
  if (packets_topic_name.size() <= suffix_len) {
    return "";
  }
  const size_t suffix_pos = packets_topic_name.size() - suffix_len;
  if (packets_topic_name.compare(suffix_pos, suffix_len, kPacketsSuffix) != 0) {
    return "";
  }
  return packets_topic_name.substr(0, suffix_pos) + kPointsSuffix;
}

std::string SensingHandler::deriveLidarPosition(const std::string & packets_topic_name)
{
  static const std::string kPrefix = "/sensing/lidar/";
  if (packets_topic_name.compare(0, kPrefix.size(), kPrefix) != 0) {
    return "";
  }
  const size_t start = kPrefix.size();
  const size_t slash = packets_topic_name.find('/', start);
  if (slash == std::string::npos) {
    return "";
  }
  return packets_topic_name.substr(start, slash - start);
}

grpc::Status SensingHandler::GetLidarCameraProjectionPreview(
  grpc::ServerContext * /* context */,
  const drs::ros2bridge::v1::GetLidarCameraProjectionPreviewRequest * request,
  drs::ros2bridge::v1::GetLidarCameraProjectionPreviewResponse * response)
{
  const std::string & topic_name = request->topic_name();
  if (topic_name.empty()) {
    return grpc::Status(grpc::StatusCode::INVALID_ARGUMENT, "topic_name is required");
  }

  const std::string points_topic_name = derivePointsTopic(topic_name);
  if (points_topic_name.empty()) {
    return grpc::Status(
      grpc::StatusCode::INVALID_ARGUMENT, "topic_name must end with \"_packets\"");
  }

  const std::string position = deriveLidarPosition(topic_name);
  const auto camera_it = kPositionToCameraNumber.find(position);
  if (camera_it == kPositionToCameraNumber.end()) {
    return grpc::Status(
      grpc::StatusCode::INVALID_ARGUMENT,
      "topic_name's LiDAR position has no configured camera mapping");
  }
  const int camera_number = camera_it->second;
  const std::string camera_topic_name =
    "/sensing/camera/camera" + std::to_string(camera_number) + "/image_raw/compressed";
  const std::string camera_info_topic_name =
    "/sensing/camera/camera" + std::to_string(camera_number) + "/camera_info";
  const std::string lidar_frame = "lidar_" + position;
  const std::string camera_optical_frame =
    "camera" + std::to_string(camera_number) + "/camera_optical_link";

  // max_points <= 0 (including the unset proto3 default) means unlimited:
  // extractPointsAndIntensities projects every point in the cloud.
  const int32_t max_points = request->max_points();

  sensor_msgs::msg::PointCloud2::SharedPtr cloud =
    point_cloud_cache_.getOrSubscribe(points_topic_name);
  const bool decoder_running = point_cloud_cache_.publisherCount(points_topic_name) > 0;
  response->set_decoder_running(decoder_running);

  if (!cloud) {
    response->set_has_data(false);
    return grpc::Status::OK;
  }

  sensor_msgs::msg::CompressedImage::SharedPtr frame =
    camera_cache_.getOrSubscribe(camera_topic_name);
  sensor_msgs::msg::CameraInfo::SharedPtr camera_info =
    camera_info_cache_.getOrSubscribe(camera_info_topic_name);
  if (!frame || !camera_info) {
    response->set_has_data(false);
    return grpc::Status::OK;
  }

  std::vector<cv::Point3f> points;
  std::vector<float> intensities;
  if (!extractPointsAndIntensities(*cloud, max_points, points, intensities)) {
    RCLCPP_WARN(
      node_->get_logger(), "PointCloud2 on %s is malformed, discarding frame",
      points_topic_name.c_str());
    response->set_has_data(false);
    return grpc::Status::OK;
  }

  cv::Mat encoded(
    1, static_cast<int>(frame->data.size()), CV_8UC1, const_cast<uint8_t *>(frame->data.data()));
  cv::Mat decoded = cv::imdecode(encoded, cv::IMREAD_COLOR);
  if (decoded.empty()) {
    RCLCPP_WARN(
      node_->get_logger(), "Failed to decode frame from topic: %s", camera_topic_name.c_str());
    response->set_has_data(false);
    return grpc::Status::OK;
  }

  // Project at the camera_info's native resolution (unscaled K) before
  // resizing, so no K-scaling arithmetic is needed.
  if (!lidar_camera_projector_.projectOntoImage(
        points, intensities, *camera_info, lidar_frame, camera_optical_frame, decoded)) {
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
    RCLCPP_WARN(
      node_->get_logger(), "Failed to encode projected frame for topic: %s", topic_name.c_str());
    response->set_has_data(false);
    return grpc::Status::OK;
  }

  response->set_has_data(true);
  response->set_content_type("image/jpeg");
  response->set_image_data(jpeg_bytes.data(), jpeg_bytes.size());
  return grpc::Status::OK;
}

void SensingHandler::sweepIdlePreviewSubscriptions()
{
  camera_cache_.sweepIdle();
  point_cloud_cache_.sweepIdle();
  camera_info_cache_.sweepIdle();
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
