import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import {
  Activity,
  BarChart3,
  AlertTriangle,
  CheckCircle,
  Settings,
  LogOut,
  Plus,
  Key,
  Trash2,
  ArrowUpRight,
  Search,
  Zap,
} from 'lucide-react'
import { clsx } from 'clsx'

const API_BASE = '/api'
const TELEMETRY_URL = 'http://telemetry_service:8080'

interface Project {
  id: string
  name: string
  created_at: string
}

interface APIKey {
  api_key: string
  project_id: string
}

interface DashboardProps {
  token: string
  onLogout: () => void
}

export default function Dashboard({ token, onLogout }: DashboardProps) {
  const [projects, setProjects] = useState<Project[]>([])
  const [apiKeys, setApiKeys] = useState<Record<string, string>>({})
  const [activeProject, setActiveProject] = useState<string | null>(null)
  const [isLoadingProjects, setIsLoadingProjects] = useState(true)
  const [showNewProject, setShowNewProject] = useState(false)
  const [newProjectName, setNewProjectName] = useState('')
  const [detectLog, setDetectLog] = useState('')
  const [detectResult, setDetectResult] = useState<{ level: string; project_id: string } | null>(null)
  const [isDetecting, setIsDetecting] = useState(false)

  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Token ${token}`,
  }

  useEffect(() => {
    fetchProjects()
  }, [])

const fetchProjects = async () => {
    try {
      const res = await fetch(`${API_BASE}/projects`, { headers })
      if (res.ok) {
        const data = await res.json()
        setProjects(data.projects || [])
        if (data.projects?.length > 0) {
          setActiveProject(data.projects[0].id)
        }
      }
    } catch (err) {
      console.error('Failed to fetch projects:', err)
    } finally {
      setIsLoadingProjects(false)
    }
  }

  const createProject = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const res = await fetch(`${API_BASE}/projects`, {
        method: 'POST',
        headers,
        body: JSON.stringify({ name: newProjectName }),
      })
      if (res.ok) {
        const data = await res.json()
        setProjects([...projects, data])
        setNewProjectName('')
        setShowNewProject(false)
      }
    } catch (err) {
      console.error('Failed to create project:', err)
    }
  }

  const deleteProject = async (projectId: string) => {
    try {
      await fetch(`${API_BASE}/projects/${projectId}`, {
        method: 'DELETE',
        headers,
      })
      setProjects(projects.filter(p => p.id !== projectId))
    } catch (err) {
      console.error('Failed to delete project:', err)
    }
  }

  const getApiKey = async (projectId: string) => {
    try {
      const res = await fetch(`${API_BASE}/projects/${projectId}/apikey`, { headers })
      if (res.ok) {
        const data = await res.json()
        setApiKeys({ ...apiKeys, [projectId]: data.api_key })
      }
    } catch (err) {
      console.error('Failed to get API key:', err)
    }
  }

  const deleteApiKey = async (apiKey: string) => {
    try {
      await fetch(`${API_BASE}/apikeys`, {
        method: 'DELETE',
        headers,
        body: JSON.stringify({ api_key: apiKey }),
      })
    } catch (err) {
      console.error('Failed to delete API key:', err)
    }
  }

  const handleDetect = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsDetecting(true)
    setDetectResult(null)

    const apiKey = Object.entries(apiKeys).find(([pid]) => pid === activeProject)?.[1] || 'test123'

    try {
      const res = await fetch(`${TELEMETRY_URL}/detect`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ log: detectLog, api_key: apiKey }),
      })
      const data = await res.json()
      setDetectResult(data)
    } catch (err) {
      console.error('Failed to detect:', err)
} finally {
      setIsDetecting(false)
    }
  }

  const getLevelColor = (level: string) => {
    switch (level) {
      case 'ERROR':
        return 'text-error bg-error/20 border-error/50'
      case 'WARNING':
        return 'text-warning bg-warning/20 border-warning/50'
      case 'INFO':
        return 'text-accent bg-accent/20 border-accent/50'
      default:
        return 'text-textMuted bg-surface border-border'
    }
  }

  return (
    <div className="flex min-h-screen bg-background">
      {/* Sidebar */}
      <motion.aside
        initial={{ x: -100 }}
        animate={{ x: 0 }}
        className="w-64 border-r border-border p-6 flex flex-col"
      >
        <div className="flex items-center gap-3 mb-8">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-primary to-accent flex items-center justify-center">
            <Activity className="w-6 h-6 text-white" />
          </div>
          <span className="text-lg font-bold gradient-text">TelemetryAI</span>
        </div>

        <nav className="flex-1 space-y-2">
          <button className="w-full flex items-center gap-3 px-4 py-3 rounded-xl bg-primary/20 text-primary">
            <BarChart3 className="w-5 h-5" />
            Dashboard
          </button>
          <button className="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-textMuted hover:text-text hover:bg-surface transition-colors">
            <Search className="w-5 h-5" />
            Logs
          </button>
        </nav>

        <button
          onClick={onLogout}
          className="flex items-center gap-3 px-4 py-3 rounded-xl text-textMuted hover:text-error hover:bg-error/10 transition-colors"
        >
          <LogOut className="w-5 h-5" />
          Sign Out
        </button>
      </motion.aside>

      {/* Main Content */}
      <main className="flex-1 p-8 overflow-auto">
        <div className="max-w-6xl mx-auto">
          {/* Header */}
          <div className="flex items-center justify-between mb-8">
            <div>
              <h1 className="text-2xl font-bold">Dashboard</h1>
              <p className="text-textMuted">Welcome back</p>
            </div>
            <button
              onClick={() => setShowNewProject(true)}
              className="flex items-center gap-2 px-4 py-2 bg-primary rounded-xl font-medium text-white"
            >
              <Plus className="w-5 h-5" />
              New Project
            </button>
          </div>

          {/* Stats */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            {[
              { label: 'Total Projects', value: projects.length, icon: BarChart3 },
              { label: 'API Keys', value: Object.keys(apiKeys).length, icon: Key },
              { label: 'Logs Analyzed', value: '0', icon: Activity },
            ].map((stat, i) => (
              <motion.div
                key={stat.label}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.1 }}
                className="p-6 rounded-2xl glass"
              >
                <div className="flex items-center gap-4">
                  <div className="w-12 h-12 rounded-xl bg-primary/20 flex items-center justify-center">
                    <stat.icon className="w-6 h-6 text-primary" />
                  </div>
                  <div>
                    <div className="text-2xl font-bold">{stat.value}</div>
                    <div className="text-textMuted text-sm">{stat.label}</div>
                  </div>
                </div>
              </motion.div>
            ))}
          </div>

          {/* Projects */}
          <motion.section
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.3 }}
            className="mb-8"
          >
            <h2 className="text-lg font-semibold mb-4">Projects</h2>
            
            {isLoadingProjects ? (
              <div className="p-8 text-center text-textMuted">Loading...</div>
            ) : projects.length === 0 ? (
              <div className="p-8 text-center text-textMuted rounded-2xl glass">
                No projects yet. Create one to get started.
              </div>
            ) : (
              <div className="grid gap-4">
                {projects.map((project) => (
                  <motion.div
                    key={project.id}
                    initial={{ opacity: 0, x: -20 }}
                    animate={{ opacity: 1, x: 0 }}
                    className={clsx(
                      'p-4 rounded-xl border cursor-pointer transition-colors',
                      activeProject === project.id
                        ? 'border-primary bg-primary/10'
                        : 'border-border hover:border-primary/50'
                    )}
                    onClick={() => setActiveProject(project.id)}
                  >
                    <div className="flex items-center justify-between">
                      <div>
                        <h3 className="font-medium">{project.name}</h3>
                        <p className="text-textMuted text-sm">{project.id}</p>
                      </div>
                      <div className="flex items-center gap-2">
                        <button
                          onClick={(e) => {
                            e.stopPropagation()
                            getApiKey(project.id)
                          }}
                          className="p-2 rounded-lg hover:bg-surface transition-colors"
                        >
                          <Key className="w-4 h-4 text-textMuted" />
                        </button>
                        <button
                          onClick={(e) => {
                            e.stopPropagation()
                            deleteProject(project.id)
                          }}
                          className="p-2 rounded-lg hover:bg-error/10 transition-colors"
                        >
                          <Trash2 className="w-4 h-4 text-error" />
                        </button>
                      </div>
                    </div>
                    {apiKeys[project.id] && (
                      <motion.div
                        initial={{ opacity: 0, height: 0 }}
                        animate={{ opacity: 1, height: 'auto' }}
                        className="mt-3 p-3 rounded-lg bg-surface"
                      >
                        <code className="text-sm text-accent">{apiKeys[project.id]}</code>
                      </motion.div>
                    )}
                  </motion.div>
                ))}
              </div>
            )}
          </motion.section>

          {/* Detect Tool */}
          <motion.section
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.4 }}
          >
            <h2 className="text-lg font-semibold mb-4">Log Analysis</h2>
            <div className="p-6 rounded-2xl glass">
              <form onSubmit={handleDetect} className="space-y-4">
                <div>
                  <textarea
                    placeholder="Paste your log message here..."
                    value={detectLog}
                    onChange={(e) => setDetectLog(e.target.value)}
                    className="w-full h-32 px-4 py-3 rounded-xl bg-surface border border-border focus:border-primary focus:outline-none transition-colors resize-none"
                    required
                  />
                </div>
                <div className="flex items-center gap-4">
                  <motion.button
                    whileHover={{ scale: 1.01 }}
                    whileTap={{ scale: 0.99 }}
                    type="submit"
                    disabled={isDetecting}
                    className="flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-primary to-primaryHover rounded-xl font-semibold text-white disabled:opacity-50"
                  >
                    <Zap className="w-5 h-5" />
                    {isDetecting ? 'Analyzing...' : 'Analyze Log'}
                  </motion.button>
                </div>
              </form>

              <AnimatePresence>
                {detectResult && (
                  <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -20 }}
                    className={clsx(
                      'mt-6 p-6 rounded-xl border',
                      getLevelColor(detectResult.level)
                    )}
                  >
                    <div className="flex items-center gap-4">
                      {detectResult.level === 'ERROR' ? (
                        <AlertTriangle className="w-8 h-8" />
                      ) : detectResult.level === 'WARNING' ? (
                        <Activity className="w-8 h-8" />
                      ) : (
                        <CheckCircle className="w-8 h-8" />
                      )}
                      <div>
                        <div className="text-2xl font-bold">{detectResult.level}</div>
                        <div className="text-textMuted">
                          Project: {detectResult.project_id || 'N/A'}
                        </div>
                      </div>
                    </div>
                  </motion.div>
                )}
              </AnimatePresence>
            </div>
          </motion.section>
        </div>
      </main>

      {/* New Project Modal */}
      <AnimatePresence>
        {showNewProject && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
            onClick={() => setShowNewProject(false)}
          >
            <motion.div
              initial={{ scale: 0.9, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0.9, opacity: 0 }}
              className="p-8 rounded-2xl glass max-w-md w-full"
              onClick={(e) => e.stopPropagation()}
            >
              <h2 className="text-xl font-bold mb-4">Create New Project</h2>
              <form onSubmit={createProject}>
                <input
                  type="text"
                  placeholder="Project Name"
                  value={newProjectName}
                  onChange={(e) => setNewProjectName(e.target.value)}
                  className="w-full px-4 py-3 rounded-xl bg-surface border border-border focus:border-primary focus:outline-none mb-4"
                  required
                />
                <div className="flex gap-4">
                  <button
                    type="button"
                    onClick={() => setShowNewProject(false)}
                    className="flex-1 py-3 rounded-xl font-medium border border-border"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="flex-1 py-3 rounded-xl font-medium bg-primary text-white"
                  >
                    Create
                  </button>
                </div>
              </form>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}