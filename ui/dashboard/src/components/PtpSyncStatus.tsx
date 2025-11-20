import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { AlertCircle, CheckCircle2, Clock, Wifi, WifiOff, type LucideIcon } from 'lucide-react'
export interface PtpStatus {
  hostname: string
  localStatus: {
    clockId: string
    masterOffsetNs: number
    gmPresent: boolean
  }
  remoteStatuses: {
    deviceName: string
    ipAddress: string
    isReachable: boolean
    offsetNs?: number
  }[]
}

interface PtpSyncStatusProps {
  ptpStatuses: PtpStatus[]
}

const getOffsetStatus = (
  offsetNs: number,
  gmPresent: boolean,
): {
  color: 'default' | 'secondary' | 'destructive'
  text: string
  icon: LucideIcon
} => {
  if (!gmPresent) return { color: 'destructive', text: 'No GM', icon: AlertCircle }

  const absOffset = Math.abs(offsetNs)
  if (absOffset < 1000) return { color: 'default', text: 'Excellent', icon: CheckCircle2 }
  if (absOffset < 10000) return { color: 'secondary', text: 'Good', icon: Clock }
  return { color: 'destructive', text: 'Poor', icon: AlertCircle }
}

const formatOffset = (offsetNs: number): string => {
  const absOffset = Math.abs(offsetNs)
  const sign = offsetNs < 0 ? '-' : '+'

  if (absOffset < 1000) {
    return `${sign}${absOffset}ns`
  } else if (absOffset < 1000000) {
    return `${sign}${(absOffset / 1000).toFixed(1)}µs`
  } else {
    return `${sign}${(absOffset / 1000000).toFixed(1)}ms`
  }
}

export function PtpSyncStatus({ ptpStatuses }: PtpSyncStatusProps) {
  // Calculate overall system status
  const systemStatus = ptpStatuses.every(
    (ptp) => ptp.localStatus.gmPresent && Math.abs(ptp.localStatus.masterOffsetNs) < 10000,
  )

  const totalDevices = ptpStatuses.reduce(
    (acc, ptp) => acc + ptp.remoteStatuses.length,
    ptpStatuses.length,
  )

  const syncedDevices = ptpStatuses.reduce((acc, ptp) => {
    const localSynced =
      ptp.localStatus.gmPresent && Math.abs(ptp.localStatus.masterOffsetNs) < 10000 ? 1 : 0
    const remoteSynced = ptp.remoteStatuses.filter(
      (r) => r.isReachable && r.offsetNs !== undefined && Math.abs(r.offsetNs) < 10000,
    ).length
    return acc + localSynced + remoteSynced
  }, 0)

  return (
    <div className="space-y-6">
      {/* System Overview Card */}
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="text-lg">System Time Synchronization</CardTitle>
            <Badge variant={systemStatus ? 'default' : 'destructive'} className="text-xs">
              {systemStatus ? 'All Synced' : 'Issues Detected'}
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">Synchronized Devices</span>
            <span className="font-medium">
              {syncedDevices} / {totalDevices}
            </span>
          </div>
          <Progress value={(syncedDevices / totalDevices) * 100} className="mt-2 h-2" />
        </CardContent>
      </Card>

      {/* ECU Cards */}
      <div className="grid gap-4 md:grid-cols-2">
        {ptpStatuses.map((ptp) => {
          const localStatus = getOffsetStatus(
            ptp.localStatus.masterOffsetNs,
            ptp.localStatus.gmPresent,
          )
          const LocalIcon = localStatus.icon

          return (
            <Card key={ptp.hostname}>
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-semibold">{ptp.hostname.toUpperCase()}</h3>
                    <p className="text-xs text-muted-foreground mt-1">
                      Clock ID: {ptp.localStatus.clockId}
                    </p>
                  </div>
                  <div className="text-right">
                    <div className="flex items-center gap-2">
                      <LocalIcon
                        className={`h-4 w-4 ${
                          localStatus.color === 'default'
                            ? 'text-green-500'
                            : localStatus.color === 'secondary'
                              ? 'text-yellow-500'
                              : 'text-red-500'
                        }`}
                      />
                      <Badge variant={localStatus.color} className="text-xs">
                        {localStatus.text}
                      </Badge>
                    </div>
                    <p className="text-xs font-mono mt-1">
                      {formatOffset(ptp.localStatus.masterOffsetNs)}
                    </p>
                  </div>
                </div>
              </CardHeader>

              <CardContent>
                {/* Remote Devices */}
                {ptp.remoteStatuses.length > 0 && (
                  <div>
                    <h4 className="text-sm font-medium mb-2">Connected Devices</h4>
                    <div className="space-y-2">
                      {ptp.remoteStatuses.map((remote) => {
                        const remoteStatus =
                          remote.isReachable && remote.offsetNs !== undefined
                            ? getOffsetStatus(remote.offsetNs, true)
                            : { color: 'destructive' as const, text: 'Offline', icon: WifiOff }

                        return (
                          <div
                            key={remote.deviceName}
                            className="flex items-center justify-between p-2 rounded-lg bg-muted/50"
                          >
                            <div className="flex items-center gap-2">
                              {remote.isReachable ? (
                                <Wifi className="h-3 w-3 text-green-500" />
                              ) : (
                                <WifiOff className="h-3 w-3 text-red-500" />
                              )}
                              <div>
                                <p className="text-sm font-medium">{remote.deviceName}</p>
                                <p className="text-xs text-muted-foreground">{remote.ipAddress}</p>
                              </div>
                            </div>
                            <div className="text-right">
                              {remote.isReachable && remote.offsetNs !== undefined ? (
                                <>
                                  <Badge variant={remoteStatus.color} className="text-xs">
                                    {formatOffset(remote.offsetNs)}
                                  </Badge>
                                </>
                              ) : (
                                <Badge variant="destructive" className="text-xs">
                                  Offline
                                </Badge>
                              )}
                            </div>
                          </div>
                        )
                      })}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          )
        })}
      </div>
    </div>
  )
}
