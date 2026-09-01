import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { tasksApi } from '../../api/tasks'
import { Sparkles, Loader2, RefreshCw } from 'lucide-react'

interface TaskSummaryCardProps {
  taskId: string
  initialSummary?: string
}

export default function TaskSummaryCard({ taskId, initialSummary }: TaskSummaryCardProps) {
  const queryClient = useQueryClient()
  const [summary, setSummary] = useState(initialSummary || '')
  const [isCached, setIsCached] = useState(false)

  const { mutate: generateSummary, isPending, error } = useMutation({
    mutationFn: () => tasksApi.summarize(taskId),
    onSuccess: (data) => {
      setSummary(data.summary)
      setIsCached(data.from_cache)
      queryClient.invalidateQueries({ queryKey: ['task', taskId] })
    },
  })

  return (
    <div className="rounded-xl border border-indigo-100 dark:border-indigo-950/60 bg-gradient-to-r from-indigo-50/50 via-purple-50/30 to-slate-50/50 dark:from-indigo-950/20 dark:via-purple-950/10 dark:to-slate-900/40 p-4 shadow-sm">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <Sparkles className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
          <span className="text-xs font-semibold text-indigo-950 dark:text-indigo-200">
            AI Ticket Summary
          </span>
          {isCached && (
            <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-indigo-100 dark:bg-indigo-900 text-indigo-700 dark:text-indigo-300">
              Cached
            </span>
          )}
        </div>

        {summary ? (
          <button
            onClick={() => generateSummary()}
            disabled={isPending}
            className="flex items-center gap-1 text-[11px] text-slate-500 hover:text-indigo-600 dark:hover:text-indigo-400"
            title="Refresh AI Summary"
          >
            <RefreshCw className={`w-3 h-3 ${isPending ? 'animate-spin' : ''}`} />
            <span>Regenerate</span>
          </button>
        ) : (
          <button
            onClick={() => generateSummary()}
            disabled={isPending}
            className="flex items-center gap-1 px-2.5 py-1 rounded-lg text-xs font-medium bg-indigo-600 hover:bg-indigo-700 text-white shadow-sm transition-colors"
          >
            {isPending ? (
              <>
                <Loader2 className="w-3 h-3 animate-spin" />
                <span>Summarizing...</span>
              </>
            ) : (
              <>
                <Sparkles className="w-3 h-3" />
                <span>✦ Summarize</span>
              </>
            )}
          </button>
        )}
      </div>

      {error && (
        <div className="text-xs text-red-600 dark:text-red-400 mt-1">
          {(error as Error).message || 'Failed to generate summary'}
        </div>
      )}

      {summary ? (
        <p className="text-xs text-slate-700 dark:text-slate-300 leading-relaxed font-sans">
          {summary}
        </p>
      ) : (
        !isPending && (
          <p className="text-xs text-slate-400 italic">
            Click '✦ Summarize' for a 1–2 sentence plain-English explanation of this ticket's purpose and scope.
          </p>
        )
      )}
    </div>
  )
}
