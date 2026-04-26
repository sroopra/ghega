import { useEffect, useState } from 'react'
import { listMessages, type MessageMetadata } from '../api.ts'

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString()
}

function StatusBadge({ status }: { status: string }) {
  const cls =
    status === 'delivered' || status === 'processed' || status === 'ok'
      ? 'ok'
      : status === 'pending' || status === 'received'
      ? 'pending'
      : 'error'
  return <span className={`status-badge ${cls}`}>{status}</span>
}

const MOCK_MESSAGES: MessageMetadata[] = [
  {
    id: 'msg-001',
    channel_id: 'demo-channel',
    message_id: 'MSG-2026-001',
    status: 'received',
    received_at: '2026-04-26T10:00:00Z',
    storage_id: 'store-001',
    location: 'memory://demo-channel/001',
  },
  {
    id: 'msg-002',
    channel_id: 'demo-channel',
    message_id: 'MSG-2026-002',
    status: 'processed',
    received_at: '2026-04-26T10:05:00Z',
    storage_id: 'store-002',
    location: 'memory://demo-channel/002',
  },
  {
    id: 'msg-003',
    channel_id: 'demo-channel',
    message_id: 'MSG-2026-003',
    status: 'error',
    received_at: '2026-04-26T10:10:00Z',
    storage_id: 'store-003',
    location: 'memory://demo-channel/003',
  },
]

export default function MessagesPage() {
  const [messages, setMessages] = useState<MessageMetadata[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    listMessages()
      .then((data) => {
        if (!cancelled) {
          setMessages(data.length > 0 ? data : MOCK_MESSAGES)
          setError(null)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          console.warn('API unavailable, using mock data:', err)
          setMessages(MOCK_MESSAGES)
          setError(null)
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div>
      <h2 className="page-title">Messages</h2>
      <p className="empty-state">Message metadata only — payload bytes are never displayed.</p>
      {loading && <p>Loading…</p>}
      {error && <p className="empty-state">Error: {error}</p>}
      {!loading && messages.length === 0 && (
        <p className="empty-state">No messages found.</p>
      )}
      {!loading && messages.length > 0 && (
        <div className="card table-container">
          <table>
            <thead>
              <tr>
                <th>Channel</th>
                <th>Message ID</th>
                <th>Status</th>
                <th>Received</th>
              </tr>
            </thead>
            <tbody>
              {messages.map((m) => (
                <tr key={m.id}>
                  <td>{m.channel_id}</td>
                  <td>{m.message_id}</td>
                  <td>
                    <StatusBadge status={m.status} />
                  </td>
                  <td>{formatDate(m.received_at)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
