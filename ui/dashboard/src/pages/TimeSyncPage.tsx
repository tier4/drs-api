import { PtpSyncStatus } from '@/components/PtpSyncStatus'
import type { PtpStatus } from '@/components/PtpSyncStatus'

interface TimeSyncPageProps {
  ptpStatuses: PtpStatus[]
}

export function TimeSyncPage({ ptpStatuses }: TimeSyncPageProps) {
  return (
    <div>
      <PtpSyncStatus ptpStatuses={ptpStatuses} />
    </div>
  )
}