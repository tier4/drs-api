import { useEffect, useRef, useState } from 'react'
import { ApiService } from '@/services/api'
import { useAutoRefresh } from '@/hooks/useAutoRefresh'

interface LidarCameraProjectionPreviewProps {
  hostname: string
  topicName: string
  enabled: boolean
}

// 0 means unlimited: the bridge projects every point in the cloud.
const MAX_POINTS = 0

// Extracts the LiDAR position segment (e.g. "front") from a topic name
// shaped like "/sensing/lidar/{position}/{vendor}_packets", for the alt
// text/empty-state copy below. Falls back to the raw topic name if it
// doesn't match that shape.
function derivePosition(topicName: string): string {
  const match = topicName.match(/^\/sensing\/lidar\/([^/]+)\//)
  return match ? match[1] : topicName
}

export function LidarCameraProjectionPreview({
  hostname,
  topicName,
  enabled,
}: LidarCameraProjectionPreviewProps) {
  const apiService = ApiService.getInstance()
  const [blobUrl, setBlobUrl] = useState<string | null>(null)
  const [hasData, setHasData] = useState(false)
  const [decoderRunning, setDecoderRunning] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const currentBlobUrlRef = useRef<string | null>(null)
  const isFetchingRef = useRef(false)

  useEffect(() => {
    return () => {
      if (currentBlobUrlRef.current) {
        URL.revokeObjectURL(currentBlobUrlRef.current)
      }
    }
  }, [])

  useAutoRefresh(
    async () => {
      // A slow request (>5s) could otherwise overlap with the next tick.
      if (isFetchingRef.current) return
      isFetchingRef.current = true
      try {
        const result = await apiService.getLidarCameraProjectionPreview(
          hostname,
          topicName,
          MAX_POINTS,
        )
        setError(null)
        setHasData(result.hasData)
        setDecoderRunning(result.decoderRunning)

        const previousBlobUrl = currentBlobUrlRef.current
        currentBlobUrlRef.current = result.blobUrl
        setBlobUrl(result.blobUrl)
        if (previousBlobUrl) {
          URL.revokeObjectURL(previousBlobUrl)
        }
      } catch {
        setError('Failed to fetch LiDAR camera projection preview')
        setHasData(false)
        if (currentBlobUrlRef.current) {
          URL.revokeObjectURL(currentBlobUrlRef.current)
          currentBlobUrlRef.current = null
        }
        setBlobUrl(null)
      } finally {
        isFetchingRef.current = false
      }
    },
    { enabled, interval: 5000 },
  )

  const position = derivePosition(topicName)
  const altText = `LiDAR points projected onto ${position} camera view`
  const emptyMessage = !decoderRunning
    ? 'LiDAR decoder not running — start it on the vehicle to preview this topic.'
    : 'Waiting for camera/LiDAR data…'

  return (
    <div className="rounded-lg border overflow-hidden">
      {hasData && blobUrl ? (
        <img src={blobUrl} alt={altText} className="w-full h-auto block" />
      ) : (
        <div className="p-4 text-xs text-center text-muted-foreground">
          {error ? <span className="text-destructive">{error}</span> : emptyMessage}
        </div>
      )}
    </div>
  )
}
