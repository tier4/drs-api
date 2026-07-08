cmake_minimum_required(VERSION 3.14)
project(ros2_bridge)

if(CMAKE_COMPILER_IS_GNUCXX OR CMAKE_CXX_COMPILER_ID MATCHES "Clang")
  add_compile_options(-Wall -Wextra -Wpedantic)
endif()

# find dependencies
find_package(ament_cmake REQUIRED)
find_package(rclcpp REQUIRED)
find_package(std_msgs REQUIRED)
find_package(geometry_msgs REQUIRED)
find_package(sensor_msgs REQUIRED)
find_package(std_srvs REQUIRED)
find_package(proto_recorder_msgs REQUIRED)
find_package(OpenCV REQUIRED)

# Find gRPC and Protobuf
find_package(PkgConfig REQUIRED)
pkg_check_modules(PROTOBUF REQUIRED protobuf)
pkg_check_modules(GRPC REQUIRED grpc++)

# Also try to find them with CONFIG
find_package(Protobuf QUIET)
find_package(gRPC QUIET)

# Proto file generation
set(PROTO_GENERATED_DIR ${CMAKE_CURRENT_SOURCE_DIR}/gen)
set(GENERATE_PROTO_SCRIPT ${CMAKE_CURRENT_SOURCE_DIR}/../../scripts/generate-proto.sh)

# Check if proto files need to be generated
if(NOT EXISTS ${PROTO_GENERATED_DIR}/drs/ros2bridge/v1/common.pb.h)
  message(STATUS "Proto files not found. Generating...")
  execute_process(
    COMMAND ${GENERATE_PROTO_SCRIPT} --cpp-only
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}/../..
    RESULT_VARIABLE PROTO_GEN_RESULT
  )
  if(NOT PROTO_GEN_RESULT EQUAL 0)
    message(FATAL_ERROR "Failed to generate proto files. Please run: ${GENERATE_PROTO_SCRIPT} --cpp-only")
  endif()
endif()

# Add generated proto files
file(GLOB_RECURSE PROTO_SOURCES
  "${PROTO_GENERATED_DIR}/*.pb.cc"
  "${PROTO_GENERATED_DIR}/*.grpc.pb.cc"
)

include_directories(
  include
  ${PROTO_GENERATED_DIR}
  ${CMAKE_CURRENT_BINARY_DIR}
)

# Create executable
add_executable(ros2_bridge_node
  src/main.cpp
  src/grpc_server.cpp
  src/sensing_handler.cpp
  src/recording_handler.cpp
  ${PROTO_SOURCES}
)

ament_target_dependencies(ros2_bridge_node
  rclcpp
  std_msgs
  geometry_msgs
  sensor_msgs
  std_srvs
  proto_recorder_msgs
)

# Link libraries
target_include_directories(ros2_bridge_node PRIVATE
  ${PROTOBUF_INCLUDE_DIRS}
  ${GRPC_INCLUDE_DIRS}
  ${OpenCV_INCLUDE_DIRS}
)

target_link_libraries(ros2_bridge_node
  ${PROTOBUF_LIBRARIES}
  ${GRPC_LIBRARIES}
  ${OpenCV_LIBRARIES}
  grpc++
  protobuf
)

# Install
install(TARGETS
  ros2_bridge_node
  DESTINATION lib/${PROJECT_NAME}
)

install(DIRECTORY
  include/
  DESTINATION include
)

if(BUILD_TESTING)
  find_package(ament_lint_auto REQUIRED)
  ament_lint_auto_find_test_dependencies()
endif()

ament_package()
