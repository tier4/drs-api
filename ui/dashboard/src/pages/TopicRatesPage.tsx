import { TopicRateStatus } from '@/components/TopicRateStatus'
import type { ModuleTopicStatus } from '@/components/TopicRateStatus'

interface TopicRatesPageProps {
  moduleTopicStatuses: ModuleTopicStatus[]
}

export function TopicRatesPage({ moduleTopicStatuses }: TopicRatesPageProps) {
  return (
    <div>
      <TopicRateStatus moduleTopicStatuses={moduleTopicStatuses} />
    </div>
  )
}
