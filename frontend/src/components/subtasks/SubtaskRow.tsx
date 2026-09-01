import { useState, useRef } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { subtasksApi } from '../../api/subtasks'
import type { Subtask } from '../../types'
import { Check, Trash2, Pencil } from 'lucide-react'
import { cn, formatDuration } from '../../lib/utils'
import TimerControls from '../timers/TimerControls'

interface SubtaskRowProps {
  subtask: Subtask
  taskId: string
  taskQueryKey: string[]
}

export default function SubtaskRow({ subtask, taskId, taskQueryKey }: SubtaskRowProps) {
  const queryClient = useQueryClient()
  const [editing, setEditing] = useState(false)
  const [title, setTitle] = useState(subtask.title)
  const inputRef = useRef<HTMLInputElement>(null)

  const { mutate: toggleDone } = useMutation({
    mutationFn: () => subtasksApi.update(subtask.id, { is_done: !subtask.is_done }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: taskQueryKey }),
  })

  const { mutate: updateTitle } = useMutation({
    mutationFn: (t: string) => subtasksApi.update(subtask.id, { title: t }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskQueryKey })
      setEditing(false)
    },
  })

  const { mutate: deleteSubtask } = useMutation({
    mutationFn: () => subtasksApi.delete(subtask.id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: taskQueryKey }),
  })

  const handleSave = () => {
    const trimmed = title.trim()
    if (!trimmed) return
    if (trimmed !== subtask.title) updateTitle(trimmed)
    else setEditing(false)
  }

  return (
    <div className={cn(
      'group flex items-center gap-3 py-2 px-3 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors',
      subtask.is_done && 'opacity-60'
    )}>
      {/* Checkbox */}
      <button
        onClick={() => toggleDone()}
        className={cn(
          'w-4 h-4 rounded border flex items-center justify-center shrink-0 transition-colors',
          subtask.is_done
            ? 'bg-green-500 border-green-500 text-white'
            : 'border-slate-300 dark:border-slate-600 hover:border-green-400'
        )}
      >
        {subtask.is_done && <Check className="w-2.5 h-2.5" />}
      </button>

      {/* Title */}
      <div className="flex-1 min-w-0">
        {editing ? (
          <input
            ref={inputRef}
            autoFocus
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            onBlur={handleSave}
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleSave()
              if (e.key === 'Escape') { setTitle(subtask.title); setEditing(false) }
            }}
            className="w-full bg-transparent border-b border-indigo-400 outline-none text-sm text-slate-800 dark:text-slate-200 py-0"
          />
        ) : (
          <span
            className={cn(
              'text-sm text-slate-800 dark:text-slate-200 truncate block cursor-default',
              subtask.is_done && 'line-through text-slate-400 dark:text-slate-500'
            )}
            onDoubleClick={() => setEditing(true)}
          >
            {subtask.title}
          </span>
        )}
        {subtask.total_time_spent_seconds > 0 && (
          <span className="text-xs text-slate-400 dark:text-slate-500">
            {formatDuration(subtask.total_time_spent_seconds)} logged
          </span>
        )}
      </div>

      {/* Actions (visible on hover) */}
      <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
        <TimerControls taskId={taskId} subtaskId={subtask.id} taskQueryKey={taskQueryKey} />
        <button
          onClick={() => { setEditing(true); setTimeout(() => inputRef.current?.focus(), 10) }}
          className="p-1 rounded text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700"
          title="Edit"
        >
          <Pencil className="w-3 h-3" />
        </button>
        <button
          onClick={() => deleteSubtask()}
          className="p-1 rounded text-slate-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-950"
          title="Delete"
        >
          <Trash2 className="w-3 h-3" />
        </button>
      </div>
    </div>
  )
}
