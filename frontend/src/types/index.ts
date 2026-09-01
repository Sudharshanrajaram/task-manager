// Project
export interface Project {
  id: string
  name: string
  key: string
  color: string
  task_counter: number
  created_at: string
  updated_at: string
}

export interface CreateProjectInput {
  name: string
  key: string
  color: string
}

// Task
export type TaskType = 'task' | 'bug' | 'improvement' | 'spike'
export type TaskStatus = 'backlog' | 'in_progress' | 'review' | 'done'
export type TaskPriority = 'p0' | 'p1' | 'p2' | 'p3'

export interface Task {
  id: string
  project_id: string
  ticket_number: number
  ticket_key: string
  type: TaskType
  title: string
  description: string
  status: TaskStatus
  priority: TaskPriority
  labels: string[]
  is_blocked?: boolean
  blocked_reason?: string
  is_archived?: boolean
  archived_at?: string
  ai_summary?: string
  steps_to_reproduce?: string
  severity?: string
  environment?: string
  subtasks: Subtask[] | null
  total_time_spent_seconds: number
  project?: Project
  created_at: string
  updated_at: string
}

export interface CreateTaskInput {
  title: string
  type?: TaskType
  priority?: TaskPriority
  description?: string
  labels?: string[]
}

export interface UpdateTaskInput {
  title?: string
  type?: TaskType
  status?: TaskStatus
  priority?: TaskPriority
  description?: string
  labels?: string[]
  steps_to_reproduce?: string
  severity?: string
  environment?: string
}

// Subtask
export interface Subtask {
  id: string
  task_id: string
  title: string
  is_done: boolean
  order_index: number
  total_time_spent_seconds: number
  created_at: string
  updated_at: string
}

export interface CreateSubtaskInput {
  title: string
}

export interface UpdateSubtaskInput {
  title?: string
  is_done?: boolean
}

// Time Entry
export interface TimeEntry {
  id: string
  task_id: string
  subtask_id?: string
  started_at: string
  ended_at?: string
  duration_seconds: number
  is_running: boolean
  created_at: string
  updated_at: string
}

// Active Timer
export interface ActiveTimerInfo {
  entry_id: string
  task_id: string
  subtask_id?: string
  task_title: string
  ticket_key: string
  project_key: string
  project_name: string
  project_color: string
  subtask_title?: string
  started_at: string
  elapsed_seconds: number
  is_paused: boolean
}

// Analytics
export interface AnalyticsByProject {
  project_id: string
  project_name: string
  project_key: string
  color: string
  total_time_seconds: number
}

export interface AnalyticsByType {
  type: string
  total_time_seconds: number
}

export interface AnalyticsSummary {
  total_time_spent_seconds: number
  by_project: AnalyticsByProject[]
  by_type: AnalyticsByType[]
  range: string
}

// RAG
export interface GroundedExample {
  source_title: string
  final_subtasks: string[]
  similarity_score: number
}

export interface SubtaskSuggestionResult {
  suggested_subtasks: string[]
  grounded_context: GroundedExample[]
  title: string
  count: number
}

// API Error
export interface APIError {
  error: boolean
  message: string
}
