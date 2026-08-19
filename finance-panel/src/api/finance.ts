import apiClient from './client'
import type { RevenueSummary } from '../types/finance'

export async function getRevenue(from: string, to: string): Promise<RevenueSummary> {
  const { data } = await apiClient.get<RevenueSummary>('/admin/finance/revenue', {
    params: { from, to },
  })
  return data
}
