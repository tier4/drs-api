import { useEffect, useRef, useState } from 'react'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { ApiService } from '@/services/api'
import { useAutoRefresh } from '@/hooks/useAutoRefresh'

interface NavSatFixPreviewProps {
  hostname: string
  enabled: boolean
}

// sensor_msgs/NavSatStatus.status values
const STATUS_NO_FIX = -1
const STATUS_FIX = 0

function statusColor(status: number): string {
  if (status <= STATUS_NO_FIX) return '#ef4444' // red - no fix
  if (status === STATUS_FIX) return '#eab308' // yellow - basic fix, no augmentation
  return '#22c55e' // green - SBAS/GBAS augmented fix
}

// 1-sigma isotropic radius (meters) from the diagonal East/North covariance terms.
function covarianceRadiusMeters(covariance: number[], covarianceType: number): number | null {
  if (covarianceType === 0) return null // UNKNOWN - not meaningful
  if (covariance.length < 5) return null
  const radius = Math.sqrt((covariance[0] + covariance[4]) / 2)
  return radius > 0 ? radius : null
}

export function NavSatFixPreview({ hostname, enabled }: NavSatFixPreviewProps) {
  const apiService = ApiService.getInstance()
  const [hasData, setHasData] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const containerRef = useRef<HTMLDivElement>(null)
  const mapRef = useRef<L.Map | null>(null)
  const markerRef = useRef<L.CircleMarker | null>(null)
  const covarianceCircleRef = useRef<L.Circle | null>(null)

  useEffect(() => {
    if (!containerRef.current || mapRef.current) return
    const map = L.map(containerRef.current, {
      zoomControl: false,
      attributionControl: false,
    }).setView([0, 0], 18)
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      maxZoom: 19,
    }).addTo(map)
    mapRef.current = map

    return () => {
      map.remove()
      mapRef.current = null
    }
  }, [])

  useAutoRefresh(
    async () => {
      try {
        const resp = await apiService.getPosition(hostname)
        setError(null)
        setHasData(resp.has_data)

        const map = mapRef.current
        if (!map || !resp.has_data || !resp.position) return

        const { latitude, longitude, status, position_covariance, position_covariance_type } =
          resp.position
        const latLng: L.LatLngExpression = [latitude, longitude]
        const color = statusColor(status)

        if (!markerRef.current) {
          markerRef.current = L.circleMarker(latLng, {
            radius: 8,
            color,
            fillColor: color,
            fillOpacity: 0.9,
            weight: 2,
          }).addTo(map)
        } else {
          markerRef.current.setLatLng(latLng)
          markerRef.current.setStyle({ color, fillColor: color })
        }

        const radiusMeters = covarianceRadiusMeters(position_covariance, position_covariance_type)
        if (radiusMeters) {
          if (!covarianceCircleRef.current) {
            covarianceCircleRef.current = L.circle(latLng, {
              radius: radiusMeters,
              color: '#3b82f6', // blue
              fill: false,
              weight: 2,
              dashArray: '6, 6',
            }).addTo(map)
          } else {
            covarianceCircleRef.current.setLatLng(latLng)
            covarianceCircleRef.current.setRadius(radiusMeters)
          }
        } else if (covarianceCircleRef.current) {
          covarianceCircleRef.current.remove()
          covarianceCircleRef.current = null
        }

        map.setView(latLng)
      } catch {
        setError('Failed to fetch position')
      }
    },
    { enabled, interval: 5000 },
  )

  return (
    <div className="rounded-lg border overflow-hidden">
      <div ref={containerRef} className="h-56 w-full bg-muted" />
      {!hasData && !error && (
        <div className="p-2 text-xs text-center text-muted-foreground">No data yet</div>
      )}
      {error && <div className="p-2 text-xs text-center text-destructive">{error}</div>}
    </div>
  )
}
