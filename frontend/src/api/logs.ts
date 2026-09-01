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

  downloadExcel: async (from?: string, to?: string, projectId?: string) => {
    const response = await apiClient.get('/logs/export', {
      params: { from, to, project_id: projectId, format: 'xlsx' },
      responseType: 'blob',
    })

    const blob = new Blob([response.data], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.setAttribute(
      'download',
      `taskflow-activity-${new Date().toISOString().split('T')[0]}.xlsx`
    )
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
  },

  triggerArchive: () =>
    apiClient.post<{ message: string; archived_count: number }>('/logs/archive-trigger').then((r) => r.data),
}
