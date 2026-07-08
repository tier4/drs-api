import { useEffect, useRef, useState } from 'react'
import * as THREE from 'three'
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js'
import { ApiService } from '@/services/api'
import { useAutoRefresh } from '@/hooks/useAutoRefresh'

interface PointCloudPreviewProps {
  hostname: string
  topicName: string
  enabled: boolean
}

const MAX_POINTS = 5000

// Turbo colormap (Google AI, MIT licensed), sampled at 11 control points and
// linearly interpolated - a low-cost approximation without pulling in a
// colormap dependency for one gradient.
const TURBO_CONTROL_POINTS: [number, number, number][] = [
  [0.18995, 0.07176, 0.23217],
  [0.25107, 0.25237, 0.63374],
  [0.27628, 0.42118, 0.89123],
  [0.15968, 0.63862, 0.95749],
  [0.08469, 0.85952, 0.75907],
  [0.28613, 0.97972, 0.4463],
  [0.64362, 0.99239, 0.20889],
  [0.90255, 0.87752, 0.14372],
  [0.99593, 0.60755, 0.13699],
  [0.90986, 0.31968, 0.05852],
  [0.66935, 0.09876, 0.03289],
]

function turboColor(t: number, out: THREE.Color): THREE.Color {
  const clamped = Math.min(1, Math.max(0, t))
  const scaled = clamped * (TURBO_CONTROL_POINTS.length - 1)
  const i = Math.min(TURBO_CONTROL_POINTS.length - 2, Math.floor(scaled))
  const frac = scaled - i
  const [r0, g0, b0] = TURBO_CONTROL_POINTS[i]
  const [r1, g1, b1] = TURBO_CONTROL_POINTS[i + 1]
  return out.setRGB(r0 + (r1 - r0) * frac, g0 + (g1 - g0) * frac, b0 + (b1 - b0) * frac)
}

export function PointCloudPreview({ hostname, topicName, enabled }: PointCloudPreviewProps) {
  const apiService = ApiService.getInstance()
  const [hasData, setHasData] = useState(false)
  const [decoderRunning, setDecoderRunning] = useState(false)
  const [pointCount, setPointCount] = useState(0)
  const [error, setError] = useState<string | null>(null)
  const isFetchingRef = useRef(false)

  const containerRef = useRef<HTMLDivElement>(null)
  const sceneRef = useRef<{
    renderer: THREE.WebGLRenderer
    scene: THREE.Scene
    camera: THREE.PerspectiveCamera
    controls: OrbitControls
    points: THREE.Points
    geometry: THREE.BufferGeometry
    material: THREE.PointsMaterial
  } | null>(null)

  // Mount: create the persistent scene once. Rotation/zoom state (via
  // OrbitControls) and the Points object survive every poll below - only
  // the position/color attribute contents are updated in place, so a
  // user's rotation doesn't reset every 5s.
  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const scene = new THREE.Scene()
    const camera = new THREE.PerspectiveCamera(60, 1, 0.1, 1000)
    // Look down the vehicle's forward axis (+X) from slightly above/behind,
    // matching how an operator would intuitively "look out" from the LiDAR.
    camera.position.set(-8, -8, 6)
    camera.up.set(0, 0, 1)
    camera.lookAt(0, 0, 0)

    const renderer = new THREE.WebGLRenderer({ antialias: true })
    renderer.setPixelRatio(window.devicePixelRatio)
    container.appendChild(renderer.domElement)

    const controls = new OrbitControls(camera, renderer.domElement)
    controls.enableDamping = true

    const geometry = new THREE.BufferGeometry()
    geometry.setAttribute(
      'position',
      new THREE.BufferAttribute(new Float32Array(MAX_POINTS * 3), 3),
    )
    geometry.setAttribute('color', new THREE.BufferAttribute(new Float32Array(MAX_POINTS * 3), 3))
    geometry.setDrawRange(0, 0)

    const material = new THREE.PointsMaterial({
      size: 2.5,
      sizeAttenuation: false,
      vertexColors: true,
    })
    const points = new THREE.Points(geometry, material)
    scene.add(points)

    sceneRef.current = { renderer, scene, camera, controls, points, geometry, material }

    const resizeObserver = new ResizeObserver(() => {
      const { clientWidth, clientHeight } = container
      if (clientWidth === 0 || clientHeight === 0) return
      camera.aspect = clientWidth / clientHeight
      camera.updateProjectionMatrix()
      renderer.setSize(clientWidth, clientHeight)
    })
    resizeObserver.observe(container)

    let animationFrame: number
    const animate = () => {
      animationFrame = requestAnimationFrame(animate)
      controls.update()
      renderer.render(scene, camera)
    }
    animate()

    return () => {
      cancelAnimationFrame(animationFrame)
      resizeObserver.disconnect()
      controls.dispose()
      renderer.dispose()
      geometry.dispose()
      material.dispose()
      container.removeChild(renderer.domElement)
      sceneRef.current = null
    }
  }, [])

  useAutoRefresh(
    async () => {
      if (isFetchingRef.current) return
      isFetchingRef.current = true
      try {
        const result = await apiService.getPointCloudPreview(hostname, topicName, MAX_POINTS)
        setError(null)
        setHasData(result.hasData)
        setDecoderRunning(result.decoderRunning)

        const sceneState = sceneRef.current
        if (!sceneState || !result.hasData || !result.points) {
          setPointCount(0)
          if (sceneState) sceneState.geometry.setDrawRange(0, 0)
          return
        }

        const { geometry } = sceneState
        const numPoints = Math.min(MAX_POINTS, Math.floor(result.points.length / 4))
        const positionAttr = geometry.attributes.position as THREE.BufferAttribute
        const colorAttr = geometry.attributes.color as THREE.BufferAttribute
        const color = new THREE.Color()

        for (let i = 0; i < numPoints; i++) {
          const x = result.points[i * 4]
          const y = result.points[i * 4 + 1]
          const z = result.points[i * 4 + 2]
          const intensity = result.points[i * 4 + 3]
          positionAttr.setXYZ(i, x, y, z)
          turboColor(intensity, color)
          colorAttr.setXYZ(i, color.r, color.g, color.b)
        }
        positionAttr.needsUpdate = true
        colorAttr.needsUpdate = true
        geometry.setDrawRange(0, numPoints)
        setPointCount(numPoints)
      } catch {
        setError('Failed to fetch point cloud preview')
        setHasData(false)
        setPointCount(0)
        sceneRef.current?.geometry.setDrawRange(0, 0)
      } finally {
        isFetchingRef.current = false
      }
    },
    { enabled, interval: 5000 },
  )

  const showEmptyState = !hasData
  const emptyMessage = !decoderRunning
    ? 'LiDAR decoder not running — start it on the vehicle to preview this topic.'
    : 'Waiting for first frame…'

  return (
    <div className="rounded-lg border overflow-hidden">
      <div className="relative h-56 w-full bg-muted">
        <div
          ref={containerRef}
          className="absolute inset-0"
          role="img"
          aria-label={`3D point cloud, ${pointCount} points, drag to rotate`}
        />
        {hasData && (
          <span className="absolute bottom-1 right-1 rounded bg-background/80 px-1.5 py-0.5 text-[10px] text-muted-foreground">
            {pointCount} pts
          </span>
        )}
      </div>
      {showEmptyState && !error && (
        <div className="p-2 text-xs text-center text-muted-foreground">{emptyMessage}</div>
      )}
      {error && <div className="p-2 text-xs text-center text-destructive">{error}</div>}
    </div>
  )
}
