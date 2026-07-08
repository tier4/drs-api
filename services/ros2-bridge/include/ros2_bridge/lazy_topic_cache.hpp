#ifndef ROS2_BRIDGE_LAZY_TOPIC_CACHE_HPP_
#define ROS2_BRIDGE_LAZY_TOPIC_CACHE_HPP_

#include <rclcpp/rclcpp.hpp>

#include <chrono>
#include <mutex>
#include <string>
#include <unordered_map>

namespace ros2_bridge
{

// Lazily subscribes to a topic on its first request, caches the latest
// message per topic, and drops topics that haven't been requested within
// idle_timeout (via periodic sweepIdle() calls, e.g. from a wall timer).
// Shared by any "preview on demand" RPC (camera, point cloud, ...) so the
// subscribe/cache/idle-unsubscribe pattern isn't duplicated per message type.
template <typename MsgT>
class LazyTopicCache
{
public:
  LazyTopicCache(
    rclcpp::Node::SharedPtr node, std::chrono::steady_clock::duration idle_timeout,
    int qos_depth = 10)
  : node_(node), idle_timeout_(idle_timeout), qos_depth_(qos_depth)
  {
  }

  // Subscribes to topic_name on first call, marks it as actively viewed so
  // sweepIdle() won't drop it, and returns the latest cached message for
  // that topic, or nullptr if none has arrived yet.
  typename MsgT::SharedPtr getOrSubscribe(const std::string & topic_name)
  {
    std::lock_guard<std::mutex> lock(mutex_);

    auto [sub_it, inserted] = subs_.try_emplace(topic_name, nullptr);
    if (inserted) {
      sub_it->second = node_->create_subscription<MsgT>(
        topic_name, qos_depth_, [this, topic_name](const typename MsgT::SharedPtr msg) {
          std::lock_guard<std::mutex> cb_lock(mutex_);
          cached_[topic_name] = msg;
        });
      RCLCPP_INFO(node_->get_logger(), "Lazily subscribed to topic: %s", topic_name.c_str());
    }

    last_request_[topic_name] = std::chrono::steady_clock::now();

    auto it = cached_.find(topic_name);
    return it != cached_.end() ? it->second : nullptr;
  }

  // Returns the number of publishers currently advertising topic_name, so
  // callers can distinguish "no publisher" from "publisher exists but no
  // message yet". 0 if the topic has never been subscribed to.
  size_t publisherCount(const std::string & topic_name)
  {
    std::lock_guard<std::mutex> lock(mutex_);
    auto it = subs_.find(topic_name);
    if (it == subs_.end()) {
      return 0;
    }
    return node_->count_publishers(topic_name);
  }

  // Unsubscribes and drops cached state for any topic not requested within
  // idle_timeout. Call periodically (e.g. from a wall timer).
  void sweepIdle()
  {
    std::lock_guard<std::mutex> lock(mutex_);
    const auto now = std::chrono::steady_clock::now();

    for (auto it = last_request_.begin(); it != last_request_.end();) {
      const std::string & topic_name = it->first;
      if (now - it->second > idle_timeout_) {
        RCLCPP_INFO(node_->get_logger(), "Unsubscribing idle topic: %s", topic_name.c_str());
        subs_.erase(topic_name);
        cached_.erase(topic_name);
        it = last_request_.erase(it);
      } else {
        ++it;
      }
    }
  }

private:
  rclcpp::Node::SharedPtr node_;
  std::chrono::steady_clock::duration idle_timeout_;
  int qos_depth_;

  std::mutex mutex_;
  std::unordered_map<std::string, typename rclcpp::Subscription<MsgT>::SharedPtr> subs_;
  std::unordered_map<std::string, typename MsgT::SharedPtr> cached_;
  std::unordered_map<std::string, std::chrono::steady_clock::time_point> last_request_;
};

}  // namespace ros2_bridge

#endif  // ROS2_BRIDGE_LAZY_TOPIC_CACHE_HPP_
