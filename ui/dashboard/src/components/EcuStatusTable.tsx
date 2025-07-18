import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Progress } from "@/components/ui/progress"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

export interface EcuModule {
  hostname: string
  moduleId?: string
  services?: {
    drs_sensor?: string
    drs_recorder?: string
  }
  recordingStatus?: string
  dataStatus: 'OK' | 'WARN' | 'ERROR'
  diskUsagePercentage: number
  diskFreeBytes: number
  diskTotalBytes: number
}

interface EcuStatusTableProps {
  modules: EcuModule[]
  onRestartSensors?: (hostname: string) => void
  onRestartMachine?: (hostname: string) => void
  onShutdownMachine?: (hostname: string) => void
}


const getServiceStatusIcon = (status?: string) => {
  if (!status) return '-'
  switch (status.toLowerCase()) {
    case 'active':
      return <span className="text-green-600">● Active</span>
    case 'inactive':
      return <span className="text-gray-400">○ Inactive</span>
    case 'failed':
      return <span className="text-red-600">○ Failed</span>
    default:
      return status
  }
}

const getRecordingStatusIcon = (status?: string) => {
  if (!status) return '-'
  switch (status.toLowerCase()) {
    case 'recording':
      return <span className="text-green-600">● Recording</span>
    case 'stopped':
      return <span className="text-gray-400">○ Stopped</span>
    case 'paused':
      return <span className="text-yellow-600">⏸ Paused</span>
    default:
      return status
  }
}

const getDataStatusIcon = (status: EcuModule['dataStatus']) => {
  switch (status) {
    case 'OK':
      return <span className="text-green-600">✓ OK</span>
    case 'WARN':
      return <span className="text-yellow-600">⚠ WARN</span>
    case 'ERROR':
      return <span className="text-red-600">✗ ERROR</span>
    default:
      return status
  }
}

const formatBytes = (bytes: number): string => {
  const gb = bytes / (1024 * 1024 * 1024)
  if (gb >= 1000) {
    const tb = gb / 1024
    return `${tb.toFixed(0)}TB`
  }
  return `${gb.toFixed(0)}GB`
}

export function EcuStatusTable({ 
  modules, 
  onRestartSensors, 
  onRestartMachine, 
  onShutdownMachine 
}: EcuStatusTableProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Module</TableHead>
            <TableHead>Module ID</TableHead>
            <TableHead>Services</TableHead>
            <TableHead>Recording</TableHead>
            <TableHead>Data Status</TableHead>
            <TableHead>Disk Usage</TableHead>
            <TableHead className="w-[50px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {modules.map((module) => (
            <TableRow key={module.hostname}>
              <TableCell className="font-medium">{module.hostname}</TableCell>
              <TableCell className="text-sm text-muted-foreground">
                {module.moduleId || '-'}
              </TableCell>
              <TableCell>
                {module.services ? (
                  <div className="space-y-1">
                    {module.services.drs_sensor && (
                      <div className="text-sm">
                        <span className="text-muted-foreground">sensor</span> {getServiceStatusIcon(module.services.drs_sensor)}
                      </div>
                    )}
                    {module.services.drs_recorder && (
                      <div className="text-sm">
                        <span className="text-muted-foreground">recorder</span> {getServiceStatusIcon(module.services.drs_recorder)}
                      </div>
                    )}
                  </div>
                ) : (
                  <span className="text-muted-foreground">-</span>
                )}
              </TableCell>
              <TableCell>
                <div className="text-sm">
                  {getRecordingStatusIcon(module.recordingStatus)}
                </div>
              </TableCell>
              <TableCell>
                <div className="text-sm">
                  {getDataStatusIcon(module.dataStatus)}
                </div>
              </TableCell>
              <TableCell>
                <div className="flex items-center space-x-2">
                  <Progress value={module.diskUsagePercentage} className="w-[100px]" />
                  <span className="text-sm text-muted-foreground">
                    {module.diskUsagePercentage.toFixed(0)}%
                  </span>
                </div>
                <div className="text-xs text-muted-foreground mt-1">
                  {formatBytes(module.diskTotalBytes - module.diskFreeBytes)}/{formatBytes(module.diskTotalBytes)}
                </div>
              </TableCell>
              <TableCell>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <button className="text-muted-foreground hover:text-foreground p-2">
                      ⋮
                    </button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent>
                    <DropdownMenuItem onClick={() => onRestartSensors?.(module.hostname)}>
                      Restart sensors
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={() => onRestartMachine?.(module.hostname)}>
                      Restart machine
                    </DropdownMenuItem>
                    <DropdownMenuItem 
                      onClick={() => onShutdownMachine?.(module.hostname)}
                      className="text-destructive"
                    >
                      Shutdown machine
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}