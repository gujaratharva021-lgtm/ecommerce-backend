import { apiClient } from './client'
import type { Address, AddressRequest } from '../types'

export const listAddresses = () =>
  apiClient.get<Address[]>('/addresses').then((r) => r.data)

export const createAddress = (data: AddressRequest) =>
  apiClient.post<Address>('/addresses', data).then((r) => r.data)

export const updateAddress = (id: number, data: AddressRequest) =>
  apiClient.put<Address>(`/addresses/${id}`, data).then((r) => r.data)

export const deleteAddress = (id: number) =>
  apiClient.delete(`/addresses/${id}`).then((r) => r.data)

export const setDefaultAddress = (id: number) =>
  apiClient.put<Address>(`/addresses/${id}/default`, {}).then((r) => r.data)
