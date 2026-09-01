import { useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { tasksApi } from '../api/tasks'
import { subtasksApi } from '../api/subtasks'
import { useTimerStore } from '../store/timerStore'
import { timersApi } from '../api/timers'
import { formatDurationHMS, STATUS_CONFIG, PRIORITY_CONFIG } from '../lib/utils'
import { Play, Pause, Square, Check, X, ArrowLeft, CheckCircle2 } from 'lucide-react'
import NotesEditor from '../components/notes/NotesEditor'

export default function FocusMode() {
  const { taskId } = useParams<{ taskId: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { activeTimers, localElapsed, updateTimer, removeTimer } = useTimerStore()

  // Esc key exits Focus Mode
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        navigate(`/tasks/${taskId}`)
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [taskId, navigate])

  const { data: task, isLoading } = useQuery({
    queryKey: ['task', taskId],
    queryFn: () => (taskId ? tasksApi.getById(taskId) : Promise.reject('No task ID')),
    enabled: !!taskId,
  })

  const { mutate: toggleSubtaskDone } = useMutation({
    mutationFn: ({ subtaskId, isDone }: { subtaskId: string; isDone: boolean }) =>
      subtasksApi.update(subtaskId, { is_done: isDone }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['task', taskId] })
    },
  })

  const currentTimer = activeTimers.find(
    (t) => t.task_id === taskId && !t.subtask_id
  )
  const elapsed = currentTimer
    ? localElapsed[currentTimer.entry_id] ?? currentTimer.elapsed_seconds
    : 0
  const isRunning = currentTimer && !currentTimer.is_paused

  const handleStart = async () => {
    if (!taskId) return
    const started = await timersApi.start(taskId)
    updateTimer(started)
    queryClient.invalidateQueries({ queryKey: ['task', taskId] })
  }

  const handlePause = async () => {
    if (!currentTimer) return
    const updated = await timersApi.pause(currentTimer.entry_id)
    updateTimer(updated)
  }

  const handleResume = async () => {
    if (!currentTimer) return
    const updated = await timersApi.resume(currentTimer.entry_id)
    updateTimer(updated)
  }

  const handleStop = async () => {
    if (!currentTimer) return
    await timersApi.stop(currentTimer.entry_id)
    removeTimer(currentTimer.entry_id)
    queryClient.invalidateQueries({ queryKey: ['task', taskId] })
  }

  if (isLoading || !task) {
    return (
      <div className="h-screen bg-slate-950 flex items-center justify-center text-slate-400">
        Loading focus mode...
      </div>
    )
  }

  const subtasks = task.subtasks || []
  const doneCount = subtasks.filter((s) => s.is_done).length

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col justify-between p-8 md:p-12 select-none">
      {/* Top Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Link
            to={`/tasks/${task.id}`}
            className="flex items-center gap-1.5 text-xs text-slate-400 hover:text-slate-200 transition-colors p-2 rounded-lg bg-slate-900 border border-slate-800"
          >
            <ArrowLeft className="w-4 h-4" />
            <span>Exit Focus (ESC)</span>
          </Link>
          <span className="font-mono text-xs px-2.5 py-1 rounded bg-slate-900 border border-slate-800 text-slate-400">
            {task.ticket_key}
          </span>
          {task.project && (
            <span className="text-xs text-slate-500">
              {task.project.name}
            </span>
          )}
        </div>

        <div className="flex items-center gap-2">
          <span
            className={`text-xs px-2.5 py-1 rounded-full font-medium ${
              STATUS_CONFIG[task.status]?.className
            }`}
          >
            {STATUS_CONFIG[task.status]?.label}
          </span>
          <span
            className={`text-xs px-2.5 py-1 rounded-full font-bold uppercase ${
              PRIORITY_CONFIG[task.priority]?.className
            }`}
          >
            {PRIORITY_CONFIG[task.priority]?.label}
          </span>
        </div>
      </div>

      {/* Main Focus Center */}
      <div className="max-w-2xl mx-auto w-full text-center my-auto py-8">
        {/* Task Title */}
        <h1 className="text-3xl md:text-4xl font-bold tracking-tight text-white mb-6">
          {task.title}
        </h1>

        {/* Big Timer */}
        <div className="my-8">
          <div className="font-mono text-7xl md:text-8xl font-black tracking-wider text-indigo-400 tabular-nums">
            {formatDurationHMS(elapsed)}
          </div>
          <p className="text-xs text-slate-500 font-mono mt-2 uppercase tracking-widest">
            {isRunning ? 'Timer Running' : currentTimer ? 'Paused' : 'Timer Inactive'}
          </p>
        </div>

        {/* Big Timer Controls */}
        <div className="flex items-center justify-center gap-4 mb-12">
          {!currentTimer ? (
            <button
              onClick={handleStart}
              className="flex items-center gap-2 px-8 py-3.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl font-semibold shadow-lg shadow-indigo-500/20 transition-all hover:scale-105 active:scale-95"
            >
              <Play className="w-5 h-5 fill-current" />
              Start Focus Session
            </button>
          ) : (
            <>
              {currentTimer.is_paused ? (
                <button
                  onClick={handleResume}
                  className="flex items-center gap-2 px-8 py-3.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl font-semibold shadow-lg shadow-indigo-500/20 transition-all hover:scale-105"
                >
                  <Play className="w-5 h-5 fill-current" />
                  Resume
                </button>
              ) : (
                <button
                  onClick={handlePause}
                  className="flex items-center gap-2 px-8 py-3.5 bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-xl font-semibold border border-slate-700 transition-all hover:scale-105"
                >
                  <Pause className="w-5 h-5 fill-current" />
                  Pause
                </button>
              )}
              <button
                onClick={handleStop}
                className="flex items-center gap-2 px-6 py-3.5 bg-red-950/40 hover:bg-red-900/60 text-red-400 border border-red-900/50 rounded-xl font-semibold transition-all hover:scale-105"
              >
                <Square className="w-4 h-4 fill-current" />
                Finish & Save
              </button>
            </>
          )}
        </div>

        {/* Distraction-Free Subtask Checklist */}
        {subtasks.length > 0 && (
          <div className="bg-slate-900/80 border border-slate-800/80 rounded-2xl p-6 text-left max-h-72 overflow-y-auto scrollbar-thin">
            <div className="flex items-center justify-between mb-3 text-xs text-slate-400 pb-2 border-b border-slate-800">
              <span className="font-semibold uppercase tracking-wider">Subtasks</span>
              <span>
                {doneCount} of {subtasks.length} completed
              </span>
            </div>

            <div className="space-y-2">
              {subtasks.map((s) => (
                <div
                  key={s.id}
                  onClick={() => toggleSubtaskDone({ subtaskId: s.id, isDone: !s.is_done })}
                  className={`flex items-center gap-3 p-2.5 rounded-lg cursor-pointer transition-colors ${
                    s.is_done
                      ? 'bg-slate-950/40 text-slate-500 line-through'
                      : 'hover:bg-slate-800/60 text-slate-200'
                  }`}
                >
                  <div
                    className={`w-4 h-4 rounded border flex items-center justify-center shrink-0 ${
                      s.is_done
                        ? 'bg-green-600 border-green-600 text-white'
                        : 'border-slate-600 hover:border-green-500'
                    }`}
                  >
                    {s.is_done && <Check className="w-3 h-3" />}
                  </div>
                  <span className="text-sm select-none">{s.title}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Obsidian-Style Notes for Focus Session (Phase 11) */}
        <div className="w-full text-left">
          <NotesEditor taskId={taskId} title="Focus Notes" />
        </div>
      </div>

      {/* Footer */}
      <div className="text-center text-xs text-slate-600 font-mono">
        TaskFlow Focus Mode · Distraction-free personal workflow
      </div>
    </div>
  )
}

