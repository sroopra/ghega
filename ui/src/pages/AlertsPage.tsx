import { useEffect, useState } from 'react'
import { listAlerts, type Alert } from '../api.ts'

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString()
}

function SeverityBadge({ severity }: { severity: string }) {
  const cls =
    severity === 'critical'
      ? 'critical'
      : severity === 'warning'
      ? 'warning'
      : 'info'
  return <span className={`status-badge ${cls}`}>{severity}</span>
}

const MOCK_ALERTS: Alert[] = [
  {
    id: 'alert-001',
    channel_id: 'demo-channel',
    severity: 'critical',
    message: 'Channel connection lost',
    created_at: '2026-04-26T10:00:00Z',
  },
  {
    id: 'alert-002',
    channel_id: 'inbound-orders',
    severity: 'warning',
    message: 'Message processing delayed',
    created_at: '2026-04-26T10:05:00Z',
  },
  {
    id: 'alert-003',
    channel_id: 'notifications',
    severity: 'info',
    message: 'System maintenance scheduled',
    created_at: '2026-04-26T10:10:00Z',
  },
]

export default function AlertsPage() {
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    listAlerts()
      .then((data) => {
        if (!cancelled) {
          setAlerts(data.length > 0 ? data : MOCK_ALERTS)
          setError(null)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          console.warn('API unavailable, using mock data:', err)
          setAlerts(MOCK_ALERTS)
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
      <h2 className="page-title">Alerts</h2>
      <p className="empty-state">Alert management — no PHI is stored in alert records.</p>
      {loading && <p>Loading…</p>}
      {error && <p className="empty-state">Error: {error}</p>}
      {!loading && alerts.length === 0 && (
        <p className="empty-state">No alerts found.</p>
      )}
      {!loading && alerts.length > 0 && (
        <div className="card table-container">
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Channel</th>
                <th>Severity</th>
                <th>Message</th>
                <th>Created At</th>
              </tr>
            </thead>
            <tbody>
              {alerts.map((a) => (
                <tr key={a.id}>
                  <td>{a.id}</td>
                  <td>{a.channel_id}</td>
                  <td>
                    <SeverityBadge severity={a.severity} />
                  </td>
                  <td>{a.message}</td>
                  <td>{formatDate(a.created_at)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
