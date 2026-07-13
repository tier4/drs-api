#include "ros2_bridge/lidar_camera_projector.hpp"

#include <opencv2/calib3d.hpp>
#include <opencv2/imgproc.hpp>

#include <tf2_geometry_msgs/tf2_geometry_msgs.hpp>

#include <tf2/LinearMath/Transform.h>
#include <tf2/exceptions.h>

#include <algorithm>
#include <cmath>
#include <string>
#include <vector>

namespace ros2_bridge
{

namespace
{

// Intensity normalization range for the JET colormap, matching
// project_lidar_to_cam.py's IMIN/IMAX.
constexpr float kIntensityMin = 0.0f;
constexpr float kIntensityMax = 40.0f;

// Point radius (pixels) and alpha-compositing opacity, matching
// project_lidar_to_cam.py's draw_points() defaults.
constexpr int kPointRadius = 1;
constexpr float kAlpha = 0.45f;

// Points behind or too close to the camera's optical center are discarded,
// matching project_lidar_to_cam.py's `front = p_col[:, 2] > 0.1` filter.
constexpr float kMinCameraFrameZ = 0.1f;

std::string transformCacheKey(
  const std::string & lidar_frame, const std::string & camera_optical_frame)
{
  return lidar_frame + "->" + camera_optical_frame;
}

cv::Mat buildCameraMatrix(const sensor_msgs::msg::CameraInfo & camera_info)
{
  cv::Mat k(3, 3, CV_64F);
  for (int row = 0; row < 3; ++row) {
    for (int col = 0; col < 3; ++col) {
      k.at<double>(row, col) = camera_info.k[static_cast<size_t>(row * 3 + col)];
    }
  }
  return k;
}

// Colors points by intensity using a JET colormap, matching
// project_lidar_to_cam.py's color_by_intensity(). Returns an Nx1 CV_8UC3
// (BGR) Mat, one row per intensity value.
cv::Mat colorByIntensity(const std::vector<float> & intensities)
{
  cv::Mat normalized(static_cast<int>(intensities.size()), 1, CV_8UC1);
  for (size_t i = 0; i < intensities.size(); ++i) {
    const float clamped =
      std::clamp((intensities[i] - kIntensityMin) / (kIntensityMax - kIntensityMin), 0.0f, 1.0f);
    normalized.at<uint8_t>(static_cast<int>(i), 0) =
      static_cast<uint8_t>(std::lround(clamped * 255.0f));
  }
  cv::Mat colors;
  cv::applyColorMap(normalized, colors, cv::COLORMAP_JET);
  return colors;
}

// Alpha-composites colored points onto image in-place, matching
// project_lidar_to_cam.py's draw_points(): each point overwrites a
// (2*radius+1)^2 pixel box in an overlay copy (so overlapping points let
// the last-drawn point win), then every touched pixel is blended once
// against the original image.
void drawPoints(
  cv::Mat & image, const std::vector<cv::Point2f> & image_points, const cv::Mat & colors,
  int radius, float alpha)
{
  cv::Mat overlay = image.clone();
  cv::Mat touched_mask = cv::Mat::zeros(image.size(), CV_8UC1);
  std::vector<cv::Point> touched_pixels;
  touched_pixels.reserve(
    image_points.size() * static_cast<size_t>((2 * radius + 1) * (2 * radius + 1)));

  const int width = image.cols;
  const int height = image.rows;

  for (size_t i = 0; i < image_points.size(); ++i) {
    const int u = static_cast<int>(std::lround(image_points[i].x));
    const int v = static_cast<int>(std::lround(image_points[i].y));
    const cv::Vec3b color = colors.at<cv::Vec3b>(static_cast<int>(i), 0);

    for (int dv = -radius; dv <= radius; ++dv) {
      const int vv = v + dv;
      if (vv < 0 || vv >= height) {
        continue;
      }
      for (int du = -radius; du <= radius; ++du) {
        const int uu = u + du;
        if (uu < 0 || uu >= width) {
          continue;
        }
        overlay.at<cv::Vec3b>(vv, uu) = color;
        if (touched_mask.at<uint8_t>(vv, uu) == 0) {
          touched_mask.at<uint8_t>(vv, uu) = 255;
          touched_pixels.emplace_back(uu, vv);
        }
      }
    }
  }

  for (const auto & pixel : touched_pixels) {
    cv::Vec3b & dst = image.at<cv::Vec3b>(pixel.y, pixel.x);
    const cv::Vec3b & src = overlay.at<cv::Vec3b>(pixel.y, pixel.x);
    for (int c = 0; c < 3; ++c) {
      dst[c] = static_cast<uint8_t>(std::lround(alpha * src[c] + (1.0f - alpha) * dst[c]));
    }
  }
}

}  // namespace

LidarCameraProjector::LidarCameraProjector(rclcpp::Node::SharedPtr node)
: node_(node),
  tf_buffer_(std::make_shared<tf2_ros::Buffer>(node_->get_clock())),
  tf_listener_(std::make_shared<tf2_ros::TransformListener>(*tf_buffer_))
{
}

std::optional<tf2::Transform> LidarCameraProjector::resolveTransform(
  const std::string & lidar_frame, const std::string & camera_optical_frame) const
{
  const std::string key = transformCacheKey(lidar_frame, camera_optical_frame);

  {
    std::lock_guard<std::mutex> lock(cache_mutex_);
    auto it = resolved_transform_cache_.find(key);
    if (it != resolved_transform_cache_.end()) {
      return it->second;
    }
  }

  try {
    const geometry_msgs::msg::TransformStamped transform_stamped =
      tf_buffer_->lookupTransform(camera_optical_frame, lidar_frame, tf2::TimePointZero);
    tf2::Transform transform;
    tf2::fromMsg(transform_stamped.transform, transform);

    std::lock_guard<std::mutex> lock(cache_mutex_);
    resolved_transform_cache_[key] = transform;
    return transform;
  } catch (const tf2::TransformException & ex) {
    RCLCPP_DEBUG(
      node_->get_logger(), "TF lookup %s -> %s failed: %s", lidar_frame.c_str(),
      camera_optical_frame.c_str(), ex.what());
    return std::nullopt;
  }
}

bool LidarCameraProjector::hasTransform(
  const std::string & lidar_frame, const std::string & camera_optical_frame) const
{
  return resolveTransform(lidar_frame, camera_optical_frame).has_value();
}

bool LidarCameraProjector::projectOntoImage(
  const std::vector<cv::Point3f> & points_lidar_frame, const std::vector<float> & intensities,
  const sensor_msgs::msg::CameraInfo & camera_info, const std::string & lidar_frame,
  const std::string & camera_optical_frame, cv::Mat & image) const
{
  const std::optional<tf2::Transform> transform =
    resolveTransform(lidar_frame, camera_optical_frame);
  if (!transform) {
    return false;
  }

  const std::string & model = camera_info.distortion_model;
  const bool is_equidistant = model == "equidistant";
  const bool is_standard_model = model == "rational_polynomial" || model == "plumb_bob";
  if (!is_equidistant && !is_standard_model) {
    RCLCPP_WARN(node_->get_logger(), "Unsupported camera distortion_model: %s", model.c_str());
    return false;
  }
  if (is_equidistant && camera_info.d.size() < 4) {
    RCLCPP_WARN(
      node_->get_logger(), "equidistant distortion_model requires at least 4 D coefficients");
    return false;
  }

  // Transform points into the camera optical frame and drop anything
  // behind/too close to the camera, matching project_lidar_to_cam.py's
  // `p_col = R_col_lr @ pts_lr + t_col_lr; front = p_col[:, 2] > 0.1`.
  std::vector<cv::Point3d> points_camera_frame;
  std::vector<float> visible_intensities;
  points_camera_frame.reserve(points_lidar_frame.size());
  visible_intensities.reserve(points_lidar_frame.size());
  for (size_t i = 0; i < points_lidar_frame.size(); ++i) {
    const cv::Point3f & p = points_lidar_frame[i];
    const tf2::Vector3 transformed = (*transform) * tf2::Vector3(p.x, p.y, p.z);
    if (transformed.z() <= kMinCameraFrameZ) {
      continue;
    }
    points_camera_frame.emplace_back(transformed.x(), transformed.y(), transformed.z());
    visible_intensities.push_back(intensities[i]);
  }

  if (points_camera_frame.empty()) {
    // Valid transform and camera_info, just nothing currently in view.
    return true;
  }

  const cv::Mat camera_matrix = buildCameraMatrix(camera_info);
  const cv::Mat zero_rvec = cv::Mat::zeros(3, 1, CV_64F);
  const cv::Mat zero_tvec = cv::Mat::zeros(3, 1, CV_64F);
  std::vector<cv::Point2d> image_points_d;

  if (is_equidistant) {
    cv::Mat dist_coeffs(4, 1, CV_64F);
    for (int i = 0; i < 4; ++i) {
      dist_coeffs.at<double>(i, 0) = camera_info.d[static_cast<size_t>(i)];
    }
    cv::fisheye::projectPoints(
      points_camera_frame, image_points_d, zero_rvec, zero_tvec, camera_matrix, dist_coeffs);
  } else {
    cv::Mat dist_coeffs(static_cast<int>(camera_info.d.size()), 1, CV_64F);
    for (size_t i = 0; i < camera_info.d.size(); ++i) {
      dist_coeffs.at<double>(static_cast<int>(i), 0) = camera_info.d[i];
    }
    cv::projectPoints(
      points_camera_frame, zero_rvec, zero_tvec, camera_matrix, dist_coeffs, image_points_d);
  }

  std::vector<cv::Point2f> image_points;
  image_points.reserve(image_points_d.size());
  for (const auto & p : image_points_d) {
    image_points.emplace_back(static_cast<float>(p.x), static_cast<float>(p.y));
  }

  const cv::Mat colors = colorByIntensity(visible_intensities);
  drawPoints(image, image_points, colors, kPointRadius, kAlpha);
  return true;
}

}  // namespace ros2_bridge
