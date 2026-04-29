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

const MOCK_MIGRATIONS: MigrationReport[] = [
  {
    id: 'adt-a01-mirth',
    channel_name: 'adt-a01',
    original_name: 'ADT A01 Inbound',
    status: 'auto-converted',
    rewrite_tasks_count: 0,
    warnings_count: 1,
  },
  {
    id: 'oru-r01-mirth',
    channel_name: 'oru-r01',
    original_name: 'ORU R01 Lab Results',
    status: 'needs-rewrite',
    rewrite_tasks_count: 3,
    warnings_count: 2,
  },
  {
    id: 'mdm-t02-mirth',
    channel_name: 'mdm-t02',
    original_name: 'MDM T02 Documents',
    status: 'unsupported',
    rewrite_tasks_count: 0,
    warnings_count: 1,
  },
]

const MOCK_DETAIL: Record<string, MigrationDetail> = {
  'adt-a01-mirth': {
    channel_name: 'adt-a01',
    original_name: 'ADT A01 Inbound',
    status: 'auto-converted',
    auto_converted: [
      { element: 'source_connector', description: 'Source mapped to type "mllp"' },
      { element: 'destination_connector', description: 'Destination mapped to type "http"' },
    ],
    needs_rewrite: [],
    unsupported: [],
    warnings: ['channel name sanitized from "ADT A01 Inbound" to "adt-a01"'],
  },
  'oru-r01-mirth': {
    channel_name: 'oru-r01',
    original_name: 'ORU R01 Lab Results',
    status: 'needs-rewrite',
    auto_converted: [
      { element: 'source_connector', description: 'Source mapped to type "mllp"' },
    ],
    needs_rewrite: [
      { severity: 'medium', description: 'Rewrite JS transformer for HL7 mapping', category: 'transformer' },
      { severity: 'high', description: 'Replace E4X XML construction with typed mapping', category: 'e4x' },
      { severity: 'low', description: 'Add validation for OBR segment', category: 'transformer' },
    ],
    unsupported: [],
    warnings: [
      'channel name sanitized from "ORU R01 Lab Results" to "oru-r01"',
      'JavaScript transformers/filters are present but not converted in this step',
    ],
  },
  'mdm-t02-mirth': {
    channel_name: 'mdm-t02',
    original_name: 'MDM T02 Documents',
    status: 'unsupported',
    auto_converted: [],
    needs_rewrite: [],
    unsupported: [
      { feature: 'destination_connector', description: 'unsupported destination connector type "com.mirth.connect.connectors.jdbc.DatabaseWriterProperties"' },
    ],
    warnings: ['channel name sanitized from "MDM T02 Documents" to "mdm-t02"'],
  },
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
          setMigrations(data.length > 0 ? data : MOCK_MIGRATIONS)
          setError(null)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          console.warn('API unavailable, using mock data:', err)
          setMigrations(MOCK_MIGRATIONS)
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
      .catch((err: unknown) => {
        if (!cancelled) {
          console.warn('API unavailable, using mock detail:', err)
          setDetail(MOCK_DETAIL[selectedId] || null)
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
