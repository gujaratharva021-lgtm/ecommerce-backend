import apiClient from './client'
import type { DailySalesSummary, RangeSalesSummary } from '../types/reports'

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

export async function getRangeSalesReport(from: string, to: string): Promise<RangeSalesSummary> {
  const { data } = await apiClient.get<{ summary: RangeSalesSummary }>('/admin/reports/range-sales', {
    params: { from, to },
  })
  return data.summary
}

export async function exportRangeSalesReport(from: string, to: string): Promise<void> {
  const response = await apiClient.get('/admin/reports/range-sales/export', {
    params: { from, to },
    responseType: 'blob',
  })
  const url = window.URL.createObjectURL(new Blob([response.data]))
  const link = document.createElement('a')
  link.href = url
  link.setAttribute('download', `sales-report-${from}_to_${to}.xlsx`)
  document.body.appendChild(link)
  link.click()
  link.remove()
  window.URL.revokeObjectURL(url)
}

export async function getSalesRegister(from: string, to: string): Promise<any[]> {
  const { data } = await apiClient.get('/admin/reports/sales-register', { params: { from, to } })
  return data.sales_register ?? []
}

export async function getPurchaseRegister(from: string, to: string): Promise<any[]> {
  const { data } = await apiClient.get('/admin/reports/purchase-register', { params: { from, to } })
  return data.purchase_register ?? []
}

export async function getRiderPayable(from: string, to: string): Promise<{ rider_payable: import('../types/reports').RiderPayableRow[]; per_delivery_rate: number }> {
  const { data } = await apiClient.get('/admin/reports/rider-payable', { params: { from, to } })
  return data
}

export async function getGatewaySettlement(from: string, to: string): Promise<{ gateway_settlement: import('../types/reports').GatewaySettlementRow[]; note: string }> {
  const { data } = await apiClient.get('/admin/reports/gateway-settlement', { params: { from, to } })
  return data
}

export async function getCashFlow(from: string, to: string): Promise<{ by_category: import('../types/reports').CashFlowRow[]; net_cash_flow: number }> {
  const { data } = await apiClient.get('/admin/reports/cash-flow', { params: { from, to } })
  return data
}

export async function getBalanceSheet(asOf?: string): Promise<import('../types/reports').BalanceSheet> {
  const { data } = await apiClient.get('/admin/reports/balance-sheet', { params: asOf ? { as_of: asOf } : {} })
  return data
}
