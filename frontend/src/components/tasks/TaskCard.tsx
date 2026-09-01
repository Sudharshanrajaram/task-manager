import { Link } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { tasksApi } from '../../api/tasks'
import type { Task, TaskStatus } from '../../types'
import {
  STATUS_CONFIG,
  STATUS_ORDER,
  PRIORITY_CONFIG,
  TYPE_CONFIG,
  formatDuration,
} from '../../lib/utils'
import TimerControls from '../timers/TimerControls'
import { CheckSquare, Clock } from 'lucide-react'

interface TaskCardProps {
  task: Task
  projectId: string
}

export default function TaskCard({ task, projectId }: TaskCardProps) {
  const queryClient = useQueryClient()

  const { mutate: updateStatus } = useMutation({
    mutationFn: (newStatus: TaskStatus) => tasksApi.update(task.id, { status: newStatus }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks', projectId] })
      queryClient.invalidateQueries({ queryKey: ['task', task.id] })
    },
  })

  const subtasksCount = task.subtasks?.length ?? 0
  const doneSubtasksCount = task.subtasks?.filter((s) => s.is_done).length ?? 0

  return (
    <div className="group bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-3.5 hover:border-slate-300 dark:hover:border-slate-700 shadow-sm hover:shadow transition-all">
      {/* Header: Key, Type, Priority, Status Dropdown */}
      <div className="flex items-center justify-between gap-2 mb-2">
        <div className="flex items-center gap-1.5 min-w-0">
          <span className="font-mono text-xs font-semibold text-slate-500 dark:text-slate-400">
            {task.ticket_key}
          </span>
          <span
            className={`text-[10px] px-1.5 py-0.5 rounded font-medium ${
              TYPE_CONFIG[task.type]?.className || 'bg-slate-100'
            }`}
          >
            {TYPE_CONFIG[task.type]?.label || task.type}
          </span>
        </div>

        <div className="flex items-center gap-1.5 shrink-0">
          {/* Priority pill */}
          <span
            className={`text-[10px] font-bold px-1.5 py-0.5 rounded flex items-center gap-1 ${
              PRIORITY_CONFIG[task.priority]?.className
            }`}
          >
            <span className={`w-1.5 h-1.5 rounded-full ${PRIORITY_CONFIG[task.priority]?.dot}`} />
            {PRIORITY_CONFIG[task.priority]?.label}
          </span>

          {/* Quick status selector */}
          <select
            value={task.status}
            onChange={(e) => updateStatus(e.target.value as TaskStatus)}
            className="text-[11px] bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded px-1.5 py-0.5 text-slate-700 dark:text-slate-300 outline-none cursor-pointer"
          >
            {STATUS_ORDER.map((s) => (
              <option key={s} value={s}>
                {STATUS_CONFIG[s].label}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Blocked indicator (independent badge per spec 1.1) */}
      {task.is_blocked && (
        <div
          title={task.blocked_reason ? `Blocked: ${task.blocked_reason}` : 'Blocked'}
          className="flex items-center gap-1 text-[11px] font-semibold px-2 py-0.5 rounded-md bg-amber-50 text-amber-800 dark:bg-amber-950/60 dark:text-amber-300 border border-amber-300 dark:border-amber-800/80 mb-2 w-fit"
        >
          <span>⚠ Blocked</span>
          {task.blocked_reason && (
            <span className="font-normal text-amber-700/80 dark:text-amber-300/80 max-w-[140px] truncate">
              · {task.blocked_reason}
            </span>
          )}
        </div>
      )}

      {/* Task Title */}
      <Link
        to={`/tasks/${task.id}`}
        className="text-sm font-medium text-slate-900 dark:text-slate-100 hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors line-clamp-2 block mb-3"
      >
        {task.title}
      </Link>

      {/* Labels */}
      {task.labels && task.labels.length > 0 && (
        <div className="flex flex-wrap gap-1 mb-3">
          {task.labels.map((l) => (
            <span
              key={l}
              className="text-[10px] bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 px-1.5 py-0.5 rounded"
            >
              {l}
            </span>
          ))}
        </div>
      )}

      {/* Footer: Subtasks count, Time spent, Timer control */}
      <div className="flex items-center justify-between pt-2 border-t border-slate-100 dark:border-slate-800/80 text-xs text-slate-500 dark:text-slate-400">
        <div className="flex items-center gap-3">
          {subtasksCount > 0 && (
            <span className="flex items-center gap-1 text-[11px]">
              <CheckSquare className="w-3 h-3 text-slate-400" />
              {doneSubtasksCount}/{subtasksCount}
            </span>
          )}

          {task.total_time_spent_seconds > 0 && (
            <span className="flex items-center gap-1 font-mono text-[11px]">
              <Clock className="w-3 h-3 text-slate-400" />
              {formatDuration(task.total_time_spent_seconds)}
            </span>
          )}
        </div>

        <TimerControls
          taskId={task.id}
          taskQueryKey={['tasks', projectId]}
        />
      </div>
    </div>
  )
}

