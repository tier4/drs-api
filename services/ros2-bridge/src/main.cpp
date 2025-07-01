#include <memory>
#include <thread>
#include <csignal>

#include "rclcpp/rclcpp.hpp"
#include "ros2_bridge/grpc_server.hpp"

std::shared_ptr<ros2_bridge::GrpcServer> g_grpc_server;

void signal_handler(int) {
    if (g_grpc_server) {
        RCLCPP_INFO(rclcpp::get_logger("ros2_bridge"), "Shutting down gRPC server...");
        g_grpc_server->Shutdown();
    }
    rclcpp::shutdown();
}

int main(int argc, char* argv[]) {
    rclcpp::init(argc, argv);

    std::signal(SIGINT, signal_handler);
    std::signal(SIGTERM, signal_handler);

    auto node = std::make_shared<rclcpp::Node>("ros2_bridge_node");
    
    int port = 50052;
    node->declare_parameter("grpc_port", port);
    node->get_parameter("grpc_port", port);

    RCLCPP_INFO(node->get_logger(), "Starting ROS2 Bridge Node");

    g_grpc_server = std::make_shared<ros2_bridge::GrpcServer>(node, port);

    std::thread grpc_thread([&]() {
        g_grpc_server->Run();
    });

    rclcpp::spin(node);

    if (grpc_thread.joinable()) {
        grpc_thread.join();
    }

    return 0;
}