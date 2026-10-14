import { DashboardStats, Job, SavedJobItem, Source } from '../types/job';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export interface JobFilterParams {
  q?: string;
  skills?: string[];
  roles?: string[];
  remote?: string;
  country?: string;
  limit?: number;
  offset?: number;
}

export async function fetchJobs(params: JobFilterParams = {}): Promise<{ jobs: Job[]; total: number }> {
  const query = new URLSearchParams();
  if (params.q) query.set('q', params.q);
  if (params.skills && params.skills.length > 0) query.set('skills', params.skills.join(','));
  if (params.roles && params.roles.length > 0) query.set('roles', params.roles.join(','));
  if (params.remote) query.set('remote', params.remote);
  if (params.country) query.set('country', params.country);
  if (params.limit) query.set('limit', params.limit.toString());
  if (params.offset) query.set('offset', params.offset.toString());

  const res = await fetch(`${API_BASE_URL}/jobs?${query.toString()}`, {
    cache: 'no-store',
  });

  if (!res.ok) {
    throw new Error(`Failed to fetch jobs: HTTP ${res.status}`);
  }

  const data = await res.json();
  return {
    jobs: data.jobs || [],
    total: data.total || 0,
  };
}

export async function fetchJobByID(id: number): Promise<Job> {
  const res = await fetch(`${API_BASE_URL}/jobs/${id}`, { cache: 'no-store' });
  if (!res.ok) throw new Error(`Failed to fetch job #${id}`);
  const data = await res.json();
  return data.job;
}

export async function saveJob(id: number, status: string = 'saved', notes: string = ''): Promise<void> {
  const res = await fetch(`${API_BASE_URL}/jobs/${id}/save`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status, notes }),
  });
  if (!res.ok) throw new Error(`Failed to save job #${id}`);
}

export async function unsaveJob(id: number): Promise<void> {
  const res = await fetch(`${API_BASE_URL}/jobs/${id}/save`, {
    method: 'DELETE',
  });
  if (!res.ok) throw new Error(`Failed to unsave job #${id}`);
}

export async function fetchSavedJobs(): Promise<SavedJobItem[]> {
  const res = await fetch(`${API_BASE_URL}/saved`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch saved jobs');
  const data = await res.json();
  return data.saved_jobs || [];
}

export async function fetchSources(): Promise<Source[]> {
  const res = await fetch(`${API_BASE_URL}/sources`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch sources');
  const data = await res.json();
  return data.sources || [];
}

export async function fetchStats(): Promise<DashboardStats> {
  const res = await fetch(`${API_BASE_URL}/stats`, { cache: 'no-store' });
  if (!res.ok) throw new Error('Failed to fetch stats');
  const data = await res.json();
  return data.stats;
}
