import { apiClient } from './client'
import type { DailyLogItem } from '../types'
export type { DailyLogItem } from '../types'

export const logsApi = {
  getDaily: (from?: string, to?: string, projectId?: string) =>
    apiClient
      .get<{ logs: DailyLogItem[] }>('/logs/daily', {
        params: { from, to, project_id: projectId },
      })
      .then((r) => r.data.logs ?? []),

  downloadExcel: async (from?: string, to?: string, projectId?: string) => {
    const tz = Intl.DateTimeFormat().resolvedOptions().timeZone
    const response = await apiClient.get('/logs/export', {
      params: { from, to, project_id: projectId, format: 'xlsx', tz },
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
