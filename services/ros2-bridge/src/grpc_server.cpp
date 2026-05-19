#include "ros2_bridge/grpc_server.hpp"

#include <rclcpp/rclcpp.hpp>

#include <string>

namespace ros2_bridge
{

GrpcServer::GrpcServer() : started_(false)
{
  // Configure server settings
  builder_.SetMaxReceiveMessageSize(64 * 1024 * 1024);  // 64MB
  builder_.SetMaxSendMessageSize(64 * 1024 * 1024);     // 64MB
}

void GrpcServer::RegisterService(grpc::Service * service)
{
  if (started_) {
    throw std::runtime_error("Cannot register service after server has started");
  }

  builder_.RegisterService(service);
}

bool GrpcServer::Start(const std::string & server_address)
{
  if (started_) {
    return false;
  }

  // Listen on the given address
  builder_.AddListeningPort(server_address, grpc::InsecureServerCredentials());

  // Build and start the server
  server_ = builder_.BuildAndStart();
  if (!server_) {
    return false;
  }

  started_ = true;
  return true;
}

void GrpcServer::Stop()
{
  if (server_ && started_) {
    server_->Shutdown();
    started_ = false;
  }
}

void GrpcServer::Wait()
{
  if (server_ && started_) {
    server_->Wait();
  }
}

}  // namespace ros2_bridge
