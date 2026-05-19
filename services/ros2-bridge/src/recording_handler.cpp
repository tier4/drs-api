#include "ros2_bridge/recording_handler.hpp"

#include <rclcpp/rclcpp.hpp>

#include <chrono>
#include <regex>
#include <string>

namespace ros2_bridge
{

RecordingHandler::RecordingHandler(rclcpp::Node::SharedPtr node) : node_(node)
{
  // Create publishers for recording control
  start_publisher_ = node_->create_publisher<std_msgs::msg::Bool>(
    "/recorder/start", rclcpp::QoS(1).transient_local());
  pause_publisher_ = node_->create_publisher<std_msgs::msg::Bool>(
    "/recorder/pause", rclcpp::QoS(1).transient_local());

  // Subscribe to recording status
  status_subscriber_ = node_->create_subscription<proto_recorder_msgs::msg::RecorderStatus>(
    "/recorder/status", 10,
    std::bind(&RecordingHandler::recorderStatusCallback, this, std::placeholders::_1));

  RCLCPP_INFO(node_->get_logger(), "RecordingHandler initialized");
}

grpc::Status RecordingHandler::GetRecording(
  grpc::ServerContext * /* context */, const drs::ros2bridge::v1::GetRecordingRequest * request,
  drs::ros2bridge::v1::GetRecordingResponse * response)
{
  std::lock_guard<std::mutex> lock(recordings_mutex_);

  auto it = recordings_cache_.find(request->hardware_id());
  if (it == recordings_cache_.end()) {
    return grpc::Status(
      grpc::StatusCode::NOT_FOUND,
      "Recording data not found for hardware_id: " + request->hardware_id());
  }

  // Check if data is too old (more than 10 seconds)
  auto update_it = last_updates_.find(request->hardware_id());
  if (update_it != last_updates_.end()) {
    auto now = std::chrono::steady_clock::now();
    auto age = std::chrono::duration_cast<std::chrono::seconds>(now - update_it->second);
    if (age.count() > 10) {
      return grpc::Status(
        grpc::StatusCode::UNAVAILABLE,
        "Recording data is stale for hardware_id: " + request->hardware_id());
    }
  }

  convertRecorderStatusToRecording(it->second, response->mutable_recording());
  return grpc::Status::OK;
}

grpc::Status RecordingHandler::ListRecordings(
  grpc::ServerContext * /* context */, const drs::ros2bridge::v1::ListRecordingsRequest * request,
  drs::ros2bridge::v1::ListRecordingsResponse * response)
{
  std::lock_guard<std::mutex> lock(recordings_mutex_);

  for (const auto & [hardware_id, status] : recordings_cache_) {
    // Apply filter if provided
    if (!request->filter().empty() && !matchesRecordingFilter(hardware_id, request->filter())) {
      continue;
    }

    auto * recording = response->add_recordings();
    convertRecorderStatusToRecording(status, recording);
  }

  return grpc::Status::OK;
}

grpc::Status RecordingHandler::ListTopicStatuses(
  grpc::ServerContext * /* context */,
  const drs::ros2bridge::v1::ListTopicStatusesRequest * request,
  drs::ros2bridge::v1::ListTopicStatusesResponse * response)
{
  std::lock_guard<std::mutex> lock(recordings_mutex_);

  auto it = recordings_cache_.find(request->hardware_id());
  if (it == recordings_cache_.end()) {
    return grpc::Status(
      grpc::StatusCode::NOT_FOUND,
      "Recording data not found for hardware_id: " + request->hardware_id());
  }

  for (const auto & topic_status : it->second.topic_statuses) {
    // Apply filter if provided
    if (!request->filter().empty() && !matchesTopicFilter(topic_status.name, request->filter())) {
      continue;
    }

    auto * proto_topic_status = response->add_topic_statuses();
    convertTopicStatus(topic_status, proto_topic_status);
  }

  return grpc::Status::OK;
}

grpc::Status RecordingHandler::StartRecording(
  grpc::ServerContext * /* context */,
  const drs::ros2bridge::v1::StartRecordingRequest * /* request */,
  drs::ros2bridge::v1::StartRecordingResponse * response)
{
  bool success = publishCommand(start_publisher_, true);
  response->set_success(success);
  response->set_message(
    success ? "Start command sent successfully" : "Failed to send start command");

  return grpc::Status::OK;
}

grpc::Status RecordingHandler::StopRecording(
  grpc::ServerContext * /* context */,
  const drs::ros2bridge::v1::StopRecordingRequest * /* request */,
  drs::ros2bridge::v1::StopRecordingResponse * response)
{
  bool success = publishCommand(start_publisher_, false);
  response->set_success(success);
  response->set_message(success ? "Stop command sent successfully" : "Failed to send stop command");

  return grpc::Status::OK;
}

grpc::Status RecordingHandler::PauseRecording(
  grpc::ServerContext * /* context */,
  const drs::ros2bridge::v1::PauseRecordingRequest * /* request */,
  drs::ros2bridge::v1::PauseRecordingResponse * response)
{
  bool success = publishCommand(pause_publisher_, true);
  response->set_success(success);
  response->set_message(
    success ? "Pause command sent successfully" : "Failed to send pause command");

  return grpc::Status::OK;
}

grpc::Status RecordingHandler::ResumeRecording(
  grpc::ServerContext * /* context */,
  const drs::ros2bridge::v1::ResumeRecordingRequest * /* request */,
  drs::ros2bridge::v1::ResumeRecordingResponse * response)
{
  bool success = publishCommand(pause_publisher_, false);
  response->set_success(success);
  response->set_message(
    success ? "Resume command sent successfully" : "Failed to send resume command");

  return grpc::Status::OK;
}

void RecordingHandler::recorderStatusCallback(
  const proto_recorder_msgs::msg::RecorderStatus::SharedPtr msg)
{
  std::lock_guard<std::mutex> lock(recordings_mutex_);

  recordings_cache_[msg->hardware_id] = *msg;
  last_updates_[msg->hardware_id] = std::chrono::steady_clock::now();

  RCLCPP_DEBUG(
    node_->get_logger(), "Updated recording cache for hardware_id: %s, is_recording: %s",
    msg->hardware_id.c_str(), msg->is_recording ? "true" : "false");
}

void RecordingHandler::convertRecorderStatusToRecording(
  const proto_recorder_msgs::msg::RecorderStatus & status,
  drs::ros2bridge::v1::Recording * recording)
{
  // Convert header
  auto * header = recording->mutable_header();
  header->mutable_stamp()->set_seconds(status.header.stamp.sec);
  header->mutable_stamp()->set_nanos(status.header.stamp.nanosec);
  header->set_frame_id(status.header.frame_id);

  // Set recording data
  recording->set_hardware_id(status.hardware_id);
  recording->set_is_recording(status.is_recording);

  // Convert error level (using actual constants from proto_recorder_msgs)
  switch (status.error_level) {
    case proto_recorder_msgs::msg::RecorderStatus::ERROR_LEVEL_OK:
      recording->set_error_level(drs::ros2bridge::v1::Recording::ERROR_LEVEL_OK);
      break;
    case proto_recorder_msgs::msg::RecorderStatus::ERROR_LEVEL_WARN:
      recording->set_error_level(drs::ros2bridge::v1::Recording::ERROR_LEVEL_WARN);
      break;
    case proto_recorder_msgs::msg::RecorderStatus::ERROR_LEVEL_ERROR:
      recording->set_error_level(drs::ros2bridge::v1::Recording::ERROR_LEVEL_ERROR);
      break;
    default:
      recording->set_error_level(drs::ros2bridge::v1::Recording::ERROR_LEVEL_OK);
      break;
  }

  // Convert topic statuses
  for (const auto & topic_status : status.topic_statuses) {
    auto * proto_topic_status = recording->add_topic_statuses();
    convertTopicStatus(topic_status, proto_topic_status);
  }
}

void RecordingHandler::convertTopicStatus(
  const proto_recorder_msgs::msg::TopicStatus & ros_topic_status,
  drs::ros2bridge::v1::TopicStatus * proto_topic_status)
{
  // Create resource name from topic name
  std::string topic_id = ros_topic_status.name;
  // Replace '/' with '_' to create valid resource ID
  std::replace(topic_id.begin(), topic_id.end(), '/', '_');
  proto_topic_status->set_name("topic_statuses/" + topic_id);

  proto_topic_status->set_topic_name(ros_topic_status.name);
  proto_topic_status->set_message_type(ros_topic_status.type);
  proto_topic_status->set_rate_hz(ros_topic_status.rate);

  // Convert rate status (using actual constants from proto_recorder_msgs)
  switch (ros_topic_status.rate_status) {
    case proto_recorder_msgs::msg::TopicStatus::RATE_STATUS_NORMAL:
      proto_topic_status->set_rate_status(drs::ros2bridge::v1::TopicStatus::RATE_STATUS_NORMAL);
      break;
    case proto_recorder_msgs::msg::TopicStatus::RATE_STATUS_UNKNOWN:
      proto_topic_status->set_rate_status(drs::ros2bridge::v1::TopicStatus::RATE_STATUS_UNKNOWN);
      break;
    case proto_recorder_msgs::msg::TopicStatus::RATE_STATUS_TOO_LOW:
      proto_topic_status->set_rate_status(drs::ros2bridge::v1::TopicStatus::RATE_STATUS_TOO_LOW);
      break;
    case proto_recorder_msgs::msg::TopicStatus::RATE_STATUS_TOO_HIGH:
      proto_topic_status->set_rate_status(drs::ros2bridge::v1::TopicStatus::RATE_STATUS_TOO_HIGH);
      break;
    case proto_recorder_msgs::msg::TopicStatus::RATE_STATUS_NO_MESSAGES:
      proto_topic_status->set_rate_status(
        drs::ros2bridge::v1::TopicStatus::RATE_STATUS_NO_MESSAGES);
      break;
    default:
      proto_topic_status->set_rate_status(
        drs::ros2bridge::v1::TopicStatus::RATE_STATUS_UNSPECIFIED);
      break;
  }
}

bool RecordingHandler::publishCommand(
  rclcpp::Publisher<std_msgs::msg::Bool>::SharedPtr publisher, bool value)
{
  try {
    auto msg = std_msgs::msg::Bool();
    msg.data = value;
    publisher->publish(msg);
    return true;
  } catch (const std::exception & e) {
    RCLCPP_ERROR(node_->get_logger(), "Failed to publish command: %s", e.what());
    return false;
  }
}

bool RecordingHandler::matchesRecordingFilter(
  const std::string & hardware_id, const std::string & filter)
{
  // Simple filter format: "hardware_id:=main_ecu"
  if (filter.find("hardware_id:=") == 0) {
    std::string target_id = filter.substr(12);  // Remove "hardware_id:="
    return hardware_id == target_id;
  }

  // Default: no filter matches all
  return true;
}

bool RecordingHandler::matchesTopicFilter(
  const std::string & topic_name, const std::string & filter)
{
  // Simple filter format: "topic_name:=/sensing/*"
  if (filter.find("topic_name:=") == 0) {
    std::string pattern = filter.substr(12);  // Remove "topic_name:="

    // Simple wildcard matching (* at the end)
    if (pattern.back() == '*') {
      std::string prefix = pattern.substr(0, pattern.length() - 1);
      return topic_name.find(prefix) == 0;
    } else {
      return topic_name == pattern;
    }
  }

  // Default: no filter matches all
  return true;
}

}  // namespace ros2_bridge
