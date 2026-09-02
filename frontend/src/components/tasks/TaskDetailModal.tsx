import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { tasksApi } from '../../api/tasks'
import { subtasksApi } from '../../api/subtasks'
import type { TaskPriority, TaskStatus, TaskType } from '../../types'
import {
  STATUS_CONFIG,
  STATUS_ORDER,
  PRIORITY_CONFIG,
  TYPE_CONFIG,
  formatDuration,
} from '../../lib/utils'
import TimerControls from '../timers/TimerControls'
import SubtaskRow from '../subtasks/SubtaskRow'
import AISubtaskPanel from '../subtasks/AISubtaskPanel'
import TaskSummaryCard from './TaskSummaryCard'
import TaskDependenciesPanel from '../dependencies/TaskDependenciesPanel'
import NotesEditor from '../notes/NotesEditor'
import CommentsSection from '../comments/CommentsSection'
import {
  X,
  Maximize2,
  Clock,
  Plus,
  Trash2,
  Save,
  Pencil,
  AlertTriangle,
} from 'lucide-react'

interface TaskDetailModalProps {
  taskId: string
  onClose: () => void
}

export default function TaskDetailModal({ taskId, onClose }: TaskDetailModalProps) {
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  const [isEditingTitle, setIsEditingTitle] = useState(false)
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [newSubtaskTitle, setNewSubtaskTitle] = useState('')

  // Body scroll lock and Esc key listener
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    const prevOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'

    return () => {
      window.removeEventListener('keydown', handleKeyDown)
      document.body.style.overflow = prevOverflow
    }
  }, [onClose])

  const { data: task, isLoading } = useQuery({
    queryKey: ['task', taskId],
    queryFn: () => tasksApi.getById(taskId),
    enabled: !!taskId,
  })

  useEffect(() => {
    if (task) {
      setTitle(task.title)
      setDescription(task.description || '')
    }
  }, [task])

  const { mutate: updateTask } = useMutation({
    mutationFn: (updates: Parameters<typeof tasksApi.update>[1]) =>
      tasksApi.update(taskId, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['task', taskId] })
      if (task?.project_id) {
        queryClient.invalidateQueries({ queryKey: ['tasks', task.project_id] })
      }
    },
  })

  const { mutate: toggleBlock } = useMutation({
    mutationFn: ({ isBlocked, reason }: { isBlocked: boolean; reason?: string }) =>
      tasksApi.block(taskId, isBlocked, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['task', taskId] })
      if (task?.project_id) {
        queryClient.invalidateQueries({ queryKey: ['tasks', task.project_id] })
      }
    },
  })

  const { mutate: addSubtask } = useMutation({
    mutationFn: (subtaskTitle: string) => subtasksApi.create(taskId, { title: subtaskTitle }),
    onSuccess: () => {
      setNewSubtaskTitle('')
      queryClient.invalidateQueries({ queryKey: ['task', taskId] })
    },
  })

  const { mutate: deleteTask } = useMutation({
    mutationFn: () => tasksApi.delete(taskId),
    onSuccess: () => {
      if (task?.project_id) {
        queryClient.invalidateQueries({ queryKey: ['tasks', task.project_id] })
      }
      onClose()
    },
  })

  if (isLoading || !task) {
    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
        <div className="w-full max-w-4xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-8 shadow-2xl animate-pulse">
          <div className="h-6 w-48 bg-slate-200 dark:bg-slate-800 rounded mb-4" />
          <div className="h-10 w-full bg-slate-200 dark:bg-slate-800 rounded mb-6" />
          <div className="h-64 w-full bg-slate-100 dark:bg-slate-800/50 rounded" />
        </div>
      </div>
    )
  }

  const subtasks = task.subtasks || []
  const doneSubtasks = subtasks.filter((s) => s.is_done)

  const handleTitleSubmit = () => {
    const trimmed = title.trim()
    if (trimmed && trimmed !== task.title) {
      updateTask({ title: trimmed })
    }
    setIsEditingTitle(false)
  }

  return (
    <div
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose()
      }}
      className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6 bg-black/60 backdrop-blur-sm overflow-y-auto animate-in fade-in duration-150"
    >
      <div className="relative w-full max-w-4xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl my-auto max-h-[92vh] flex flex-col overflow-hidden">
        {/* Top Header Bar */}
        <div className="flex items-center justify-between gap-3 px-6 py-4 border-b border-slate-200 dark:border-slate-800 bg-slate-50/80 dark:bg-slate-900/80 shrink-0">
          <div className="flex items-center gap-2 min-w-0">
            {task.project && (
              <span
                className="w-2.5 h-2.5 rounded-full shrink-0"
                style={{ backgroundColor: task.project.color }}
              />
            )}
            <span className="font-mono text-xs font-semibold px-2 py-0.5 rounded bg-slate-200 dark:bg-slate-800 text-slate-700 dark:text-slate-300">
              {task.ticket_key}
            </span>
            <span className="text-xs text-slate-500 dark:text-slate-400 truncate">
              {task.project?.name}
            </span>
          </div>

          <div className="flex items-center gap-2 shrink-0">
            <Link
              to={`/tasks/${task.id}/focus`}
              className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-slate-700 dark:text-slate-300 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors shadow-sm"
              title="Open full Focus Mode"
            >
              <Maximize2 className="w-3.5 h-3.5" />
              <span>Focus Mode</span>
            </Link>

            <button
              onClick={() => {
                if (confirm(`Delete ticket ${task.ticket_key}?`)) deleteTask()
              }}
              className="p-1.5 rounded-lg border border-slate-200 dark:border-slate-700 text-slate-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-950/40 transition-colors"
              title="Delete ticket"
            >
              <Trash2 className="w-4 h-4" />
            </button>

            <button
              onClick={onClose}
              className="p-1.5 rounded-lg border border-slate-200 dark:border-slate-700 text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
              title="Close modal (Esc)"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>

        {/* Scrollable Content Body */}
        <div className="p-6 overflow-y-auto space-y-6 scrollbar-thin">
          {/* Blocked banner */}
          {task.is_blocked && (
            <div className="p-3 rounded-lg bg-amber-50 dark:bg-amber-950/40 border border-amber-300 dark:border-amber-800/80 flex items-center justify-between gap-3">
              <div className="flex items-center gap-2 text-xs text-amber-800 dark:text-amber-300">
                <AlertTriangle className="w-4 h-4 shrink-0 text-amber-600" />
                <span className="font-bold">Blocked:</span>
                <span>{task.blocked_reason || 'No reason specified'}</span>
              </div>
              <button
                onClick={() => toggleBlock({ isBlocked: false })}
                className="text-xs font-medium underline text-amber-700 hover:text-amber-900 dark:text-amber-300"
              >
                Remove blocker
              </button>
            </div>
          )}

          {/* Ticket Title */}
          <div>
            {isEditingTitle ? (
              <div className="flex items-center gap-2">
                <input
                  autoFocus
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') handleTitleSubmit()
                    if (e.key === 'Escape') {
                      setTitle(task.title)
                      setIsEditingTitle(false)
                    }
                  }}
                  className="flex-1 text-xl font-bold bg-transparent border-b-2 border-indigo-500 outline-none pb-1 text-slate-900 dark:text-slate-100"
                />
                <button
                  onClick={handleTitleSubmit}
                  className="px-3 py-1 bg-indigo-600 text-white rounded text-xs font-medium"
                >
                  Save
                </button>
              </div>
            ) : (
              <div className="flex items-center gap-2 group">
                <h1
                  onClick={() => setIsEditingTitle(true)}
                  className="text-xl font-bold text-slate-900 dark:text-slate-100 cursor-pointer hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors"
                  title="Click to edit title"
                >
                  {task.title}
                </h1>
                <button
                  onClick={() => setIsEditingTitle(true)}
                  className="opacity-0 group-hover:opacity-100 p-1 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-opacity"
                  title="Edit title"
                >
                  <Pencil className="w-3.5 h-3.5" />
                </button>
              </div>
            )}
          </div>

          {/* Meta Controls (Status, Priority, Type, Timers) */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 p-4 rounded-xl bg-slate-50 dark:bg-slate-800/50 border border-slate-200/80 dark:border-slate-800">
            <div>
              <label className="block text-[10px] uppercase font-bold text-slate-400 mb-1">Status</label>
              <select
                value={task.status}
                onChange={(e) => updateTask({ status: e.target.value as TaskStatus })}
                className="w-full text-xs font-medium bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg px-2.5 py-1.5 text-slate-800 dark:text-slate-200 outline-none cursor-pointer"
              >
                {STATUS_ORDER.map((s) => (
                  <option key={s} value={s}>
                    {STATUS_CONFIG[s].label}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-[10px] uppercase font-bold text-slate-400 mb-1">Priority</label>
              <select
                value={task.priority}
                onChange={(e) => updateTask({ priority: e.target.value as TaskPriority })}
                className="w-full text-xs font-medium bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg px-2.5 py-1.5 text-slate-800 dark:text-slate-200 outline-none cursor-pointer"
              >
                <option value="p0">P0 - Critical</option>
                <option value="p1">P1 - High</option>
                <option value="p2">P2 - Medium</option>
                <option value="p3">P3 - Low</option>
              </select>
            </div>

            <div>
              <label className="block text-[10px] uppercase font-bold text-slate-400 mb-1">Type</label>
              <select
                value={task.type}
                onChange={(e) => updateTask({ type: e.target.value as TaskType })}
                className="w-full text-xs font-medium bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg px-2.5 py-1.5 text-slate-800 dark:text-slate-200 outline-none cursor-pointer"
              >
                <option value="task">Task</option>
                <option value="bug">Bug</option>
                <option value="improvement">Improvement</option>
                <option value="spike">Spike</option>
              </select>
            </div>

            <div>
              <label className="block text-[10px] uppercase font-bold text-slate-400 mb-1">Logged Time</label>
              <div className="flex items-center gap-1.5 text-xs font-mono font-medium text-slate-700 dark:text-slate-300 py-1.5">
                <Clock className="w-3.5 h-3.5 text-slate-400" />
                <span>{formatDuration(task.total_time_spent_seconds || 0)}</span>
              </div>
            </div>
          </div>

          {/* Active Timer Controls */}
          <div className="flex items-center justify-between p-4 bg-white dark:bg-slate-800/70 rounded-xl border border-slate-200 dark:border-slate-700/80">
            <div className="flex items-center gap-2 text-xs font-medium text-slate-700 dark:text-slate-300">
              <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
              <span>Timer Session</span>
            </div>
            <TimerControls taskId={task.id} taskQueryKey={['task', task.id]} />
          </div>

          {/* AI Task Summary Card */}
          <TaskSummaryCard
            taskId={task.id}
            initialSummary={task.ai_summary}
          />

          {/* Description */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 shadow-sm">
            <div className="flex items-center justify-between mb-2">
              <h2 className="text-xs font-bold uppercase tracking-wider text-slate-400">
                Description
              </h2>
              {description !== (task.description || '') && (
                <button
                  onClick={() => updateTask({ description })}
                  className="flex items-center gap-1 text-xs text-indigo-600 hover:text-indigo-700 font-medium"
                >
                  <Save className="w-3 h-3" />
                  Save description
                </button>
              )}
            </div>
            <textarea
              rows={3}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Add description..."
              className="w-full text-sm bg-transparent border-0 outline-none text-slate-900 dark:text-slate-100 placeholder:text-slate-400 resize-y"
            />
          </div>

          {/* Subtasks */}
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 shadow-sm space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <h2 className="text-xs font-bold uppercase tracking-wider text-slate-400">
                  Subtasks
                </h2>
                <span className="text-xs font-mono text-slate-500">
                  ({doneSubtasks.length}/{subtasks.length})
                </span>
              </div>
              <AISubtaskPanel
                taskId={task.id}
                taskTitle={task.title}
                projectId={task.project_id}
                onAccepted={() => queryClient.invalidateQueries({ queryKey: ['task', taskId] })}
              />
            </div>

            <form
              onSubmit={(e) => {
                e.preventDefault()
                if (newSubtaskTitle.trim()) addSubtask(newSubtaskTitle.trim())
              }}
              className="flex items-center gap-2"
            >
              <input
                type="text"
                value={newSubtaskTitle}
                onChange={(e) => setNewSubtaskTitle(e.target.value)}
                placeholder="Add a new subtask..."
                className="flex-1 text-xs px-3 py-2 rounded-lg bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-900 dark:text-slate-100 outline-none focus:border-indigo-500"
              />
              <button
                type="submit"
                disabled={!newSubtaskTitle.trim()}
                className="flex items-center gap-1 px-3 py-2 text-xs font-medium text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 rounded-lg transition-colors"
              >
                <Plus className="w-3.5 h-3.5" />
                Add
              </button>
            </form>

            <div className="space-y-1.5">
              {subtasks.map((subtask) => (
                <SubtaskRow
                  key={subtask.id}
                  subtask={subtask}
                  taskId={task.id}
                  taskQueryKey={['task', task.id]}
                />
              ))}
            </div>
          </div>

          {/* Task Dependencies (Phase 13) */}
          <TaskDependenciesPanel taskId={task.id} projectId={task.project_id} />

          {/* Focus Mode Notes (Phase 11 & 19) */}
          <NotesEditor
            taskId={task.id}
            mode="task"
            ticketKey={task.ticket_key}
            ticketTitle={task.title}
          />

          {/* Comments Section (Phase 17) */}
          <CommentsSection taskId={task.id} />
        </div>
      </div>
    </div>
  )
}
