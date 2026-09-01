import { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { LayoutDashboard, Search, Layers, Plus, Calendar, FileText, LogIn, LogOut, User as UserIcon } from 'lucide-react'
import { projectsApi } from '../../api/projects'
import { useUIStore } from '../../store/uiStore'
import { useAuthStore } from '../../store/authStore'
import { cn } from '../../lib/utils'
import CreateProjectDialog from '../projects/CreateProjectDialog'
import AuthDialog from '../auth/AuthDialog'
import NotesEditor from '../notes/NotesEditor'

export default function Sidebar() {
  const { data: projects = [] } = useQuery({
    queryKey: ['projects'],
    queryFn: projectsApi.list,
  })
  const { openCommandPalette } = useUIStore()
  const { user, isAuthenticated, logout } = useAuthStore()
  const location = useLocation()
  const [showCreateProject, setShowCreateProject] = useState(false)
  const [showAuthDialog, setShowAuthDialog] = useState(false)
  const [showScratchpad, setShowScratchpad] = useState(false)

  return (
    <aside className="w-56 shrink-0 flex flex-col border-r border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 h-full">
      {/* Logo */}
      <div className="h-12 flex items-center px-4 border-b border-slate-200 dark:border-slate-800">
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
          className="w-full flex items-center gap-2 px-3 py-1.5 rounded-md text-xs text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
        >
          <Search className="w-3.5 h-3.5" />
          <span className="flex-1 text-left">Search...</span>
          <kbd className="text-[10px] bg-white dark:bg-slate-700 px-1 rounded border border-slate-200 dark:border-slate-600">
            ⌘K
          </kbd>
        </button>
      </div>

      {/* Main Nav */}
      <nav className="px-2 pb-2 space-y-0.5">
        <SidebarLink
          to="/dashboard"
          icon={<LayoutDashboard className="w-4 h-4" />}
          label="Dashboard"
          active={location.pathname === '/dashboard'}
        />
        <SidebarLink
          to="/logs"
          icon={<Calendar className="w-4 h-4" />}
          label="Daily Logs"
          active={location.pathname === '/logs'}
        />
        <button
          onClick={() => setShowScratchpad(true)}
          className="w-full flex items-center gap-2 px-2 py-1.5 rounded-md text-xs text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors text-left"
        >
          <FileText className="w-4 h-4" />
          <span>Scratchpad</span>
        </button>
      </nav>

      {/* Projects List */}
      <div className="px-3 py-2 flex-1 overflow-y-auto">
        <div className="flex items-center justify-between mb-1.5">
          <span className="text-[11px] font-semibold text-slate-400 dark:text-slate-500 uppercase tracking-wider">
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
                'flex items-center gap-2 px-2 py-1.5 rounded-md text-xs transition-colors',
                location.pathname === `/projects/${p.id}`
                  ? 'bg-indigo-50 dark:bg-indigo-950 text-indigo-700 dark:text-indigo-300 font-medium'
                  : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800'
              )}
            >
              <span
                className="w-2 h-2 rounded-full shrink-0"
                style={{ backgroundColor: p.color }}
              />
              <span className="truncate flex-1">{p.name}</span>
              <span className="text-[10px] text-slate-400 dark:text-slate-500 font-mono">
                {p.key}
              </span>
            </Link>
          ))}
        </div>
      </div>

      {/* Bottom Auth & User Account Section */}
      <div className="p-3 border-t border-slate-200 dark:border-slate-800">
        {isAuthenticated && user ? (
          <div className="flex items-center justify-between gap-2">
            <div className="flex items-center gap-2 min-w-0">
              <div className="w-7 h-7 rounded-full bg-indigo-600 text-white font-bold text-xs flex items-center justify-center shrink-0">
                {user.name ? user.name[0].toUpperCase() : 'U'}
              </div>
              <div className="min-w-0">
                <p className="text-xs font-semibold text-slate-900 dark:text-slate-100 truncate">
                  {user.name}
                </p>
                <p className="text-[10px] text-slate-400 truncate">{user.email}</p>
              </div>
            </div>
            <button
              onClick={() => logout()}
              className="p-1 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 rounded"
              title="Sign Out"
            >
              <LogOut className="w-3.5 h-3.5" />
            </button>
          </div>
        ) : (
          <button
            onClick={() => setShowAuthDialog(true)}
            className="w-full flex items-center justify-center gap-2 py-1.5 px-3 rounded-lg text-xs font-medium border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300 transition-colors shadow-sm"
          >
            <LogIn className="w-3.5 h-3.5" />
            <span>Sign In / Register</span>
          </button>
        )}
      </div>

      {/* Dialogs */}
      {showCreateProject && (
        <CreateProjectDialog onClose={() => setShowCreateProject(false)} />
      )}
      {showAuthDialog && (
        <AuthDialog onClose={() => setShowAuthDialog(false)} />
      )}
      {showScratchpad && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
          <div className="bg-white dark:bg-slate-900 rounded-xl shadow-2xl border border-slate-200 dark:border-slate-800 w-full max-w-2xl h-[550px] p-6 flex flex-col">
            <div className="flex justify-end mb-2">
              <button
                onClick={() => setShowScratchpad(false)}
                className="text-xs text-slate-400 hover:text-slate-600 dark:hover:text-slate-200"
              >
                Close ✕
              </button>
            </div>
            <div className="flex-1 overflow-hidden">
              <NotesEditor title="Global Scratchpad" />
            </div>
          </div>
        </div>
      )}
    </aside>
  )
}

function SidebarLink({
  to,
  icon,
  label,
  active,
}: {
  to: string
  icon: React.ReactNode
  label: string
  active: boolean
}) {
  return (
    <Link
      to={to}
      className={cn(
        'flex items-center gap-2 px-2 py-1.5 rounded-md text-xs transition-colors',
        active
          ? 'bg-indigo-50 dark:bg-indigo-950 text-indigo-700 dark:text-indigo-300 font-semibold'
          : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800'
      )}
    >
      {icon}
      {label}
    </Link>
  )
}
