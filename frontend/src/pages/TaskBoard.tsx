import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { projectsApi } from '../api/projects'
import { tasksApi } from '../api/tasks'
import { useUIStore } from '../store/uiStore'
import { STATUS_CONFIG, STATUS_ORDER } from '../lib/utils'
import type { TaskStatus } from '../types'
import TaskCard from '../components/tasks/TaskCard'
import CreateTaskDialog from '../components/tasks/CreateTaskDialog'
import { Plus, ArrowLeft } from 'lucide-react'

export default function TaskBoard() {
  const { projectId } = useParams<{ projectId: string }>()
  const { setSelectedProject } = useUIStore()
  const [showCreateDialog, setShowCreateDialog] = useState(false)

  useEffect(() => {
    if (projectId) {
      setSelectedProject(projectId)
    }
    return () => setSelectedProject(null)
  }, [projectId, setSelectedProject])

  const { data: project } = useQuery({
    queryKey: ['project', projectId],
    queryFn: () => (projectId ? projectsApi.getById(projectId) : Promise.reject('No project ID')),
    enabled: !!projectId,
  })

  const { data: tasks = [], isLoading } = useQuery({
    queryKey: ['tasks', projectId],
    queryFn: () => (projectId ? tasksApi.listByProject(projectId) : Promise.resolve([])),
    enabled: !!projectId,
  })

  if (!projectId) return <div>Project not found</div>

  return (
    <div className="p-6 h-full flex flex-col min-w-0">
      {/* Top Header */}
      <div className="flex items-center justify-between mb-6 pb-4 border-b border-slate-200 dark:border-slate-800 shrink-0">
        <div className="flex items-center gap-3">
          <Link
            to="/dashboard"
            className="p-1.5 rounded-lg border border-slate-200 dark:border-slate-700 text-slate-500 hover:text-slate-900 dark:hover:text-slate-100 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
          </Link>

          {project ? (
            <div className="flex items-center gap-2">
              <span
                className="w-3 h-3 rounded-full shrink-0"
                style={{ backgroundColor: project.color }}
              />
              <h1 className="text-lg font-semibold text-slate-900 dark:text-slate-100">
                {project.name}
              </h1>
              <span className="font-mono text-xs px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400">
                {project.key}
              </span>
              <span className="text-xs text-slate-400 ml-2">
                {tasks.length} ticket{tasks.length !== 1 ? 's' : ''}
              </span>
            </div>
          ) : (
            <div className="h-6 w-36 bg-slate-200 dark:bg-slate-800 rounded animate-pulse" />
          )}
        </div>

        <button
          onClick={() => setShowCreateDialog(true)}
          className="flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-medium bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg transition-colors shadow-sm"
        >
          <Plus className="w-3.5 h-3.5" />
          New Ticket
        </button>
      </div>

      {/* Kanban Columns */}
      {isLoading ? (
        <div className="grid grid-cols-5 gap-4 flex-1">
          {STATUS_ORDER.map((s) => (
            <div
              key={s}
              className="bg-slate-100 dark:bg-slate-800/50 rounded-xl p-3 animate-pulse h-96"
            />
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-5 gap-4 flex-1 items-start overflow-x-auto pb-4 scrollbar-thin">
          {STATUS_ORDER.map((status: TaskStatus) => {
            const columnTasks = tasks.filter((t) => t.status === status)
            const config = STATUS_CONFIG[status]

            return (
              <div
                key={status}
                className="flex flex-col bg-slate-100/70 dark:bg-slate-950/40 rounded-xl border border-slate-200/80 dark:border-slate-800/80 p-3 min-w-[250px] max-h-full"
              >
                {/* Column Header */}
                <div className="flex items-center justify-between mb-3 px-1">
                  <div className="flex items-center gap-2">
                    <span
                      className={`text-xs font-semibold px-2 py-0.5 rounded-md ${config.className}`}
                    >
                      {config.label}
                    </span>
                    <span className="text-xs font-mono text-slate-400">
                      {columnTasks.length}
                    </span>
                  </div>
                </div>

                {/* Task List */}
                <div className="space-y-2.5 overflow-y-auto flex-1 pr-1 scrollbar-thin">
                  {columnTasks.map((task) => (
                    <TaskCard key={task.id} task={task} projectId={projectId} />
                  ))}

                  {columnTasks.length === 0 && (
                    <div className="py-8 text-center border-2 border-dashed border-slate-200 dark:border-slate-800 rounded-lg">
                      <span className="text-xs text-slate-400">No tickets</span>
                    </div>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      )}

      {showCreateDialog && (
        <CreateTaskDialog
          projectId={projectId}
          onClose={() => setShowCreateDialog(false)}
        />
      )}
    </div>
  )
}
