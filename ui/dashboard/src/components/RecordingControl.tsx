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
  status: 'recording' | 'stopped' | 'paused'
  active: boolean
  hardwareId?: string
}

interface RecordingControlProps {
  recordingStatuses: RecordingStatus[]
  globalRecordingEnabled: boolean
  onGlobalRecordingToggle: (enabled: boolean) => void
}

const getRecordingStatusColor = (status: RecordingStatus['status']): 'default' | 'secondary' | 'destructive' => {
  switch (status) {
    case 'recording':
      return 'default'
    case 'paused':
      return 'secondary'
    case 'stopped':
      return 'destructive'
    default:
      return 'destructive'
  }
}

const getRecordingStatusText = (status: RecordingStatus['status']): string => {
  switch (status) {
    case 'recording':
      return 'Recording'
    case 'paused':
      return 'Paused'
    case 'stopped':
      return 'Stopped'
    default:
      return 'Unknown'
  }
}

export function RecordingControl({ 
  recordingStatuses, 
  globalRecordingEnabled,
  onGlobalRecordingToggle
}: RecordingControlProps) {
  const isAnyRecording = recordingStatuses.some(r => r.status === 'recording')
  const isAnyPaused = recordingStatuses.some(r => r.status === 'paused')

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
          <div className="flex items-center space-x-2">
            <span className="text-sm font-medium">Recording Status:</span>
            <Badge variant={isAnyRecording ? 'default' : 'secondary'}>
              {isAnyRecording ? 'RECORDING' : isAnyPaused ? 'PAUSED' : 'STOPPED'}
            </Badge>
          </div>

          {/* Per-Module Recording Status */}
          <div>
            <h4 className="text-sm font-medium mb-2">Module Status</h4>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Module</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Active</TableHead>
                  <TableHead>Hardware ID</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {recordingStatuses.map((recording) => (
                  <TableRow key={recording.hostname}>
                    <TableCell className="font-medium">{recording.hostname}</TableCell>
                    <TableCell>
                      <Badge variant={getRecordingStatusColor(recording.status)}>
                        {getRecordingStatusText(recording.status)}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant={recording.active ? 'default' : 'secondary'}>
                        {recording.active ? 'Active' : 'Inactive'}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {recording.hardwareId || 'N/A'}
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