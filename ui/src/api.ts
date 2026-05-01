const API_BASE = import.meta.env.VITE_API_BASE_URL || ''

export interface MessageMetadata {
  id: string
  channel_id: string
  message_id: string
  status: string
  received_at: string
  storage_id: string
  location: string
}

export interface ApiError {
  error: string
}

export interface User {
  email: string
  name: string
  roles: string[]
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    if (res.status === 401) {
      window.location.href = '/auth/login'
      // Prevent further execution by throwing; the redirect will navigate away.
      throw new Error('Unauthorized')
    }
    const err = (await res.json().catch(() => ({}))) as ApiError
    throw new Error(err.error || `HTTP ${res.status}`)
  }
  return res.json() as Promise<T>
}

export async function listMessages(
  channelID?: string,
  limit = 50,
  offset = 0
): Promise<MessageMetadata[]> {
  const params = new URLSearchParams()
  if (channelID) params.set('channel_id', channelID)
  params.set('limit', String(limit))
  params.set('offset', String(offset))
  const res = await fetch(`${API_BASE}/api/v1/messages?${params.toString()}`, {
    credentials: 'include',
  })
  return handleResponse<MessageMetadata[]>(res)
}

export async function getMessage(id: string): Promise<MessageMetadata> {
  const res = await fetch(`${API_BASE}/api/v1/messages/${id}`, {
    credentials: 'include',
  })
  return handleResponse<MessageMetadata>(res)
}

export interface Channel {
  id: string
  name: string
}

export async function listChannels(): Promise<Channel[]> {
  const res = await fetch(`${API_BASE}/api/v1/channels`, {
    credentials: 'include',
  })
  return handleResponse<Channel[]>(res)
}

export interface Alert {
  id: string
  channel_id: string
  severity: string
  message: string
  created_at: string
  resolved_at?: string
  acknowledged_at?: string
}

export async function listAlerts(): Promise<Alert[]> {
  const res = await fetch(`${API_BASE}/api/v1/alerts`, {
    credentials: 'include',
  })
  return handleResponse<Alert[]>(res)
}

export interface MigrationReport {
  id: string
  channel_name: string
  original_name: string
  status: string
  rewrite_tasks_count: number
  warnings_count: number
}

export interface MigrationDetail {
  channel_name: string
  original_name: string
  status: string
  auto_converted: { element: string; description: string }[]
  needs_rewrite: { severity: string; description: string; category?: string }[]
  unsupported: { feature: string; description: string }[]
  warnings: string[]
}

export async function listMigrations(): Promise<MigrationReport[]> {
  const res = await fetch(`${API_BASE}/api/v1/migrations`, {
    credentials: 'include',
  })
  return handleResponse<MigrationReport[]>(res)
}

export async function getMigration(id: string): Promise<MigrationDetail> {
  const res = await fetch(`${API_BASE}/api/v1/migrations/${id}`, {
    credentials: 'include',
  })
  return handleResponse<MigrationDetail>(res)
}

export async function getMe(): Promise<User> {
  const res = await fetch(`${API_BASE}/api/v1/me`, {
    credentials: 'include',
  })
  return handleResponse<User>(res)
}

export function login(): void {
  window.location.href = '/auth/login'
}

export function logout(): void {
  window.location.href = '/auth/logout'
}
