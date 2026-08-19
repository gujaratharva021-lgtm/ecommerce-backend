import apiClient from './client'
import type { DailySalesSummary } from '../types/reports'

export async function getDailySalesReport(date: string): Promise<DailySalesSummary> {
  const { data } = await apiClient.get<{ summary: DailySalesSummary }>('/admin/reports/daily-sales', {
    params: { date },
  })
  return data.summary
}

export async function exportDailySalesReport(date: string): Promise<void> {
  const response = await apiClient.get('/admin/reports/daily-sales/export', {
    params: { date },
    responseType: 'blob',
  })
  const url = window.URL.createObjectURL(new Blob([response.data]))
  const link = document.createElement('a')
  link.href = url
  link.setAttribute('download', `daily-sales-${date}.xlsx`)
  document.body.appendChild(link)
  link.click()
  link.remove()
  window.URL.revokeObjectURL(url)
}
