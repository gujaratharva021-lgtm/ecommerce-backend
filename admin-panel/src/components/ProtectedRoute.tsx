import type { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

export default function ProtectedRoute({ children }: { children: ReactNode }) {
  const { isAuthenticated, isLoading, user } = useAuth()

  if (isLoading) {
    return <div style={{ padding: 40, textAlign: 'center' }}>Loading...</div>
  }

  // Check role explicitly here too, not just isAuthenticated — the data
  // itself is already protected server-side by AdminOnly() middleware, but
  // this keeps the panel UI from rendering for a non-admin session.
  if (!isAuthenticated || user?.role !== 'admin') {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}
