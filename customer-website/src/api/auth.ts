import { apiClient } from './client'
import type { AuthResponse, User } from '../types'

export const sendOTP = (phone: string) =>
  apiClient.post('/auth/send-otp', { phone }).then((r) => r.data)

export const verifyOTP = (phone: string, otp: string) =>
  apiClient.post<AuthResponse>('/auth/verify-otp', { phone, otp }).then((r) => r.data)

export const getMe = () => apiClient.get<User>('/auth/me').then((r) => r.data)

export const updateProfile = (name: string) =>
  apiClient.put<User>('/auth/me', { name }).then((r) => r.data)
