export interface DailySalesSummary {
  date: string
  total_orders: number
  delivered_orders: number
  cancelled_orders: number
  pending_orders: number
  total_revenue: number
  cod_revenue: number
  online_revenue: number
  cod_orders: number
  online_orders: number
  total_delivery_charge: number
  total_wallet_used: number
  avg_order_value: number
}

export interface RangeSalesSummary {
  from: string
  to: string
  total_orders: number
  delivered_orders: number
  cancelled_orders: number
  pending_orders: number
  total_revenue: number
  cod_revenue: number
  online_revenue: number
  total_delivery_charge: number
  total_wallet_used: number
  avg_order_value: number
  taxable_amount: number
  cgst_amount: number
  sgst_amount: number
  igst_amount: number
  total_output_gst: number
  total_vendor_gst: number
}
