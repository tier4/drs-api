import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { 
  HardDrive, 
  Activity, 
  Cpu, 
  Clock, 
  MoreVertical,
  CheckCircle2,
  AlertCircle,
  AlertTriangle,
  XCircle,
  Circle
} from "lucide-react"

export interface Module {
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
  ptpStatus?: {
    gmPresent: boolean
    offsetNs: number
  }
}

interface ModuleStatusProps {
  modules: Module[]
  onRestartSensors?: (hostname: string) => void
  onRestartMachine?: (hostname: string) => void
  onShutdownMachine?: (hostname: string) => void
}

const getServiceIcon = (status?: string) => {
  if (!status) return { icon: Circle, color: 'text-gray-400' }
  switch (status.toLowerCase()) {
    case 'active':
      return { icon: CheckCircle2, color: 'text-green-500' }
    case 'inactive':
      return { icon: Circle, color: 'text-gray-400' }
    case 'failed':
      return { icon: XCircle, color: 'text-red-500' }
    default:
      return { icon: AlertCircle, color: 'text-yellow-500' }
  }
}

const getRecordingStatusColor = (status?: string) => {
  if (!status || status === '') return ''
  switch (status.toLowerCase()) {
    case 'recording':
      return 'text-green-500'
    case 'stopped':
      return 'text-gray-500'
    case 'paused':
      return 'text-yellow-500'
    default:
      return 'text-gray-500'
  }
}

const getDataStatusIcon = (status: Module['dataStatus']) => {
  if (!status || (status as any) === '') return null
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

const getPtpStatusColor = (offsetNs: number, gmPresent: boolean): string => {
  if (!gmPresent) return 'text-red-500'
  
  const absOffset = Math.abs(offsetNs)
  if (absOffset < 1000) return 'text-green-500' // < 1µs
  if (absOffset < 10000) return 'text-yellow-500' // < 10µs
  return 'text-red-500' // > 10µs
}

const formatBytes = (bytes: number): string => {
  const gb = bytes / (1024 * 1024 * 1024)
  if (gb >= 1000) {
    const tb = gb / 1024
    return `${tb.toFixed(1)}TB`
  }
  return `${gb.toFixed(0)}GB`
}

const getDiskUsageColor = (percentage: number): string => {
  if (percentage >= 90) return 'bg-red-500'
  if (percentage >= 80) return 'bg-yellow-500'
  return ''
}

export function ModuleStatus({ 
  modules, 
  onRestartSensors, 
  onRestartMachine, 
  onShutdownMachine 
}: ModuleStatusProps) {
  const [dialogState, setDialogState] = useState<{
    isOpen: boolean
    action: 'restart-sensors' | 'restart-machine' | 'shutdown-machine' | null
    hostname: string | null
  }>({
    isOpen: false,
    action: null,
    hostname: null
  })

  const handleAction = () => {
    if (!dialogState.hostname || !dialogState.action) return

    switch (dialogState.action) {
      case 'restart-sensors':
        onRestartSensors?.(dialogState.hostname)
        break
      case 'restart-machine':
        onRestartMachine?.(dialogState.hostname)
        break
      case 'shutdown-machine':
        onShutdownMachine?.(dialogState.hostname)
        break
    }
    
    setDialogState({ isOpen: false, action: null, hostname: null })
  }

  const openDialog = (action: 'restart-sensors' | 'restart-machine' | 'shutdown-machine', hostname: string) => {
    setDialogState({ isOpen: true, action, hostname })
  }

  const getDialogContent = () => {
    switch (dialogState.action) {
      case 'restart-sensors':
        return {
          title: 'Restart Sensors',
          description: `Are you sure you want to restart sensors on ${dialogState.hostname}? This will temporarily stop sensor data collection.`
        }
      case 'restart-machine':
        return {
          title: 'Restart Machine',
          description: `Are you sure you want to restart ${dialogState.hostname}? This will stop all services and reboot the machine.`
        }
      case 'shutdown-machine':
        return {
          title: 'Shutdown Machine',
          description: `Are you sure you want to shutdown ${dialogState.hostname}? This will stop all services and power off the machine.`
        }
      default:
        return { title: '', description: '' }
    }
  }

  return (
    <>
      <div className="space-y-6">
        {/* Module Cards */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {modules.map((module) => {
            const sensorIcon = getServiceIcon(module.services?.drs_sensor)
            const recorderIcon = getServiceIcon(module.services?.drs_recorder)
            const SensorIcon = sensorIcon.icon
            const RecorderIcon = recorderIcon.icon
            const dataStatusInfo = getDataStatusIcon(module.dataStatus)
            const DataStatusIcon = dataStatusInfo?.icon

            return (
              <Card key={module.hostname}>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div>
                      <CardTitle className="text-lg">{module.hostname.toUpperCase()}</CardTitle>
                      {module.moduleId && (
                        <p className="text-xs text-muted-foreground mt-1">ID: {module.moduleId}</p>
                      )}
                    </div>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon" className="h-8 w-8">
                          <MoreVertical className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => openDialog('restart-sensors', module.hostname)}>
                          Restart sensors
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => openDialog('restart-machine', module.hostname)}>
                          Restart machine
                        </DropdownMenuItem>
                        <DropdownMenuItem 
                          onClick={() => openDialog('shutdown-machine', module.hostname)}
                          className="text-destructive"
                        >
                          Shutdown machine
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  {/* Services */}
                  {module.services && (
                    <div>
                      <div className="flex items-center gap-2 mb-2">
                        <Cpu className="h-4 w-4 text-muted-foreground" />
                        <span className="text-sm font-medium">Services</span>
                      </div>
                      <div className="grid grid-cols-2 gap-2 ml-6">
                        <div className="flex items-center gap-2">
                          <SensorIcon className={`h-4 w-4 ${sensorIcon.color}`} />
                          <span className="text-sm">Sensor</span>
                        </div>
                        <div className="flex items-center gap-2">
                          <RecorderIcon className={`h-4 w-4 ${recorderIcon.color}`} />
                          <span className="text-sm">Recorder</span>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Recording & Data Status */}
                  {module.recordingStatus && (
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Activity className="h-4 w-4 text-muted-foreground" />
                        <span className="text-sm font-medium">Recording</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className={`text-sm font-medium ${getRecordingStatusColor(module.recordingStatus)}`}>
                          {module.recordingStatus.charAt(0).toUpperCase() + module.recordingStatus.slice(1)}
                        </span>
                        {DataStatusIcon && (
                          <DataStatusIcon className={`h-4 w-4 ${dataStatusInfo.color}`} />
                        )}
                      </div>
                    </div>
                  )}

                  {/* Time Sync */}
                  {module.ptpStatus && (
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Clock className="h-4 w-4 text-muted-foreground" />
                        <span className="text-sm font-medium">Time Sync</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className={`text-sm font-medium ${getPtpStatusColor(module.ptpStatus.offsetNs, module.ptpStatus.gmPresent)}`}>
                          {module.ptpStatus.gmPresent ? 'Synced' : 'No GM'}
                        </span>
                      </div>
                    </div>
                  )}

                  {/* Disk Usage */}
                  <div>
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center gap-2">
                        <HardDrive className="h-4 w-4 text-muted-foreground" />
                        <span className="text-sm font-medium">Disk Usage</span>
                      </div>
                      <span className="text-sm font-medium">{module.diskUsagePercentage.toFixed(0)}%</span>
                    </div>
                    <Progress 
                      value={module.diskUsagePercentage} 
                      className={`h-2 ${getDiskUsageColor(module.diskUsagePercentage)}`}
                    />
                    <div className="flex justify-between mt-1">
                      <span className="text-xs text-muted-foreground">
                        {formatBytes(module.diskTotalBytes - module.diskFreeBytes)} used
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {formatBytes(module.diskTotalBytes)} total
                      </span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      </div>

      <AlertDialog open={dialogState.isOpen} onOpenChange={(open) => !open && setDialogState({ isOpen: false, action: null, hostname: null })}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{getDialogContent().title}</AlertDialogTitle>
            <AlertDialogDescription>
              {getDialogContent().description}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleAction}>Continue</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}