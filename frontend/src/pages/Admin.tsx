import { useEffect, useState } from 'react';
import { adminAPI } from '../api';
import './Admin.css';

interface Stats {
  total_users: number;
  total_projects: number;
  total_reviews: number;
  users_by_day: { date: string; count: number }[];
  projects_by_day: { date: string; count: number }[];
}

interface User {
  id: number;
  email: string;
  full_name: string | null;
  role: string;
  is_active: boolean;
  created_at: string;
}

export default function Admin() {
  const [stats, setStats] = useState<Stats | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [statsRes, usersRes] = await Promise.all([
        adminAPI.getStats(),
        adminAPI.listUsers(),
      ]);
      setStats(statsRes.data);
      setUsers(usersRes.data);
    } catch (err) {
      console.error('Failed to load admin data');
    } finally {
      setLoading(false);
    }
  };

  const toggleUserActive = async (userId: number) => {
    try {
      await adminAPI.toggleUserActive(userId);
      setUsers(users.map((u) => u.id === userId ? { ...u, is_active: !u.is_active } : u));
    } catch (err) {
      console.error('Failed to toggle user active');
    }
  };

  const deleteUser = async (userId: number) => {
    if (!confirm('Are you sure you want to delete this user?')) return;
    try {
      await adminAPI.deleteUser(userId);
      setUsers(users.filter((u) => u.id !== userId));
    } catch (err) {
      console.error('Failed to delete user');
    }
  };

  const maxCount = stats ? Math.max(
    ...stats.users_by_day.map((d) => d.count),
    ...stats.projects_by_day.map((d) => d.count),
    1
  ) : 1;

  if (loading) return <div className="loading">Loading admin panel...</div>;

  return (
    <div className="admin animate-fadeIn">
      <h1 className="page-title">Admin Dashboard</h1>
      <p className="page-subtitle">Manage users and view analytics</p>

      <div className="stats-grid">
        <div className="stat-card">
          <span className="stat-value">{stats?.total_users || 0}</span>
          <span className="stat-label">Total Users</span>
        </div>
        <div className="stat-card">
          <span className="stat-value">{stats?.total_projects || 0}</span>
          <span className="stat-label">Total Projects</span>
        </div>
        <div className="stat-card">
          <span className="stat-value">{stats?.total_reviews || 0}</span>
          <span className="stat-label">Total Reviews</span>
        </div>
      </div>

      {stats && (
        <div className="charts-section">
          <div className="chart-card">
            <h3>Users (Last 30 Days)</h3>
            <div className="chart">
              {stats.users_by_day.map((d, i) => (
                <div key={i} className="chart-bar">
                  <div
                    className="chart-bar-fill users"
                    style={{ height: `${(d.count / maxCount) * 100}%` }}
                    title={`${d.date}: ${d.count}`}
                  />
                </div>
              ))}
            </div>
          </div>
          <div className="chart-card">
            <h3>Projects (Last 30 Days)</h3>
            <div className="chart">
              {stats.projects_by_day.map((d, i) => (
                <div key={i} className="chart-bar">
                  <div
                    className="chart-bar-fill projects"
                    style={{ height: `${(d.count / maxCount) * 100}%` }}
                    title={`${d.date}: ${d.count}`}
                  />
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      <div className="users-section">
        <h3>Users</h3>
        <div className="users-table">
          <div className="table-header">
            <span>Email</span>
            <span>Name</span>
            <span>Role</span>
            <span>Status</span>
            <span>Joined</span>
            <span>Actions</span>
          </div>
          {users.map((user) => (
            <div key={user.id} className="table-row">
              <span className="user-email">{user.email}</span>
              <span>{user.full_name || '-'}</span>
              <span className={`badge ${user.role === 'admin' ? 'badge-success' : ''}`}>
                {user.role}
              </span>
              <span className={`badge ${user.is_active ? 'badge-success' : ''}`}>
                {user.is_active ? 'Active' : 'Inactive'}
              </span>
              <span>{new Date(user.created_at).toLocaleDateString()}</span>
              <div className="user-actions">
                <button onClick={() => toggleUserActive(user.id)} className="btn-action">
                  {user.is_active ? 'Disable' : 'Enable'}
                </button>
                <button onClick={() => deleteUser(user.id)} className="btn-action btn-danger">
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}