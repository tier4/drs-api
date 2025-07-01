#include "ros2_bridge/grpc_server.hpp"
#include "ros2_bridge/ros2_handler.hpp"

#include <grpcpp/server_builder.h>

namespace ros2_bridge {

GrpcServer::GrpcServer(std::shared_ptr<rclcpp::Node> node, int port)
    : node_(node), port_(port) {
    ros2_handler_ = std::make_shared<ROS2Handler>(node_);
}

GrpcServer::~GrpcServer() {
    Shutdown();
}

void GrpcServer::Run() {
    std::string server_address = "0.0.0.0:" + std::to_string(port_);

    grpc::ServerBuilder builder;
    builder.AddListeningPort(server_address, grpc::InsecureServerCredentials());
    
    // Register the ROS2 bridge service
    builder.RegisterService(ros2_handler_.get());

    server_ = builder.BuildAndStart();
    
    if (server_) {
        RCLCPP_INFO(node_->get_logger(), 
                    "gRPC server listening on %s", server_address.c_str());
        server_->Wait();
    } else {
        RCLCPP_ERROR(node_->get_logger(), 
                     "Failed to start gRPC server on %s", server_address.c_str());
    }
}

void GrpcServer::Shutdown() {
    if (server_) {
        server_->Shutdown();
    }
}

} // namespace ros2_bridge