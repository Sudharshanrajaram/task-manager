import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { projectsApi } from '../../api/projects'
import { X, AlertCircle } from 'lucide-react'

const PROJECT_COLORS = [
  '#4F46E5', '#7C3AED', '#DB2777', '#DC2626',
  '#D97706', '#16A34A', '#0891B2', '#475569',
]

export default function CreateProjectDialog({ onClose }: { onClose: () => void }) {
  const queryClient = useQueryClient()
  const [name, setName] = useState('')
  const [key, setKey] = useState('')
  const [color, setColor] = useState(PROJECT_COLORS[0])

  const { mutate, isPending, error } = useMutation({
    mutationFn: () =>
      projectsApi.create({
        name: name.trim(),
        key: key.trim().toUpperCase(),
        color,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
      onClose()
    },
  })

  const handleKeyAuto = (inputName: string) => {
    setName(inputName)
    const words = inputName.trim().split(/\s+/).filter(Boolean)
    if (words.length === 1) {
      const clean = words[0].replace(/[^a-zA-Z0-9]/g, '').toUpperCase()
      if (clean.length >= 2) {
        setKey(clean.slice(0, 4))
      } else {
        setKey(clean)
      }
    } else if (words.length > 1) {
      const initials = words.map((w) => w[0]).join('').replace(/[^a-zA-Z0-9]/g, '').toUpperCase()
      if (initials.length >= 2) {
        setKey(initials.slice(0, 5))
      } else {
        const clean = words[0].replace(/[^a-zA-Z0-9]/g, '').toUpperCase()
        setKey(clean.slice(0, 4))
      }
    }
  }

  const isValidKey = /^[A-Z0-9]{2,10}$/.test(key.trim().toUpperCase())

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
      <div className="bg-white dark:bg-slate-900 rounded-xl shadow-2xl border border-slate-200 dark:border-slate-700 w-full max-w-sm p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="font-semibold text-slate-900 dark:text-slate-100">New Project</h2>
          <button onClick={onClose} className="p-1 rounded hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-400">
            <X className="w-4 h-4" />
          </button>
        </div>

        {error && (
          <div className="mb-4 p-3 rounded-lg bg-red-50 dark:bg-red-950/40 border border-red-200 dark:border-red-900/50 flex items-start gap-2 text-xs text-red-600 dark:text-red-400">
            <AlertCircle className="w-4 h-4 shrink-0 mt-0.5" />
            <span>{(error as Error).message || 'Failed to create project'}</span>
          </div>
        )}

        <div className="space-y-3">
          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
              Project Name <span className="text-red-500">*</span>
            </label>
            <input
              autoFocus
              value={name}
              onChange={(e) => handleKeyAuto(e.target.value)}
              placeholder="e.g. Auth Service"
              className="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">
              Key (short ticket prefix) <span className="text-red-500">*</span>
            </label>
            <input
              value={key}
              onChange={(e) => setKey(e.target.value.toUpperCase().replace(/[^A-Z0-9]/g, '').slice(0, 10))}
              placeholder="e.g. AUTH"
              className="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 font-mono uppercase"
            />
            <span className="text-[11px] text-slate-400 dark:text-slate-500 mt-0.5 block">
              2–10 uppercase letters or numbers (e.g. AUTH, BILL)
            </span>
          </div>
          <div>
            <label className="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Color</label>
            <div className="flex gap-2 flex-wrap">
              {PROJECT_COLORS.map((c) => (
                <button
                  key={c}
                  type="button"
                  onClick={() => setColor(c)}
                  className={`w-6 h-6 rounded-full transition-all ${color === c ? 'ring-2 ring-offset-2 ring-indigo-500 scale-110' : 'hover:scale-105'}`}
                  style={{ backgroundColor: c }}
                />
              ))}
            </div>
          </div>
        </div>

        <div className="flex justify-end gap-2 mt-6 pt-4 border-t border-slate-100 dark:border-slate-800">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={() => mutate()}
            disabled={!name.trim() || !isValidKey || isPending}
            className="px-4 py-2 text-sm rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-40 font-medium transition-all shadow-sm"
          >
            {isPending ? 'Creating...' : 'Create Project'}
          </button>
        </div>
      </div>
    </div>
  )
}
