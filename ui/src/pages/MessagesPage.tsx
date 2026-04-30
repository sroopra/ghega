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
          setMessages(data)
          setError(null)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          const msg = err instanceof Error ? err.message : String(err)
          setError(msg)
          setMessages([])
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
