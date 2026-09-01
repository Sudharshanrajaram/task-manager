import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { ragApi } from '../../api/rag'
import type { SubtaskSuggestionResult } from '../../types'
import { Sparkles, ChevronDown, Check, X, Loader2 } from 'lucide-react'

interface AISubtaskPanelProps {
  taskId: string
  taskTitle: string
  projectId: string
  onAccepted: () => void
}

export default function AISubtaskPanel({ taskId, taskTitle, projectId, onAccepted }: AISubtaskPanelProps) {
  const queryClient = useQueryClient()
  const [result, setResult] = useState<SubtaskSuggestionResult | null>(null)
  const [editedSubtasks, setEditedSubtasks] = useState<string[]>([])
  const [count, setCount] = useState(5)
  const [showContext, setShowContext] = useState(false)

  const { mutate: suggest, isPending: isSuggesting } = useMutation({
    mutationFn: () => ragApi.suggestForTask(taskId, count),
    onSuccess: (data) => {
      setResult(data)
      setEditedSubtasks([...data.suggested_subtasks])
    },
  })

  const { mutate: accept, isPending: isAccepting } = useMutation({
    mutationFn: () => ragApi.accept(taskId, editedSubtasks.filter((s) => s.trim())),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['task', taskId] })
      setResult(null)
      setEditedSubtasks([])
      onAccepted()
    },
  })

  const updateSubtask = (idx: number, value: string) => {
    setEditedSubtasks((prev) => prev.map((s, i) => (i === idx ? value : s)))
  }

  const removeSubtask = (idx: number) => {
    setEditedSubtasks((prev) => prev.filter((_, i) => i !== idx))
  }

  return (
    <div className="border border-slate-200 dark:border-slate-700 rounded-xl overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-gradient-to-r from-indigo-50 to-purple-50 dark:from-indigo-950/40 dark:to-purple-950/40 border-b border-slate-200 dark:border-slate-700">
        <div className="flex items-center gap-2">
          <Sparkles className="w-4 h-4 text-indigo-500" />
          <span className="text-sm font-medium text-slate-800 dark:text-slate-200">
            AI Subtask Suggestions
          </span>
          {result && (
            <span className="text-xs text-slate-500 dark:text-slate-400">
              grounded on {result.grounded_context.length} similar task{result.grounded_context.length !== 1 ? 's' : ''}
            </span>
          )}
        </div>
        {!result && (
          <div className="flex items-center gap-2">
            <select
              value={count}
              onChange={(e) => setCount(Number(e.target.value))}
              className="text-xs bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded px-2 py-1 text-slate-700 dark:text-slate-300"
            >
              {[3, 4, 5, 6, 8].map((n) => (
                <option key={n} value={n}>{n} subtasks</option>
              ))}
            </select>
            <button
              onClick={() => suggest()}
              disabled={isSuggesting}
              className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg disabled:opacity-60 transition-colors"
            >
              {isSuggesting ? (
                <Loader2 className="w-3 h-3 animate-spin" />
              ) : (
                <Sparkles className="w-3 h-3" />
              )}
              {isSuggesting ? 'Generating...' : 'Suggest'}
            </button>
          </div>
        )}
      </div>

      {/* Suggestions */}
      {result && (
        <div className="p-4 space-y-3">
          <div className="space-y-2">
            {editedSubtasks.map((sub, idx) => (
              <div key={idx} className="flex items-center gap-2 group">
                <span className="w-5 h-5 rounded-full bg-indigo-100 dark:bg-indigo-900 text-indigo-700 dark:text-indigo-300 text-xs flex items-center justify-center font-medium shrink-0">
                  {idx + 1}
                </span>
                <input
                  value={sub}
                  onChange={(e) => updateSubtask(idx, e.target.value)}
                  className="flex-1 text-sm bg-transparent border-0 border-b border-transparent focus:border-indigo-400 dark:focus:border-indigo-600 outline-none py-0.5 text-slate-800 dark:text-slate-200 placeholder-slate-400 transition-colors"
                  placeholder="Subtask title..."
                />
                <button
                  onClick={() => removeSubtask(idx)}
                  className="p-0.5 rounded opacity-0 group-hover:opacity-100 text-slate-400 hover:text-red-500 transition-all"
                >
                  <X className="w-3.5 h-3.5" />
                </button>
              </div>
            ))}
          </div>

          {/* Grounded context accordion */}
          {result.grounded_context.length > 0 && (
            <div className="mt-3">
              <button
                onClick={() => setShowContext((v) => !v)}
                className="flex items-center gap-1 text-xs text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 transition-colors"
              >
                <ChevronDown className={`w-3 h-3 transition-transform ${showContext ? 'rotate-180' : ''}`} />
                View {result.grounded_context.length} source task{result.grounded_context.length !== 1 ? 's' : ''} used for context
              </button>
              {showContext && (
                <div className="mt-2 space-y-2">
                  {result.grounded_context.map((ctx, i) => (
                    <div key={i} className="bg-slate-50 dark:bg-slate-800/60 rounded-lg p-3">
                      <div className="flex items-center justify-between mb-1.5">
                        <span className="text-xs font-medium text-slate-700 dark:text-slate-300">
                          {ctx.source_title}
                        </span>
                        <span className="text-xs text-slate-400 font-mono">
                          {(ctx.similarity_score * 100).toFixed(0)}% similar
                        </span>
                      </div>
                      <ul className="space-y-0.5">
                        {ctx.final_subtasks.map((s, j) => (
                          <li key={j} className="text-xs text-slate-500 dark:text-slate-400 flex items-center gap-1.5">
                            <span className="w-1 h-1 rounded-full bg-slate-400 shrink-0" />
                            {s}
                          </li>
                        ))}
                      </ul>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* Actions */}
          <div className="flex items-center justify-between pt-2 border-t border-slate-100 dark:border-slate-800">
            <button
              onClick={() => { setResult(null); setEditedSubtasks([]) }}
              className="text-xs text-slate-500 hover:text-slate-700 dark:hover:text-slate-300"
            >
              Dismiss
            </button>
            <div className="flex gap-2">
              <button
                onClick={() => suggest()}
                disabled={isSuggesting}
                className="flex items-center gap-1 px-3 py-1.5 text-xs border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800"
              >
                Regenerate
              </button>
              <button
                onClick={() => accept()}
                disabled={isAccepting || editedSubtasks.filter((s) => s.trim()).length === 0}
                className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg disabled:opacity-60 transition-colors"
              >
                {isAccepting ? (
                  <Loader2 className="w-3 h-3 animate-spin" />
                ) : (
                  <Check className="w-3 h-3" />
                )}
                Accept {editedSubtasks.filter((s) => s.trim()).length} subtasks
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
