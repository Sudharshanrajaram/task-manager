import { Outlet } from 'react-router-dom'
import Sidebar from './Sidebar'
import ActiveTimerBar from '../timers/ActiveTimerBar'

export default function Layout() {
  return (
    <div className="flex h-screen bg-slate-50 dark:bg-slate-900 overflow-hidden">
      <Sidebar />
      <div className="flex flex-col flex-1 min-w-0">
        <ActiveTimerBar />
        <main className="flex-1 overflow-y-auto scrollbar-thin">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
