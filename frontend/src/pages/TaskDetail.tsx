import { useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { tasksApi } from '../api/tasks'
import { subtasksApi } from '../api/subtasks'
import type { TaskPriority, TaskStatus, TaskType } from '../types'
import {
  STATUS_CONFIG,
  STATUS_ORDER,
  PRIORITY_CONFIG,
  TYPE_CONFIG,
  formatDuration,
} from '../lib/utils'
import TimerControls from '../components/timers/TimerControls'
import SubtaskRow from '../components/subtasks/SubtaskRow'
import AISubtaskPanel from '../components/subtasks/AISubtaskPanel'
import TaskSummaryCard from '../components/tasks/TaskSummaryCard'
import TaskDependenciesPanel from '../components/dependencies/TaskDependenciesPanel'
import NotesEditor from '../components/notes/NotesEditor'
import CommentsSection from '../components/comments/CommentsSection'
import {
  ArrowLeft,
  Clock,
  Maximize2,
  Plus,
  Trash2,
  CheckCircle2,
  AlertCircle,
  Save,
  Pencil,
} from 'lucide-react'

export default function TaskDetail() {
  const { taskId } = useParams<{ taskId: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const [isEditingTitle, setIsEditingTitle] = useState(false)
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [newSubtaskTitle, setNewSubtaskTitle] = useState('')

  // Bug fields
  const [stepsToReproduce, setStepsToReproduce] = useState('')
  const [severity, setSeverity] = useState('')
  const [environment, setEnvironment] = useState('')

  const { data: task, isLoading } = useQuery({
    queryKey: ['task', taskId],
    queryFn: () => (taskId ? tasksApi.getById(taskId) : Promise.reject('No task ID')),
    enabled: !!taskId,
  })

  useEffect(() => {
    if (task) {
      setTitle(task.title)
      setDescription(task.description || '')
      setStepsToReproduce(task.steps_to_reproduce || '')
      setSeverity(task.severity || 'medium')
      setEnvironment(task.environment || '')
    }
  }, [task])

  const { mutate: updateTask } = useMutation({
    mutationFn: (updates: Parameters<typeof tasksApi.update>[1]) =>
      taskId ? tasksApi.update(taskId, updates) : Promise.reject(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['task', taskId] })
      if (task?.project_id) {
        queryClient.invalidateQueries({ queryKey: ['tasks', task.project_id] })
      }
    },
  })

  const { mutate: toggleBlock } = useMutation({
    mutationFn: ({ isBlocked, reason }: { isBlocked: boolean; reason?: string }) =>
      taskId ? tasksApi.block(taskId, isBlocked, reason) : Promise.reject(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['task', taskId] })
      if (task?.project_id) {
        queryClient.invalidateQueries({ queryKey: ['tasks', task.project_id] })
      }
    },
  })

  const { mutate: toggleArchive } = useMutation({
    mutationFn: (isArchived: boolean) =>
      taskId ? tasksApi.archive(taskId, isArchived) : Promise.reject(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['task', taskId] })
      if (task?.project_id) {
        queryClient.invalidateQueries({ queryKey: ['tasks', task.project_id] })
      }
    },
  })

  const { mutate: deleteTask } = useMutation({
    mutationFn: () => (taskId ? tasksApi.delete(taskId) : Promise.reject()),
    onSuccess: () => {
      if (task?.project_id) {
        queryClient.invalidateQueries({ queryKey: ['tasks', task.project_id] })
        navigate(`/projects/${task.project_id}`)
      } else {
        navigate('/dashboard')
      }
    },
  })

  const { mutate: addSubtask } = useMutation({
    mutationFn: () =>
      taskId && newSubtaskTitle.trim()
        ? subtasksApi.create(taskId, { title: newSubtaskTitle.trim() })
        : Promise.reject(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['task', taskId] })
      setNewSubtaskTitle('')
    },
  })

  if (isLoading) {
    return (
      <div className="p-8 max-w-4xl mx-auto animate-pulse space-y-6">
        <div className="h-6 w-32 bg-slate-200 dark:bg-slate-800 rounded" />
        <div className="h-10 w-full bg-slate-100 dark:bg-slate-800 rounded" />
        <div className="h-48 w-full bg-slate-100 dark:bg-slate-800 rounded" />
      </div>
    )
  }

  if (!task) {
    return (
      <div className="p-8 text-center">
        <p className="text-slate-500">Task not found</p>
        <Link to="/dashboard" className="text-indigo-600 hover:underline mt-2 inline-block">
          Return to dashboard
        </Link>
      </div>
    )
  }

  const subtasks = task.subtasks || []
  const doneSubtasks = subtasks.filter((s) => s.is_done)

  return (
    <div className="p-6 md:p-8 max-w-4xl mx-auto min-w-0 pb-16">
      {/* Top Navigation & Meta */}
      <div className="flex items-center justify-between gap-4 mb-6 pb-4 border-b border-slate-200 dark:border-slate-800">
        <div className="flex items-center gap-3">
          <button
            onClick={() => (task.project_id ? navigate(`/projects/${task.project_id}`) : navigate('/dashboard'))}
            className="p-1.5 rounded-lg border border-slate-200 dark:border-slate-700 text-slate-500 hover:text-slate-900 dark:hover:text-slate-100 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
          </button>

          {task.project && (
            <Link
              to={`/projects/${task.project.id}`}
              className="flex items-center gap-1.5 text-xs text-slate-500 hover:text-indigo-600 transition-colors"
            >
              <span
                className="w-2 h-2 rounded-full"
                style={{ backgroundColor: task.project.color }}
              />
              <span>{task.project.name}</span>
            </Link>
          )}

          <span className="font-mono text-xs font-semibold px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
            {task.ticket_key}
          </span>
        </div>

        {/* Header Actions */}
        <div className="flex items-center gap-2">
          {task.is_blocked ? (
            <button
              onClick={() => toggleBlock({ isBlocked: false })}
              className="px-3 py-1.5 text-xs font-semibold rounded-lg bg-amber-500 hover:bg-amber-600 text-white transition-colors"
            >
              Unblock Task
            </button>
          ) : (
            <button
              onClick={() => {
                const reason = prompt('Enter reason for blocking (optional):')
                if (reason !== null) {
                  toggleBlock({ isBlocked: true, reason: reason.trim() })
                }
              }}
              className="px-3 py-1.5 text-xs font-medium rounded-lg border border-amber-300 dark:border-amber-700 text-amber-700 dark:text-amber-400 hover:bg-amber-50 dark:hover:bg-amber-950/40 transition-colors"
            >
              ⚠ Block Task
            </button>
          )}

          <button
            onClick={() => toggleArchive(!task.is_archived)}
            className={`px-3 py-1.5 text-xs font-medium rounded-lg border border-slate-200 dark:border-slate-700 transition-colors ${
              task.is_archived
                ? 'bg-slate-200 dark:bg-slate-700 text-slate-900 dark:text-slate-100'
                : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800'
            }`}
          >
            {task.is_archived ? 'Unarchive' : 'Archive'}
          </button>

          <Link
            to={`/tasks/${task.id}/focus`}
            className="flex items-center gap-1 px-3 py-1.5 text-xs font-medium text-slate-700 dark:text-slate-300 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors shadow-sm"
          >
            <Maximize2 className="w-3.5 h-3.5" />
            Focus Mode
          </Link>

          <button
            onClick={() => {
              if (confirm(`Delete ${task.ticket_key}?`)) deleteTask()
            }}
            className="p-1.5 rounded-lg border border-slate-200 dark:border-slate-700 text-slate-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-950/40 transition-colors"
            title="Delete ticket"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Main Ticket Card */}
      <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm mb-6">
        {/* Blocked Alert Banner if task is blocked */}
        {task.is_blocked && (
          <div className="mb-4 p-3 rounded-lg bg-amber-50 dark:bg-amber-950/40 border border-amber-300 dark:border-amber-800/80 flex items-center justify-between gap-3">
            <div className="flex items-center gap-2 text-xs text-amber-800 dark:text-amber-300">
              <span className="font-bold">⚠ Blocked:</span>
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
        {/* Title row */}
        <div className="mb-4">
          {isEditingTitle ? (
            <div className="flex items-center gap-2">
              <input
                autoFocus
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                className="flex-1 text-xl font-bold bg-transparent border-b-2 border-indigo-500 outline-none pb-1 text-slate-900 dark:text-slate-100"
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    const trimmed = title.trim()
                    if (trimmed && trimmed !== task.title) updateTask({ title: trimmed })
                    setIsEditingTitle(false)
                  }
                  if (e.key === 'Escape') {
                    setTitle(task.title)
                    setIsEditingTitle(false)
                  }
                }}
              />
              <button
                onClick={() => {
                  const trimmed = title.trim()
                  if (trimmed && trimmed !== task.title) updateTask({ title: trimmed })
                  setIsEditingTitle(false)
                }}
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

        {/* Controls Bar: Type, Priority, Status, Timer */}
        <div className="flex flex-wrap items-center justify-between gap-4 py-3 px-4 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-200/80 dark:border-slate-800 mb-6">
          <div className="flex flex-wrap items-center gap-3">
            {/* Type */}
            <div>
              <select
                value={task.type}
                onChange={(e) => updateTask({ type: e.target.value as TaskType })}
                className={`text-xs font-medium px-2 py-1 rounded border border-slate-200 dark:border-slate-700 outline-none cursor-pointer ${
                  TYPE_CONFIG[task.type]?.className
                }`}
              >
                <option value="task">Task</option>
                <option value="bug">Bug</option>
                <option value="improvement">Improvement</option>
                <option value="spike">Spike</option>
              </select>
            </div>

            {/* Priority */}
            <div>
              <select
                value={task.priority}
                onChange={(e) => updateTask({ priority: e.target.value as TaskPriority })}
                className="text-xs font-bold px-2 py-1 rounded border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 outline-none cursor-pointer uppercase"
              >
                <option value="p0">P0 - Critical</option>
                <option value="p1">P1 - High</option>
                <option value="p2">P2 - Medium</option>
                <option value="p3">P3 - Low</option>
              </select>
            </div>

            {/* Status */}
            <div>
              <select
                value={task.status}
                onChange={(e) => updateTask({ status: e.target.value as TaskStatus })}
                className={`text-xs font-medium px-2.5 py-1 rounded border border-slate-200 dark:border-slate-700 outline-none cursor-pointer ${
                  STATUS_CONFIG[task.status]?.className
                }`}
              >
                {STATUS_ORDER.map((s) => (
                  <option key={s} value={s}>
                    {STATUS_CONFIG[s].label}
                  </option>
                ))}
              </select>
            </div>
          </div>

          {/* Time Tracking Section */}
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-1.5 text-xs text-slate-500 font-mono">
              <Clock className="w-3.5 h-3.5" />
              <span>Total: {formatDuration(task.total_time_spent_seconds)}</span>
            </div>

            <div className="pl-3 border-l border-slate-200 dark:border-slate-700">
              <TimerControls taskId={task.id} taskQueryKey={['task', task.id]} />
            </div>
          </div>
        </div>

        {/* AI Ticket Summary Card (Phase 12) */}
        <div className="mb-6">
          <TaskSummaryCard taskId={task.id} initialSummary={task.ai_summary} />
        </div>

        {/* Description Section */}
        <div className="mb-6">
          <label className="block text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-2">
            Description
          </label>
          <textarea
            rows={4}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            onBlur={() => {
              if (description !== (task.description || '')) {
                updateTask({ description })
              }
            }}
            placeholder="Add detailed task description or requirements..."
            className="w-full text-sm p-3 rounded-lg border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/30 text-slate-800 dark:text-slate-200 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white dark:focus:bg-slate-900 transition-all"
          />
        </div>

        {/* Bug Specific Details (if bug) */}
        {task.type === 'bug' && (
          <div className="p-4 bg-red-50/40 dark:bg-red-950/20 border border-red-200 dark:border-red-900/50 rounded-xl space-y-4 mb-6">
            <div className="flex items-center gap-1.5 text-xs font-bold text-red-600 dark:text-red-400 uppercase tracking-wider">
              <AlertCircle className="w-4 h-4" />
              <span>Bug Details</span>
            </div>

            <div>
              <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
                Steps to Reproduce
              </label>
              <textarea
                rows={3}
                value={stepsToReproduce}
                onChange={(e) => setStepsToReproduce(e.target.value)}
                onBlur={() => updateTask({ steps_to_reproduce: stepsToReproduce })}
                className="w-full text-xs p-2.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-200 outline-none focus:ring-2 focus:ring-red-400"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
                  Severity
                </label>
                <select
                  value={severity}
                  onChange={(e) => {
                    setSeverity(e.target.value)
                    updateTask({ severity: e.target.value })
                  }}
                  className="w-full text-xs p-2 rounded border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-200 outline-none"
                >
                  <option value="critical">Critical</option>
                  <option value="major">Major</option>
                  <option value="medium">Medium</option>
                  <option value="minor">Minor</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
                  Environment
                </label>
                <input
                  value={environment}
                  onChange={(e) => setEnvironment(e.target.value)}
                  onBlur={() => updateTask({ environment })}
                  className="w-full text-xs p-2 rounded border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-200 outline-none"
                />
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Subtasks Section */}
      <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-6 shadow-sm space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <CheckCircle2 className="w-5 h-5 text-indigo-600" />
            <h2 className="font-semibold text-base text-slate-900 dark:text-slate-100">Subtasks</h2>
            <span className="text-xs text-slate-400 font-mono">
              ({doneSubtasks.length}/{subtasks.length})
            </span>
          </div>
        </div>

        {/* AI Subtask Suggestions Panel */}
        <AISubtaskPanel
          taskId={task.id}
          taskTitle={task.title}
          projectId={task.project_id}
          onAccepted={() => queryClient.invalidateQueries({ queryKey: ['task', task.id] })}
        />

        {/* Existing Subtask Rows */}
        <div className="space-y-1 divide-y divide-slate-100 dark:divide-slate-800/60">
          {subtasks.map((subtask) => (
            <SubtaskRow
              key={subtask.id}
              subtask={subtask}
              taskId={task.id}
              taskQueryKey={['task', task.id]}
            />
          ))}

          {subtasks.length === 0 && (
            <div className="py-6 text-center text-xs text-slate-400">
              No subtasks yet. Add one below or click "Suggest" in the AI panel above.
            </div>
          )}
        </div>

        {/* Manual Subtask Input */}
        <form
          onSubmit={(e) => {
            e.preventDefault()
            addSubtask()
          }}
          className="flex items-center gap-2 pt-2"
        >
          <input
            value={newSubtaskTitle}
            onChange={(e) => setNewSubtaskTitle(e.target.value)}
            placeholder="Add a new subtask..."
            className="flex-1 text-sm px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
          <button
            type="submit"
            disabled={!newSubtaskTitle.trim()}
            className="flex items-center gap-1 px-3.5 py-2 text-xs font-medium bg-slate-900 dark:bg-slate-100 text-white dark:text-slate-900 rounded-lg hover:opacity-90 disabled:opacity-40 transition-opacity"
          >
            <Plus className="w-3.5 h-3.5" />
            Add
          </button>
        </form>
      </div>

      {/* Dependencies Section (Phase 13) */}
      <TaskDependenciesPanel taskId={task.id} projectId={task.project_id} />

      {/* Obsidian-Style Notes Editor (Phase 11) */}
      <NotesEditor taskId={task.id} title="Ticket Notes & Scratchpad" />

      {/* Ticket Comments & Activity Thread (Phase 17) */}
      <CommentsSection taskId={task.id} />
    </div>
  )
}

