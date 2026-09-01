import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { projectsApi } from '../api/projects'
import { tasksApi } from '../api/tasks'
import { timersApi } from '../api/timers'
import { useTimerStore } from '../store/timerStore'
import { formatDuration, STATUS_CONFIG, TYPE_CONFIG } from '../lib/utils'
import { Plus, Clock, CheckSquare, Layers, Sparkles } from 'lucide-react'
import { useState } from 'react'
import type { Project } from '../types'
import CreateTaskDialog from '../components/tasks/CreateTaskDialog'
import StandupDialog from '../components/analytics/StandupDialog'

export default function Dashboard() {
  const { data: projects = [], isLoading } = useQuery({
    queryKey: ['projects'],
    queryFn: projectsApi.list,
  })
  const { data: analytics } = useQuery({
    queryKey: ['analytics', 'week'],
    queryFn: () => timersApi.getAnalytics('week'),
  })
  const { activeTimers, localElapsed } = useTimerStore()
  const [createTaskProjectId, setCreateTaskProjectId] = useState<string | null>(null)
  const [showStandup, setShowStandup] = useState(false)

  if (isLoading) return <LoadingState />

  const timeByProject: Record<string, number> = {}
  analytics?.by_project?.forEach((p) => { timeByProject[p.project_id] = p.total_time_seconds })

  return (
    <div className="p-6 max-w-6xl mx-auto">
      {/* Page header */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-slate-900 dark:text-slate-100">Dashboard</h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
            {projects.length} project{projects.length !== 1 ? 's' : ''} · {activeTimers.length} timer{activeTimers.length !== 1 ? 's' : ''} running
          </p>
        </div>
        <button
          onClick={() => setShowStandup(true)}
          className="flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-semibold rounded-lg bg-indigo-50 dark:bg-indigo-950/60 border border-indigo-200 dark:border-indigo-800 text-indigo-700 dark:text-indigo-300 hover:bg-indigo-100 dark:hover:bg-indigo-900 transition-colors shadow-sm"
        >
          <Sparkles className="w-3.5 h-3.5" />
          <span>Daily Standup</span>
        </button>
      </div>

      {/* Weekly analytics summary bar */}
      {analytics && analytics.total_time_spent_seconds > 0 && (
        <div className="mb-6 p-4 rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900">
          <div className="flex items-center justify-between mb-3">
            <span className="text-sm font-medium text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
              <Clock className="w-4 h-4 text-indigo-500" />
              This week
            </span>
            <span className="text-lg font-semibold font-mono text-slate-900 dark:text-slate-100 tabular-nums">
              {formatDuration(analytics.total_time_spent_seconds)}
            </span>
          </div>
          <div className="flex gap-2 flex-wrap">
            {analytics.by_type?.filter((t) => t.total_time_seconds > 0).map((t) => (
              <span
                key={t.type}
                className="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400"
              >
                <span className={`w-1.5 h-1.5 rounded-full ${TYPE_CONFIG[t.type as keyof typeof TYPE_CONFIG]?.className || 'bg-slate-400'}`} />
                {t.type}: {formatDuration(t.total_time_seconds)}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Projects grid */}
      {projects.length === 0 ? (
        <EmptyProjects />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {projects.map((project) => (
            <ProjectCard
              key={project.id}
              project={project}
              weekSeconds={timeByProject[project.id] ?? 0}
              activeTimerCount={activeTimers.filter((t) => t.project_key === project.key).length}
              onCreateTask={() => setCreateTaskProjectId(project.id)}
            />
          ))}
        </div>
      )}

      {createTaskProjectId && (
        <CreateTaskDialog
          projectId={createTaskProjectId}
          onClose={() => setCreateTaskProjectId(null)}
        />
      )}

      {showStandup && (
        <StandupDialog onClose={() => setShowStandup(false)} />
      )}
    </div>
  )
}

function ProjectCard({
  project, weekSeconds, activeTimerCount, onCreateTask,
}: {
  project: Project
  weekSeconds: number
  activeTimerCount: number
  onCreateTask: () => void
}) {
  const { data: tasks = [] } = useQuery({
    queryKey: ['tasks', project.id],
    queryFn: () => tasksApi.listByProject(project.id),
  })

  const doneTasks = tasks.filter((t) => t.status === 'done').length
  const openTasks = tasks.length - doneTasks

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden hover:border-slate-300 dark:hover:border-slate-600 transition-colors">
      {/* Color header */}
      <div className="h-1" style={{ backgroundColor: project.color }} />
      <div className="p-4">
        <div className="flex items-start justify-between mb-3">
          <div>
            <Link
              to={`/projects/${project.id}`}
              className="font-semibold text-slate-900 dark:text-slate-100 hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors"
            >
              {project.name}
            </Link>
            <span className="ml-2 font-mono text-xs text-slate-400 dark:text-slate-500">{project.key}</span>
          </div>
          {activeTimerCount > 0 && (
            <span className="flex items-center gap-1 text-xs text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-950 px-2 py-0.5 rounded-full">
              <span className="w-1.5 h-1.5 rounded-full bg-indigo-500 animate-pulse" />
              {activeTimerCount} running
            </span>
          )}
        </div>

        <div className="flex items-center gap-4 text-sm text-slate-500 dark:text-slate-400 mb-4">
          <span className="flex items-center gap-1">
            <CheckSquare className="w-3.5 h-3.5" />
            {openTasks} open
          </span>
          {weekSeconds > 0 && (
            <span className="flex items-center gap-1 font-mono">
              <Clock className="w-3.5 h-3.5" />
              {formatDuration(weekSeconds)} this week
            </span>
          )}
        </div>

        {/* Recent tasks */}
        {tasks.slice(0, 3).map((task) => (
          <Link
            key={task.id}
            to={`/tasks/${task.id}`}
            className="flex items-center gap-2 py-1 hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors group"
          >
            <span className="font-mono text-xs text-slate-400 dark:text-slate-500 w-16 shrink-0">{task.ticket_key}</span>
            <span className="text-xs text-slate-700 dark:text-slate-300 truncate group-hover:text-indigo-600 dark:group-hover:text-indigo-400">
              {task.title}
            </span>
            <span className={`ml-auto text-xs px-1.5 py-0.5 rounded-full shrink-0 ${STATUS_CONFIG[task.status]?.className}`}>
              {STATUS_CONFIG[task.status]?.label}
            </span>
          </Link>
        ))}

        <div className="flex items-center justify-between mt-3 pt-3 border-t border-slate-100 dark:border-slate-800">
          <Link
            to={`/projects/${project.id}`}
            className="text-xs text-slate-500 dark:text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-400"
          >
            View all {tasks.length} tasks →
          </Link>
          <button
            onClick={onCreateTask}
            className="flex items-center gap-1 text-xs text-slate-500 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-950 px-2 py-1 rounded transition-colors"
          >
            <Plus className="w-3 h-3" />
            New task
          </button>
        </div>
      </div>
    </div>
  )
}

function EmptyProjects() {
  return (
    <div className="flex flex-col items-center justify-center py-24 text-center">
      <Layers className="w-10 h-10 text-slate-300 dark:text-slate-600 mb-4" />
      <h3 className="text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">No projects yet</h3>
      <p className="text-xs text-slate-500 dark:text-slate-400">
        Click the + button in the sidebar to create your first project.
      </p>
    </div>
  )
}

function LoadingState() {
  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="h-6 w-32 bg-slate-200 dark:bg-slate-700 rounded animate-pulse mb-6" />
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="h-48 bg-slate-100 dark:bg-slate-800 rounded-xl animate-pulse" />
        ))}
      </div>
    </div>
  )
}
