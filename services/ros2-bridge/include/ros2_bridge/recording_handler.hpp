#ifndef ROS2_BRIDGE_RECORDING_HANDLER_HPP
#define ROS2_BRIDGE_RECORDING_HANDLER_HPP

#include "drs/ros2bridge/v1/recording_service.grpc.pb.h"

#include <rclcpp/rclcpp.hpp>

#include <proto_recorder_msgs/msg/recorder_status.hpp>
#include <std_msgs/msg/bool.hpp>

#include <memory>
#include <mutex>
#include <unordered_map>

namespace ros2_bridge
{

class RecordingHandler : public drs::ros2bridge::v1::RecordingService::Service
{
public:
  explicit RecordingHandler(rclcpp::Node::SharedPtr node);
  ~RecordingHandler() override = default;

  // gRPC service methods
  grpc::Status GetRecording(
    grpc::ServerContext * context, const drs::ros2bridge::v1::GetRecordingRequest * request,
    drs::ros2bridge::v1::GetRecordingResponse * response) override;

  grpc::Status ListRecordings(
    grpc::ServerContext * context, const drs::ros2bridge::v1::ListRecordingsRequest * request,
    drs::ros2bridge::v1::ListRecordingsResponse * response) override;

  grpc::Status ListTopicStatuses(
    grpc::ServerContext * context, const drs::ros2bridge::v1::ListTopicStatusesRequest * request,
    drs::ros2bridge::v1::ListTopicStatusesResponse * response) override;

  grpc::Status StartRecording(
    grpc::ServerContext * context, const drs::ros2bridge::v1::StartRecordingRequest * request,
    drs::ros2bridge::v1::StartRecordingResponse * response) override;

  grpc::Status StopRecording(
    grpc::ServerContext * context, const drs::ros2bridge::v1::StopRecordingRequest * request,
    drs::ros2bridge::v1::StopRecordingResponse * response) override;

  grpc::Status PauseRecording(
    grpc::ServerContext * context, const drs::ros2bridge::v1::PauseRecordingRequest * request,
    drs::ros2bridge::v1::PauseRecordingResponse * response) override;

  grpc::Status ResumeRecording(
    grpc::ServerContext * context, const drs::ros2bridge::v1::ResumeRecordingRequest * request,
    drs::ros2bridge::v1::ResumeRecordingResponse * response) override;

private:
  // ROS2 callback for recorder status messages
  void recorderStatusCallback(const proto_recorder_msgs::msg::RecorderStatus::SharedPtr msg);

  // Convert ROS2 RecorderStatus to protobuf Recording
  void convertRecorderStatusToRecording(
    const proto_recorder_msgs::msg::RecorderStatus & status,
    drs::ros2bridge::v1::Recording * recording);

  // Convert ROS2 TopicStatus to protobuf TopicStatus
  void convertTopicStatus(
    const proto_recorder_msgs::msg::TopicStatus & ros_topic_status,
    drs::ros2bridge::v1::TopicStatus * proto_topic_status);

  // Helper to publish boolean commands
  bool publishCommand(rclcpp::Publisher<std_msgs::msg::Bool>::SharedPtr publisher, bool value);

  // Filter recordings based on filter string
  bool matchesRecordingFilter(const std::string & hardware_id, const std::string & filter);

  // Filter topic statuses based on filter string
  bool matchesTopicFilter(const std::string & topic_name, const std::string & filter);

  rclcpp::Node::SharedPtr node_;

  // ROS2 publishers for recording control
  rclcpp::Publisher<std_msgs::msg::Bool>::SharedPtr start_publisher_;
  rclcpp::Publisher<std_msgs::msg::Bool>::SharedPtr pause_publisher_;

  // ROS2 subscriber for recording status
  rclcpp::Subscription<proto_recorder_msgs::msg::RecorderStatus>::SharedPtr status_subscriber_;

  // Multi-ECU cache: hardware_id -> RecorderStatus
  std::mutex recordings_mutex_;
  std::unordered_map<std::string, proto_recorder_msgs::msg::RecorderStatus> recordings_cache_;
  std::unordered_map<std::string, std::chrono::steady_clock::time_point> last_updates_;
};

}  // namespace ros2_bridge

#endif  // ROS2_BRIDGE_RECORDING_HANDLER_HPP
