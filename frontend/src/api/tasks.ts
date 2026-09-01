import { apiClient } from './client'
import type { Task, CreateTaskInput, UpdateTaskInput } from '../types'

export const tasksApi = {
  listByProject: (projectId: string) =>
    apiClient.get<Task[]>(`/projects/${projectId}/tasks`).then((r) => r.data),
  getById: (idOrKey: string) =>
    apiClient.get<Task>(`/tasks/${idOrKey}`).then((r) => r.data),
  create: (projectId: string, input: CreateTaskInput) =>
    apiClient.post<Task>(`/projects/${projectId}/tasks`, input).then((r) => r.data),
  update: (id: string, input: UpdateTaskInput) =>
    apiClient.patch<Task>(`/tasks/${id}`, input).then((r) => r.data),
  delete: (id: string) => apiClient.delete(`/tasks/${id}`),
  block: (id: string, isBlocked: boolean, reason?: string) =>
    apiClient.patch<Task>(`/tasks/${id}/block`, { is_blocked: isBlocked, blocked_reason: reason }).then((r) => r.data),
  archive: (id: string, isArchived: boolean = true) =>
    apiClient.post<Task>(`/tasks/${id}/archive`, { is_archived: isArchived }).then((r) => r.data),
}
