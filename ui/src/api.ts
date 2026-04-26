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

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
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
  const res = await fetch(`${API_BASE}/api/v1/messages?${params.toString()}`)
  return handleResponse<MessageMetadata[]>(res)
}

export async function getMessage(id: string): Promise<MessageMetadata> {
  const res = await fetch(`${API_BASE}/api/v1/messages/${id}`)
  return handleResponse<MessageMetadata>(res)
}

export interface Channel {
  id: string
  name: string
}

export async function listChannels(): Promise<Channel[]> {
  const res = await fetch(`${API_BASE}/api/v1/channels`)
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
  const res = await fetch(`${API_BASE}/api/v1/alerts`)
  return handleResponse<Alert[]>(res)
}
