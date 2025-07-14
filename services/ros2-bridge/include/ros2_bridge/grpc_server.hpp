#ifndef ROS2_BRIDGE_GRPC_SERVER_HPP
#define ROS2_BRIDGE_GRPC_SERVER_HPP

#include <grpcpp/grpcpp.h>
#include <memory>
#include <string>

namespace ros2_bridge {

class GrpcServer {
public:
    GrpcServer();
    ~GrpcServer() = default;

    // Register a service with the server
    void RegisterService(grpc::Service* service);

    // Start the server on the specified port
    bool Start(const std::string& server_address);

    // Stop the server
    void Stop();

    // Wait for the server to finish
    void Wait();

private:
    std::unique_ptr<grpc::Server> server_;
    grpc::ServerBuilder builder_;
    bool started_;
};

} // namespace ros2_bridge

#endif // ROS2_BRIDGE_GRPC_SERVER_HPP