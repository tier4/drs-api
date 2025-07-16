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

# Find gRPC and Protobuf
find_package(PkgConfig REQUIRED)
pkg_check_modules(PROTOBUF REQUIRED protobuf)
pkg_check_modules(GRPC REQUIRED grpc++)

# Also try to find them with CONFIG
find_package(Protobuf QUIET)
find_package(gRPC QUIET)

# Proto file generation
set(PROTO_DIR ${CMAKE_CURRENT_SOURCE_DIR}/../../proto)
set(PROTO_GENERATED_DIR ${CMAKE_CURRENT_BINARY_DIR}/gen)

# Find protoc and grpc_cpp_plugin
find_program(PROTOC protoc)
find_program(GRPC_CPP_PLUGIN grpc_cpp_plugin)

if(NOT PROTOC)
  message(FATAL_ERROR "protoc not found")
endif()

if(NOT GRPC_CPP_PLUGIN)
  message(FATAL_ERROR "grpc_cpp_plugin not found")
endif()

# Define proto files
set(PROTO_FILES
  ${PROTO_DIR}/drs/ros2bridge/v1/common.proto
  ${PROTO_DIR}/drs/ros2bridge/v1/sensing_service.proto
  ${PROTO_DIR}/drs/ros2bridge/v1/recording_service.proto
)

# Generate C++ files from proto
set(PROTO_SOURCES)
set(PROTO_HEADERS)

foreach(proto_file ${PROTO_FILES})
  get_filename_component(proto_name ${proto_file} NAME_WE)
  get_filename_component(proto_dir ${proto_file} DIRECTORY)
  file(RELATIVE_PATH proto_rel_dir ${PROTO_DIR} ${proto_dir})
  
  set(proto_gen_dir ${PROTO_GENERATED_DIR}/${proto_rel_dir})
  file(MAKE_DIRECTORY ${proto_gen_dir})
  
  set(proto_pb_cc ${proto_gen_dir}/${proto_name}.pb.cc)
  set(proto_pb_h ${proto_gen_dir}/${proto_name}.pb.h)
  set(proto_grpc_cc ${proto_gen_dir}/${proto_name}.grpc.pb.cc)
  set(proto_grpc_h ${proto_gen_dir}/${proto_name}.grpc.pb.h)
  
  list(APPEND PROTO_SOURCES ${proto_pb_cc} ${proto_grpc_cc})
  list(APPEND PROTO_HEADERS ${proto_pb_h} ${proto_grpc_h})
  
  add_custom_command(
    OUTPUT ${proto_pb_cc} ${proto_pb_h} ${proto_grpc_cc} ${proto_grpc_h}
    COMMAND ${PROTOC}
    ARGS -I ${PROTO_DIR}
         --cpp_out=${PROTO_GENERATED_DIR}
         --grpc_out=${PROTO_GENERATED_DIR}
         --plugin=protoc-gen-grpc=${GRPC_CPP_PLUGIN}
         ${proto_file}
    DEPENDS ${proto_file}
    COMMENT "Generating C++ code from ${proto_file}"
  )
endforeach()

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
)

target_link_libraries(ros2_bridge_node
  ${PROTOBUF_LIBRARIES}
  ${GRPC_LIBRARIES}
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