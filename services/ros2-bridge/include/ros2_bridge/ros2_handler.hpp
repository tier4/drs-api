#pragma once

#include <memory>
#include <string>
#include <unordered_map>
#include <mutex>

#include "rclcpp/rclcpp.hpp"
#include "geometry_msgs/msg/twist.hpp"
#include "sensor_msgs/msg/point_cloud2.hpp"
#include "std_srvs/srv/trigger.hpp"
#include "ros2bridge/v1/bridge.grpc.pb.h"

namespace ros2_bridge {

class ROS2Handler : public ros2bridge::v1::ROS2BridgeService::Service {
public:
    explicit ROS2Handler(std::shared_ptr<rclcpp::Node> node);
    ~ROS2Handler() = default;

    // gRPC service implementations
    grpc::Status GetTopicData(grpc::ServerContext* context,
                             const ros2bridge::v1::GetTopicDataRequest* request,
                             ros2bridge::v1::GetTopicDataResponse* response) override;
    
    grpc::Status CallService(grpc::ServerContext* context,
                           const ros2bridge::v1::CallServiceRequest* request,
                           ros2bridge::v1::CallServiceResponse* response) override;
    
    grpc::Status SubscribeToTopic(grpc::ServerContext* context,
                                const ros2bridge::v1::SubscribeToTopicRequest* request,
                                grpc::ServerWriter<ros2bridge::v1::TopicDataUpdate>* writer) override;

    // Topic handling
    void SubscribeToTopic(const std::string& topic_name, int queue_size);
    bool GetLatestTopicData(const std::string& topic_name, std::string& data);

    // Service handling
    bool CallTriggerService(const std::string& service_name, 
                          bool& success, std::string& message);

private:
    std::shared_ptr<rclcpp::Node> node_;
    
    // Topic subscribers and data storage
    std::unordered_map<std::string, rclcpp::SubscriptionBase::SharedPtr> subscribers_;
    std::unordered_map<std::string, std::string> latest_topic_data_;
    std::mutex topic_data_mutex_;

    // Callback handlers for different message types
    void HandleTwistMessage(const geometry_msgs::msg::Twist::SharedPtr msg,
                          const std::string& topic_name);
    void HandlePointCloud2Message(const sensor_msgs::msg::PointCloud2::SharedPtr msg,
                                const std::string& topic_name);
};

} // namespace ros2_bridge