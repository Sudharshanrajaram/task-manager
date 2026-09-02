import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { timersApi } from '../../api/timers'
import { Clock, X, Check } from 'lucide-react'

interface EditTimerModalProps {
  entryId: string
  initialDurationSeconds: number
  taskKey?: string
  taskTitle?: string
  onClose: () => void
}

export default function EditTimerModal({
  entryId,
  initialDurationSeconds,
  taskKey,
  taskTitle,
  onClose,
}: EditTimerModalProps) {
  const queryClient = useQueryClient()

  const initialHours = Math.floor(initialDurationSeconds / 3600)
  const initialMins = Math.floor((initialDurationSeconds % 3600) / 60)

  const [hours, setHours] = useState(initialHours)
  const [minutes, setMinutes] = useState(initialMins)
  const [error, setError] = useState('')

  const { mutate: updateTimer, isPending } = useMutation({
    mutationFn: (newSeconds: number) =>
      timersApi.update(entryId, { duration_seconds: newSeconds }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['daily-logs'] })
      queryClient.invalidateQueries({ queryKey: ['task'] })
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
      queryClient.invalidateQueries({ queryKey: ['analytics'] })
      onClose()
    },
    onError: (err: any) => {
      setError(err?.response?.data?.message || 'Failed to update timer duration')
    },
  })

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault()
    const totalSeconds = hours * 3600 + minutes * 60
    if (totalSeconds < 0) {
      setError('Duration cannot be negative')
      return
    }
    updateTimer(totalSeconds)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-150">
      <div className="w-full max-w-sm bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl shadow-xl overflow-hidden">
        {/* Modal Header */}
        <div className="flex items-center justify-between p-4 border-b border-slate-100 dark:border-slate-800">
          <div className="flex items-center gap-2">
            <Clock className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
            <h3 className="text-sm font-semibold text-slate-900 dark:text-slate-100">
              Edit Logged Duration
            </h3>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        <form onSubmit={handleSave} className="p-5 space-y-4">
          {taskKey && (
            <div className="text-xs text-slate-500 dark:text-slate-400 bg-slate-50 dark:bg-slate-800/60 p-2.5 rounded-lg border border-slate-200/60 dark:border-slate-700/60">
              <span className="font-mono font-semibold text-slate-700 dark:text-slate-300">
                {taskKey}
              </span>
              {taskTitle && <span className="ml-1.5 truncate">· {taskTitle}</span>}
            </div>
          )}

          {error && (
            <div className="p-2.5 text-xs text-red-600 bg-red-50 dark:bg-red-950/40 border border-red-200 dark:border-red-900/50 rounded-lg">
              {error}
            </div>
          )}

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
                Hours
              </label>
              <input
                type="number"
                min="0"
                max="240"
                value={hours}
                onChange={(e) => setHours(Math.max(0, parseInt(e.target.value) || 0))}
                className="w-full px-3 py-2 text-sm bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded-lg focus:ring-2 focus:ring-indigo-500 text-slate-900 dark:text-slate-100 outline-none font-mono"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">
                Minutes
              </label>
              <input
                type="number"
                min="0"
                max="59"
                value={minutes}
                onChange={(e) => setMinutes(Math.max(0, Math.min(59, parseInt(e.target.value) || 0)))}
                className="w-full px-3 py-2 text-sm bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded-lg focus:ring-2 focus:ring-indigo-500 text-slate-900 dark:text-slate-100 outline-none font-mono"
              />
            </div>
          </div>

          <div className="text-[11px] text-slate-400">
            Total:{' '}
            <span className="font-mono font-semibold text-slate-600 dark:text-slate-300">
              {hours}h {String(minutes).padStart(2, '0')}m
            </span>{' '}
            ({hours * 3600 + minutes * 60} seconds)
          </div>

          {/* Actions */}
          <div className="flex items-center justify-end gap-2 pt-2 border-t border-slate-100 dark:border-slate-800">
            <button
              type="button"
              onClick={onClose}
              className="px-3.5 py-1.5 text-xs font-medium text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isPending}
              className="flex items-center gap-1.5 px-4 py-1.5 text-xs font-medium text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 rounded-lg transition-colors shadow-sm"
            >
              <Check className="w-3.5 h-3.5" />
              {isPending ? 'Saving...' : 'Save Duration'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

