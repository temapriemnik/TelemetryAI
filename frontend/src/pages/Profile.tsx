import { useState } from 'react';
import { useAuthStore } from '../store';
import { userAPI } from '../api';
import './Profile.css';

export default function Profile() {
  const { user, setUser } = useAuthStore();
  const [form, setForm] = useState({
    full_name: user?.full_name || '',
    bio: user?.bio || '',
    avatar_url: user?.avatar_url || '',
  });
  const [password, setPassword] = useState('');
  const [success, setSuccess] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');
    setLoading(true);
    try {
      const data: any = {};
      if (form.full_name) data.full_name = form.full_name;
      if (form.bio) data.bio = form.bio;
      if (form.avatar_url) data.avatar_url = form.avatar_url;
      if (password) data.password = password;
      const { data: updated } = await userAPI.updateMe(data);
      setUser(updated);
      setPassword('');
      setSuccess('Profile updated successfully');
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Failed to update profile');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="profile animate-fadeIn">
      <h1 className="page-title">Profile Settings</h1>
      <p className="page-subtitle">Manage your account information</p>

      <div className="profile-content">
        <div className="profile-card card">
          <div className="profile-avatar-section">
            <div className="avatar-large">
              {user?.full_name?.[0] || user?.email[0]}
            </div>
            <div>
              <h3 className="profile-name">{user?.full_name || 'No name set'}</h3>
              <p className="profile-email">{user?.email}</p>
              <span className={`badge ${user?.role === 'admin' ? 'badge-success' : ''}`}>
                {user?.role}
              </span>
            </div>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="profile-form card">
          <h3>Edit Profile</h3>
          {success && <div className="success-message">{success}</div>}
          {error && <div className="auth-error">{error}</div>}

          <div className="form-group">
            <label>Full Name</label>
            <input
              type="text"
              className="input"
              value={form.full_name}
              onChange={(e) => setForm({ ...form, full_name: e.target.value })}
              placeholder="John Doe"
            />
          </div>

          <div className="form-group">
            <label>Bio</label>
            <textarea
              className="input"
              value={form.bio}
              onChange={(e) => setForm({ ...form, bio: e.target.value })}
              rows={3}
              placeholder="Tell us about yourself..."
            />
          </div>

          <div className="form-group">
            <label>Avatar URL</label>
            <input
              type="url"
              className="input"
              value={form.avatar_url}
              onChange={(e) => setForm({ ...form, avatar_url: e.target.value })}
              placeholder="https://..."
            />
          </div>

          <div className="form-divider"></div>

          <div className="form-group">
            <label>New Password</label>
            <input
              type="password"
              className="input"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Leave blank to keep current"
            />
          </div>

          <button type="submit" className="btn btn-primary" disabled={loading}>
            {loading ? 'Saving...' : 'Save Changes'}
          </button>
        </form>
      </div>
    </div>
  );
}