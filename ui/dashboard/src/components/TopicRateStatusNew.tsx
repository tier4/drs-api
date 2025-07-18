import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { AlertCircle, AlertTriangle, CheckCircle2, Activity } from "lucide-react"
import type { ModuleTopicStatus } from "./TopicRateStatus"

interface TopicRateStatusNewProps {
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

const getRateIndicator = (actualRate: number, topicName: string) => {
  // Expected rates based on topic type
  const expectedRates: Record<string, { min: number, max: number }> = {
    camera: { min: 20, max: 40 },
    lidar: { min: 8, max: 12 },
    imu: { min: 80, max: 120 },
    gps: { min: 0.5, max: 2 },
    radar: { min: 10, max: 20 },
    vehicle: { min: 40, max: 60 }
  }

  let expected = null
  for (const [key, range] of Object.entries(expectedRates)) {
    if (topicName.toLowerCase().includes(key)) {
      expected = range
      break
    }
  }

  if (!expected) return null

  const percentage = ((actualRate - expected.min) / (expected.max - expected.min)) * 100
  return {
    percentage: Math.max(0, Math.min(100, percentage)),
    expected
  }
}

export function TopicRateStatusNew({ moduleTopicStatuses }: TopicRateStatusNewProps) {
  // Calculate summary statistics
  const totalTopics = moduleTopicStatuses.reduce((acc, m) => acc + m.topics.length, 0)
  const errorTopics = moduleTopicStatuses.reduce((acc, m) => 
    acc + m.topics.filter(t => t.status === 'ERROR').length, 0
  )
  const warnTopics = moduleTopicStatuses.reduce((acc, m) => 
    acc + m.topics.filter(t => t.status === 'WARN').length, 0
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
      <div className="grid gap-4 md:grid-cols-2">
        {moduleTopicStatuses.map((module) => {
          const moduleErrors = module.topics.filter(t => t.status === 'ERROR').length
          const moduleWarnings = module.topics.filter(t => t.status === 'WARN').length

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
                    const rateInfo = getRateIndicator(topic.rateHz, topic.topicName)

                    return (
                      <div
                        key={topic.topicName}
                        className="flex items-center justify-between p-2 rounded-lg bg-muted/50 hover:bg-muted/70 transition-colors"
                      >
                        <div className="flex items-center gap-2 flex-1 min-w-0">
                          <StatusIcon className={`h-4 w-4 flex-shrink-0 ${statusInfo.color}`} />
                          <span className="text-sm font-mono truncate" title={topic.topicName}>
                            {topic.topicName}
                          </span>
                        </div>
                        <div className="flex items-center gap-3">
                          <div className="text-right">
                            <div className="text-sm font-medium">
                              {formatRate(topic.rateHz)} Hz
                            </div>
                            {rateInfo && (
                              <div className="text-xs text-muted-foreground">
                                Expected: {rateInfo.expected.min}-{rateInfo.expected.max} Hz
                              </div>
                            )}
                          </div>
                          <Badge 
                            variant={getStatusColor(topic.status)}
                            className="text-xs min-w-[50px] justify-center"
                          >
                            {topic.status}
                          </Badge>
                        </div>
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