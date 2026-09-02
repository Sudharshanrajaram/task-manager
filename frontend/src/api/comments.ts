import { apiClient } from './client'
import type { Comment } from '../types'

export const commentsApi = {
  listByTask: (taskId: string) =>
    apiClient.get<Comment[]>(`/tasks/${taskId}/comments`).then((r) => r.data),
  create: (taskId: string, content: string) =>
    apiClient.post<Comment>(`/tasks/${taskId}/comments`, { content }).then((r) => r.data),
  delete: (taskId: string, commentId: string) =>
    apiClient.delete(`/tasks/${taskId}/comments/${commentId}`),
}
