import apiClient from './client'
import type { Settings } from '../types/settings'

export async function getSettings(): Promise<Settings> {
  const { data } = await apiClient.get<{ settings: Settings }>('/admin/settings')
  return data.settings
}

export async function updateSettings(settings: Partial<Settings>): Promise<Record<string, string>> {
  const { data } = await apiClient.put<{ success: boolean; updated: Record<string, string> }>('/admin/settings', {
    settings,
  })
  return data.updated
}
