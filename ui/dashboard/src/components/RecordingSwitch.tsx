import { Switch } from "@/components/ui/switch"
import { Loader2 } from "lucide-react"

interface RecordingSwitchProps {
  isRecording: boolean
  isLoading?: boolean
  onToggle: (enabled: boolean) => void
}

export function RecordingSwitch({ isRecording, isLoading, onToggle }: RecordingSwitchProps) {
  return (
    <div className="flex items-center space-x-2">
      <span className="text-sm text-muted-foreground">Recording</span>
      <div className="relative">
        <Switch 
          checked={isRecording} 
          onCheckedChange={onToggle}
          disabled={isLoading}
        />
        {isLoading && (
          <div className="absolute inset-0 flex items-center justify-center bg-background/50 rounded-full">
            <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
          </div>
        )}
      </div>
    </div>
  )
}