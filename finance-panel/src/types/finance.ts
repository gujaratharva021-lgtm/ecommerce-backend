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
