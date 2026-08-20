import apiClient from './client'
import type { RevenueSummary, Expense, ExpenseListResponse, ExpenseFormInput, Payroll, PayrollListResponse, PayrollFormInput, ProfitLoss, PaymentReconciliation, GSTSummary } from '../types/finance'

export async function getRevenue(from: string, to: string): Promise<RevenueSummary> {
  const { data } = await apiClient.get<RevenueSummary>('/admin/finance/revenue', {
    params: { from, to },
  })
  return data
}

export async function listExpenses(params: {
  category?: string
  warehouse_id?: number
  from?: string
  to?: string
  page?: number
  limit?: number
}): Promise<ExpenseListResponse> {
  const { data } = await apiClient.get<ExpenseListResponse>('/admin/finance/expenses', { params })
  return data
}

export async function createExpense(body: ExpenseFormInput): Promise<Expense> {
  const { data } = await apiClient.post<Expense>('/admin/finance/expenses', body)
  return data
}

export async function updateExpense(id: number, body: ExpenseFormInput): Promise<Expense> {
  const { data } = await apiClient.put<Expense>(`/admin/finance/expenses/${id}`, body)
  return data
}

export async function deleteExpense(id: number): Promise<void> {
  await apiClient.delete(`/admin/finance/expenses/${id}`)
}

export async function listPayroll(params: {
  staff_id?: number
  status?: string
  month?: number
  year?: number
  page?: number
  limit?: number
}): Promise<PayrollListResponse> {
  const { data } = await apiClient.get<PayrollListResponse>('/admin/finance/payroll', { params })
  return data
}

export async function createPayroll(body: PayrollFormInput): Promise<Payroll> {
  const { data } = await apiClient.post<Payroll>('/admin/finance/payroll', body)
  return data
}

export async function updatePayroll(id: number, body: PayrollFormInput): Promise<Payroll> {
  const { data } = await apiClient.put<Payroll>(`/admin/finance/payroll/${id}`, body)
  return data
}

export async function deletePayroll(id: number): Promise<void> {
  await apiClient.delete(`/admin/finance/payroll/${id}`)
}



export async function getProfitLoss(from: string, to: string): Promise<ProfitLoss> {
  const { data } = await apiClient.get<ProfitLoss>('/admin/finance/profit-loss', {
    params: { from, to },
  })
  return data
}


export async function getPaymentReconciliation(dateFrom: string, dateTo: string): Promise<PaymentReconciliation> {
  const { data } = await apiClient.get<PaymentReconciliation>('/admin/payments/reconciliation', {
    params: { date_from: dateFrom, date_to: dateTo },
  })
  return data
}


export async function getGSTSummary(from: string, to: string): Promise<GSTSummary> {
  const { data } = await apiClient.get<GSTSummary>('/admin/finance/gst', {
    params: { from, to },
  })
  return data
}
