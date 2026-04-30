import { useEffect } from 'react'
import { login } from '../api.ts'

export default function LoginPage() {
  useEffect(() => {
    login()
  }, [])

  return (
    <div className="card">
      <p>Redirecting to login...</p>
    </div>
  )
}
