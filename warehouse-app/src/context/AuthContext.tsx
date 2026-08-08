import { createContext, useContext, useState, useEffect } from 'react'
import type { ReactNode } from 'react'
import type { WarehouseStaff } from '../types/warehouse'
import { sendOtp as sendOtpApi, verifyOtp as verifyOtpApi } from '../api/auth'

interface AuthContextValue {
  staff: WarehouseStaff | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
  sendOtp: (phone: string) => Promise<void>
  verifyOtp: (phone: string, otp: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [staff, setStaff] = useState<WarehouseStaff | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    const storedToken = localStorage.getItem('warehouse_token')
    const storedStaff = localStorage.getItem('warehouse_staff')
    if (storedToken && storedStaff) {
      try {
        setToken(storedToken)
        setStaff(JSON.parse(storedStaff))
      } catch {
        localStorage.removeItem('warehouse_token')
        localStorage.removeItem('warehouse_staff')
      }
    }
    setIsLoading(false)
  }, [])

  async function sendOtp(phone: string) {
    await sendOtpApi(phone)
  }

  async function verifyOtp(phone: string, otp: string) {
    const res = await verifyOtpApi(phone, otp)
    localStorage.setItem('warehouse_token', res.token)
    localStorage.setItem('warehouse_staff', JSON.stringify(res.warehouse_staff))
    setToken(res.token)
    setStaff(res.warehouse_staff)
  }

  function logout() {
    localStorage.removeItem('warehouse_token')
    localStorage.removeItem('warehouse_staff')
    setToken(null)
    setStaff(null)
  }

  return (
    <AuthContext.Provider
      value={{
        staff,
        token,
        isAuthenticated: !!token && !!staff,
        isLoading,
        sendOtp,
        verifyOtp,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
