import { create } from 'zustand'

interface UIState {
  selectedProjectId: string | null
  commandPaletteOpen: boolean
  focusMode: boolean
  setSelectedProject: (id: string | null) => void
  openCommandPalette: () => void
  closeCommandPalette: () => void
  setFocusMode: (v: boolean) => void
}

export const useUIStore = create<UIState>((set) => ({
  selectedProjectId: null,
  commandPaletteOpen: false,
  focusMode: false,
  setSelectedProject: (id) => set({ selectedProjectId: id }),
  openCommandPalette: () => set({ commandPaletteOpen: true }),
  closeCommandPalette: () => set({ commandPaletteOpen: false }),
  setFocusMode: (v) => set({ focusMode: v }),
}))
