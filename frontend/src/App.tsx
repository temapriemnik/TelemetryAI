import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import Landing from './pages/Landing'
import Dashboard from './pages/Dashboard'

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [token, setToken] = useState<string | null>(null)

  useEffect(() => {
    const storedToken = localStorage.getItem('token')
    if (storedToken) {
      setToken(storedToken)
      setIsAuthenticated(true)
    }
  }, [])

  const handleLogin = (newToken: string) => {
    localStorage.setItem('token', newToken)
    setToken(newToken)
    setIsAuthenticated(true)
  }

  const handleLogout = () => {
    localStorage.removeItem('token')
    setToken(null)
    setIsAuthenticated(false)
  }

  return (
    <div className="min-h-screen bg-background">
      <AnimatePresence mode="wait">
        {isAuthenticated ? (
          <Dashboard key="dashboard" token={token!} onLogout={handleLogout} />
        ) : (
          <Landing key="landing" onLogin={handleLogin} />
        )}
      </AnimatePresence>
    </div>
  )
}

export default App