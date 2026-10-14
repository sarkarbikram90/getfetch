'use client';

import React, { useState } from 'react';
import { BookmarkCheck, ExternalLink, Trash2, CheckCircle2 } from 'lucide-react';
import { SavedJobItem } from '../types/job';

interface TrackerViewProps {
  savedJobs: SavedJobItem[];
  onUpdateStatus: (jobId: number, status: string, notes?: string) => Promise<void>;
  onUnsave: (jobId: number) => Promise<void>;
}

export const TrackerView: React.FC<TrackerViewProps> = ({
  savedJobs,
  onUpdateStatus,
  onUnsave,
}) => {
  const [activeFilter, setActiveFilter] = useState<string>('all');

  const filteredJobs = savedJobs.filter((item) => {
    if (activeFilter === 'all') return true;
    return item.status === activeFilter;
  });

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'applied':
        return 'bg-blue-500/20 text-blue-300 border-blue-500/30';
      case 'interviewing':
        return 'bg-amber-500/20 text-amber-300 border-amber-500/30';
      case 'rejected':
        return 'bg-rose-500/20 text-rose-300 border-rose-500/30';
      default:
        return 'bg-slate-800 text-slate-300 border-slate-700';
    }
  };

  return (
    <div className="mb-12">
      {/* Subheader & Status Tabs */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-4 border-b border-slate-800 gap-4 mb-6">
        <div>
          <h2 className="text-xl font-bold text-white flex items-center space-x-2">
            <BookmarkCheck className="h-5 w-5 text-sky-400" />
            <span>My Applications Tracker</span>
          </h2>
          <p className="text-xs text-slate-400 mt-0.5">
            Manage your personal shortlist, application progress, and recruiter status.
          </p>
        </div>

        {/* Filter by Status */}
        <div className="flex items-center space-x-1.5 bg-slate-900 border border-slate-800 p-1 rounded-lg">
          {['all', 'saved', 'applied', 'interviewing', 'rejected'].map((st) => (
            <button
              key={st}
              onClick={() => setActiveFilter(st)}
              className={`px-3 py-1 rounded-md text-xs font-medium capitalize transition-colors ${
                activeFilter === st
                  ? 'bg-sky-500 text-slate-950 font-semibold'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              {st}
            </button>
          ))}
        </div>
      </div>

      {filteredJobs.length === 0 ? (
        <div className="text-center py-16 bg-slate-900/30 border border-slate-800/80 rounded-xl">
          <BookmarkCheck className="h-10 w-10 text-slate-600 mx-auto mb-3" />
          <h3 className="text-base font-semibold text-slate-300">No tracked roles in this view</h3>
          <p className="text-xs text-slate-500 max-w-sm mx-auto mt-1">
            Browse listings in the Explore tab and click the bookmark button on roles you want to apply to or track.
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          {filteredJobs.map((item) => (
            <div
              key={item.id}
              className="bg-slate-900/60 border border-slate-800 rounded-xl p-5 flex flex-col md:flex-row md:items-center justify-between gap-4"
            >
              {/* Job Info */}
              <div className="flex-1">
                <div className="flex items-center space-x-2 mb-1">
                  <span className="text-xs font-semibold text-sky-400">{item.job.company_name}</span>
                  <span className="text-xs text-slate-500">&bull;</span>
                  <span className="text-xs text-slate-400">{item.job.location}</span>
                  <span
                    className={`text-[10px] uppercase font-bold px-2 py-0.5 rounded-full border ${getStatusBadge(
                      item.status
                    )}`}
                  >
                    {item.status}
                  </span>
                </div>

                <h4 className="text-base font-bold text-white mb-2">
                  <a
                    href={item.job.canonical_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="hover:underline hover:text-sky-300"
                  >
                    {item.job.title}
                  </a>
                </h4>

                <div className="flex flex-wrap gap-1.5">
                  {item.job.matched_skills.slice(0, 5).map((s) => (
                    <span
                      key={s}
                      className="text-[11px] px-2 py-0.5 rounded bg-slate-800/80 border border-slate-700/60 text-slate-300"
                    >
                      {s}
                    </span>
                  ))}
                </div>
              </div>

              {/* Status Switcher & Actions */}
              <div className="flex items-center space-x-3 shrink-0">
                <select
                  value={item.status}
                  onChange={(e) => onUpdateStatus(item.job_id, e.target.value)}
                  className="bg-slate-950 border border-slate-700 text-xs text-slate-200 rounded-lg px-2.5 py-1.5 focus:outline-none focus:border-sky-500"
                >
                  <option value="saved">Saved</option>
                  <option value="applied">Applied</option>
                  <option value="interviewing">Interviewing</option>
                  <option value="rejected">Rejected</option>
                </select>

                <a
                  href={item.job.canonical_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="p-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700"
                  title="Open Job Posting"
                >
                  <ExternalLink className="h-4 w-4" />
                </a>

                <button
                  onClick={() => onUnsave(item.job_id)}
                  className="p-2 rounded-lg bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 border border-rose-500/30"
                  title="Remove from Tracker"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
