export interface Job {
  id: number;
  source_id: number;
  external_id: string;
  canonical_url: string;
  company_name: string;
  title: string;
  location: string;
  remote_type: 'remote' | 'hybrid' | 'onsite' | 'unknown';
  country: string;
  department: string;
  description_text: string;
  matched_skills: string[];
  matched_roles: string[];
  snippet_match: string;
  content_hash: string;
  posted_at?: string;
  first_seen_at: string;
  last_seen_at: string;
  status: string;
}

export interface SavedJobItem {
  id: number;
  job_id: number;
  status: 'saved' | 'applied' | 'interviewing' | 'rejected' | 'archived';
  notes: string;
  applied_at?: string;
  created_at: string;
  updated_at: string;
  job: Job;
}

export interface Source {
  id: number;
  source_type: 'greenhouse' | 'lever' | 'ashby' | string;
  company_name: string;
  board_identifier: string;
  source_url: string;
  enabled: boolean;
  crawl_interval_minutes: number;
  last_success_at?: string;
  last_error?: string;
}

export interface DashboardStats {
  total_active_jobs: number;
  new_jobs_last_24h: number;
  total_sources_active: number;
  top_skills: Record<string, number>;
}
