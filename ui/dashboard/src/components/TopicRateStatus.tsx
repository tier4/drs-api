import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

export interface TopicStatus {
  topicName: string
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

const getTopicStatusColor = (status: TopicStatus['status']): 'default' | 'secondary' | 'destructive' => {
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

// Status is determined by the backend based on topic rate analysis

const formatRate = (rate: number): string => {
  if (rate === 0) return '0.0'
  return rate.toFixed(1)
}

export function TopicRateStatus({ moduleTopicStatuses }: TopicRateStatusProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Topic Rate Status</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          {moduleTopicStatuses.map((module) => (
            <div key={module.hostname} className="space-y-3">
              <h4 className="font-medium">{module.hostname}</h4>
              
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Topic</TableHead>
                    <TableHead>Rate</TableHead>
                    <TableHead>Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {module.topics.map((topic) => (
                    <TableRow key={topic.topicName}>
                      <TableCell className="font-medium text-sm">
                        {topic.topicName}
                      </TableCell>
                      <TableCell className="text-sm">
                        {formatRate(topic.rateHz)} Hz
                      </TableCell>
                      <TableCell>
                        <Badge variant={getTopicStatusColor(topic.status)}>
                          {topic.status}
                        </Badge>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}