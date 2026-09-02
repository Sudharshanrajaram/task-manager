import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { logsApi } from '../api/logs'
import { projectsApi } from '../api/projects'
import { Link } from 'react-router-dom'
import { Calendar, Download, Clock, Archive, Loader2, Pencil, Sparkles } from 'lucide-react'
import EditTimerModal from '../components/timers/EditTimerModal'

export default function DailyLogs() {
  const queryClient = useQueryClient()
  const [selectedProjectId, setSelectedProjectId] = useState<string>('')
  const [dateRange, setDateRange] = useState<'today' | 'week' | 'month' | 'all'>('all')
  const [isExporting, setIsExporting] = useState(false)
  const [exportError, setExportError] = useState<string | null>(null)
  const [editingItem, setEditingItem] = useState<{
    entryId: string
    taskKey: string
    taskTitle: string
    durationSeconds: number
  } | null>(null)

  // Calculate from & to dates based on filter
  const today = new Date().toISOString().split('T')[0]
  let fromDate: string | undefined = undefined
  let toDate: string | undefined = undefined

  if (dateRange === 'today') {
    fromDate = today
    toDate = today
  } else if (dateRange === 'week') {
    const d = new Date()
    d.setDate(d.getDate() - 7)
    fromDate = d.toISOString().split('T')[0]
    toDate = today
  } else if (dateRange === 'month') {
    const d = new Date()
    d.setDate(d.getDate() - 30)
    fromDate = d.toISOString().split('T')[0]
    toDate = today
  }

  const { data: projects = [] } = useQuery({
    queryKey: ['projects'],
    queryFn: projectsApi.list,
  })

  const { data: logs = [], isLoading } = useQuery({
    queryKey: ['daily-logs', fromDate, toDate, selectedProjectId],
    queryFn: () => logsApi.getDaily(fromDate, toDate, selectedProjectId || undefined),
  })

  const { mutate: triggerArchive, isPending: isArchiving } = useMutation({
    mutationFn: logsApi.triggerArchive,
    onSuccess: (data) => {
      alert(`Auto-archiver executed! ${data.archived_count} completed tasks older than 14 days were archived.`)
      queryClient.invalidateQueries({ queryKey: ['daily-logs'] })
    },
  })

  const handleExport = async () => {
    try {
      setIsExporting(true)
      setExportError(null)
      await logsApi.downloadExcel(fromDate, toDate, selectedProjectId || undefined)
    } catch (err) {
      setExportError((err as Error).message || 'Failed to download Excel file')
    } finally {
      setIsExporting(false)
    }
  }

  const totalSeconds = logs.reduce((acc, l) => acc + (l.total_duration_seconds || 0), 0)
  const totalHours = Math.floor(totalSeconds / 3600)
  const totalMinutes = Math.floor((totalSeconds % 3600) / 60)

  return (
    <div className="max-w-7xl mx-auto space-y-6">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <Calendar className="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
            <span>Live Activity Grid ("Live Excel")</span>
          </h1>
          <p className="text-xs text-slate-500 mt-1">
            Spreadsheet-grid view of ticket sessions. Correct logged durations inline; download as formatted .xlsx anytime.
          </p>
        </div>

        {/* Action Controls */}
        <div className="flex items-center gap-2.5">
          <button
            onClick={() => triggerArchive()}
            disabled={isArchiving}
            title="Manual trigger for 14-day Done task auto-archiver worker"
            className="flex items-center gap-1.5 px-3 py-2 text-xs font-medium rounded-lg border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300 transition-colors"
          >
            <Archive className="w-3.5 h-3.5 text-amber-500" />
            <span>{isArchiving ? 'Archiving...' : 'Run Auto-Archiver'}</span>
          </button>

          <button
            onClick={handleExport}
            disabled={isExporting}
            className="flex items-center gap-2 px-3.5 py-2 text-xs font-semibold rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white shadow-sm transition-all disabled:opacity-50"
            title="Download snapshot as clean Excel .xlsx (Phases 18 & 20)"
          >
            {isExporting ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                <span>Exporting...</span>
              </>
            ) : (
              <>
                <Download className="w-4 h-4" />
                <span>Export Excel (.xlsx)</span>
              </>
            )}
          </button>
        </div>
      </div>

      {exportError && (
        <div className="p-3 bg-red-50 dark:bg-red-950/50 border border-red-200 dark:border-red-900/50 rounded-xl text-xs text-red-600 dark:text-red-400">
          {exportError}
        </div>
      )}

      {/* Filter Toolbar */}
      <div className="flex flex-wrap items-center justify-between gap-3 p-4 bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm">
        {/* Date Presets */}
        <div className="flex items-center gap-1 bg-slate-100 dark:bg-slate-800 p-1 rounded-lg text-xs">
          {(['today', 'week', 'month', 'all'] as const).map((range) => (
            <button
              key={range}
              onClick={() => setDateRange(range)}
              className={`px-3 py-1.5 rounded-md font-medium capitalize transition-colors ${
                dateRange === range
                  ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 shadow-sm'
                  : 'text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'
              }`}
            >
              {range === 'all' ? 'All Time' : range}
            </button>
          ))}
        </div>

        {/* Project Selector */}
        <div className="flex items-center gap-2">
          <select
            value={selectedProjectId}
            onChange={(e) => setSelectedProjectId(e.target.value)}
            className="text-xs bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 text-slate-800 dark:text-slate-200 outline-none cursor-pointer"
          >
            <option value="">All Projects</option>
            {projects.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name} ({p.key})
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Summary Stat Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm">
          <span className="text-xs text-slate-400 font-medium">Total Time Tracked</span>
          <div className="flex items-baseline gap-2 mt-1">
            <Clock className="w-4 h-4 text-indigo-500" />
            <span className="text-xl font-bold font-mono text-slate-900 dark:text-slate-100">
              {totalHours}h {String(totalMinutes).padStart(2, '0')}m
            </span>
          </div>
        </div>
        <div className="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm">
          <span className="text-xs text-slate-400 font-medium">Ticket Entries</span>
          <div className="text-xl font-bold text-slate-900 dark:text-slate-100 mt-1">
            {logs.length}
          </div>
        </div>
        <div className="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm">
          <span className="text-xs text-slate-400 font-medium">Auto-Archiving Status</span>
          <div className="text-xs text-emerald-600 dark:text-emerald-400 font-semibold mt-1">
            Active · Daily 24h ticker
          </div>
        </div>
      </div>

      {/* Live Editable Spreadsheet Grid */}
      <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs border-collapse">
            <thead className="bg-slate-50 dark:bg-slate-800/60 border-b border-slate-200 dark:border-slate-800 text-slate-500 font-semibold uppercase tracking-wider text-[10px]">
              <tr>
                <th className="py-3 px-4">Date</th>
                <th className="py-3 px-4">Project</th>
                <th className="py-3 px-4">Ticket</th>
                <th className="py-3 px-4">Task Title</th>
                <th className="py-3 px-4">
                  <div className="flex items-center gap-1 text-indigo-600 dark:text-indigo-400">
                    <Clock className="w-3 h-3" />
                    <span>Duration (Editable)</span>
                  </div>
                </th>
                <th className="py-3 px-4">
                  <div className="flex items-center gap-1">
                    <Sparkles className="w-3 h-3 text-amber-500" />
                    <span>AI Summary</span>
                  </div>
                </th>
                <th className="py-3 px-4">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
              {isLoading ? (
                <tr>
                  <td colSpan={7} className="py-8 text-center text-slate-400">
                    Loading live spreadsheet grid...
                  </td>
                </tr>
              ) : logs.length === 0 ? (
                <tr>
                  <td colSpan={7} className="py-8 text-center text-slate-400">
                    No activity recorded for this filter. Start a timer on any ticket to see live logs!
                  </td>
                </tr>
              ) : (
                logs.map((log, i) => {
                  const h = Math.floor(log.total_duration_seconds / 3600)
                  const m = Math.floor((log.total_duration_seconds % 3600) / 60)
                  const cleanDuration = `${h}h ${String(m).padStart(2, '0')}m`

                  return (
                    <tr
                      key={i}
                      className="hover:bg-slate-50/80 dark:hover:bg-slate-800/40 transition-colors group"
                    >
                      {/* Date */}
                      <td className="py-3 px-4 font-mono text-slate-600 dark:text-slate-400 whitespace-nowrap">
                        {log.date}
                      </td>

                      {/* Project */}
                      <td className="py-3 px-4 whitespace-nowrap">
                        <div className="flex items-center gap-1.5">
                          <span
                            className="w-2 h-2 rounded-full shrink-0"
                            style={{ backgroundColor: log.project_color || '#4F46E5' }}
                          />
                          <span className="font-medium text-slate-800 dark:text-slate-200">
                            {log.project_name}
                          </span>
                        </div>
                      </td>

                      {/* Ticket */}
                      <td className="py-3 px-4 whitespace-nowrap">
                        <Link
                          to={`/tasks/${log.task_id}`}
                          className="font-mono font-semibold text-indigo-600 dark:text-indigo-400 hover:underline"
                        >
                          {log.ticket_key}
                        </Link>
                      </td>

                      {/* Task Title */}
                      <td className="py-3 px-4 max-w-xs truncate">
                        <span className="font-medium text-slate-900 dark:text-slate-100">
                          {log.task_title}
                        </span>
                      </td>

                      {/* Duration (Editable Cell) */}
                      <td className="py-3 px-4 whitespace-nowrap">
                        <button
                          type="button"
                          onClick={() => {
                            if (log.latest_entry_id) {
                              setEditingItem({
                                entryId: log.latest_entry_id,
                                taskKey: log.ticket_key,
                                taskTitle: log.task_title,
                                durationSeconds: log.total_duration_seconds,
                              })
                            }
                          }}
                          className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-slate-100 dark:bg-slate-800 hover:bg-indigo-50 dark:hover:bg-indigo-950/50 hover:text-indigo-600 dark:hover:text-indigo-400 font-mono font-semibold text-slate-800 dark:text-slate-200 border border-slate-200/80 dark:border-slate-700 transition-colors cursor-pointer group-hover:border-indigo-400"
                          title="Click to edit duration directly in grid"
                        >
                          <span>{cleanDuration}</span>
                          <Pencil className="w-3 h-3 opacity-0 group-hover:opacity-100 text-indigo-500 transition-opacity" />
                        </button>
                      </td>

                      {/* AI Summary */}
                      <td className="py-3 px-4 max-w-sm truncate text-slate-600 dark:text-slate-400">
                        {log.ai_summary ? (
                          <span className="truncate block" title={log.ai_summary}>
                            {log.ai_summary}
                          </span>
                        ) : (
                          <span className="text-slate-300 dark:text-slate-600 italic">—</span>
                        )}
                      </td>

                      {/* Status */}
                      <td className="py-3 px-4 whitespace-nowrap">
                        <span className="px-2 py-0.5 rounded text-[10px] font-medium uppercase bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
                          {log.status}
                        </span>
                      </td>
                    </tr>
                  )
                })
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Edit Duration Modal for Grid */}
      {editingItem && (
        <EditTimerModal
          entryId={editingItem.entryId}
          initialDurationSeconds={editingItem.durationSeconds}
          taskKey={editingItem.taskKey}
          taskTitle={editingItem.taskTitle}
          onClose={() => setEditingItem(null)}
        />
      )}
    </div>
  )
}
