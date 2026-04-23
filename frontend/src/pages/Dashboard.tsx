import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useProjectStore, useAuthStore } from '../store';
import { projectAPI } from '../api';
import './Dashboard.css';

export default function Dashboard() {
  const { user } = useAuthStore();
  const { projects, setProjects, addProject } = useProjectStore();
  const [showModal, setShowModal] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadProjects();
  }, []);

  const loadProjects = async () => {
    try {
      const { data } = await projectAPI.list();
      setProjects(data);
    } catch (err) {
      console.error('Failed to load projects');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateProject = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const name = (form.elements.namedItem('name') as HTMLInputElement).value;
    const description = (form.elements.namedItem('description') as HTMLTextAreaElement).value;
    const webhookUrl = (form.elements.namedItem('webhook_url') as HTMLInputElement).value;
    
    try {
      const { data } = await projectAPI.create({ name, description, webhook_url: webhookUrl });
      addProject(data);
      setShowModal(false);
      form.reset();
    } catch (err) {
      console.error('Failed to create project');
    }
  };

  return (
    <div className="dashboard">
      <div className="dashboard-header">
        <div>
          <h1 className="page-title">Projects</h1>
          <p className="page-subtitle">Welcome back, {user?.full_name || user?.email}</p>
        </div>
        <button onClick={() => setShowModal(true)} className="btn btn-primary">
          <span>+</span> New Project
        </button>
      </div>

      {loading ? (
        <div className="loading">Loading projects...</div>
      ) : projects.length === 0 ? (
        <div className="empty-state">
          <div className="empty-icon">📦</div>
          <h3>No projects yet</h3>
          <p>Create your first project to get started</p>
          <button onClick={() => setShowModal(true)} className="btn btn-primary">
            Create Project
          </button>
        </div>
      ) : (
        <div className="projects-grid">
          {projects.map((project) => (
            <Link key={project.id} to={`/project/${project.id}`} className="project-card">
              <div className="project-header">
                <h3 className="project-name">{project.name}</h3>
                <span className={`badge ${project.status === 'active' ? 'badge-success' : ''}`}>
                  {project.status}
                </span>
              </div>
              {project.description && (
                <p className="project-description">{project.description}</p>
              )}
              <div className="project-meta">
                <span>{project.api_keys?.length || 0} API keys</span>
                <span>{new Date(project.created_at).toLocaleDateString()}</span>
              </div>
            </Link>
          ))}
        </div>
      )}

      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h2 className="modal-title">Create New Project</h2>
            <form onSubmit={handleCreateProject} className="modal-form">
              <div className="form-group">
                <label>Project Name</label>
                <input type="text" name="name" className="input" required placeholder="My Project" />
              </div>
              <div className="form-group">
                <label>Description</label>
                <textarea name="description" className="input" rows={3} placeholder="Project description..." />
              </div>
              <div className="form-group">
                <label>Webhook URL (optional)</label>
                <input type="url" name="webhook_url" className="input" placeholder="https://..." />
              </div>
              <div className="modal-actions">
                <button type="button" onClick={() => setShowModal(false)} className="btn btn-secondary">
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary">
                  Create Project
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}