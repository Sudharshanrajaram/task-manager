import { useEffect, useRef, useState } from 'react'
import { Command } from 'cmdk'
import { useUIStore } from '../../store/uiStore'
import { useQuery } from '@tanstack/react-query'
import { projectsApi } from '../../api/projects'
import { tasksApi } from '../../api/tasks'
import { useNavigate } from 'react-router-dom'
import { LayoutDashboard, Layers, FileText, Maximize2, Search } from 'lucide-react'

export default function CommandPalette() {
  const { commandPaletteOpen, closeCommandPalette, selectedProjectId, setFocusMode } = useUIStore()
  const navigate = useNavigate()
  const [query, setQuery] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  const { data: projects = [] } = useQuery({
    queryKey: ['projects'],
    queryFn: projectsApi.list,
    enabled: commandPaletteOpen,
  })

  const { data: tasks = [] } = useQuery({
    queryKey: ['tasks', selectedProjectId],
    queryFn: () => selectedProjectId ? tasksApi.listByProject(selectedProjectId) : Promise.resolve([]),
    enabled: commandPaletteOpen && !!selectedProjectId,
  })

  // Open on Cmd+K
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        commandPaletteOpen ? closeCommandPalette() : useUIStore.getState().openCommandPalette()
      }
      if (e.key === 'Escape') closeCommandPalette()
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [commandPaletteOpen, closeCommandPalette])

  useEffect(() => {
    if (commandPaletteOpen) {
      setQuery('')
      setTimeout(() => inputRef.current?.focus(), 10)
    }
  }, [commandPaletteOpen])

  if (!commandPaletteOpen) return null

  const run = (fn: () => void) => {
    fn()
    closeCommandPalette()
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center pt-[15vh] bg-black/40 backdrop-blur-sm"
      onClick={closeCommandPalette}
    >
      <div
        className="w-full max-w-lg bg-white dark:bg-slate-900 rounded-xl shadow-2xl border border-slate-200 dark:border-slate-700 overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        <Command className="flex flex-col" shouldFilter>
          <div className="flex items-center gap-2 border-b border-slate-200 dark:border-slate-700 px-4 py-3">
            <Search className="w-4 h-4 text-slate-400 shrink-0" />
            <Command.Input
              ref={inputRef}
              value={query}
              onValueChange={setQuery}
              placeholder="Search tasks, projects, or run a command..."
              className="flex-1 bg-transparent outline-none text-sm text-slate-900 dark:text-slate-100 placeholder-slate-400"
            />
            <kbd className="text-xs bg-slate-100 dark:bg-slate-800 text-slate-500 px-1.5 py-0.5 rounded border border-slate-200 dark:border-slate-700">
              ESC
            </kbd>
          </div>

          <Command.List className="max-h-80 overflow-y-auto py-2 scrollbar-thin">
            <Command.Empty className="text-center text-sm text-slate-400 py-8">
              No results found.
            </Command.Empty>

            <Command.Group heading="Navigation" className="px-2">
              <CmdItem
                icon={<LayoutDashboard className="w-4 h-4" />}
                label="Go to Dashboard"
                onSelect={() => run(() => navigate('/dashboard'))}
              />
              <CmdItem
                icon={<Maximize2 className="w-4 h-4" />}
                label="Enter Focus Mode"
                onSelect={() => run(() => setFocusMode(true))}
              />
            </Command.Group>

            {projects.length > 0 && (
              <Command.Group heading="Projects" className="px-2">
                {projects.map((p) => (
                  <CmdItem
                    key={p.id}
                    icon={
                      <span
                        className="w-3 h-3 rounded-full shrink-0"
                        style={{ backgroundColor: p.color }}
                      />
                    }
                    label={`${p.name} (${p.key})`}
                    onSelect={() => run(() => navigate(`/projects/${p.id}`))}
                  />
                ))}
              </Command.Group>
            )}

            {tasks.length > 0 && (
              <Command.Group heading="Tasks" className="px-2">
                {tasks.map((t) => (
                  <CmdItem
                    key={t.id}
                    icon={<FileText className="w-4 h-4 text-slate-400" />}
                    label={t.title}
                    subtitle={t.ticket_key}
                    onSelect={() => run(() => navigate(`/tasks/${t.id}`))}
                  />
                ))}
              </Command.Group>
            )}
          </Command.List>
        </Command>
      </div>
    </div>
  )
}

function CmdItem({
  icon, label, subtitle, onSelect,
}: {
  icon: React.ReactNode; label: string; subtitle?: string; onSelect: () => void
}) {
  return (
    <Command.Item
      onSelect={onSelect}
      className="flex items-center gap-3 px-3 py-2 rounded-lg cursor-pointer text-sm text-slate-700 dark:text-slate-300 data-[selected=true]:bg-indigo-50 dark:data-[selected=true]:bg-indigo-950 data-[selected=true]:text-indigo-700 dark:data-[selected=true]:text-indigo-300 transition-colors"
    >
      <span className="text-slate-400">{icon}</span>
      <span className="flex-1 truncate">{label}</span>
      {subtitle && (
        <span className="text-xs font-mono text-slate-400">{subtitle}</span>
      )}
    </Command.Item>
  )
}
