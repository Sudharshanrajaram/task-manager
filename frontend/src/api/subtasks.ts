import { apiClient } from './client'
import type { Subtask, CreateSubtaskInput, UpdateSubtaskInput } from '../types'

export const subtasksApi = {
  create: (taskId: string, input: CreateSubtaskInput) =>
    apiClient.post<Subtask>(`/tasks/${taskId}/subtasks`, input).then((r) => r.data),
  update: (id: string, input: UpdateSubtaskInput) =>
    apiClient.patch<Subtask>(`/subtasks/${id}`, input).then((r) => r.data),
  delete: (id: string) => apiClient.delete(`/subtasks/${id}`),
  reorder: (taskId: string, orderedIds: string[]) =>
    apiClient.put(`/tasks/${taskId}/subtasks/reorder`, { ordered_ids: orderedIds }),
}
