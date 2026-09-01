import { apiClient } from './client'
import type { Project, CreateProjectInput } from '../types'

export const projectsApi = {
  list: () => apiClient.get<Project[]>('/projects').then((r) => r.data),
  getById: (id: string) => apiClient.get<Project>(`/projects/${id}`).then((r) => r.data),
  create: (input: CreateProjectInput) =>
    apiClient.post<Project>('/projects', input).then((r) => r.data),
  update: (id: string, input: Partial<CreateProjectInput>) =>
    apiClient.patch<Project>(`/projects/${id}`, input).then((r) => r.data),
  delete: (id: string) => apiClient.delete(`/projects/${id}`),
}
