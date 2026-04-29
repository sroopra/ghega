import { useEffect, useState } from 'react'
import { listMigrations, getMigration, type MigrationReport, type MigrationDetail } from '../api.ts'

function StatusBadge({ status }: { status: string }) {
  const cls =
    status === 'auto-converted'
      ? 'ok'
      : status === 'needs-rewrite' || status === 'mixed'
      ? 'warning'
      : status === 'unsupported'
      ? 'error'
      : 'info'
  return <span className={`status-badge ${cls}`}>{status}</span>
}

export default function MigrationsPage() {
  const [migrations, setMigrations] = useState<MigrationReport[]>([])
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [detail, setDetail] = useState<MigrationDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    listMigrations()
      .then((data) => {
        if (!cancelled) {
          setMigrations(data)
          setError(null)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          const msg = err instanceof Error ? err.message : String(err)
          setError(msg)
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    if (!selectedId) {
      setDetail(null)
      return
    }
    let cancelled = false
    getMigration(selectedId)
      .then((data) => {
        if (!cancelled) setDetail(data)
      })
      .catch(() => {
        if (!cancelled) {
          setDetail(null)
        }
      })
    return () => {
      cancelled = true
    }
  }, [selectedId])

  return (
    <div>
      <h2 className="page-title">Migrations</h2>
      <p className="empty-state">Migration reports are synthetic — no PHI is stored.</p>
      {loading && <p>Loading…</p>}
      {error && <p className="empty-state">Error: {error}</p>}
      {!loading && migrations.length === 0 && (
        <div className="card">
          <p className="empty-state">No migrations found. Run <code>ghega migrate mirth</code> to generate reports.</p>
        </div>
      )}
      {!loading && migrations.length > 0 && (
        <div className="card table-container">
          <table>
            <thead>
              <tr>
                <th>Channel</th>
                <th>Original Name</th>
                <th>Status</th>
                <th>Rewrite Tasks</th>
                <th>Warnings</th>
              </tr>
            </thead>
            <tbody>
              {migrations.map((m) => (
                <>
                  <tr
                    key={m.id}
                    onClick={() => setSelectedId(selectedId === m.id ? null : m.id)}
                    style={{ cursor: 'pointer' }}
                  >
                    <td>{m.channel_name}</td>
                    <td>{m.original_name}</td>
                    <td>
                      <StatusBadge status={m.status} />
                    </td>
                    <td>{m.rewrite_tasks_count}</td>
                    <td>{m.warnings_count}</td>
                  </tr>
                  {selectedId === m.id && detail && (
                    <tr key={`${m.id}-detail`}>
                      <td colSpan={5}>
                        <div className="card" style={{ margin: '0.5rem 0' }}>
                          <h4>Migration Details: {detail.channel_name}</h4>
                          {detail.warnings.length > 0 && (
                            <div style={{ marginBottom: '0.75rem' }}>
                              <strong>Warnings:</strong>
                              <ul>
                                {detail.warnings.map((w, i) => (
                                  <li key={i}>{w}</li>
                                ))}
                              </ul>
                            </div>
                          )}
                          {detail.auto_converted.length > 0 && (
                            <div style={{ marginBottom: '0.75rem' }}>
                              <strong>Auto-Converted ({detail.auto_converted.length}):</strong>
                              <ul>
                                {detail.auto_converted.map((c, i) => (
                                  <li key={i}>{c.element} — {c.description}</li>
                                ))}
                              </ul>
                            </div>
                          )}
                          {detail.needs_rewrite.length > 0 && (
                            <div style={{ marginBottom: '0.75rem' }}>
                              <strong>Needs Rewrite ({detail.needs_rewrite.length}):</strong>
                              <ul>
                                {detail.needs_rewrite.map((t, i) => (
                                  <li key={i}>
                                    <span className={`status-badge ${t.severity === 'high' ? 'error' : t.severity === 'medium' ? 'warning' : 'info'}`}>
                                      {t.severity}
                                    </span>{' '}
                                    {t.description}
                                    {t.category && <em> ({t.category})</em>}
                                  </li>
                                ))}
                              </ul>
                            </div>
                          )}
                          {detail.unsupported.length > 0 && (
                            <div>
                              <strong>Unsupported ({detail.unsupported.length}):</strong>
                              <ul>
                                {detail.unsupported.map((u, i) => (
                                  <li key={i}>
                                    {u.feature} — {u.description}
                                  </li>
                                ))}
                              </ul>
                            </div>
                          )}
                        </div>
                      </td>
                    </tr>
                  )}
                </>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
