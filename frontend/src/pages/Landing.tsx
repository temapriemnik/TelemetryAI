import { useState } from 'react'
import { motion } from 'framer-motion'
import { 
  Shield, 
  Zap, 
  BarChart3, 
  Lock, 
  MessageSquare, 
  ArrowRight,
  Cpu,
  Activity,
  Eye
} from 'lucide-react'
import { clsx } from 'clsx'

interface LandingProps {
  onLogin: (token: string) => void
}

export default function Landing({ onLogin }: LandingProps) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState('')

  const handleLogin = async () => {
    setIsLoading(true)
    setError('')

    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      })

      const data = await res.json()
      
      if (res.ok && data.token) {
        onLogin(data.token)
      } else {
        setError(data.error || 'Login failed')
      }
    } catch (err) {
      setError('Connection error')
    } finally {
      setIsLoading(false)
    }
  }

  const handleRegister = async () => {
    setIsLoading(true)
    setError('')

    try {
      const res = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      })

      const data = await res.json()
      
      if (res.ok && data.token) {
        onLogin(data.token)
      } else {
        setError(data.error || 'Registration failed')
      }
    } catch (err) {
      setError('Connection error')
    } finally {
      setIsLoading(false)
    }
  }

  const features = [
    {
      icon: Shield,
      title: 'Real-time Protection',
      description: 'Detect security threats instantly with AI-powered analysis',
    },
    {
      icon: Zap,
      title: 'Lightning Fast',
      description: 'Process millions of logs in milliseconds',
    },
    {
      icon: BarChart3,
      title: 'Deep Analytics',
      description: 'Insights that drive decisions',
    },
  ]

  return (
    <div className="relative min-h-screen overflow-hidden">
      {/* Animated Background */}
      <div className="absolute inset-0 grid-pattern opacity-50" />
      <div className="absolute top-0 left-1/4 w-96 h-96 bg-primary/20 rounded-full blur-3xl animate-pulse-slow" />
      <div className="absolute bottom-0 right-1/4 w-96 h-96 bg-accent/10 rounded-full blur-3xl animate-pulse-slow" style={{ animationDelay: '2s' }} />

      {/* Navigation */}
      <motion.nav
        initial={{ y: -100 }}
        animate={{ y: 0 }}
        className="relative z-10 flex items-center justify-between px-8 py-6"
      >
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-primary to-accent flex items-center justify-center">
            <Activity className="w-6 h-6 text-white" />
          </div>
          <span className="text-xl font-bold gradient-text">TelemetryAI</span>
        </div>
        <button
          onClick={() => document.getElementById('auth')?.scrollIntoView({ behavior: 'smooth' })}
          className="px-4 py-2 text-sm font-medium text-textMuted hover:text-text transition-colors"
        >
          Get Started
        </button>
      </motion.nav>

      {/* Hero Section */}
      <div className="relative z-10 flex flex-col items-center justify-center min-h-[80vh] px-4">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
          className="text-center max-w-4xl"
        >
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: 0.2 }}
            className="inline-flex items-center gap-2 px-4 py-2 mb-6 rounded-full glass text-sm"
          >
            <span className="w-2 h-2 rounded-full bg-success animate-pulse" />
            <span className="text-textMuted">AI-Powered Log Analysis</span>
          </motion.div>

          <h1 className="text-5xl md:text-7xl font-bold tracking-tight mb-6">
            Understand your{' '}
            <span className="gradient-text">logs</span>
            <br />
            in real-time
          </h1>

          <p className="text-xl text-textMuted max-w-2xl mx-auto mb-10">
            Detect errors, analyze patterns, and get actionable insights from your telemetry data 
            with our AI-powered analysis engine.
          </p>

          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <motion.button
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              onClick={() => document.getElementById('auth')?.scrollIntoView({ behavior: 'smooth' })}
              className="group px-8 py-4 bg-gradient-to-r from-primary to-primaryHover rounded-xl font-semibold text-white flex items-center gap-3 glow"
            >
              Start Free
              <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
            </motion.button>
            <button className="px-8 py-4 rounded-xl font-medium text-textMuted hover:text-text border border-border hover:border-primary transition-colors">
              View Demo
            </button>
          </div>
        </motion.div>

        {/* Stats */}
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4 }}
          className="grid grid-cols-3 gap-8 mt-20"
        >
          {[
            { value: '10M+', label: 'Logs/day' },
            { value: '<10ms', label: 'Latency' },
            { value: '99.9%', label: 'Uptime' },
          ].map((stat, i) => (
            <div key={i} className="text-center">
              <div className="text-3xl md:text-4xl font-bold gradient-text">{stat.value}</div>
              <div className="text-textMuted text-sm">{stat.label}</div>
            </div>
          ))}
        </motion.div>
      </div>

      {/* Features */}
      <div className="relative z-10 py-32 px-8">
        <div className="max-w-6xl mx-auto">
          <motion.div
            initial={{ opacity: 0 }}
            whileInView={{ opacity: 1 }}
            viewport={{ once: true }}
            className="grid md:grid-cols-3 gap-6"
          >
            {features.map((feature, i) => (
              <motion.div
                key={feature.title}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ delay: i * 0.1 }}
                className="p-8 rounded-2xl glass hover:border-primary/50 transition-colors group"
              >
                <div className="w-12 h-12 rounded-xl bg-primary/20 flex items-center justify-center mb-4 group-hover:bg-primary/30 transition-colors">
                  <feature.icon className="w-6 h-6 text-primary" />
                </div>
                <h3 className="text-xl font-semibold mb-2">{feature.title}</h3>
                <p className="text-textMuted">{feature.description}</p>
              </motion.div>
            ))}
          </motion.div>
        </div>
      </div>

      {/* Auth Section */}
      <motion.div
        id="auth"
        initial={{ opacity: 0 }}
        whileInView={{ opacity: 1 }}
        viewport={{ once: true }}
        className="relative z-10 py-32 px-8"
      >
        <div className="max-w-md mx-auto">
          <div className="p-8 rounded-2xl glass">
            <h2 className="text-2xl font-bold text-center mb-2">Get Started</h2>
            <p className="text-textMuted text-center mb-8">Sign in to access your dashboard</p>

            <form className="space-y-4">
              <div>
                <input
                  type="email"
                  name="email"
                  placeholder="Email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full px-4 py-3 rounded-xl bg-surface border border-border focus:border-primary focus:outline-none transition-colors"
                  autoFocus
                  required
                />
              </div>
              <div>
                <input
                  type="password"
                  name="password"
                  placeholder="Password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full px-4 py-3 rounded-xl bg-surface border border-border focus:border-primary focus:outline-none transition-colors"
                  required
                />
              </div>

              {error && (
                <motion.p
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  className="text-error text-sm bg-error/10 p-3 rounded-lg"
                >
                  {error}
                </motion.p>
              )}

              <div className="grid grid-cols-2 gap-4">
                <motion.button
                  whileHover={{ scale: 1.01 }}
                  whileTap={{ scale: 0.99 }}
                  disabled={isLoading}
                  type="button"
                  onClick={handleLogin}
                  className="py-3 bg-surface border border-border rounded-xl font-medium text-text hover:border-primary transition-colors disabled:opacity-50"
                >
                  {isLoading ? 'Loading...' : 'Sign In'}
                </motion.button>
                
                <motion.button
                  whileHover={{ scale: 1.01 }}
                  whileTap={{ scale: 0.99 }}
                  disabled={isLoading}
                  type="button"
                  onClick={handleRegister}
                  className="py-3 bg-gradient-to-r from-primary to-primaryHover rounded-xl font-semibold text-white disabled:opacity-50"
                >
                  {isLoading ? 'Loading...' : 'Sign Up'}
                </motion.button>
              </div>
            </form>
          </div>
        </div>
      </motion.div>

      {/* Footer */}
      <footer className="relative z-10 py-8 px-8 text-center text-textMuted text-sm">
        <p>Powered by NATS JetStream</p>
      </footer>
    </div>
  )
}