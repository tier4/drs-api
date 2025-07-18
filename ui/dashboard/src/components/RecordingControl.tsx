import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Switch } from "@/components/ui/switch"
import { Badge } from "@/components/ui/badge"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

export interface RecordingStatus {
  hostname: string
  recording_status: 'recording' | 'stopped'
  health_status: 'OK' | 'WARN' | 'ERROR'
  hardware_id?: string
}

interface RecordingControlProps {
  recordingStatuses: RecordingStatus[]
  globalRecordingEnabled: boolean
  onGlobalRecordingToggle: (enabled: boolean) => void
}

const getRecordingStatusColor = (status: RecordingStatus['recording_status']): 'default' | 'secondary' | 'destructive' => {
  switch (status) {
    case 'recording':
      return 'default'
    case 'stopped':
      return 'destructive'
    default:
      return 'destructive'
  }
}

const getRecordingStatusText = (status: RecordingStatus['recording_status']): string => {
  switch (status) {
    case 'recording':
      return 'Recording'
    case 'stopped':
      return 'Stopped'
    default:
      return 'Unknown'
  }
}

const getHealthStatusColor = (status: RecordingStatus['health_status']): 'default' | 'secondary' | 'destructive' => {
  switch (status) {
    case 'OK':
      return 'default'
    case 'WARN':
      return 'secondary'
    case 'ERROR':
      return 'destructive'
    default:
      return 'secondary'
  }
}

export function RecordingControl({ 
  recordingStatuses, 
  globalRecordingEnabled,
  onGlobalRecordingToggle
}: RecordingControlProps) {
  const isAnyRecording = recordingStatuses.some(r => r.recording_status === 'recording')
  const hasAnyError = recordingStatuses.some(r => r.health_status === 'ERROR')
  const hasAnyWarn = recordingStatuses.some(r => r.health_status === 'WARN')

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <span>Recording Control</span>
          <div className="flex items-center space-x-2">
            <span className="text-sm font-normal">Record</span>
            <Switch 
              checked={globalRecordingEnabled} 
              onCheckedChange={onGlobalRecordingToggle}
            />
          </div>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {/* Recording Status Display */}
          <div className="flex items-center space-x-4">
            <div className="flex items-center space-x-2">
              <span className="text-sm font-medium">Recording Status:</span>
              <Badge variant={isAnyRecording ? 'default' : 'destructive'}>
                {isAnyRecording ? 'RECORDING' : 'STOPPED'}
              </Badge>
            </div>
            <div className="flex items-center space-x-2">
              <span className="text-sm font-medium">Health Status:</span>
              <Badge variant={hasAnyError ? 'destructive' : hasAnyWarn ? 'secondary' : 'default'}>
                {hasAnyError ? 'ERROR' : hasAnyWarn ? 'WARNING' : 'OK'}
              </Badge>
            </div>
          </div>

          {/* Per-Module Recording Status */}
          <div>
            <h4 className="text-sm font-medium mb-2">Module Status</h4>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Module</TableHead>
                  <TableHead>Recording Status</TableHead>
                  <TableHead>Health Status</TableHead>
                  <TableHead>Hardware ID</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {recordingStatuses.map((recording) => (
                  <TableRow key={recording.hostname}>
                    <TableCell className="font-medium">{recording.hostname}</TableCell>
                    <TableCell>
                      <Badge variant={getRecordingStatusColor(recording.recording_status)}>
                        {getRecordingStatusText(recording.recording_status)}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant={getHealthStatusColor(recording.health_status)}>
                        {recording.health_status}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {recording.hardware_id || 'N/A'}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}