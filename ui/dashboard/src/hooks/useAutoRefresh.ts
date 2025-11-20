import { useEffect, useRef } from 'react'

interface UseAutoRefreshOptions {
  enabled?: boolean
  interval?: number
}

export function useAutoRefresh(
  callback: () => void | Promise<void>,
  options: UseAutoRefreshOptions = {},
) {
  const { enabled = true, interval = 5000 } = options
  const savedCallback = useRef<() => void | Promise<void>>(callback)

  // Remember the latest callback
  useEffect(() => {
    savedCallback.current = callback
  }, [callback])

  // Set up the interval
  useEffect(() => {
    if (!enabled) return

    const tick = () => {
      if (savedCallback.current) {
        savedCallback.current()
      }
    }

    // Call immediately on mount
    tick()

    const id = setInterval(tick, interval)
    return () => clearInterval(id)
  }, [enabled, interval])
}
