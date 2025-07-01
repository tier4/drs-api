#include "ros2_bridge/ros2_handler.hpp"

#include <sstream>
#include <chrono>
#include <thread>
#include <grpcpp/grpcpp.h>

namespace ros2_bridge {

ROS2Handler::ROS2Handler(std::shared_ptr<rclcpp::Node> node)
    : node_(node) {
    // Initialize with some example topics
    SubscribeToTopic("/cmd_vel", 10);
    SubscribeToTopic("/pointcloud", 1);
}

void ROS2Handler::SubscribeToTopic(const std::string& topic_name, int queue_size) {
    // Example implementation for specific message types
    if (topic_name == "/cmd_vel") {
        auto callback = [this, topic_name](const geometry_msgs::msg::Twist::SharedPtr msg) {
            HandleTwistMessage(msg, topic_name);
        };
        
        subscribers_[topic_name] = node_->create_subscription<geometry_msgs::msg::Twist>(
            topic_name, queue_size, callback);
            
        RCLCPP_INFO(node_->get_logger(), "Subscribed to topic: %s", topic_name.c_str());
    }
    else if (topic_name == "/pointcloud") {
        auto callback = [this, topic_name](const sensor_msgs::msg::PointCloud2::SharedPtr msg) {
            HandlePointCloud2Message(msg, topic_name);
        };
        
        subscribers_[topic_name] = node_->create_subscription<sensor_msgs::msg::PointCloud2>(
            topic_name, queue_size, callback);
            
        RCLCPP_INFO(node_->get_logger(), "Subscribed to topic: %s", topic_name.c_str());
    }
}

bool ROS2Handler::GetLatestTopicData(const std::string& topic_name, std::string& data) {
    std::lock_guard<std::mutex> lock(topic_data_mutex_);
    
    auto it = latest_topic_data_.find(topic_name);
    if (it != latest_topic_data_.end()) {
        data = it->second;
        return true;
    }
    
    return false;
}

bool ROS2Handler::CallTriggerService(const std::string& service_name, 
                                   bool& success, std::string& message) {
    auto client = node_->create_client<std_srvs::srv::Trigger>(service_name);
    
    if (!client->wait_for_service(std::chrono::seconds(5))) {
        RCLCPP_ERROR(node_->get_logger(), 
                     "Service %s not available", service_name.c_str());
        return false;
    }
    
    auto request = std::make_shared<std_srvs::srv::Trigger::Request>();
    auto future = client->async_send_request(request);
    
    if (rclcpp::spin_until_future_complete(node_, future) == 
        rclcpp::FutureReturnCode::SUCCESS) {
        auto response = future.get();
        success = response->success;
        message = response->message;
        return true;
    }
    
    return false;
}

void ROS2Handler::HandleTwistMessage(const geometry_msgs::msg::Twist::SharedPtr msg,
                                   const std::string& topic_name) {
    std::stringstream ss;
    ss << "linear: [" << msg->linear.x << ", " << msg->linear.y << ", " << msg->linear.z << "], ";
    ss << "angular: [" << msg->angular.x << ", " << msg->angular.y << ", " << msg->angular.z << "]";
    
    std::lock_guard<std::mutex> lock(topic_data_mutex_);
    latest_topic_data_[topic_name] = ss.str();
}

void ROS2Handler::HandlePointCloud2Message(const sensor_msgs::msg::PointCloud2::SharedPtr msg,
                                         const std::string& topic_name) {
    std::stringstream ss;
    ss << "width: " << msg->width << ", height: " << msg->height;
    ss << ", point_step: " << msg->point_step << ", row_step: " << msg->row_step;
    ss << ", data_size: " << msg->data.size() << " bytes";
    
    std::lock_guard<std::mutex> lock(topic_data_mutex_);
    latest_topic_data_[topic_name] = ss.str();
}

// gRPC service implementations
grpc::Status ROS2Handler::GetTopicData(grpc::ServerContext* context,
                                      const ros2bridge::v1::GetTopicDataRequest* request,
                                      ros2bridge::v1::GetTopicDataResponse* response) {
    std::string data;
    bool success = GetLatestTopicData(request->topic_name(), data);
    
    response->set_success(success);
    if (success) {
        response->set_message("Topic data retrieved successfully");
        response->set_topic_type("unknown"); // TODO: Implement type detection
        // response->mutable_data()->PackFrom(...); // TODO: Implement proper data packing
    } else {
        response->set_message("Failed to get topic data or no data available");
    }
    
    return grpc::Status::OK;
}

grpc::Status ROS2Handler::CallService(grpc::ServerContext* context,
                                    const ros2bridge::v1::CallServiceRequest* request,
                                    ros2bridge::v1::CallServiceResponse* response) {
    // Simple implementation for trigger service
    bool success;
    std::string message;
    bool result = CallTriggerService(request->service_name(), success, message);
    
    response->set_success(result && success);
    response->set_message(message);
    
    return grpc::Status::OK;
}

grpc::Status ROS2Handler::SubscribeToTopic(grpc::ServerContext* context,
                                          const ros2bridge::v1::SubscribeToTopicRequest* request,
                                          grpc::ServerWriter<ros2bridge::v1::TopicDataUpdate>* writer) {
    // Simple implementation - subscribe and stream for a limited time
    SubscribeToTopic(request->topic_name(), request->queue_size());
    
    // Stream for 10 seconds as example
    auto start_time = std::chrono::steady_clock::now();
    while (std::chrono::steady_clock::now() - start_time < std::chrono::seconds(10)) {
        std::string data;
        if (GetLatestTopicData(request->topic_name(), data)) {
            ros2bridge::v1::TopicDataUpdate update;
            update.set_topic_name(request->topic_name());
            // update.mutable_data()->PackFrom(...); // TODO: Implement proper data packing
            update.set_timestamp_ns(std::chrono::duration_cast<std::chrono::nanoseconds>(
                std::chrono::steady_clock::now().time_since_epoch()).count());
            
            if (!writer->Write(update)) {
                break; // Client disconnected
            }
        }
        std::this_thread::sleep_for(std::chrono::milliseconds(100));
    }
    
    return grpc::Status::OK;
}

} // namespace ros2_bridge