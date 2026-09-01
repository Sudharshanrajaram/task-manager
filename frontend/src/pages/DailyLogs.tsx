import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { logsApi } from '../api/logs'
import { projectsApi } from '../api/projects'
import { Link } from 'react-router-dom'
import { formatDuration } from '../lib/utils'
import { Calendar, Download, Filter, Clock } from 'lucide-react'

export default function DailyLogs() {
  const [selectedProjectId, setSelectedProjectId] = useState<string>('')
  const [dateRange, setDateRange] = useState<'today' | 'week' | 'month' | 'all'>('week')

  // Calculate from & to dates based on filter
  const today = new Date().toISOString().split('T')[0]
  let fromDate: string | undefined = undefined
  const toDate = today

  if (dateRange === 'today') {
    fromDate = today
  } else if (dateRange === 'week') {
    const d = new Date()
    d.setDate(d.getDate() - 7)
    fromDate = d.toISOString().split('T')[0]
  } else if (dateRange === 'month') {
    const d = new Date()
    d.setDate(d.getDate() - 30)
    fromDate = d.toISOString().split('T')[0]
  }

  const { data: projects = [] } = useQuery({
    queryKey: ['projects'],
    queryFn: projectsApi.list,
  })

  const { data: logs = [], isLoading } = useQuery({
    queryKey: ['daily-logs', fromDate, toDate, selectedProjectId],
    queryFn: () => logsApi.getDaily(fromDate, toDate, selectedProjectId || undefined),
  })

  const totalSeconds = logs.reduce((acc, l) => acc + l.total_duration_seconds, 0)
  const exportUrl = logsApi.getExportUrl(fromDate, toDate, selectedProjectId || undefined)

  return (
    <div className="max-w-7xl mx-auto space-y-6">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <Calendar className="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
            <span>Daily Activity Logs</span>
          </h1>
          <p className="text-xs text-slate-500 mt-1">
            Track and export developer work sessions aggregated by day, project, and ticket.
          </p>
        </div>

        {/* Action Controls */}
        <div className="flex items-center gap-3">
          <a
            href={exportUrl}
            download
            className="flex items-center gap-2 px-3.5 py-2 text-xs font-semibold rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white shadow-sm transition-all"
          >
            <Download className="w-3.5 h-3.5" />
            <span>Export to Excel (.xlsx)</span>
          </a>
        </div>
      </div>

      {/* Filter Toolbar */}
      <div className="flex flex-wrap items-center justify-between gap-3 p-3.5 bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm">
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-xs text-slate-400 font-medium flex items-center gap-1 mr-1">
            <Filter className="w-3 h-3" />
            Time Range:
          </span>
          {(['today', 'week', 'month', 'all'] as const).map((r) => (
            <button
              key={r}
              onClick={() => setDateRange(r)}
              className={`px-3 py-1 text-xs rounded-lg font-medium capitalize transition-colors ${
                dateRange === r
                  ? 'bg-indigo-50 text-indigo-600 dark:bg-indigo-950 dark:text-indigo-400 border border-indigo-200 dark:border-indigo-800'
                  : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800'
              }`}
            >
              {r === 'today' ? 'Today' : r === 'week' ? 'Past 7 Days' : r === 'month' ? 'Past 30 Days' : 'All Time'}
            </button>
          ))}
        </div>

        <div className="flex items-center gap-2">
          <label className="text-xs text-slate-400 font-medium">Project:</label>
          <select
            value={selectedProjectId}
            onChange={(e) => setSelectedProjectId(e.target.value)}
            className="text-xs px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-800 dark:text-slate-200 outline-none cursor-pointer"
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

      {/* Summary Stat */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm">
          <span className="text-xs text-slate-400 font-medium">Total Time Tracked</span>
          <div className="flex items-baseline gap-2 mt-1">
            <Clock className="w-4 h-4 text-indigo-500" />
            <span className="text-xl font-bold font-mono text-slate-900 dark:text-slate-100">
              {formatDuration(totalSeconds)}
            </span>
          </div>
        </div>
        <div className="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm">
          <span className="text-xs text-slate-400 font-medium">Sessions Logged</span>
          <div className="text-xl font-bold text-slate-900 dark:text-slate-100 mt-1">
            {logs.length}
          </div>
        </div>
        <div className="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm">
          <span className="text-xs text-slate-400 font-medium">Auto-Archiving Status</span>
          <div className="text-xs text-green-600 dark:text-green-400 font-semibold mt-1">
            Active (Tasks &gt;14d auto-archived)
          </div>
        </div>
      </div>

      {/* Activity Table */}
      <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-slate-50 dark:bg-slate-800/60 border-b border-slate-200 dark:border-slate-800 text-slate-500 font-semibold uppercase">
              <tr>
                <th className="py-3 px-4">Date</th>
                <th className="py-3 px-4">Project</th>
                <th className="py-3 px-4">Ticket</th>
                <th className="py-3 px-4">Task & Subtask</th>
                <th className="py-3 px-4">Duration</th>
                <th className="py-3 px-4">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
              {isLoading ? (
                <tr>
                  <td colSpan={6} className="py-8 text-center text-slate-400">
                    Loading activity logs...
                  </td>
                </tr>
              ) : logs.length === 0 ? (
                <tr>
                  <td colSpan={6} className="py-8 text-center text-slate-400">
                    No time entries recorded for this period. Start tracking time to populate daily logs!
                  </td>
                </tr>
              ) : (
                logs.map((log, i) => (
                  <tr key={i} className="hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors">
                    <td className="py-3 px-4 font-mono text-slate-600 dark:text-slate-400 whitespace-nowrap">
                      {log.date}
                    </td>
                    <td className="py-3 px-4 whitespace-nowrap">
                      <div className="flex items-center gap-1.5">
                        <span
                          className="w-2 h-2 rounded-full"
                          style={{ backgroundColor: log.project_color || '#4F46E5' }}
                        />
                        <span className="font-medium text-slate-800 dark:text-slate-200">
                          {log.project_name}
                        </span>
                      </div>
                    </td>
                    <td className="py-3 px-4 whitespace-nowrap">
                      <Link
                        to={`/tasks/${log.task_id}`}
                        className="font-mono font-semibold text-indigo-600 dark:text-indigo-400 hover:underline"
                      >
                        {log.ticket_key}
                      </Link>
                    </td>
                    <td className="py-3 px-4 max-w-xs truncate">
                      <div className="font-medium text-slate-900 dark:text-slate-100">{log.task_title}</div>
                      {log.subtask_title && (
                        <div className="text-[11px] text-slate-400 truncate">↳ {log.subtask_title}</div>
                      )}
                    </td>
                    <td className="py-3 px-4 font-mono font-semibold text-slate-800 dark:text-slate-200 whitespace-nowrap">
                      {formatDuration(log.total_duration_seconds)}
                    </td>
                    <td className="py-3 px-4 whitespace-nowrap">
                      <span className="px-2 py-0.5 rounded text-[10px] font-medium uppercase bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
                        {log.status}
                      </span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
