#pragma once

#include <memory>
#include <string>

#include <grpcpp/grpcpp.h>
#include "rclcpp/rclcpp.hpp"

namespace ros2_bridge {

class ROS2Handler;

class GrpcServer {
public:
    GrpcServer(std::shared_ptr<rclcpp::Node> node, int port);
    ~GrpcServer();

    void Run();
    void Shutdown();

private:
    std::shared_ptr<rclcpp::Node> node_;
    int port_;
    std::unique_ptr<grpc::Server> server_;
    std::shared_ptr<ROS2Handler> ros2_handler_;
};

} // namespace ros2_bridge