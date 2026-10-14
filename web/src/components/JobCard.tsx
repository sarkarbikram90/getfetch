'use client';

import React, { useState } from 'react';
import { ExternalLink, Bookmark, CheckCircle, MapPin, Building, Clock, ChevronDown } from 'lucide-react';
import { Job } from '../types/job';

interface JobCardProps {
  job: Job;
  isSaved?: boolean;
  savedStatus?: string;
  onSave?: (id: number, status: string, notes?: string) => Promise<void>;
  onUnsave?: (id: number) => Promise<void>;
}

export const JobCard: React.FC<JobCardProps> = ({
  job,
  isSaved = false,
  savedStatus = '',
  onSave,
  onUnsave,
}) => {
  const [loading, setLoading] = useState(false);
  const [showStatusMenu, setShowStatusMenu] = useState(false);

  const handleStatusChange = async (status: string) => {
    if (!onSave) return;
    setLoading(true);
    try {
      await onSave(job.id, status);
    } finally {
      setLoading(false);
      setShowStatusMenu(false);
    }
  };

  const handleToggleBookmark = async () => {
    setLoading(true);
    try {
      if (isSaved && onUnsave) {
        await onUnsave(job.id);
      } else if (onSave) {
        await onSave(job.id, 'saved');
      }
    } finally {
      setLoading(false);
    }
  };

  const formatRelativeTime = (dateStr?: string) => {
    if (!dateStr) return 'Recently';
    const date = new Date(dateStr);
    const diffMs = Date.now() - date.getTime();
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    if (diffHours < 1) return 'Just now';
    if (diffHours < 24) return `${diffHours}h ago`;
    const diffDays = Math.floor(diffHours / 24);
    if (diffDays === 1) return 'Yesterday';
    return `${diffDays}d ago`;
  };

  return (
    <div className="bg-slate-900/50 border border-slate-800/80 hover:border-slate-700/80 transition-all rounded-xl p-5 mb-4 shadow-sm hover:shadow-md">
      {/* Top Meta Header */}
      <div className="flex items-start justify-between gap-4 mb-2.5">
        <div>
          <div className="flex items-center space-x-2 text-xs text-slate-400 mb-1">
            <span className="font-semibold text-slate-200 flex items-center space-x-1">
              <Building className="h-3.5 w-3.5 text-slate-400" />
              <span>{job.company_name}</span>
            </span>
            <span>&bull;</span>
            <span className="flex items-center space-x-1 text-slate-400">
              <Clock className="h-3 w-3" />
              <span>{formatRelativeTime(job.first_seen_at)}</span>
            </span>
          </div>

          <h3 className="text-base sm:text-lg font-bold text-white group-hover:text-sky-400 transition-colors">
            <a
              href={job.canonical_url}
              target="_blank"
              rel="noopener noreferrer"
              className="hover:underline flex items-center gap-1.5"
            >
              <span>{job.title}</span>
              <ExternalLink className="h-3.5 w-3.5 text-slate-500 inline-block" />
            </a>
          </h3>
        </div>

        {/* Bookmark & Tracker Actions */}
        <div className="relative flex items-center space-x-1.5 shrink-0">
          <button
            onClick={handleToggleBookmark}
            disabled={loading}
            className={`p-2 rounded-lg border transition-colors ${
              isSaved
                ? 'bg-amber-500/10 border-amber-500/40 text-amber-400'
                : 'bg-slate-950/60 border-slate-800 text-slate-400 hover:text-white hover:border-slate-700'
            }`}
            title={isSaved ? 'Remove from Tracker' : 'Save to Tracker'}
          >
            <Bookmark className={`h-4 w-4 ${isSaved ? 'fill-amber-400' : ''}`} />
          </button>

          {isSaved && (
            <div className="relative">
              <button
                onClick={() => setShowStatusMenu(!showStatusMenu)}
                className="px-2.5 py-1.5 rounded-lg text-xs font-semibold bg-slate-800 border border-slate-700 text-sky-300 flex items-center space-x-1"
              >
                <span className="capitalize">{savedStatus || 'Saved'}</span>
                <ChevronDown className="h-3 w-3" />
              </button>

              {showStatusMenu && (
                <div className="absolute right-0 mt-1 w-36 bg-slate-900 border border-slate-700 rounded-lg shadow-xl z-20 py-1 text-xs">
                  {['saved', 'applied', 'interviewing', 'rejected'].map((st) => (
                    <button
                      key={st}
                      onClick={() => handleStatusChange(st)}
                      className={`w-full text-left px-3 py-1.5 hover:bg-slate-800 capitalize flex items-center justify-between ${
                        savedStatus === st ? 'text-sky-400 font-semibold' : 'text-slate-300'
                      }`}
                    >
                      <span>{st}</span>
                      {savedStatus === st && <CheckCircle className="h-3 w-3" />}
                    </button>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Location & Remote Tags */}
      <div className="flex flex-wrap items-center gap-2 mb-3.5 text-xs text-slate-400">
        <span className="flex items-center space-x-1">
          <MapPin className="h-3.5 w-3.5 text-slate-500" />
          <span>{job.location || 'Location Not Specified'}</span>
        </span>
        {job.remote_type && job.remote_type !== 'unknown' && (
          <span className="px-2 py-0.5 rounded-full text-[11px] font-medium bg-purple-500/10 border border-purple-500/30 text-purple-300 capitalize">
            {job.remote_type}
          </span>
        )}
        {job.country && (
          <span className="px-2 py-0.5 rounded-full text-[11px] font-medium bg-slate-800 text-slate-300">
            {job.country}
          </span>
        )}
      </div>

      {/* Matched Skill Pills */}
      {job.matched_skills && job.matched_skills.length > 0 && (
        <div className="flex flex-wrap gap-1.5 mb-3.5">
          {job.matched_skills.map((skill) => (
            <span
              key={skill}
              className={`text-xs px-2.5 py-0.5 rounded font-medium ${
                skill === 'Kubernetes'
                  ? 'bg-sky-500/10 text-sky-300 border border-sky-500/30 font-semibold'
                  : 'bg-slate-800/80 text-slate-300 border border-slate-700/60'
              }`}
            >
              {skill}
            </span>
          ))}
        </div>
      )}

      {/* Snippet Proof Highlight */}
      {job.snippet_match && (
        <div className="bg-slate-950/70 border-l-2 border-sky-500 px-3.5 py-2.5 rounded-r-lg text-xs text-slate-300 font-mono italic mb-3.5 leading-relaxed">
          {job.snippet_match}
        </div>
      )}

      {/* Card Footer */}
      <div className="flex items-center justify-between pt-2 border-t border-slate-800/60">
        <span className="text-[11px] text-slate-500">
          Source ID: <span className="font-mono text-slate-400">{job.external_id}</span>
        </span>
        <a
          href={job.canonical_url}
          target="_blank"
          rel="noopener noreferrer"
          className="text-xs font-semibold text-sky-400 hover:text-sky-300 flex items-center space-x-1"
        >
          <span>Open Job Listing</span>
          <ExternalLink className="h-3.5 w-3.5" />
        </a>
      </div>
    </div>
  );
};
