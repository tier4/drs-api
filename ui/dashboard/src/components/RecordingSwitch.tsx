import { Switch } from '@/components/ui/switch'
import { Loader2 } from 'lucide-react'

interface RecordingSwitchProps {
  isRecording: boolean
  isLoading?: boolean
  onToggle: (enabled: boolean) => void
}

export function RecordingSwitch({ isRecording, isLoading, onToggle }: RecordingSwitchProps) {
  return (
    <div className="flex items-center space-x-2">
      <span
        className={`text-sm transition-all duration-200 ${
          isLoading
            ? 'text-muted-foreground/50'
            : isRecording
              ? 'text-red-600 font-medium'
              : 'text-muted-foreground'
        }`}
      >
        {isLoading ? 'Processing...' : isRecording ? 'Recording' : 'Stopped'}
      </span>
      <div className="relative">
        <Switch
          checked={isRecording}
          onCheckedChange={onToggle}
          disabled={isLoading}
          className={isLoading ? 'opacity-70' : ''}
        />
        {isLoading && (
          <div className="absolute inset-0 flex items-center justify-center bg-background/80 rounded-full">
            <Loader2 className="h-4 w-4 animate-spin text-primary" />
          </div>
        )}
      </div>
    </div>
  )
}
