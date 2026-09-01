import { useTimerStore } from '../../store/timerStore'
import { timersApi } from '../../api/timers'
import { useQueryClient } from '@tanstack/react-query'
import { formatDurationHMS } from '../../lib/utils'
import { Play, Pause, Square } from 'lucide-react'
import { Link } from 'react-router-dom'

export default function ActiveTimerBar() {
  const { activeTimers, localElapsed, updateTimer, removeTimer } = useTimerStore()
  const queryClient = useQueryClient()

  if (activeTimers.length === 0) return null

  const handlePause = async (entryId: string) => {
    const updated = await timersApi.pause(entryId)
    updateTimer(updated)
  }

  const handleResume = async (entryId: string) => {
    const updated = await timersApi.resume(entryId)
    updateTimer(updated)
  }

  const handleStop = async (entryId: string, taskId: string) => {
    await timersApi.stop(entryId)
    removeTimer(entryId)
    queryClient.invalidateQueries({ queryKey: ['task', taskId] })
  }

  return (
    <div className="border-b border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-4 py-1.5 flex items-center gap-3 overflow-x-auto scrollbar-thin shrink-0">
      <span className="text-xs font-medium text-slate-400 dark:text-slate-500 shrink-0">
        Active
      </span>
      {activeTimers.map((timer) => {
        const elapsed = localElapsed[timer.entry_id] ?? timer.elapsed_seconds
        return (
          <div
            key={timer.entry_id}
            className="flex items-center gap-2 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-md px-2.5 py-1 shrink-0"
          >
            {/* Color dot */}
            <span
              className="w-1.5 h-1.5 rounded-full shrink-0"
              style={{ backgroundColor: timer.project_color || '#6366f1' }}
            />
            {/* Task info */}
            <Link
              to={`/tasks/${timer.task_id}`}
              className="flex items-center gap-1 text-xs hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors"
            >
              <span className="font-mono text-slate-400 dark:text-slate-500">
                {timer.ticket_key}
              </span>
              <span className="text-slate-700 dark:text-slate-300 max-w-[120px] truncate">
                {timer.subtask_title || timer.task_title}
              </span>
            </Link>
            {/* Live elapsed */}
            <span className={`font-mono text-xs tabular-nums font-semibold ${timer.is_paused ? 'text-slate-400' : 'text-indigo-600 dark:text-indigo-400'}`}>
              {formatDurationHMS(elapsed)}
            </span>
            {/* Controls */}
            <div className="flex items-center gap-0.5">
              {timer.is_paused ? (
                <button
                  onClick={() => handleResume(timer.entry_id)}
                  className="p-0.5 rounded hover:bg-indigo-100 dark:hover:bg-indigo-900 text-slate-500 hover:text-indigo-600 dark:hover:text-indigo-400"
                  title="Resume"
                >
                  <Play className="w-3 h-3" />
                </button>
              ) : (
                <button
                  onClick={() => handlePause(timer.entry_id)}
                  className="p-0.5 rounded hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-500"
                  title="Pause"
                >
                  <Pause className="w-3 h-3" />
                </button>
              )}
              <button
                onClick={() => handleStop(timer.entry_id, timer.task_id)}
                className="p-0.5 rounded hover:bg-red-50 dark:hover:bg-red-950 text-slate-400 hover:text-red-500"
                title="Stop"
              >
                <Square className="w-3 h-3" />
              </button>
            </div>
          </div>
        )
      })}
    </div>
  )
}
