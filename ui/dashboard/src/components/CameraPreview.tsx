import { useEffect, useRef, useState } from 'react'
import { ApiService } from '@/services/api'
import { useAutoRefresh } from '@/hooks/useAutoRefresh'

interface CameraPreviewProps {
  hostname: string
  topicName: string
  enabled: boolean
}

export function CameraPreview({ hostname, topicName, enabled }: CameraPreviewProps) {
  const apiService = ApiService.getInstance()
  const [blobUrl, setBlobUrl] = useState<string | null>(null)
  const [hasData, setHasData] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const currentBlobUrlRef = useRef<string | null>(null)

  useEffect(() => {
    return () => {
      if (currentBlobUrlRef.current) {
        URL.revokeObjectURL(currentBlobUrlRef.current)
      }
    }
  }, [])

  useAutoRefresh(
    async () => {
      try {
        const result = await apiService.getCameraPreview(hostname, topicName)
        setError(null)
        setHasData(result.hasData)

        const previousBlobUrl = currentBlobUrlRef.current
        currentBlobUrlRef.current = result.blobUrl
        setBlobUrl(result.blobUrl)
        if (previousBlobUrl) {
          URL.revokeObjectURL(previousBlobUrl)
        }
      } catch {
        setError('Failed to fetch camera preview')
      }
    },
    { enabled, interval: 5000 },
  )

  return (
    <div className="rounded-lg border overflow-hidden">
      {hasData && blobUrl ? (
        <img src={blobUrl} alt={topicName} className="w-full h-auto block" />
      ) : (
        <div className="p-4 text-xs text-center text-muted-foreground">
          {error ? <span className="text-destructive">{error}</span> : 'No data yet'}
        </div>
      )}
    </div>
  )
}
