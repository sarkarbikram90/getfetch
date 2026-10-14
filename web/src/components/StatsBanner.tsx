'use client';

import React from 'react';
import { Sparkles, ShieldCheck, Layers, Cpu } from 'lucide-react';
import { DashboardStats } from '../types/job';

interface StatsBannerProps {
  stats: DashboardStats | null;
  onSkillClick?: (skill: string) => void;
  selectedSkills?: string[];
}

export const StatsBanner: React.FC<StatsBannerProps> = ({ stats, onSkillClick, selectedSkills = [] }) => {
  if (!stats) return null;

  return (
    <div className="mb-8">
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        {/* Total Active Jobs */}
        <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-4 flex items-center space-x-3.5">
          <div className="h-10 w-10 rounded-lg bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-400">
            <ShieldCheck className="h-5 w-5" />
          </div>
          <div>
            <div className="text-2xl font-bold text-white tracking-tight">{stats.total_active_jobs}</div>
            <div className="text-xs text-slate-400">Active Roles</div>
          </div>
        </div>

        {/* Discovered Last 24h */}
        <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-4 flex items-center space-x-3.5">
          <div className="h-10 w-10 rounded-lg bg-sky-500/10 border border-sky-500/20 flex items-center justify-center text-sky-400">
            <Sparkles className="h-5 w-5" />
          </div>
          <div>
            <div className="text-2xl font-bold text-sky-300 tracking-tight">{stats.new_jobs_last_24h}</div>
            <div className="text-xs text-slate-400">Discovered Today</div>
          </div>
        </div>

        {/* Monitored Boards */}
        <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-4 flex items-center space-x-3.5">
          <div className="h-10 w-10 rounded-lg bg-purple-500/10 border border-purple-500/20 flex items-center justify-center text-purple-400">
            <Layers className="h-5 w-5" />
          </div>
          <div>
            <div className="text-2xl font-bold text-purple-300 tracking-tight">{stats.total_sources_active}</div>
            <div className="text-xs text-slate-400">Company Boards</div>
          </div>
        </div>

        {/* Skills Tracked */}
        <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-4 flex items-center space-x-3.5">
          <div className="h-10 w-10 rounded-lg bg-amber-500/10 border border-amber-500/20 flex items-center justify-center text-amber-400">
            <Cpu className="h-5 w-5" />
          </div>
          <div>
            <div className="text-2xl font-bold text-amber-300 tracking-tight">
              {Object.keys(stats.top_skills || {}).length}
            </div>
            <div className="text-xs text-slate-400">Cloud Skills Tracked</div>
          </div>
        </div>
      </div>

      {/* Top Skills Quick Filter Bar */}
      {stats.top_skills && Object.keys(stats.top_skills).length > 0 && (
        <div className="flex flex-wrap items-center gap-2 pt-2 pb-1 border-t border-slate-800/60">
          <span className="text-xs font-semibold text-slate-500 uppercase tracking-wider mr-1">Trending Skills:</span>
          {Object.entries(stats.top_skills).slice(0, 8).map(([skill, count]) => {
            const isSelected = selectedSkills.includes(skill);
            return (
              <button
                key={skill}
                onClick={() => onSkillClick && onSkillClick(skill)}
                className={`text-xs px-2.5 py-1 rounded-full font-medium transition-all flex items-center space-x-1.5 ${
                  isSelected
                    ? 'bg-sky-500 text-slate-950 font-semibold shadow-sm'
                    : 'bg-slate-800/80 text-slate-300 hover:bg-slate-700 border border-slate-700/60'
                }`}
              >
                <span>{skill}</span>
                <span className={`text-[10px] ${isSelected ? 'text-slate-900' : 'text-slate-500'}`}>
                  {count}
                </span>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
};
