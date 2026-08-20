import apiClient from './client'
import type { RevenueSummary, Expense, ExpenseListResponse, ExpenseFormInput, Payroll, PayrollListResponse, PayrollFormInput, ProfitLoss, PaymentReconciliation, GSTSummary, Vendor, VendorRequest, VendorBill, VendorBillRequest, Account, AccountRequest, LedgerEntry, ManualJournalEntryRequest, TrialBalance, BankTransaction, BankTransactionRequest, BankTransactionMatchRequest } from '../types/finance'

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


// ---- Vendors ----

export async function getVendors(isActive?: boolean): Promise<{ vendors: Vendor[] }> {
  const { data } = await apiClient.get('/admin/finance/vendors', {
    params: isActive === undefined ? {} : { is_active: isActive },
  })
  return data
}

export async function createVendor(payload: VendorRequest): Promise<Vendor> {
  const { data } = await apiClient.post<Vendor>('/admin/finance/vendors', payload)
  return data
}

export async function updateVendor(id: number, payload: VendorRequest): Promise<Vendor> {
  const { data } = await apiClient.put<Vendor>(`/admin/finance/vendors/${id}`, payload)
  return data
}

export async function deleteVendor(id: number): Promise<void> {
  await apiClient.delete(`/admin/finance/vendors/${id}`)
}

// ---- Vendor Bills ----

export async function getVendorBills(params: {
  vendor_id?: number
  status?: string
  page?: number
}): Promise<{ bills: VendorBill[]; total: number; total_pages: number; total_outstanding: number }> {
  const { data } = await apiClient.get('/admin/finance/vendor-bills', { params })
  return data
}

export async function createVendorBill(payload: VendorBillRequest): Promise<VendorBill> {
  const { data } = await apiClient.post<VendorBill>('/admin/finance/vendor-bills', payload)
  return data
}

export async function payVendorBill(id: number, amount: number): Promise<VendorBill> {
  const { data } = await apiClient.post<VendorBill>(`/admin/finance/vendor-bills/${id}/pay`, { amount })
  return data
}

export async function deleteVendorBill(id: number): Promise<void> {
  await apiClient.delete(`/admin/finance/vendor-bills/${id}`)
}

// ---- Chart of Accounts ----

export async function getAccounts(type?: string): Promise<{ accounts: Account[] }> {
  const { data } = await apiClient.get('/admin/finance/accounts', {
    params: type ? { type } : {},
  })
  return data
}

export async function createAccount(payload: AccountRequest): Promise<Account> {
  const { data } = await apiClient.post<Account>('/admin/finance/accounts', payload)
  return data
}

export async function updateAccount(id: number, payload: AccountRequest): Promise<Account> {
  const { data } = await apiClient.put<Account>(`/admin/finance/accounts/${id}`, payload)
  return data
}

// ---- Ledger ----

export async function getLedgerEntries(params: {
  account_id?: number
  from?: string
  to?: string
  page?: number
}): Promise<{ entries: LedgerEntry[]; total: number; total_pages: number }> {
  const { data } = await apiClient.get('/admin/finance/ledger', { params })
  return data
}

export async function createManualJournalEntry(
  payload: ManualJournalEntryRequest
): Promise<{ transaction_ref: string; entries: LedgerEntry[] }> {
  const { data } = await apiClient.post('/admin/finance/ledger', payload)
  return data
}

export async function getTrialBalance(asOf?: string): Promise<TrialBalance> {
  const { data } = await apiClient.get<TrialBalance>('/admin/finance/ledger/trial-balance', {
    params: asOf ? { as_of: asOf } : {},
  })
  return data
}

// ---- Bank Transactions ----

export async function getBankTransactions(params: {
  status?: string
  from?: string
  to?: string
  page?: number
}): Promise<{ transactions: BankTransaction[]; total: number; total_pages: number; unmatched_count: number }> {
  const { data } = await apiClient.get('/admin/finance/bank-transactions', { params })
  return data
}

export async function createBankTransaction(payload: BankTransactionRequest): Promise<BankTransaction> {
  const { data } = await apiClient.post<BankTransaction>('/admin/finance/bank-transactions', payload)
  return data
}

export async function matchBankTransaction(
  id: number,
  payload: BankTransactionMatchRequest
): Promise<BankTransaction> {
  const { data } = await apiClient.post<BankTransaction>(`/admin/finance/bank-transactions/${id}/match`, payload)
  return data
}

export async function ignoreBankTransaction(id: number): Promise<BankTransaction> {
  const { data } = await apiClient.post<BankTransaction>(`/admin/finance/bank-transactions/${id}/ignore`)
  return data
}

export async function deleteBankTransaction(id: number): Promise<void> {
  await apiClient.delete(`/admin/finance/bank-transactions/${id}`)
}
