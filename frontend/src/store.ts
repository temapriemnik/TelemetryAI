import { create } from 'zustand';
import { persist } from 'zustand/middleware';


interface User {
  id: number;
  email: string;
  full_name: string | null;
  avatar_url: string | null;
  bio: string | null;
  role: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

interface Project {
  id: number;
  name: string;
  description: string | null;
  webhook_url: string | null;
  status: string;
  owner_id: number;
  created_at: string;
  updated_at: string;
  api_keys: APIKey[];
}

interface APIKey {
  id: number;
  key: string;
  name: string | null;
  is_active: boolean;
  created_at: string;
}

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  setUser: (user: User | null) => void;
  setTokens: (access: string, refresh: string) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      setUser: (user) => set({ user, isAuthenticated: !!user }),
      setTokens: (accessToken, refreshToken) => set({ accessToken, refreshToken }),
      logout: () => set({ user: null, accessToken: null, refreshToken: null, isAuthenticated: false }),
    }),
    { name: 'auth-storage' }
  )
);

interface ProjectState {
  projects: Project[];
  currentProject: Project | null;
  setProjects: (projects: Project[]) => void;
  setCurrentProject: (project: Project | null) => void;
  addProject: (project: Project) => void;
  updateProject: (id: number, data: Partial<Project>) => void;
  removeProject: (id: number) => void;
}

export const useProjectStore = create<ProjectState>()((set) => ({
  projects: [],
  currentProject: null,
  setProjects: (projects) => set({ projects }),
  setCurrentProject: (currentProject) => set({ currentProject }),
  addProject: (project) => set((state) => ({ projects: [...state.projects, project] })),
  updateProject: (id, data) => set((state) => ({
    projects: state.projects.map((p) => p.id === id ? { ...p, ...data } : p),
  })),
  removeProject: (id) => set((state) => ({
    projects: state.projects.filter((p) => p.id !== id),
  })),
}));