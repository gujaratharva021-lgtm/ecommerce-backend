import apiClient from './client'
import type {
  SendOtpRequest,
  SendOtpResponse,
  VerifyOtpRequest,
  AuthResponse,
} from '../types/auth'

// POST /api/v1/auth/send-otp
// phone must be exactly 10 digits, numeric, no +91 prefix, no spaces
export async function sendOtp(phone: string): Promise<SendOtpResponse> {
  const payload: SendOtpRequest = { phone }
  const { data } = await apiClient.post<SendOtpResponse>('/auth/send-otp', payload)
  return data
}

// POST /api/v1/auth/verify-otp
// Returns a JWT + user object. The JWT's "role" claim must be "admin" for
// admin panel access — this is checked client-side after login, and
// enforced server-side by AdminOnly() middleware on every /admin/* call.
export async function verifyOtp(phone: string, otp: string): Promise<AuthResponse> {
  const payload: VerifyOtpRequest = { phone, otp }
  const { data } = await apiClient.post<AuthResponse>('/auth/verify-otp', payload)
  return data
}
