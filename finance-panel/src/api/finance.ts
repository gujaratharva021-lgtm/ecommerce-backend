import apiClient from './client'
import type { RevenueResponse } from '../types/finance'

export async function getRevenue(from: string, to: string): Promise<RevenueResponse> {
  const { data } = await apiClient.get<RevenueResponse>('/finance/revenue', {
    params: { from, to },
  })
  return data
}
