import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Progress } from "@/components/ui/progress"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

export interface EcuModule {
  hostname: string
  status: 'OK' | 'WARN' | 'ERROR'
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

const getStatusColor = (status: EcuModule['status']) => {
  switch (status) {
    case 'OK':
      return 'default'
    case 'WARN':
      return 'secondary'
    case 'ERROR':
      return 'destructive'
    default:
      return 'default'
  }
}

const formatBytes = (bytes: number): string => {
  const gb = bytes / (1024 * 1024 * 1024)
  return `${gb.toFixed(1)}GB`
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
            <TableHead>Status</TableHead>
            <TableHead>Disk</TableHead>
            <TableHead className="w-[50px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {modules.map((module) => (
            <TableRow key={module.hostname}>
              <TableCell className="font-medium">{module.hostname}</TableCell>
              <TableCell>
                <Badge variant={getStatusColor(module.status)}>
                  {module.status}
                </Badge>
              </TableCell>
              <TableCell>
                <div className="flex items-center space-x-2">
                  <Progress value={module.diskUsagePercentage} className="w-[100px]" />
                  <span className="text-sm text-muted-foreground">
                    {module.diskUsagePercentage.toFixed(1)}%
                  </span>
                </div>
                <div className="text-xs text-muted-foreground mt-1">
                  {formatBytes(module.diskFreeBytes)} / {formatBytes(module.diskTotalBytes)} free
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