import { useEffect, useRef, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { AlertCircle, AlertTriangle, CheckCircle2, Activity, Eye, Pause, Play } from 'lucide-react'
import { NavSatFixPreview } from '@/components/NavSatFixPreview'
import { CameraPreview } from '@/components/CameraPreview'
import { LidarCameraProjectionPreview } from '@/components/LidarCameraProjectionPreview'

// message_type values with a built preview. Topics with any other type show
// no inspect icon.
const PREVIEWABLE_TYPES = {
  'sensor_msgs/msg/NavSatFix': 'navsat',
  'sensor_msgs/msg/CompressedImage': 'camera',
  // Both map to 'lidar-camera-projection': seyond/msg/SeyondScan is what
  // real hardware reports (verified via `ros2 topic type` on
  // /sensing/lidar/front/seyond_packets); nebula_msgs/msg/NebulaPackets is
  // kept for other decoder configs that use it.
  'seyond/msg/SeyondScan': 'lidar-camera-projection',
  'nebula_msgs/msg/NebulaPackets': 'lidar-camera-projection',
} as const

type PreviewKind = (typeof PREVIEWABLE_TYPES)[keyof typeof PREVIEWABLE_TYPES]

function getPreviewKind(messageType: string): PreviewKind | null {
  return PREVIEWABLE_TYPES[messageType as keyof typeof PREVIEWABLE_TYPES] ?? null
}

// An open preview (camera or nav_sat_fix) keeps polling the backend every 5s
// as long as refresh is enabled - for camera topics this also keeps a ROS2
// subscription (and, for remote cameras, a cross-ECU network relay) alive.
// To avoid an operator leaving one open and forgetting about it, refresh
// auto-disables after this long for every preview kind. For camera topics
// the backend then unsubscribes ~30s after that (see
// sweepIdleCameraSubscriptions in sensing_handler.cpp).
const PREVIEW_AUTO_OFF_MS = 60_000

export interface TopicStatus {
  topicName: string
  messageType: string
  rateHz: number
  status: 'OK' | 'WARN' | 'ERROR'
}

export interface ModuleTopicStatus {
  hostname: string
  topics: TopicStatus[]
}

interface TopicRateStatusProps {
  moduleTopicStatuses: ModuleTopicStatus[]
}

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'OK':
      return { icon: CheckCircle2, color: 'text-green-500' }
    case 'WARN':
      return { icon: AlertTriangle, color: 'text-yellow-500' }
    case 'ERROR':
      return { icon: AlertCircle, color: 'text-red-500' }
    default:
      return { icon: AlertCircle, color: 'text-gray-500' }
  }
}

const getStatusColor = (status: string): 'default' | 'secondary' | 'destructive' => {
  switch (status) {
    case 'OK':
      return 'default'
    case 'WARN':
      return 'secondary'
    case 'ERROR':
      return 'destructive'
    default:
      return 'destructive'
  }
}

const formatRate = (rate: number): string => {
  if (rate === 0) return '0.0'
  if (rate >= 100) return rate.toFixed(0)
  return rate.toFixed(1)
}

export function TopicRateStatus({ moduleTopicStatuses }: TopicRateStatusProps) {
  const [expandedKeys, setExpandedKeys] = useState<Set<string>>(new Set())
  const [refreshEnabled, setRefreshEnabled] = useState<Record<string, boolean>>({})
  const previewAutoOffTimers = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map())

  useEffect(() => {
    const timers = previewAutoOffTimers.current
    return () => {
      timers.forEach((timer) => clearTimeout(timer))
      timers.clear()
    }
  }, [])

  const schedulePreviewAutoOff = (key: string) => {
    const existing = previewAutoOffTimers.current.get(key)
    if (existing) clearTimeout(existing)
    const timer = setTimeout(() => {
      setRefreshEnabled((prev) => ({ ...prev, [key]: false }))
      previewAutoOffTimers.current.delete(key)
    }, PREVIEW_AUTO_OFF_MS)
    previewAutoOffTimers.current.set(key, timer)
  }

  const clearPreviewAutoOff = (key: string) => {
    const existing = previewAutoOffTimers.current.get(key)
    if (existing) {
      clearTimeout(existing)
      previewAutoOffTimers.current.delete(key)
    }
  }

  const toggleExpand = (key: string) => {
    setExpandedKeys((prev) => {
      const next = new Set(prev)
      if (next.has(key)) {
        next.delete(key)
        clearPreviewAutoOff(key)
      } else {
        next.add(key)
        setRefreshEnabled((prevEnabled) => {
          const nextEnabled = prevEnabled[key] ?? true
          if (nextEnabled) {
            schedulePreviewAutoOff(key)
          }
          return { ...prevEnabled, [key]: nextEnabled }
        })
      }
      return next
    })
  }

  const toggleRefresh = (key: string) => {
    setRefreshEnabled((prev) => {
      const nextValue = !(prev[key] ?? true)
      if (nextValue) {
        schedulePreviewAutoOff(key)
      } else {
        clearPreviewAutoOff(key)
      }
      return { ...prev, [key]: nextValue }
    })
  }

  // Calculate summary statistics
  const totalTopics = moduleTopicStatuses.reduce((acc, m) => acc + m.topics.length, 0)
  const errorTopics = moduleTopicStatuses.reduce(
    (acc, m) => acc + m.topics.filter((t) => t.status === 'ERROR').length,
    0,
  )
  const warnTopics = moduleTopicStatuses.reduce(
    (acc, m) => acc + m.topics.filter((t) => t.status === 'WARN').length,
    0,
  )

  return (
    <div className="space-y-6">
      {/* Summary Card */}
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="text-lg flex items-center gap-2">
              <Activity className="h-5 w-5" />
              Topic Rate Monitor
            </CardTitle>
            <div className="flex gap-2">
              {errorTopics > 0 && (
                <Badge variant="destructive" className="text-xs">
                  {errorTopics} Error{errorTopics > 1 ? 's' : ''}
                </Badge>
              )}
              {warnTopics > 0 && (
                <Badge variant="secondary" className="text-xs">
                  {warnTopics} Warning{warnTopics > 1 ? 's' : ''}
                </Badge>
              )}
              {errorTopics === 0 && warnTopics === 0 && (
                <Badge variant="default" className="text-xs">
                  All Healthy
                </Badge>
              )}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="text-sm text-muted-foreground">
            Monitoring {totalTopics} topics across {moduleTopicStatuses.length} modules
          </div>
        </CardContent>
      </Card>

      {/* Module Cards */}
      <div className="space-y-4">
        {moduleTopicStatuses.map((module) => {
          const moduleErrors = module.topics.filter((t) => t.status === 'ERROR').length
          const moduleWarnings = module.topics.filter((t) => t.status === 'WARN').length

          return (
            <Card key={module.hostname}>
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base">{module.hostname.toUpperCase()}</CardTitle>
                  <div className="flex gap-1">
                    {moduleErrors > 0 && (
                      <Badge variant="destructive" className="text-xs">
                        {moduleErrors} Error{moduleErrors > 1 ? 's' : ''}
                      </Badge>
                    )}
                    {moduleWarnings > 0 && (
                      <Badge variant="secondary" className="text-xs">
                        {moduleWarnings} Warning{moduleWarnings > 1 ? 's' : ''}
                      </Badge>
                    )}
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  {module.topics.map((topic) => {
                    const statusInfo = getStatusIcon(topic.status)
                    const StatusIcon = statusInfo.icon
                    const previewKind = getPreviewKind(topic.messageType)
                    const key = `${module.hostname}:${topic.topicName}`
                    const isExpanded = expandedKeys.has(key)
                    const isRefreshing = refreshEnabled[key] ?? true

                    return (
                      <div key={topic.topicName} className="rounded-lg bg-muted/50">
                        <div className="flex items-center justify-between p-2 hover:bg-muted/70 transition-colors">
                          <div className="flex items-center gap-2 flex-1 min-w-0">
                            <StatusIcon className={`h-4 w-4 flex-shrink-0 ${statusInfo.color}`} />
                            <span className="text-sm font-mono truncate" title={topic.topicName}>
                              {topic.topicName}
                            </span>
                          </div>
                          <div className="flex items-center gap-3">
                            <div className="text-sm font-medium">{formatRate(topic.rateHz)} Hz</div>
                            <Badge
                              variant={getStatusColor(topic.status)}
                              className="text-xs min-w-[50px] justify-center"
                            >
                              {topic.status}
                            </Badge>
                            {previewKind && (
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-6 w-6"
                                title="Inspect data"
                                onClick={() => toggleExpand(key)}
                              >
                                <Eye
                                  className={`h-4 w-4 ${isExpanded ? 'text-primary' : 'text-muted-foreground'}`}
                                />
                              </Button>
                            )}
                          </div>
                        </div>
                        {isExpanded && previewKind && (
                          <div className="p-2 pt-0 space-y-2">
                            <div className="flex justify-end">
                              <Button
                                variant="outline"
                                size="sm"
                                className="h-6 text-xs gap-1"
                                onClick={() => toggleRefresh(key)}
                              >
                                {isRefreshing ? (
                                  <>
                                    <Pause className="h-3 w-3" /> Pause
                                  </>
                                ) : (
                                  <>
                                    <Play className="h-3 w-3" /> Resume
                                  </>
                                )}
                              </Button>
                            </div>
                            {previewKind === 'navsat' && (
                              <NavSatFixPreview hostname={module.hostname} enabled={isRefreshing} />
                            )}
                            {previewKind === 'camera' && (
                              <CameraPreview
                                hostname={module.hostname}
                                topicName={topic.topicName}
                                enabled={isRefreshing}
                              />
                            )}
                            {previewKind === 'lidar-camera-projection' && (
                              <LidarCameraProjectionPreview
                                hostname={module.hostname}
                                topicName={topic.topicName}
                                enabled={isRefreshing}
                              />
                            )}
                          </div>
                        )}
                      </div>
                    )
                  })}
                </div>
              </CardContent>
            </Card>
          )
        })}
      </div>
    </div>
  )
}
