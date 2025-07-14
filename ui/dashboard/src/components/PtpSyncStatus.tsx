import { useState } from "react"
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

const getPtpStatusColor = (offsetNs: number, gmPresent: boolean): 'default' | 'secondary' | 'destructive' => {
  if (!gmPresent) return 'destructive'
  
  const absOffset = Math.abs(offsetNs)
  if (absOffset < 1000) return 'default' // < 1µs
  if (absOffset < 10000) return 'secondary' // < 10µs
  return 'destructive' // > 10µs
}

const formatOffset = (offsetNs: number): string => {
  const absOffset = Math.abs(offsetNs)
  if (absOffset < 1000) {
    return `${offsetNs}ns`
  } else if (absOffset < 1000000) {
    return `${(offsetNs / 1000).toFixed(1)}µs`
  } else {
    return `${(offsetNs / 1000000).toFixed(1)}ms`
  }
}

const hasRemoteIssues = (remoteStatuses: PtpStatus['remoteStatuses']) => {
  return remoteStatuses.some(remote => !remote.isReachable)
}

export function PtpSyncStatus({ ptpStatuses }: PtpSyncStatusProps) {
  // Initialize with modules that have issues automatically expanded
  const [expandedModules, setExpandedModules] = useState<Set<string>>(() => {
    const initialExpanded = new Set<string>()
    ptpStatuses.forEach(ptp => {
      if (hasRemoteIssues(ptp.remoteStatuses)) {
        initialExpanded.add(ptp.hostname)
      }
    })
    return initialExpanded
  })

  const toggleExpanded = (hostname: string) => {
    const newExpanded = new Set(expandedModules)
    if (newExpanded.has(hostname)) {
      newExpanded.delete(hostname)
    } else {
      newExpanded.add(hostname)
    }
    setExpandedModules(newExpanded)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Time Synchronization Status</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          {ptpStatuses.map((ptp) => {
            const isExpanded = expandedModules.has(ptp.hostname)
            const hasIssues = hasRemoteIssues(ptp.remoteStatuses)
            const showExpanded = isExpanded

            return (
              <div key={ptp.hostname} className="space-y-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <h4 className="font-medium">{ptp.hostname}</h4>
                    <span className="text-xs text-muted-foreground">
                      {ptp.localStatus.clockId}
                    </span>
                  </div>
                  <div className="flex items-center space-x-2">
                    <Badge variant={getPtpStatusColor(ptp.localStatus.masterOffsetNs, ptp.localStatus.gmPresent)}>
                      {ptp.localStatus.gmPresent ? 'SYNCED' : 'NO GM'}
                    </Badge>
                    <span className="text-sm text-muted-foreground">
                      {formatOffset(ptp.localStatus.masterOffsetNs)}
                    </span>
                  </div>
                </div>

                {ptp.remoteStatuses.length > 0 && (
                  <div>
                    <button
                      onClick={() => toggleExpanded(ptp.hostname)}
                      className="flex items-center space-x-1 text-xs text-muted-foreground hover:text-foreground mb-2"
                    >
                      <span>Remote Devices ({ptp.remoteStatuses.length})</span>
                      {hasIssues && (
                        <Badge variant="destructive" className="text-xs ml-1">
                          Issues
                        </Badge>
                      )}
                      <span className="text-xs">
                        {showExpanded ? '▼' : '▶'}
                      </span>
                    </button>
                    
                    {showExpanded && (
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead className="text-xs">Device</TableHead>
                            <TableHead className="text-xs">IP Address</TableHead>
                            <TableHead className="text-xs">Status</TableHead>
                            <TableHead className="text-xs">Offset</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {ptp.remoteStatuses.map((remote) => (
                            <TableRow key={remote.deviceName}>
                              <TableCell className="text-xs">{remote.deviceName}</TableCell>
                              <TableCell className="text-xs">{remote.ipAddress}</TableCell>
                              <TableCell>
                                <Badge variant={remote.isReachable ? 'default' : 'destructive'} className="text-xs">
                                  {remote.isReachable ? 'Online' : 'Offline'}
                                </Badge>
                              </TableCell>
                              <TableCell className="text-xs">
                                {remote.offsetNs !== undefined ? formatOffset(remote.offsetNs) : 'N/A'}
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
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
}