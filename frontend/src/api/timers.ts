import { apiClient } from './client'
import type { ActiveTimerInfo, TimeEntry, AnalyticsSummary } from '../types'

export const timersApi = {
  start: (taskId: string, subtaskId?: string) =>
    apiClient
      .post<ActiveTimerInfo>('/timers/start', { task_id: taskId, subtask_id: subtaskId })
      .then((r) => r.data),
  pause: (entryId: string) =>
    apiClient.post<ActiveTimerInfo>(`/timers/${entryId}/pause`).then((r) => r.data),
  resume: (entryId: string) =>
    apiClient.post<ActiveTimerInfo>(`/timers/${entryId}/resume`).then((r) => r.data),
  stop: (entryId: string) =>
    apiClient.post<TimeEntry>(`/timers/${entryId}/stop`).then((r) => r.data),
  adjust: (entryId: string, deltaSeconds: number) =>
    apiClient.post<TimeEntry>(`/timers/${entryId}/adjust`, { delta_seconds: deltaSeconds }).then((r) => r.data),
  getActive: () =>
    apiClient.get<ActiveTimerInfo[]>('/timers/active').then((r) => r.data),
  getAnalytics: (range: 'today' | 'week' | 'month') =>
    apiClient.get<AnalyticsSummary>(`/analytics/summary?range=${range}`).then((r) => r.data),
}
