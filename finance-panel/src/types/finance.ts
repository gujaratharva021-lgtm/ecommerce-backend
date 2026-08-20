export interface WarehouseRevenue {
  warehouse_id: number | null
  warehouse_name: string
  revenue: number
  order_count: number
}

export interface ProductRevenue {
  product_id: number
  product_name: string
  revenue: number
  quantity: number
}

export interface RevenueSummary {
  from: string
  to: string
  gross_sales: number
  net_sales: number
  discounts: number
  delivery_charge: number
  platform_fee: number
  order_count: number
  by_warehouse: WarehouseRevenue[]
  by_product: ProductRevenue[]
}

export interface Expense {
  id: number
  amount: number
  category: string
  expense_date: string
  warehouse_id?: number | null
  warehouse?: { id: number; name: string }
  note?: string
  receipt_url?: string
  added_by_id: number
  created_at: string
  updated_at: string
}

export const EXPENSE_CATEGORIES = [
  'rent',
  'utilities',
  'salaries',
  'packaging',
  'marketing',
  'maintenance',
  'misc',
] as const

export interface ExpenseListResponse {
  expenses: Expense[]
  page: number
  limit: number
  total: number
  total_pages: number
  total_amount: number
}

export interface ExpenseFormInput {
  amount: number
  category: string
  expense_date: string
  warehouse_id?: number | null
  note?: string
  receipt_url?: string
}

export interface Payroll {
  id: number
  staff_id: number
  staff?: { id: number; name: string; phone: string; warehouse_id: number }
  amount: number
  month: number
  year: number
  status: string
  payment_method?: string
  note?: string
  paid_by_id?: number | null
  paid_at?: string | null
  created_at: string
  updated_at: string
}

export const PAYROLL_PAYMENT_METHODS = ['cash', 'bank', 'upi'] as const

export interface PayrollListResponse {
  payroll: Payroll[]
  page: number
  limit: number
  total: number
  total_pages: number
  total_pending: number
  total_paid: number
}

export interface PayrollFormInput {
  staff_id: number
  amount: number
  month: number
  year: number
  status?: string
  payment_method?: string
  note?: string
}

export interface ProfitLoss {
  from: string
  to: string
  revenue: number
  cogs: number
  cost_price_coverage: number
  gross_profit: number
  operating_expenses: number
  ebitda: number
  net_profit: number
}

export interface PaymentReconciliation {
  total_collected: number
  total_pending: number
  total_refunded: number
  count_paid: number
  count_pending: number
  count_failed: number
  count_refunded: number
  online_collected: number
  cod_collected: number
}

export interface GSTByHSN {
  hsn_code: string
  taxable_amount: number
  gst_amount: number
  quantity: number
}

export interface GSTByRate {
  gst_percent: number
  taxable_amount: number
  gst_amount: number
}

export interface GSTSummary {
  from: string
  to: string
  taxable_amount: number
  cgst_amount: number
  sgst_amount: number
  igst_amount: number
  total_gst: number
  invoice_count: number
  by_hsn: GSTByHSN[]
  by_rate: GSTByRate[]
}

export interface Vendor {
  id: number
  name: string
  contact_name?: string
  phone?: string
  email?: string
  gstin?: string
  address?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface VendorRequest {
  name: string
  contact_name?: string
  phone?: string
  email?: string
  gstin?: string
  address?: string
  is_active?: boolean
}

export interface VendorBill {
  id: number
  vendor_id: number
  vendor?: Vendor
  bill_number?: string
  amount: number
  amount_paid: number
  bill_date: string
  due_date?: string
  note?: string
  created_by_id: number
  created_at: string
  updated_at: string
  status: 'unpaid' | 'partially_paid' | 'paid'
}

export interface VendorBillRequest {
  vendor_id: number
  bill_number?: string
  amount: number
  bill_date: string
  due_date?: string
  note?: string
}

export interface Account {
  id: number
  code: string
  name: string
  type: 'asset' | 'liability' | 'equity' | 'revenue' | 'expense'
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface AccountRequest {
  code: string
  name: string
  type: string
  is_active?: boolean
}

export interface LedgerEntry {
  id: number
  transaction_ref: string
  account_id: number
  account?: Account
  type: 'debit' | 'credit'
  amount: number
  description?: string
  reference_type?: string
  reference_id?: number
  entry_date: string
  created_by_id: number
  created_at: string
}

export interface LedgerEntryLine {
  account_id: number
  type: 'debit' | 'credit'
  amount: number
  description?: string
}

export interface ManualJournalEntryRequest {
  entry_date: string
  lines: LedgerEntryLine[]
}

export interface TrialBalanceRow {
  account_id: number
  account_code: string
  account_name: string
  account_type: string
  total_debit: number
  total_credit: number
}

export interface TrialBalance {
  as_of: string
  accounts: TrialBalanceRow[]
  total_debit: number
  total_credit: number
  is_balanced: boolean
}

export interface BankTransaction {
  id: number
  transaction_date: string
  description?: string
  amount: number
  reference_number?: string
  status: 'unmatched' | 'matched' | 'ignored'
  matched_type?: string
  matched_id?: number
  matched_at?: string
  matched_by_id?: number
  note?: string
  created_by_id: number
  created_at: string
  updated_at: string
}

export interface BankTransactionRequest {
  transaction_date: string
  description?: string
  amount: number
  reference_number?: string
}

export interface BankTransactionMatchRequest {
  matched_type: string
  matched_id?: number
  note?: string
}
