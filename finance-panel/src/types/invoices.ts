export interface InvoiceItem {
  id: number
  invoice_id: number
  product_id: number
  product_name: string
  sku?: string
  hsn_code?: string
  quantity: number
  price: number
  gst_percent: number
  gst_amount: number
}

export interface Invoice {
  id: number
  invoice_number: string
  order_id: number
  customer_name: string
  customer_phone: string
  address_line1: string
  address_line2?: string
  address_city: string
  address_state: string
  address_pincode: string
  items_amount: number
  discount_amount: number
  delivery_charge: number
  platform_fee: number
  wallet_amount_used: number
  is_inter_state: boolean
  taxable_amount: number
  cgst_amount: number
  sgst_amount: number
  igst_amount: number
  total_amount: number
  payment_method: string
  payment_reference?: string
  items?: InvoiceItem[]
  generated_at: string
  created_at: string
}

export interface InvoiceSearchParams {
  invoice_number?: string
  order_id?: string
  payment_status?: string
  date_from?: string
  date_to?: string
  page?: number
  limit?: number
}

export interface InvoiceSearchResponse {
  invoices: Invoice[]
  page: number
  limit: number
  total: number
  total_pages: number
}
