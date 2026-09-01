import { apiClient } from './client'

export interface DailyLogItem {
  date: string
  project_id: string
  project_name: string
  project_key: string
  project_color: string
  task_id: string
  ticket_key: string
  task_title: string
  subtask_title: string
  total_duration_seconds: number
  status: string
}

export const logsApi = {
  getDaily: (from?: string, to?: string, projectId?: string) =>
    apiClient
      .get<{ logs: DailyLogItem[] }>('/logs/daily', {
        params: { from, to, project_id: projectId },
      })
      .then((r) => r.data.logs ?? []),

  getExportUrl: (from?: string, to?: string, projectId?: string) => {
    const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
    const params = new URLSearchParams()
    if (from) params.set('from', from)
    if (to) params.set('to', to)
    if (projectId) params.set('project_id', projectId)
    params.set('format', 'xlsx')
    return `${baseURL}/api/logs/export?${params.toString()}`
  },
}
