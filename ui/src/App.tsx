import { Routes, Route, NavLink } from 'react-router-dom'
import { useAuth } from './context/AuthContext.tsx'
import WelcomePage from './pages/WelcomePage.tsx'
import MessagesPage from './pages/MessagesPage.tsx'
import ChannelsPage from './pages/ChannelsPage.tsx'
import OperationsPage from './pages/OperationsPage.tsx'
import MigrationsPage from './pages/MigrationsPage.tsx'
import SettingsPage from './pages/SettingsPage.tsx'
import AlertsPage from './pages/AlertsPage.tsx'
import LoginPage from './pages/LoginPage.tsx'
import './App.css'

function App() {
  const { user, loading, login, logout } = useAuth()

  return (
    <div className="app">
      <header className="app-header">
        <div className="brand">
          <h1>Ghega Console</h1>
          <span className="tagline">Typed healthcare integration. AI-assisted operations.</span>
        </div>
        <nav className="app-nav">
          <NavLink to="/" end>Home</NavLink>
          <NavLink to="/channels">Channels</NavLink>
          <NavLink to="/messages">Messages</NavLink>
          <NavLink to="/alerts">Alerts</NavLink>
          <NavLink to="/operations">Operations</NavLink>
          <NavLink to="/migrations">Migrations</NavLink>
          <NavLink to="/settings">Settings</NavLink>
        </nav>
        <div className="auth-controls">
          {loading ? (
            <span className="auth-loading">Loading…</span>
          ) : user ? (
            <>
              <span className="auth-user">{user.name || user.email}</span>
              <button className="auth-button" onClick={logout}>
                Logout
              </button>
            </>
          ) : (
            <button className="auth-button" onClick={login}>
              Login
            </button>
          )}
        </div>
      </header>
      <main className="app-main">
        <Routes>
          <Route path="/" element={<WelcomePage />} />
          <Route path="/channels" element={<ChannelsPage />} />
          <Route path="/messages" element={<MessagesPage />} />
          <Route path="/alerts" element={<AlertsPage />} />
          <Route path="/operations" element={<OperationsPage />} />
          <Route path="/migrations" element={<MigrationsPage />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="/login" element={<LoginPage />} />
        </Routes>
      </main>
      <footer className="app-footer">
        <p>© Ghega — Open-source healthcare integration engine</p>
      </footer>
    </div>
  )
}

export default App
