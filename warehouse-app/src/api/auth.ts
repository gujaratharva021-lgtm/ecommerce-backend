import apiClient from './client'
import type { WarehouseStaff } from '../types/warehouse'

export const sendOtp = (phone: string) =>
  apiClient.post('/warehouse/send-otp', { phone }).then((r) => r.data)

export const verifyOtp = (phone: string, otp: string) =>
  apiClient
    .post('/warehouse/verify-otp', { phone, otp })
    .then((r) => r.data as { token: string; warehouse_staff: WarehouseStaff })
