import { useEffect, useRef, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useTimerStore } from '../store/timerStore'
import type { ActiveTimerInfo } from '../types'

interface TimerEventMessage {
  type: string
  entry_id: string
  task_id: string
  subtask_id?: string
  timer_info?: ActiveTimerInfo
  elapsed_seconds?: number
  timestamp: string
}

export function useTimerWebSocket() {
  const queryClient = useQueryClient()
  const { updateTimer, removeTimer } = useTimerStore()
  const [isConnected, setIsConnected] = useState(false)
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const retryCountRef = useRef(0)

  useEffect(() => {
    let unmounted = false

    const connect = () => {
      if (unmounted) return

      // Derive ws:// from VITE_API_BASE_URL or window.location
      const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
      const wsURL = baseURL.replace(/^http/, 'ws') + '/ws/timers'

      try {
        const ws = new WebSocket(wsURL)
        wsRef.current = ws

        ws.onopen = () => {
          if (unmounted) {
            ws.close()
            return
          }
          setIsConnected(true)
          retryCountRef.current = 0
        }

        ws.onmessage = (event) => {
          try {
            const data: TimerEventMessage = JSON.parse(event.data)

            switch (data.type) {
              case 'timer.started':
              case 'timer.resumed':
              case 'timer.paused':
                if (data.timer_info) {
                  updateTimer(data.timer_info)
                }
                break

              case 'timer.stopped':
                if (data.timer_info) {
                  removeTimer(data.timer_info.entry_id)
                  queryClient.invalidateQueries({ queryKey: ['task', data.timer_info.task_id] })
                  queryClient.invalidateQueries({ queryKey: ['analytics'] })
                }
                break

              case 'timer.adjusted':
                if (data.timer_info) {
                  updateTimer(data.timer_info)
                  queryClient.invalidateQueries({ queryKey: ['task', data.timer_info.task_id] })
                }
                break
            }
          } catch {
            // Ignore malformed ping/text messages
          }
        }

        ws.onclose = () => {
          setIsConnected(false)
          wsRef.current = null

          if (!unmounted) {
            // Exponential backoff reconnect: 1s, 2s, 4s, max 10s
            const delay = Math.min(1000 * Math.pow(1.5, retryCountRef.current), 10000)
            retryCountRef.current += 1
            reconnectTimeoutRef.current = setTimeout(connect, delay)
          }
        }

        ws.onerror = () => {
          ws.close()
        }
      } catch {
        if (!unmounted) {
          reconnectTimeoutRef.current = setTimeout(connect, 3000)
        }
      }
    }

    connect()

    return () => {
      unmounted = true
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [updateTimer, removeTimer, queryClient])

  return { isConnected }
}

