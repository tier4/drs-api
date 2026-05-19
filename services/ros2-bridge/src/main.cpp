#include "ros2_bridge/grpc_server.hpp"
#include "ros2_bridge/recording_handler.hpp"
#include "ros2_bridge/sensing_handler.hpp"

#include <rclcpp/rclcpp.hpp>

#include <signal.h>

#include <memory>
#include <thread>

std::unique_ptr<ros2_bridge::GrpcServer> g_server;

void signalHandler(int signum)
{
  RCLCPP_INFO(rclcpp::get_logger("ros2_bridge"), "Interrupt signal (%d) received", signum);

  if (g_server) {
    g_server->Stop();
  }

  rclcpp::shutdown();
}

int main(int argc, char ** argv)
{
  // Initialize ROS2
  rclcpp::init(argc, argv);

  // Create ROS2 node
  auto node = rclcpp::Node::make_shared("ros2_bridge_node");

  // Declare parameters
  node->declare_parameter("grpc_port", 50052);
  node->declare_parameter("grpc_address", "0.0.0.0");

  // Get parameters
  int grpc_port = node->get_parameter("grpc_port").as_int();
  std::string grpc_address = node->get_parameter("grpc_address").as_string();
  std::string server_address = grpc_address + ":" + std::to_string(grpc_port);

  RCLCPP_INFO(node->get_logger(), "Starting ROS2 Bridge Node");
  RCLCPP_INFO(node->get_logger(), "gRPC server will listen on: %s", server_address.c_str());

  // Set up signal handling
  signal(SIGINT, signalHandler);
  signal(SIGTERM, signalHandler);

  try {
    // Create gRPC server
    g_server = std::make_unique<ros2_bridge::GrpcServer>();

    // Create service handlers
    auto sensing_handler = std::make_unique<ros2_bridge::SensingHandler>(node);
    auto recording_handler = std::make_unique<ros2_bridge::RecordingHandler>(node);

    // Register services
    g_server->RegisterService(sensing_handler.get());
    g_server->RegisterService(recording_handler.get());

    // Start gRPC server
    if (!g_server->Start(server_address)) {
      RCLCPP_ERROR(node->get_logger(), "Failed to start gRPC server on %s", server_address.c_str());
      return 1;
    }

    RCLCPP_INFO(
      node->get_logger(), "gRPC server started successfully on %s", server_address.c_str());

    // Run gRPC server in separate thread
    std::thread server_thread([&]() { g_server->Wait(); });

    // Spin ROS2 node
    RCLCPP_INFO(node->get_logger(), "ROS2 Bridge Node is running...");
    rclcpp::spin(node);

    // Clean shutdown
    RCLCPP_INFO(node->get_logger(), "Shutting down ROS2 Bridge Node");
    g_server->Stop();

    if (server_thread.joinable()) {
      server_thread.join();
    }

  } catch (const std::exception & e) {
    RCLCPP_ERROR(node->get_logger(), "Exception in main: %s", e.what());
    return 1;
  }

  RCLCPP_INFO(node->get_logger(), "ROS2 Bridge Node stopped");
  return 0;
}
