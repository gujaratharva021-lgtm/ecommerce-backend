// Matches internal/models/user.go and internal/models/AuthResponse on the backend

export interface User {
  id: number
  name: string
  phone: string
  role: 'customer' | 'admin' | 'delivery_partner' | 'warehouse_staff'
  created_at: string
  updated_at: string
}

export interface AuthResponse {
  token: string
  user: User
}

// POST /api/v1/auth/send-otp
// Phone must be exactly 10 digits, numeric, no country code (binding:"required,len=10,numeric")
export interface SendOtpRequest {
  phone: string
}

export interface SendOtpResponse {
  message: string
  expires_in_minutes: number
  otp?: string // only present when backend GIN_MODE != release (dev only)
}

// POST /api/v1/auth/verify-otp
export interface VerifyOtpRequest {
  phone: string
  otp: string
}
