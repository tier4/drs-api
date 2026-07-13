#ifndef ROS2_BRIDGE_LIDAR_CAMERA_PROJECTOR_HPP_
#define ROS2_BRIDGE_LIDAR_CAMERA_PROJECTOR_HPP_

#include <opencv2/core.hpp>
#include <rclcpp/rclcpp.hpp>

#include <sensor_msgs/msg/camera_info.hpp>

#include <tf2_ros/buffer.h>
#include <tf2_ros/transform_listener.h>

#include <memory>
#include <mutex>
#include <optional>
#include <string>
#include <unordered_map>
#include <vector>

namespace ros2_bridge
{

// Projects LiDAR points (already extracted from a PointCloud2, in the
// LiDAR's own sensor frame) onto a camera image, mirroring the "distort"
// mode of my-rosbag-util's tools/lidar-camera/project_lidar_to_cam.py:
// the original (still-distorted) image is left untouched and only the
// point projection itself accounts for the camera's lens distortion.
// Points are colored by intensity with a JET colormap and alpha-composited
// onto the image in-place.
class LidarCameraProjector
{
public:
  explicit LidarCameraProjector(rclcpp::Node::SharedPtr node);

  // Returns true if lidar_frame -> camera_optical_frame currently resolves
  // in the TF tree. Used to validate the fixed position->camera number
  // table (see SensingHandler) against the live TF tree without deriving
  // the mapping from it.
  bool hasTransform(
    const std::string & lidar_frame, const std::string & camera_optical_frame) const;

  // Projects points_lidar_frame (parallel to intensities) onto image using
  // camera_info's intrinsics and the lidar_frame -> camera_optical_frame
  // TF transform, compositing colored points in-place. Points with
  // camera-frame z <= 0.1 or that land outside image bounds are dropped.
  // Returns false if the transform can't be resolved or camera_info's
  // distortion_model is unrecognized; the caller should treat either as
  // has_data:false rather than as an exception.
  bool projectOntoImage(
    const std::vector<cv::Point3f> & points_lidar_frame, const std::vector<float> & intensities,
    const sensor_msgs::msg::CameraInfo & camera_info, const std::string & lidar_frame,
    const std::string & camera_optical_frame, cv::Mat & image) const;

private:
  // Resolves (and caches, read-only after first success) the lidar_frame ->
  // camera_optical_frame transform. The pairing is fixed vehicle hardware,
  // so a transform that resolved once is assumed valid for the node's
  // lifetime.
  std::optional<tf2::Transform> resolveTransform(
    const std::string & lidar_frame, const std::string & camera_optical_frame) const;

  rclcpp::Node::SharedPtr node_;
  std::shared_ptr<tf2_ros::Buffer> tf_buffer_;
  std::shared_ptr<tf2_ros::TransformListener> tf_listener_;

  mutable std::mutex cache_mutex_;
  mutable std::unordered_map<std::string, tf2::Transform> resolved_transform_cache_;
};

}  // namespace ros2_bridge

#endif  // ROS2_BRIDGE_LIDAR_CAMERA_PROJECTOR_HPP_
