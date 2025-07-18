import { TopicRateStatusNew } from '@/components/TopicRateStatusNew'
import type { ModuleTopicStatus } from '@/components/TopicRateStatus'

interface TopicRatesPageProps {
  moduleTopicStatuses: ModuleTopicStatus[]
}

export function TopicRatesPage({ moduleTopicStatuses }: TopicRatesPageProps) {
  return (
    <div>
      <TopicRateStatusNew moduleTopicStatuses={moduleTopicStatuses} />
    </div>
  )
}