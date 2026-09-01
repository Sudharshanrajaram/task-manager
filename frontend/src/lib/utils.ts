import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'
import type { TaskPriority, TaskStatus, TaskType } from '../types'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatDuration(totalSeconds: number): string {
  const h = Math.floor(totalSeconds / 3600)
  const m = Math.floor((totalSeconds % 3600) / 60)
  const s = totalSeconds % 60
  if (h > 0) return `${h}h ${m.toString().padStart(2, '0')}m`
  if (m > 0) return `${m}m ${s.toString().padStart(2, '0')}s`
  return `${s}s`
}

export function formatDurationHMS(totalSeconds: number): string {
  const h = Math.floor(totalSeconds / 3600)
  const m = Math.floor((totalSeconds % 3600) / 60)
  const s = totalSeconds % 60
  return [
    h.toString().padStart(2, '0'),
    m.toString().padStart(2, '0'),
    s.toString().padStart(2, '0'),
  ].join(':')
}

export function formatRelativeTime(isoString: string): string {
  const date = new Date(isoString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  if (diffMins < 1) return 'just now'
  if (diffMins < 60) return `${diffMins}m ago`
  const diffHours = Math.floor(diffMins / 60)
  if (diffHours < 24) return `${diffHours}h ago`
  const diffDays = Math.floor(diffHours / 24)
  return `${diffDays}d ago`
}

// Status display config (4 columns per spec 1.1)
export const STATUS_CONFIG: Record<TaskStatus, { label: string; className: string }> = {
  backlog: { label: 'Backlog', className: 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400' },
  in_progress: { label: 'In Progress', className: 'bg-indigo-50 text-indigo-700 dark:bg-indigo-950 dark:text-indigo-300' },
  review: { label: 'Review', className: 'bg-blue-50 text-blue-700 dark:bg-blue-950 dark:text-blue-300' },
  done: { label: 'Done', className: 'bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-400' },
}

export const BLOCKED_CONFIG = {
  label: 'Blocked',
  badge: 'bg-amber-50 text-amber-700 dark:bg-amber-950/60 dark:text-amber-300 border border-amber-200 dark:border-amber-800/80',
}

// Priority display config
export const PRIORITY_CONFIG: Record<TaskPriority, { label: string; className: string; dot: string }> = {
  p0: { label: 'P0', className: 'text-red-600 dark:text-red-400', dot: 'bg-red-500' },
  p1: { label: 'P1', className: 'text-amber-600 dark:text-amber-400', dot: 'bg-amber-500' },
  p2: { label: 'P2', className: 'text-indigo-600 dark:text-indigo-400', dot: 'bg-indigo-500' },
  p3: { label: 'P3', className: 'text-slate-500 dark:text-slate-400', dot: 'bg-slate-400' },
}

// Type display config
export const TYPE_CONFIG: Record<TaskType, { label: string; className: string }> = {
  task: { label: 'Task', className: 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300' },
  bug: { label: 'Bug', className: 'bg-red-50 text-red-700 dark:bg-red-950 dark:text-red-300' },
  improvement: { label: 'Improvement', className: 'bg-teal-50 text-teal-700 dark:bg-teal-950 dark:text-teal-300' },
  spike: { label: 'Spike', className: 'bg-purple-50 text-purple-700 dark:bg-purple-950 dark:text-purple-300' },
}

export const STATUS_ORDER: TaskStatus[] = ['backlog', 'in_progress', 'review', 'done']
