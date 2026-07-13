import { Switch } from '@/components/ui/switch'
import { buttonVariants } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { Loader2 } from 'lucide-react'

interface RecordingSwitchProps {
  isRecording: boolean
  isLoading?: boolean
  onToggle: (enabled: boolean) => void
}

export function RecordingSwitch({ isRecording, isLoading, onToggle }: RecordingSwitchProps) {
  return (
    <label
      className={cn(
        buttonVariants({ variant: isRecording ? 'destructive' : 'secondary' }),
        'h-11 rounded-full cursor-pointer',
        'has-[:focus-visible]:ring-[3px] has-[:focus-visible]:ring-ring/50',
        isLoading && 'opacity-70 cursor-not-allowed',
      )}
    >
      <Switch
        checked={isRecording}
        onCheckedChange={onToggle}
        disabled={isLoading}
        className="sr-only !size-px !min-h-0 !min-w-0"
      />
      {isLoading ? (
        <Loader2 className="size-3.5 animate-spin" />
      ) : (
        <span className={cn('size-2.5 rounded-full bg-current', isRecording && 'animate-pulse')} />
      )}
      {isLoading ? 'Processing...' : isRecording ? 'Recording' : 'Start Recording'}
    </label>
  )
}
