export interface RevenueDataPoint {
  date: string
  gross_revenue: number
  net_revenue: number
  orders_count: number
}

export interface RevenueSummary {
  total_gross_revenue: number
  total_net_revenue: number
  total_orders: number
  average_order_value: number
}

export interface RevenueResponse {
  summary: RevenueSummary
  daily: RevenueDataPoint[]
}
