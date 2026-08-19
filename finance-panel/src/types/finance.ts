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
