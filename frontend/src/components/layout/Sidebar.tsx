import { Link, useLocation } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { LayoutDashboard, Search, Layers, Plus } from 'lucide-react'
import { projectsApi } from '../../api/projects'
import { useUIStore } from '../../store/uiStore'
import { cn } from '../../lib/utils'
import CreateProjectDialog from '../projects/CreateProjectDialog'
import { useState } from 'react'

export default function Sidebar() {
  const { data: projects = [] } = useQuery({
    queryKey: ['projects'],
    queryFn: projectsApi.list,
  })
  const { openCommandPalette } = useUIStore()
  const location = useLocation()
  const [showCreateProject, setShowCreateProject] = useState(false)

  return (
    <aside className="w-56 shrink-0 flex flex-col border-r border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 h-full">
      {/* Logo */}
      <div className="h-12 flex items-center px-4 border-b border-slate-200 dark:border-slate-700">
        <Link to="/dashboard" className="flex items-center gap-2">
          <Layers className="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
          <span className="font-semibold text-sm text-slate-900 dark:text-slate-100 tracking-tight">
            TaskFlow
          </span>
        </Link>
      </div>

      {/* Search / Command palette trigger */}
      <div className="px-3 py-2">
        <button
          onClick={openCommandPalette}
          className="w-full flex items-center gap-2 px-3 py-1.5 rounded-md text-sm text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
        >
          <Search className="w-3.5 h-3.5" />
          <span className="flex-1 text-left">Search...</span>
          <kbd className="text-xs bg-white dark:bg-slate-700 px-1 rounded border border-slate-200 dark:border-slate-600">
            ⌘K
          </kbd>
        </button>
      </div>

      {/* Nav */}
      <nav className="px-2 pb-2">
        <SidebarLink to="/dashboard" icon={<LayoutDashboard className="w-4 h-4" />} label="Dashboard" active={location.pathname === '/dashboard'} />
      </nav>

      <div className="px-3 py-1">
        <div className="flex items-center justify-between mb-1">
          <span className="text-xs font-medium text-slate-400 dark:text-slate-500 uppercase tracking-wider">
            Projects
          </span>
          <button
            onClick={() => setShowCreateProject(true)}
            className="p-0.5 rounded hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 transition-colors"
            title="New project"
          >
            <Plus className="w-3.5 h-3.5" />
          </button>
        </div>

        <div className="space-y-0.5">
          {projects.map((p) => (
            <Link
              key={p.id}
              to={`/projects/${p.id}`}
              className={cn(
                'flex items-center gap-2 px-2 py-1.5 rounded-md text-sm transition-colors',
                location.pathname === `/projects/${p.id}`
                  ? 'bg-indigo-50 dark:bg-indigo-950 text-indigo-700 dark:text-indigo-300 font-medium'
                  : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800'
              )}
            >
              <span
                className="w-2 h-2 rounded-full shrink-0"
                style={{ backgroundColor: p.color }}
              />
              <span className="truncate">{p.name}</span>
              <span className="ml-auto text-xs text-slate-400 dark:text-slate-500 font-mono">
                {p.key}
              </span>
            </Link>
          ))}
        </div>
      </div>

      {showCreateProject && (
        <CreateProjectDialog onClose={() => setShowCreateProject(false)} />
      )}
    </aside>
  )
}

function SidebarLink({
  to, icon, label, active,
}: {
  to: string; icon: React.ReactNode; label: string; active: boolean
}) {
  return (
    <Link
      to={to}
      className={cn(
        'flex items-center gap-2 px-2 py-1.5 rounded-md text-sm transition-colors',
        active
          ? 'bg-indigo-50 dark:bg-indigo-950 text-indigo-700 dark:text-indigo-300 font-medium'
          : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800'
      )}
    >
      {icon}
      {label}
    </Link>
  )
}
