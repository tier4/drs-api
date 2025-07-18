import { NavLink } from 'react-router-dom'
import { cn } from '@/lib/utils'

export function Navigation() {
  const linkClass = ({ isActive }: { isActive: boolean }) => cn(
    "px-4 py-2 rounded-md text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground",
    isActive ? "bg-accent text-accent-foreground" : "text-muted-foreground"
  )

  return (
    <nav className="flex space-x-2 mb-6">
      <NavLink to="/" className={linkClass}>
        Modules
      </NavLink>
      <NavLink to="/time-sync" className={linkClass}>
        Time Sync
      </NavLink>
      <NavLink to="/topic-rates" className={linkClass}>
        Topic Rates
      </NavLink>
    </nav>
  )
}