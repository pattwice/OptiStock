import { apiClient, type ApiEnvelope } from './client'

export interface ReportPageResult {
  headers: string[]
  rows: string[][]
  total: number
  page: number
  page_size: number
}

async function unwrap<T>(promise: Promise<{ data: ApiEnvelope<T> }>) {
  const { data } = await promise
  return data.data
}

export function fetchReport(
  reportKey: string,
  params: Record<string, string | number | boolean | undefined>,
) {
  const query: Record<string, string> = {}
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== '') query[k] = String(v)
  }
  return unwrap<ReportPageResult>(apiClient.get(`/reports/${reportKey}`, { params: query }))
}

export async function downloadReportExport(
  reportKey: string,
  format: 'csv' | 'xlsx',
  params: Record<string, string | number | boolean | undefined>,
) {
  const query: Record<string, string> = { format }
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== '') query[k] = String(v)
  }
  const response = await apiClient.get(`/reports/${reportKey}/export`, {
    params: query,
    responseType: 'blob',
  })
  const blob = response.data as Blob
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${reportKey}.${format === 'xlsx' ? 'xlsx' : 'csv'}`
  link.click()
  window.URL.revokeObjectURL(url)
}
