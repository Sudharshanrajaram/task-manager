import { apiClient } from './client'

export interface Note {
  id?: string
  task_id?: string
  user_id?: string
  content: string
  created_at?: string
  updated_at?: string
}

export const notesApi = {
  getTaskNote: (taskId: string) =>
    apiClient.get<Note>(`/tasks/${taskId}/notes`).then((r) => r.data),

  saveTaskNote: (taskId: string, content: string) =>
    apiClient.put<Note>(`/tasks/${taskId}/notes`, { content }).then((r) => r.data),

  getScratchpad: () =>
    apiClient.get<Note>('/notes/scratchpad').then((r) => r.data),

  saveScratchpad: (content: string) =>
    apiClient.put<Note>('/notes/scratchpad', { content }).then((r) => r.data),
}

