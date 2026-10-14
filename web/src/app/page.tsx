'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { Header } from '../components/Header';
import { StatsBanner } from '../components/StatsBanner';
import { FilterBar } from '../components/FilterBar';
import { JobCard } from '../components/JobCard';
import { TrackerView } from '../components/TrackerView';
import { SourcesView } from '../components/SourcesView';
import {
  fetchJobs,
  fetchStats,
  fetchSavedJobs,
  fetchSources,
  saveJob,
  unsaveJob,
} from '../lib/api';
import { Job, DashboardStats, SavedJobItem, Source } from '../types/job';
import { Loader2, AlertCircle, RefreshCw } from 'lucide-react';

export default function Home() {
  const [activeTab, setActiveTab] = useState<'explore' | 'tracker' | 'sources'>('explore');

  // Data states
  const [jobs, setJobs] = useState<Job[]>([]);
  const [totalJobs, setTotalJobs] = useState<number>(0);
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [savedJobs, setSavedJobs] = useState<SavedJobItem[]>([]);
  const [sources, setSources] = useState<Source[]>([]);

  // Loading & error states
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  // Filter states
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [debouncedQuery, setDebouncedQuery] = useState<string>('');
  const [selectedSkills, setSelectedSkills] = useState<string[]>([]);
  const [selectedRoles, setSelectedRoles] = useState<string[]>([]);
  const [remoteType, setRemoteType] = useState<string>('');
  const [indiaOnly, setIndiaOnly] = useState<boolean>(false);
  const [page, setPage] = useState<number>(1);
  const pageSize = 20;

  // Debounce search query input
  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedQuery(searchQuery);
      setPage(1);
    }, 300);
    return () => clearTimeout(handler);
  }, [searchQuery]);

  // Load Dashboard Stats, Saved Jobs, and Sources initially
  const loadInitialData = useCallback(async () => {
    try {
      const [statsData, savedData, sourcesData] = await Promise.allSettled([
        fetchStats(),
        fetchSavedJobs(),
        fetchSources(),
      ]);

      if (statsData.status === 'fulfilled') setStats(statsData.value);
      if (savedData.status === 'fulfilled') setSavedJobs(savedData.value);
      if (sourcesData.status === 'fulfilled') setSources(sourcesData.value);
    } catch (err) {
      console.warn('Initial data load warning:', err);
    }
  }, []);

  useEffect(() => {
    loadInitialData();
  }, [loadInitialData]);

  // Load Jobs whenever filters or pagination change
  const loadJobs = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const offset = (page - 1) * pageSize;
      const res = await fetchJobs({
        q: debouncedQuery,
        skills: selectedSkills,
        roles: selectedRoles,
        remote: remoteType,
        country: indiaOnly ? 'India' : undefined,
        limit: pageSize,
        offset: offset,
      });
      setJobs(res.jobs);
      setTotalJobs(res.total);
    } catch (err: any) {
      console.error('Failed to load jobs:', err);
      setError(err.message || 'Failed to connect to getfetch backend API');
    } finally {
      setLoading(false);
    }
  }, [debouncedQuery, selectedSkills, selectedRoles, remoteType, indiaOnly, page]);

  useEffect(() => {
    loadJobs();
  }, [loadJobs]);

  // Toggle skill filter
  const toggleSkill = (skill: string) => {
    setSelectedSkills((prev) =>
      prev.includes(skill) ? prev.filter((s) => s !== skill) : [...prev, skill]
    );
    setPage(1);
  };

  // Toggle role filter
  const toggleRole = (role: string) => {
    setSelectedRoles((prev) =>
      prev.includes(role) ? prev.filter((r) => r !== role) : [...prev, role]
    );
    setPage(1);
  };

  const clearAllFilters = () => {
    setSearchQuery('');
    setDebouncedQuery('');
    setSelectedSkills([]);
    setSelectedRoles([]);
    setRemoteType('');
    setIndiaOnly(false);
    setPage(1);
  };

  // Bookmark / status handlers
  const handleSaveJob = async (id: number, status: string, notes: string = '') => {
    await saveJob(id, status, notes);
    const updatedSaved = await fetchSavedJobs();
    setSavedJobs(updatedSaved);
  };

  const handleUnsaveJob = async (id: number) => {
    await unsaveJob(id);
    const updatedSaved = await fetchSavedJobs();
    setSavedJobs(updatedSaved);
  };

  const savedJobMap = new Map(savedJobs.map((item) => [item.job_id, item.status]));

  const totalPages = Math.ceil(totalJobs / pageSize) || 1;

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col">
      {/* Header */}
      <Header
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        savedCount={savedJobs.length}
      />

      {/* Main Content Area */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Tracker View */}
        {activeTab === 'tracker' && (
          <TrackerView
            savedJobs={savedJobs}
            onUpdateStatus={handleSaveJob}
            onUnsave={handleUnsaveJob}
          />
        )}

        {/* Monitored Sources View */}
        {activeTab === 'sources' && (
          <SourcesView sources={sources} loading={loading} />
        )}

        {/* Explore Jobs View */}
        {activeTab === 'explore' && (
          <>
            {/* Stats Metrics Banner */}
            <StatsBanner
              stats={stats}
              onSkillClick={toggleSkill}
              selectedSkills={selectedSkills}
            />

            {/* Filter Bar */}
            <FilterBar
              searchQuery={searchQuery}
              setSearchQuery={setSearchQuery}
              selectedSkills={selectedSkills}
              toggleSkill={toggleSkill}
              selectedRoles={selectedRoles}
              toggleRole={toggleRole}
              remoteType={remoteType}
              setRemoteType={setRemoteType}
              indiaOnly={indiaOnly}
              setIndiaOnly={setIndiaOnly}
              onClearAll={clearAllFilters}
            />

            {/* Header above job listings */}
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-sm font-semibold text-slate-400">
                Found <span className="text-white font-bold">{totalJobs}</span> matching roles
              </h2>

              <button
                onClick={loadJobs}
                disabled={loading}
                className="text-xs text-slate-400 hover:text-sky-400 flex items-center space-x-1"
                title="Refresh listings"
              >
                <RefreshCw className={`h-3.5 w-3.5 ${loading ? 'animate-spin' : ''}`} />
                <span>Refresh</span>
              </button>
            </div>

            {/* Error Message */}
            {error && (
              <div className="bg-rose-950/40 border border-rose-800/80 rounded-xl p-4 mb-6 flex items-start space-x-3 text-rose-300">
                <AlertCircle className="h-5 w-5 shrink-0 text-rose-400 mt-0.5" />
                <div className="text-xs">
                  <div className="font-semibold text-sm text-rose-200">Unable to reach getfetch API</div>
                  <p className="mt-1">
                    Ensure the Go API backend is running (`go run ./cmd/api -port 8080`) or start the Docker environment.
                  </p>
                </div>
              </div>
            )}

            {/* Loading Indicator */}
            {loading && jobs.length === 0 && (
              <div className="flex flex-col items-center justify-center py-20 text-slate-500">
                <Loader2 className="h-8 w-8 animate-spin text-sky-400 mb-3" />
                <span className="text-sm font-medium">Scanning verified ATS listings...</span>
              </div>
            )}

            {/* Empty State */}
            {!loading && jobs.length === 0 && !error && (
              <div className="text-center py-16 bg-slate-900/30 border border-slate-800/80 rounded-xl">
                <div className="text-base font-semibold text-slate-300">No matching jobs found</div>
                <p className="text-xs text-slate-500 max-w-sm mx-auto mt-1 mb-4">
                  Try adjusting your skill filters, search query, or work mode.
                </p>
                <button
                  onClick={clearAllFilters}
                  className="px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-sky-500 text-slate-950 hover:bg-sky-400 transition-colors"
                >
                  Clear all filters
                </button>
              </div>
            )}

            {/* Job Listings Cards */}
            <div className="space-y-4">
              {jobs.map((job) => (
                <JobCard
                  key={job.id}
                  job={job}
                  isSaved={savedJobMap.has(job.id)}
                  savedStatus={savedJobMap.get(job.id)}
                  onSave={handleSaveJob}
                  onUnsave={handleUnsaveJob}
                />
              ))}
            </div>

            {/* Pagination Controls */}
            {totalPages > 1 && (
              <div className="flex items-center justify-center space-x-2 pt-8 pb-12">
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page === 1}
                  className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-slate-900 border border-slate-800 text-slate-300 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-slate-800 transition-colors"
                >
                  Previous
                </button>

                <span className="text-xs text-slate-400 px-3">
                  Page <span className="text-white font-bold">{page}</span> of{' '}
                  <span className="text-white font-bold">{totalPages}</span>
                </span>

                <button
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page === totalPages}
                  className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-slate-900 border border-slate-800 text-slate-300 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-slate-800 transition-colors"
                >
                  Next
                </button>
              </div>
            )}
          </>
        )}
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-900 bg-slate-950 py-6 text-center text-xs text-slate-600">
        <p>KubeJobs &bull; Powered by getfetch &bull; Direct ATS public ingestion</p>
      </footer>
    </div>
  );
}
