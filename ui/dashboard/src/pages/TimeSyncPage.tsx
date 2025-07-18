import { PtpSyncStatusNew } from '@/components/PtpSyncStatusNew'
import type { PtpStatus } from '@/components/PtpSyncStatus'

interface TimeSyncPageProps {
  ptpStatuses: PtpStatus[]
}

export function TimeSyncPage({ ptpStatuses }: TimeSyncPageProps) {
  return (
    <div>
      <PtpSyncStatusNew ptpStatuses={ptpStatuses} />
    </div>
  )
}