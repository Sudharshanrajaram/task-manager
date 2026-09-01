import { create } from 'zustand'
import type { ActiveTimerInfo } from '../types'

interface TimerState {
  activeTimers: ActiveTimerInfo[]
  localElapsed: Record<string, number> // entryId -> elapsed seconds
  setActiveTimers: (timers: ActiveTimerInfo[]) => void
  tickAll: () => void
  updateTimer: (updated: ActiveTimerInfo) => void
  removeTimer: (entryId: string) => void
}

export const useTimerStore = create<TimerState>((set, get) => ({
  activeTimers: [],
  localElapsed: {},

  setActiveTimers: (timers) => {
    const elapsed: Record<string, number> = {}
    for (const t of timers) {
      elapsed[t.entry_id] = t.elapsed_seconds
    }
    set({ activeTimers: timers, localElapsed: elapsed })
  },

  tickAll: () => {
    const { activeTimers, localElapsed } = get()
    const next = { ...localElapsed }
    for (const t of activeTimers) {
      if (!t.is_paused) {
        next[t.entry_id] = (next[t.entry_id] ?? t.elapsed_seconds) + 1
      }
    }
    set({ localElapsed: next })
  },

  updateTimer: (updated) => {
    set((state) => {
      const exists = state.activeTimers.find((t) => t.entry_id === updated.entry_id)
      const timers = exists
        ? state.activeTimers.map((t) => (t.entry_id === updated.entry_id ? updated : t))
        : [...state.activeTimers, updated]
      return {
        activeTimers: timers,
        localElapsed: { ...state.localElapsed, [updated.entry_id]: updated.elapsed_seconds },
      }
    })
  },

  removeTimer: (entryId) => {
    set((state) => {
      const { [entryId]: _, ...rest } = state.localElapsed
      return {
        activeTimers: state.activeTimers.filter((t) => t.entry_id !== entryId),
        localElapsed: rest,
      }
    })
  },
}))
