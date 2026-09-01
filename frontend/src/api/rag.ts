import { apiClient } from './client'
import type { SubtaskSuggestionResult, Task } from '../types'

export const ragApi = {
  suggest: (title: string, projectId?: string, count?: number) =>
    apiClient
      .post<SubtaskSuggestionResult>('/tasks/suggest-subtasks', {
        title,
        project_id: projectId,
        count: count ?? 5,
      })
      .then((r) => r.data),
  suggestForTask: (taskId: string, count?: number) =>
    apiClient
      .post<SubtaskSuggestionResult>(`/tasks/${taskId}/suggest-subtasks`, {
        count: count ?? 5,
      })
      .then((r) => r.data),
  accept: (taskId: string, subtasks: string[]) =>
    apiClient
      .post<Task>(`/tasks/${taskId}/accept-subtasks`, { subtasks })
      .then((r) => r.data),
}
