import apiClient from './client'
import type { InvoiceSearchParams, InvoiceSearchResponse } from '../types/invoices'

export async function searchInvoices(params: InvoiceSearchParams): Promise<InvoiceSearchResponse> {
  const { data } = await apiClient.get<InvoiceSearchResponse>('/admin/invoices', { params })
  return data
}

export async function downloadInvoicePDF(id: number, invoiceNumber: string): Promise<void> {
  const response = await apiClient.get(`/admin/invoices/${id}/pdf`, { responseType: 'blob' })
  const url = window.URL.createObjectURL(new Blob([response.data]))
  const link = document.createElement('a')
  link.href = url
  link.setAttribute('download', `${invoiceNumber}.pdf`)
  document.body.appendChild(link)
  link.click()
  link.remove()
  window.URL.revokeObjectURL(url)
}
