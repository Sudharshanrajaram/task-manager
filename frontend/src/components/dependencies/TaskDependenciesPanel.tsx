import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { dependenciesApi } from '../../api/dependencies'
import { tasksApi } from '../../api/tasks'
import { Link } from 'react-router-dom'
import { Link2, Plus, Trash2, AlertCircle } from 'lucide-react'

interface TaskDependenciesPanelProps {
  taskId: string
  projectId: string
}

export default function TaskDependenciesPanel({ taskId, projectId }: TaskDependenciesPanelProps) {
  const queryClient = useQueryClient()
  const [selectedTargetId, setSelectedTargetId] = useState('')
  const [showAddForm, setShowAddForm] = useState(false)

  // Fetch dependencies for this task
  const { data, isLoading } = useQuery({
    queryKey: ['dependencies', taskId],
    queryFn: () => dependenciesApi.get(taskId),
  })

  // Fetch all tasks in this project to pick targets from
  const { data: projectTasks = [] } = useQuery({
    queryKey: ['tasks', projectId],
    queryFn: () => tasksApi.listByProject(projectId),
    enabled: !!projectId,
  })

  const { mutate: addDep, isPending: isAdding, error: addError } = useMutation({
    mutationFn: (targetId: string) => dependenciesApi.add(taskId, targetId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dependencies', taskId] })
      setSelectedTargetId('')
      setShowAddForm(false)
    },
  })

  const { mutate: removeDep } = useMutation({
    mutationFn: (depId: string) => dependenciesApi.remove(taskId, depId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dependencies', taskId] })
    },
  })

  const blockedBy = data?.blocked_by ?? []
  const blocks = data?.blocks ?? []

  // Candidates for "Blocked By" (excluding self and already added)
  const candidateTasks = projectTasks.filter(
    (t) => t.id !== taskId && !blockedBy.some((b) => b.depends_on_task_id === t.id)
  )

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-4 shadow-sm space-y-4">
      <div className="flex items-center justify-between pb-3 border-b border-slate-100 dark:border-slate-800">
        <div className="flex items-center gap-2">
          <Link2 className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
          <h3 className="text-sm font-semibold text-slate-900 dark:text-slate-100">Dependencies</h3>
        </div>

        <button
          onClick={() => setShowAddForm(!showAddForm)}
          className="flex items-center gap-1 text-xs font-medium text-indigo-600 dark:text-indigo-400 hover:text-indigo-700"
        >
          <Plus className="w-3.5 h-3.5" />
          <span>Add Dependency</span>
        </button>
      </div>

      {showAddForm && (
        <div className="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-700 space-y-2">
          <label className="block text-xs font-medium text-slate-600 dark:text-slate-400">
            This task is blocked by:
          </label>
          {addError && (
            <div className="p-2 rounded bg-red-50 dark:bg-red-950/40 border border-red-200 dark:border-red-900 text-[11px] text-red-600 dark:text-red-400 flex items-center gap-1">
              <AlertCircle className="w-3.5 h-3.5 shrink-0" />
              <span>{(addError as Error).message}</span>
            </div>
          )}
          <div className="flex gap-2">
            <select
              value={selectedTargetId}
              onChange={(e) => setSelectedTargetId(e.target.value)}
              className="flex-1 text-xs px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 outline-none"
            >
              <option value="">Select ticket...</option>
              {candidateTasks.map((t) => (
                <option key={t.id} value={t.id}>
                  [{t.ticket_key}] {t.title}
                </option>
              ))}
            </select>
            <button
              onClick={() => selectedTargetId && addDep(selectedTargetId)}
              disabled={!selectedTargetId || isAdding}
              className="px-3 py-1.5 text-xs font-medium bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg disabled:opacity-40 transition-colors"
            >
              {isAdding ? 'Adding...' : 'Link'}
            </button>
          </div>
        </div>
      )}

      {isLoading ? (
        <div className="h-16 bg-slate-100 dark:bg-slate-800/40 rounded-lg animate-pulse" />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Blocked by */}
          <div className="space-y-2">
            <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider block">
              Blocked by ({blockedBy.length})
            </span>
            {blockedBy.length === 0 ? (
              <p className="text-xs text-slate-400 italic">No blockers</p>
            ) : (
              <div className="space-y-1.5">
                {blockedBy.map((dep) => (
                  <div
                    key={dep.id}
                    className="flex items-center justify-between p-2 rounded-lg bg-amber-50/50 dark:bg-amber-950/20 border border-amber-200/60 dark:border-amber-900/40 text-xs"
                  >
                    <Link
                      to={`/tasks/${dep.depends_on_task_id}`}
                      className="font-mono font-semibold text-amber-800 dark:text-amber-300 hover:underline truncate max-w-[180px]"
                    >
                      {dep.depends_on_task?.ticket_key || 'Ticket'}:{' '}
                      <span className="font-normal font-sans text-slate-700 dark:text-slate-300">
                        {dep.depends_on_task?.title}
                      </span>
                    </Link>
                    <button
                      onClick={() => removeDep(dep.id)}
                      className="p-1 text-slate-400 hover:text-red-500 rounded"
                      title="Unlink dependency"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Blocks */}
          <div className="space-y-2">
            <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider block">
              Blocks ({blocks.length})
            </span>
            {blocks.length === 0 ? (
              <p className="text-xs text-slate-400 italic">Does not block any tasks</p>
            ) : (
              <div className="space-y-1.5">
                {blocks.map((dep) => (
                  <div
                    key={dep.id}
                    className="flex items-center justify-between p-2 rounded-lg bg-slate-50 dark:bg-slate-800/40 border border-slate-200 dark:border-slate-800 text-xs"
                  >
                    <Link
                      to={`/tasks/${dep.task_id}`}
                      className="font-mono font-semibold text-indigo-600 dark:text-indigo-400 hover:underline truncate max-w-[180px]"
                    >
                      {dep.task?.ticket_key || 'Ticket'}:{' '}
                      <span className="font-normal font-sans text-slate-700 dark:text-slate-300">
                        {dep.task?.title}
                      </span>
                    </Link>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
