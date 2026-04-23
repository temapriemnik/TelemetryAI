import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { projectAPI, reviewAPI } from '../api';
import './ProjectPage.css';

interface Review {
  id: number;
  rating: number;
  content: string;
  user_id: number;
  project_id: number;
  created_at: string;
  user: { email: string; full_name: string | null };
}

export default function ProjectPage() {
  const { id } = useParams();
  const [project, setProject] = useState<any>(null);
  const [reviews, setReviews] = useState<Review[]>([]);
  const [showReviewModal, setShowReviewModal] = useState(false);
  const [showKeyModal, setShowKeyModal] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (id) loadProject();
  }, [id]);

  const loadProject = async () => {
    try {
      const { data: proj } = await projectAPI.get(Number(id));
      setProject(proj);
      const { data: rev } = await reviewAPI.list(Number(id));
      setReviews(rev);
    } catch (err) {
      console.error('Failed to load project');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateApiKey = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const name = (form.elements.namedItem('name') as HTMLInputElement).value;
    try {
      const { data } = await projectAPI.createApiKey(Number(id), { name });
      setProject({ ...project, api_keys: [...project.api_keys, data] });
      setShowKeyModal(false);
      form.reset();
    } catch (err) {
      console.error('Failed to create API key');
    }
  };

  const handleCreateReview = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const rating = parseInt((form.elements.namedItem('rating') as HTMLSelectElement).value);
    const content = (form.elements.namedItem('content') as HTMLTextAreaElement).value;
    try {
      const { data } = await reviewAPI.create({ project_id: Number(id), rating, content });
      setReviews([...reviews, data]);
      setShowReviewModal(false);
      form.reset();
    } catch (err) {
      console.error('Failed to create review');
    }
  };

  if (loading) return <div className="loading">Loading project...</div>;
  if (!project) return <div className="loading">Project not found</div>;

  return (
    <div className="project-page animate-fadeIn">
      <div className="project-header-section">
        <div>
          <h1 className="page-title">{project.name}</h1>
          {project.description && <p className="page-subtitle">{project.description}</p>}
        </div>
        <span className={`badge ${project.status === 'active' ? 'badge-success' : ''}`}>
          {project.status}
        </span>
      </div>

      <div className="project-sections">
        <section className="section">
          <div className="section-header">
            <h2>API Keys</h2>
            <button onClick={() => setShowKeyModal(true)} className="btn btn-secondary">
              Generate Key
            </button>
          </div>
          {project.api_keys.length === 0 ? (
            <p className="empty-text">No API keys yet</p>
          ) : (
            <div className="api-keys-list">
              {project.api_keys.map((key: any) => (
                <div key={key.id} className="api-key-item">
                  <div className="api-key-info">
                    <code className="api-key-value">{key.key}</code>
                    {key.name && <span className="api-key-name">{key.name}</span>}
                  </div>
                  <span className={`badge ${key.is_active ? 'badge-success' : ''}`}>
                    {key.is_active ? 'Active' : 'Inactive'}
                  </span>
                </div>
              ))}
            </div>
          )}
        </section>

        <section className="section">
          <div className="section-header">
            <h2>Reviews</h2>
            <button onClick={() => setShowReviewModal(true)} className="btn btn-primary">
              Write Review
            </button>
          </div>
          {reviews.length === 0 ? (
            <p className="empty-text">No reviews yet</p>
          ) : (
            <div className="reviews-list">
              {reviews.map((review) => (
                <div key={review.id} className="review-item">
                  <div className="review-header">
                    <span className="review-author">{review.user.full_name || review.user.email}</span>
                    <div className="review-rating">
                      {Array.from({ length: 5 }).map((_, i) => (
                        <span key={i} className={i < review.rating ? 'star filled' : 'star'}>★</span>
                      ))}
                    </div>
                  </div>
                  <p className="review-content">{review.content}</p>
                  <span className="review-date">{new Date(review.created_at).toLocaleDateString()}</span>
                </div>
              ))}
            </div>
          )}
        </section>
      </div>

      {showKeyModal && (
        <div className="modal-overlay" onClick={() => setShowKeyModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h2 className="modal-title">Generate API Key</h2>
            <form onSubmit={handleCreateApiKey} className="modal-form">
              <div className="form-group">
                <label>Key Name</label>
                <input type="text" name="name" className="input" placeholder="Production Key" />
              </div>
              <div className="modal-actions">
                <button type="button" onClick={() => setShowKeyModal(false)} className="btn btn-secondary">Cancel</button>
                <button type="submit" className="btn btn-primary">Generate</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {showReviewModal && (
        <div className="modal-overlay" onClick={() => setShowReviewModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h2 className="modal-title">Write Review</h2>
            <form onSubmit={handleCreateReview} className="modal-form">
              <div className="form-group">
                <label>Rating</label>
                <select name="rating" className="input">
                  <option value="5">5 - Excellent</option>
                  <option value="4">4 - Good</option>
                  <option value="3">3 - Average</option>
                  <option value="2">2 - Poor</option>
                  <option value="1">1 - Terrible</option>
                </select>
              </div>
              <div className="form-group">
                <label>Review</label>
                <textarea name="content" className="input" rows={4} placeholder="Share your experience..." required />
              </div>
              <div className="modal-actions">
                <button type="button" onClick={() => setShowReviewModal(false)} className="btn btn-secondary">Cancel</button>
                <button type="submit" className="btn btn-primary">Submit</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}