import { useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { tasksApi } from '../api/tasks'
import { subtasksApi } from '../api/subtasks'
import { useTimerStore } from '../store/timerStore'
import { timersApi } from '../api/timers'
import { formatDurationHMS, STATUS_CONFIG, PRIORITY_CONFIG } from '../lib/utils'
import { Play, Pause, Square, Check, X, CheckCircle2, ListTodo, Sparkles } from 'lucide-react'
import NotesEditor from '../components/notes/NotesEditor'

export default function FocusMode() {
  const { taskId } = useParams<{ taskId: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { activeTimers, localElapsed, updateTimer, removeTimer } = useTimerStore()

  // Esc key exits Focus Mode and locks body scroll
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        navigate(task?.project_id ? `/projects/${task.project_id}` : '/dashboard')
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    const prevOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'

    return () => {
      window.removeEventListener('keydown', handleKeyDown)
      document.body.style.overflow = prevOverflow
    }
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

  const handleExit = () => {
    if (task?.project_id) {
      navigate(`/projects/${task.project_id}`)
    } else {
      navigate('/dashboard')
    }
  }

  if (isLoading || !task) {
    return (
      <div className="fixed inset-0 z-50 bg-slate-950 flex items-center justify-center text-slate-400 font-mono text-sm">
        Entering Focus Mode...
      </div>
    )
  }

  const subtasks = task.subtasks || []
  const doneCount = subtasks.filter((s) => s.is_done).length

  return (
    <div className="fixed inset-0 z-50 bg-slate-950 text-slate-100 flex flex-col justify-between p-6 md:p-10 w-screen h-screen overflow-y-auto">
      {/* Top Header Bar */}
      <div className="flex items-center justify-between mb-4 shrink-0 max-w-6xl mx-auto w-full">
        <div className="flex items-center gap-3">
          <span className="font-mono text-xs px-2.5 py-1 rounded bg-slate-900 border border-slate-800 text-slate-400 font-semibold">
            {task.ticket_key}
          </span>
          {task.project && (
            <div className="flex items-center gap-1.5 text-xs text-slate-400">
              <span
                className="w-2 h-2 rounded-full"
                style={{ backgroundColor: task.project.color }}
              />
              <span>{task.project.name}</span>
            </div>
          )}
        </div>

        <div className="flex items-center gap-3">
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

          <button
            onClick={handleExit}
            className="flex items-center gap-1.5 text-xs text-slate-400 hover:text-slate-100 transition-colors px-3 py-1.5 rounded-lg bg-slate-900 hover:bg-slate-800 border border-slate-800 ml-2"
            title="Exit Focus Mode (Esc)"
          >
            <span>Exit</span>
            <kbd className="font-mono text-[10px] text-slate-500 bg-slate-800 px-1 rounded">ESC</kbd>
            <X className="w-3.5 h-3.5 ml-0.5" />
          </button>
        </div>
      </div>

      {/* Center Stopwatch Area */}
      <div className="text-center py-6 shrink-0 max-w-3xl mx-auto w-full">
        <h1 className="text-xl md:text-2xl font-bold tracking-tight text-white mb-3 max-w-2xl mx-auto line-clamp-2">
          {task.title}
        </h1>

        {/* Big Stopwatch Display */}
        <div className="my-5">
          <div className="font-mono text-6xl sm:text-7xl md:text-8xl font-black tracking-wider text-slate-100 tabular-nums select-none drop-shadow-sm">
            {formatDurationHMS(elapsed)}
          </div>

          {/* Status badge with pulse */}
          <div className="flex items-center justify-center gap-2 mt-3">
            {isRunning ? (
              <span className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-mono font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                <span className="relative flex h-2 w-2">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
                </span>
                Active Session
              </span>
            ) : currentTimer ? (
              <span className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-mono font-medium bg-amber-500/10 text-amber-400 border border-amber-500/20">
                <span className="w-2 h-2 rounded-full bg-amber-400" />
                Session Paused
              </span>
            ) : (
              <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-mono font-medium bg-slate-900 text-slate-400 border border-slate-800">
                Session Ready
              </span>
            )}
          </div>
        </div>

        {/* Big Controls */}
        <div className="flex items-center justify-center gap-3">
          {!currentTimer ? (
            <button
              onClick={handleStart}
              className="flex items-center gap-2 px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl font-semibold shadow-lg shadow-indigo-600/20 transition-all hover:scale-105 active:scale-95 text-sm"
            >
              <Play className="w-4 h-4 fill-current" />
              Start Focus Session
            </button>
          ) : (
            <>
              {currentTimer.is_paused ? (
                <button
                  onClick={handleResume}
                  className="flex items-center gap-2 px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl font-semibold shadow-lg shadow-indigo-600/20 transition-all hover:scale-105 active:scale-95 text-sm"
                >
                  <Play className="w-4 h-4 fill-current" />
                  Resume
                </button>
              ) : (
                <button
                  onClick={handlePause}
                  className="flex items-center gap-2 px-6 py-2.5 bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-xl font-semibold border border-slate-700 transition-all hover:scale-105 active:scale-95 text-sm"
                >
                  <Pause className="w-4 h-4 fill-current" />
                  Pause
                </button>
              )}
              <button
                onClick={handleStop}
                className="flex items-center gap-2 px-5 py-2.5 bg-rose-950/40 hover:bg-rose-900/60 text-rose-400 border border-rose-900/50 rounded-xl font-semibold transition-all hover:scale-105 active:scale-95 text-sm"
              >
                <Square className="w-3.5 h-3.5 fill-current" />
                Finish & Save
              </button>
            </>
          )}
        </div>
      </div>

      {/* Sibling Two-Column Layout (Subtasks + Notes side by side) */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 w-full max-w-6xl mx-auto items-stretch flex-1 pb-4">
        {/* Left Column: Distraction-Free Subtask Checklist */}
        <div className="bg-slate-900 rounded-2xl border border-slate-800 p-5 shadow-sm flex flex-col h-[360px]">
          <div className="flex items-center justify-between pb-3 border-b border-slate-800 mb-3 shrink-0">
            <div className="flex items-center gap-2">
              <ListTodo className="w-4 h-4 text-indigo-400" />
              <span className="text-sm font-semibold text-slate-100">
                Checklist Subtasks
              </span>
            </div>
            <span className="text-xs font-mono px-2 py-0.5 rounded-full bg-slate-800 text-slate-400">
              {doneCount} / {subtasks.length} done
            </span>
          </div>

          <div className="flex-1 overflow-y-auto space-y-2 pr-1 scrollbar-thin">
            {subtasks.length === 0 ? (
              <div className="h-full flex flex-col items-center justify-center text-xs text-slate-400 text-center py-8">
                <CheckCircle2 className="w-6 h-6 text-slate-600 mb-2" />
                <span>No subtasks for this ticket.</span>
                <span className="text-slate-500 mt-1">Use the Notes panel on the right for scratchpad thoughts!</span>
              </div>
            ) : (
              subtasks.map((s) => (
                <div
                  key={s.id}
                  onClick={() => toggleSubtaskDone({ subtaskId: s.id, isDone: !s.is_done })}
                  className={`flex items-center gap-3 p-3 rounded-xl cursor-pointer transition-colors border ${
                    s.is_done
                      ? 'bg-slate-950/40 border-slate-800/60 text-slate-500 line-through'
                      : 'bg-slate-800/40 hover:bg-slate-800 border-slate-800 text-slate-100'
                  }`}
                >
                  <div
                    className={`w-4 h-4 rounded border flex items-center justify-center shrink-0 transition-colors ${
                      s.is_done
                        ? 'bg-emerald-600 border-emerald-600 text-white'
                        : 'border-slate-600 hover:border-emerald-500'
                    }`}
                  >
                    {s.is_done && <Check className="w-3 h-3" />}
                  </div>
                  <span className="text-xs font-medium select-none flex-1 leading-snug">
                    {s.title}
                  </span>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Right Column: Obsidian-Style Notes Editor */}
        <NotesEditor
          taskId={taskId}
          title="Focus Notes"
          className="h-[360px]"
        />
      </div>

      {/* Footer */}
      <div className="text-center text-xs text-slate-600 font-mono py-2 shrink-0">
        TaskFlow Focus Mode · Distraction-free full-screen environment
      </div>
    </div>
  )
}
