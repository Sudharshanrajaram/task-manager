import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { analyticsApi } from '../../api/analytics'
import ReactMarkdown from 'react-markdown'
import { X, Sparkles, Copy, Check, Loader2 } from 'lucide-react'

export default function StandupDialog({
  projectId,
  onClose,
}: {
  projectId?: string
  onClose: () => void
}) {
  const [report, setReport] = useState<string>('')
  const [copied, setCopied] = useState(false)

  const { mutate: generate, isPending, error } = useMutation({
    mutationFn: () => analyticsApi.getStandup(projectId),
    onSuccess: (data) => {
      setReport(data.report)
    },
  })

  const handleCopy = () => {
    if (!report) return
    navigator.clipboard.writeText(report)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
      <div className="bg-white dark:bg-slate-900 rounded-xl shadow-2xl border border-slate-200 dark:border-slate-800 w-full max-w-2xl p-6 flex flex-col max-h-[85vh]">
        <div className="flex items-center justify-between pb-4 border-b border-slate-100 dark:border-slate-800">
          <div className="flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
            <h2 className="font-semibold text-slate-900 dark:text-slate-100">Daily Standup Summary</h2>
          </div>
          <button onClick={onClose} className="p-1 rounded hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-400">
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="py-4 flex-1 overflow-y-auto">
          {isPending ? (
            <div className="py-16 flex flex-col items-center justify-center gap-3 text-slate-400">
              <Loader2 className="w-8 h-8 animate-spin text-indigo-600" />
              <p className="text-xs">Analyzing tracked time, completed tickets, and active blockers...</p>
            </div>
          ) : error ? (
            <div className="p-3 rounded-lg bg-red-50 dark:bg-red-950/40 border border-red-200 dark:border-red-900 text-xs text-red-600 dark:text-red-400">
              {(error as Error).message || 'Failed to generate standup report'}
            </div>
          ) : report ? (
            <div className="p-4 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-200 dark:border-slate-800 prose dark:prose-invert prose-sm max-w-none text-slate-800 dark:text-slate-200">
              <ReactMarkdown>{report}</ReactMarkdown>
            </div>
          ) : (
            <div className="text-center py-12 space-y-3">
              <p className="text-sm text-slate-600 dark:text-slate-400">
                Click below to synthesize today's completed tickets, in-progress tasks, and blockers into a structured engineering standup report.
              </p>
              <button
                onClick={() => generate()}
                className="px-4 py-2 text-xs font-semibold rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white shadow-sm transition-colors inline-flex items-center gap-2"
              >
                <Sparkles className="w-3.5 h-3.5" />
                <span>Generate Report Now</span>
              </button>
            </div>
          )}
        </div>

        {report && (
          <div className="flex items-center justify-between pt-4 border-t border-slate-100 dark:border-slate-800">
            <button
              onClick={() => generate()}
              disabled={isPending}
              className="px-3 py-1.5 text-xs text-slate-500 hover:text-slate-800 dark:hover:text-slate-200 font-medium"
            >
              Regenerate
            </button>
            <div className="flex gap-2">
              <button
                onClick={handleCopy}
                className="flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-semibold rounded-lg bg-slate-900 dark:bg-white text-white dark:text-slate-900 hover:opacity-90 transition-opacity"
              >
                {copied ? (
                  <>
                    <Check className="w-3.5 h-3.5 text-green-400 dark:text-green-600" />
                    <span>Copied!</span>
                  </>
                ) : (
                  <>
                    <Copy className="w-3.5 h-3.5" />
                    <span>Copy Markdown</span>
                  </>
                )}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
