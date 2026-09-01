import { apiClient } from './client'
import type { AnalyticsSummary } from '../types'

export const analyticsApi = {
  getSummary: () =>
    apiClient.get<AnalyticsSummary>('/analytics/summary').then((r) => r.data),

  getStandup: (projectId?: string) =>
    apiClient
      .post<{ report: string }>('/analytics/standup', null, {
        params: projectId ? { project_id: projectId } : undefined,
      })
      .then((r) => r.data),
}

