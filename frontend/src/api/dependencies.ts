import { apiClient } from './client'
import type { Task } from '../types'

export interface TaskDependency {
  id: string
  task_id: string
  depends_on_task_id: string
  created_at: string
  task?: Task
  depends_on_task?: Task
}

export const dependenciesApi = {
  get: (taskId: string) =>
    apiClient
      .get<{ blocked_by: TaskDependency[]; blocks: TaskDependency[] }>(`/tasks/${taskId}/dependencies`)
      .then((r) => r.data),

  add: (taskId: string, dependsOnTaskId: string) =>
    apiClient
      .post<TaskDependency>(`/tasks/${taskId}/dependencies`, {
        depends_on_task_id: dependsOnTaskId,
      })
      .then((r) => r.data),

  remove: (taskId: string, depId: string) =>
    apiClient.delete(`/tasks/${taskId}/dependencies/${depId}`),
}

