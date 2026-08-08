import { createContext, useContext, useState, useEffect, type ReactNode } from 'react'
import { getMe } from '../api/auth'
import type { User } from '../types'

interface AuthContextType {
  user: User | null
  isLoading: boolean
  login: (token: string, user: User) => void
  logout: () => void
  setUser: (user: User) => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    const token = localStorage.getItem('customer_token')
    if (!token) {
      setIsLoading(false)
      return
    }
    getMe()
      .then((u) => setUser(u))
      .catch(() => {
        localStorage.removeItem('customer_token')
        localStorage.removeItem('customer_user')
      })
      .finally(() => setIsLoading(false))
  }, [])

  function login(token: string, user: User) {
    localStorage.setItem('customer_token', token)
    localStorage.setItem('customer_user', JSON.stringify(user))
    setUser(user)
  }

  function logout() {
    localStorage.removeItem('customer_token')
    localStorage.removeItem('customer_user')
    setUser(null)
  }

  function updateUser(updated: User) {
    localStorage.setItem('customer_user', JSON.stringify(updated))
    setUser(updated)
  }

  return (
    <AuthContext.Provider value={{ user, isLoading, login, logout, setUser: updateUser }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
