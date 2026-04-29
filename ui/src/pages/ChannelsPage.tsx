import { useEffect, useState } from 'react'
import { listChannels, type Channel } from '../api.ts'

export default function ChannelsPage() {
  const [channels, setChannels] = useState<Channel[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    listChannels()
      .then((data) => {
        if (!cancelled) {
          setChannels(data)
          setError(null)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          const message = err instanceof Error ? err.message : String(err)
          setError(message)
          setChannels([])
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
      <h2 className="page-title">Channels</h2>
      {loading && <p>Loading…</p>}
      {error && <p className="empty-state">Error: {error}</p>}
      {!loading && channels.length === 0 && (
        <p className="empty-state">No channels found.</p>
      )}
      {!loading && channels.length > 0 && (
        <div className="card table-container">
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Name</th>
              </tr>
            </thead>
            <tbody>
              {channels.map((c) => (
                <tr key={c.id}>
                  <td>{c.id}</td>
                  <td>{c.name}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
