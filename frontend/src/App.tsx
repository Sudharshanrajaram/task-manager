import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useEffect } from 'react'
import { useTimerStore } from './store/timerStore'
import { timersApi } from './api/timers'
import { useTimerWebSocket } from './hooks/useTimerWebSocket'
import Layout from './components/layout/Layout'
import CommandPalette from './components/command/CommandPalette'
import Dashboard from './pages/Dashboard'
import DailyLogs from './pages/DailyLogs'
import TaskBoard from './pages/TaskBoard'
import TaskDetail from './pages/TaskDetail'
import FocusMode from './pages/FocusMode'

export default function App() {
  const { setActiveTimers, tickAll } = useTimerStore()

  // Realtime WebSocket sync across all open tabs & devices
  useTimerWebSocket()

  // Poll active timers every 15s as fallback and tick locally every second
  useEffect(() => {
    const syncTimers = async () => {
      try {
        const timers = await timersApi.getActive()
        setActiveTimers(timers ?? [])
      } catch {
        // Silently fail if backend is not yet running
      }
    }

    syncTimers()
    const pollInterval = setInterval(syncTimers, 10_000)
    const tickInterval = setInterval(tickAll, 1_000)

    return () => {
      clearInterval(pollInterval)
      clearInterval(tickInterval)
    }
  }, [setActiveTimers, tickAll])

  return (
    <BrowserRouter>
      <CommandPalette />
      <Routes>
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route element={<Layout />}>
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/logs" element={<DailyLogs />} />
          <Route path="/projects/:projectId" element={<TaskBoard />} />
          <Route path="/tasks/:taskId" element={<TaskDetail />} />
        </Route>
        <Route path="/tasks/:taskId/focus" element={<FocusMode />} />
      </Routes>
    </BrowserRouter>
  )
}
