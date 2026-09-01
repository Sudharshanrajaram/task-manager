import { useTimerStore } from '../../store/timerStore'
import { timersApi } from '../../api/timers'
import { useQueryClient } from '@tanstack/react-query'
import { formatDurationHMS } from '../../lib/utils'
import { Play, Pause, Square, Timer } from 'lucide-react'
import { useState } from 'react'

interface TimerControlsProps {
  taskId: string
  subtaskId?: string
  taskQueryKey: string[]
}

export default function TimerControls({ taskId, subtaskId, taskQueryKey }: TimerControlsProps) {
  const { activeTimers, localElapsed, updateTimer, removeTimer } = useTimerStore()
  const queryClient = useQueryClient()
  const [loading, setLoading] = useState(false)

  const timer = activeTimers.find((t) =>
    subtaskId ? t.subtask_id === subtaskId : (t.task_id === taskId && !t.subtask_id)
  )
  const elapsed = timer ? (localElapsed[timer.entry_id] ?? timer.elapsed_seconds) : 0
  const isRunning = !!timer && !timer.is_paused

  const handleStart = async () => {
    setLoading(true)
    try {
      const started = await timersApi.start(taskId, subtaskId)
      updateTimer(started)
      queryClient.invalidateQueries({ queryKey: taskQueryKey })
    } finally {
      setLoading(false)
    }
  }

  const handlePause = async () => {
    if (!timer) return
    setLoading(true)
    try {
      const updated = await timersApi.pause(timer.entry_id)
      updateTimer(updated)
    } finally {
      setLoading(false)
    }
  }

  const handleResume = async () => {
    if (!timer) return
    setLoading(true)
    try {
      const updated = await timersApi.resume(timer.entry_id)
      updateTimer(updated)
    } finally {
      setLoading(false)
    }
  }

  const handleStop = async () => {
    if (!timer) return
    setLoading(true)
    try {
      await timersApi.stop(timer.entry_id)
      removeTimer(timer.entry_id)
      queryClient.invalidateQueries({ queryKey: taskQueryKey })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex items-center gap-1.5">
      {timer && (
        <span className={`font-mono text-sm tabular-nums font-semibold ${isRunning ? 'text-indigo-600 dark:text-indigo-400' : 'text-slate-500'}`}>
          {formatDurationHMS(elapsed)}
        </span>
      )}
      {!timer ? (
        <button
          onClick={handleStart}
          disabled={loading}
          className="flex items-center gap-1 px-2 py-1 rounded text-xs text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 hover:bg-indigo-50 dark:hover:bg-indigo-950 hover:text-indigo-600 dark:hover:text-indigo-400 border border-slate-200 dark:border-slate-700 transition-colors disabled:opacity-50"
        >
          <Timer className="w-3 h-3" />
          Start
        </button>
      ) : (
        <>
          {timer.is_paused ? (
            <button
              onClick={handleResume}
              disabled={loading}
              className="p-1 rounded hover:bg-indigo-50 dark:hover:bg-indigo-950 text-slate-500 hover:text-indigo-600 dark:hover:text-indigo-400 disabled:opacity-50"
              title="Resume"
            >
              <Play className="w-3.5 h-3.5" />
            </button>
          ) : (
            <button
              onClick={handlePause}
              disabled={loading}
              className="p-1 rounded hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-500 disabled:opacity-50"
              title="Pause"
            >
              <Pause className="w-3.5 h-3.5" />
            </button>
          )}
          <button
            onClick={handleStop}
            disabled={loading}
            className="p-1 rounded hover:bg-red-50 dark:hover:bg-red-950 text-slate-400 hover:text-red-500 disabled:opacity-50"
            title="Stop"
          >
            <Square className="w-3.5 h-3.5" />
          </button>
        </>
      )}
    </div>
  )
}
