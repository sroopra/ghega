import { Routes, Route, NavLink } from 'react-router-dom'
import WelcomePage from './pages/WelcomePage.tsx'
import MessagesPage from './pages/MessagesPage.tsx'
import ChannelsPage from './pages/ChannelsPage.tsx'
import OperationsPage from './pages/OperationsPage.tsx'
import MigrationsPage from './pages/MigrationsPage.tsx'
import SettingsPage from './pages/SettingsPage.tsx'
import AlertsPage from './pages/AlertsPage.tsx'
import './App.css'

function App() {
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
        </Routes>
      </main>
      <footer className="app-footer">
        <p>© Ghega — Open-source healthcare integration engine</p>
      </footer>
    </div>
  )
}

export default App
