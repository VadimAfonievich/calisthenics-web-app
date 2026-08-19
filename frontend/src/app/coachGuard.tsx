import type { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useSessionStore } from '../store/session'

export function CoachGuard({children}:{children:ReactNode}) {
  const {status,user}=useSessionStore()
  if(status==='loading') return <div className="notice skeleton">Проверяем права…</div>
  if(!user?.available_modes.includes('coach')) return <Navigate to="/profile" replace />
  return children
}
