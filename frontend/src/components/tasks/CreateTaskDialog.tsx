import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { tasksApi } from '../../api/tasks'
import type { TaskPriority, TaskType } from '../../types'
import { X } from 'lucide-react'

interface CreateTaskDialogProps {
  projectId: string
  onClose: () => void
  onCreated?: (taskId: string) => void
}

export default function CreateTaskDialog({ projectId, onClose, onCreated }: CreateTaskDialogProps) {
  const queryClient = useQueryClient()
  const [title, setTitle] = useState('')
  const [type, setType] = useState<TaskType>('task')
  const [priority, setPriority] = useState<TaskPriority>('p2')
  const [description, setDescription] = useState('')
  const [labelInput, setLabelInput] = useState('')
  const [labels, setLabels] = useState<string[]>([])

  // Bug-specific fields
  const [stepsToReproduce, setStepsToReproduce] = useState('')
  const [severity, setSeverity] = useState('medium')
  const [environment, setEnvironment] = useState('production')

  const { mutate, isPending } = useMutation({
    mutationFn: () =>
      tasksApi.create(projectId, {
        title: title.trim(),
        type,
        priority,
        description: description.trim(),
        labels,
      }),
    onSuccess: (createdTask) => {
      // If bug, we can update the bug-specific fields
      if (type === 'bug' && (stepsToReproduce || severity || environment)) {
        tasksApi.update(createdTask.id, {
          steps_to_reproduce: stepsToReproduce,
          severity,
          environment,
        })
      }
      queryClient.invalidateQueries({ queryKey: ['tasks', projectId] })
      onClose()
      if (onCreated) {
        onCreated(createdTask.id)
      }
    },
  })

  const handleAddLabel = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault()
      const trimmed = labelInput.trim().replace(/^,|,$/g, '')
      if (trimmed && !labels.includes(trimmed)) {
        setLabels([...labels, trimmed])
        setLabelInput('')
      }
    }
  }

  const removeLabel = (labelToRemove: string) => {
    setLabels(labels.filter((l) => l !== labelToRemove))
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
      <div className="bg-white dark:bg-slate-900 rounded-xl shadow-2xl border border-slate-200 dark:border-slate-700 w-full max-w-lg overflow-hidden flex flex-col max-h-[90vh]">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700">
          <h2 className="font-semibold text-base text-slate-900 dark:text-slate-100">Create New Ticket</h2>
          <button onClick={onClose} className="p-1 rounded hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-400">
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Form Body */}
        <div className="p-6 space-y-4 overflow-y-auto scrollbar-thin">
          {/* Title */}
          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
              Title <span className="text-red-500">*</span>
            </label>
            <input
              autoFocus
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="e.g. Implement OAuth2 Refresh Token Rotation"
              className="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>

          {/* Type and Priority */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Type</label>
              <select
                value={type}
                onChange={(e) => setType(e.target.value as TaskType)}
                className="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500 capitalize"
              >
                <option value="task">Task</option>
                <option value="bug">Bug</option>
                <option value="improvement">Improvement</option>
                <option value="spike">Spike</option>
              </select>
            </div>

            <div>
              <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Priority</label>
              <select
                value={priority}
                onChange={(e) => setPriority(e.target.value as TaskPriority)}
                className="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500 uppercase"
              >
                <option value="p0">P0 - Critical</option>
                <option value="p1">P1 - High</option>
                <option value="p2">P2 - Medium</option>
                <option value="p3">P3 - Low</option>
              </select>
            </div>
          </div>

          {/* Description */}
          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Description</label>
            <textarea
              rows={3}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Provide technical context or requirements..."
              className="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>

          {/* Bug-Specific Fields */}
          {type === 'bug' && (
            <div className="p-3 bg-red-50/50 dark:bg-red-950/20 border border-red-200 dark:border-red-900/50 rounded-lg space-y-3">
              <span className="text-xs font-semibold text-red-700 dark:red-400 uppercase tracking-wider block">
                Bug Details
              </span>

              <div>
                <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
                  Steps to Reproduce
                </label>
                <textarea
                  rows={2}
                  value={stepsToReproduce}
                  onChange={(e) => setStepsToReproduce(e.target.value)}
                  placeholder="1. Login\n2. Click expired session..."
                  className="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-red-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Severity</label>
                  <select
                    value={severity}
                    onChange={(e) => setSeverity(e.target.value)}
                    className="w-full px-3 py-1.5 text-xs rounded border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100"
                  >
                    <option value="critical">Critical</option>
                    <option value="major">Major</option>
                    <option value="medium">Medium</option>
                    <option value="minor">Minor</option>
                  </select>
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Environment</label>
                  <input
                    value={environment}
                    onChange={(e) => setEnvironment(e.target.value)}
                    placeholder="e.g. staging, prod-us-east"
                    className="w-full px-3 py-1.5 text-xs rounded border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100"
                  />
                </div>
              </div>
            </div>
          )}

          {/* Labels */}
          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
              Labels (press Enter or comma)
            </label>
            <div className="flex flex-wrap gap-1.5 mb-2">
              {labels.map((l) => (
                <span
                  key={l}
                  className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300"
                >
                  {l}
                  <button type="button" onClick={() => removeLabel(l)} className="hover:text-red-500">
                    <X className="w-3 h-3" />
                  </button>
                </span>
              ))}
            </div>
            <input
              value={labelInput}
              onChange={(e) => setLabelInput(e.target.value)}
              onKeyDown={handleAddLabel}
              placeholder="e.g. backend, auth, p1"
              className="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-2 px-6 py-4 border-t border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-950/50">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={() => mutate()}
            disabled={!title.trim() || isPending}
            className="px-4 py-2 text-sm rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-50 font-medium transition-colors"
          >
            {isPending ? 'Creating...' : 'Create Ticket'}
          </button>
        </div>
      </div>
    </div>
  )
}

