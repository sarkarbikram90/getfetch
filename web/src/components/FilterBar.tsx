'use client';

import React from 'react';
import { Search, X, MapPin, Globe } from 'lucide-react';

interface FilterBarProps {
  searchQuery: string;
  setSearchQuery: (query: string) => void;
  selectedSkills: string[];
  toggleSkill: (skill: string) => void;
  selectedRoles: string[];
  toggleRole: (role: string) => void;
  remoteType: string;
  setRemoteType: (remote: string) => void;
  indiaOnly: boolean;
  setIndiaOnly: (val: boolean) => void;
  onClearAll: () => void;
}

const COMMON_SKILLS = [
  'Kubernetes',
  'Helm',
  'Terraform',
  'Go',
  'Docker',
  'Prometheus',
  'Grafana',
  'Cilium',
  'eBPF',
  'GCP',
  'AWS',
  'Azure',
];

const COMMON_ROLES = [
  'Site Reliability Engineer',
  'Platform Engineer',
  'DevOps Engineer',
  'Cloud Engineer',
  'Infrastructure Engineer',
];

export const FilterBar: React.FC<FilterBarProps> = ({
  searchQuery,
  setSearchQuery,
  selectedSkills,
  toggleSkill,
  selectedRoles,
  toggleRole,
  remoteType,
  setRemoteType,
  indiaOnly,
  setIndiaOnly,
  onClearAll,
}) => {
  const hasActiveFilters =
    searchQuery ||
    selectedSkills.length > 0 ||
    selectedRoles.length > 0 ||
    remoteType !== '' ||
    indiaOnly;

  return (
    <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-5 mb-8 shadow-sm">
      {/* Search Bar & Quick Location Toggles */}
      <div className="flex flex-col sm:flex-row gap-3 items-stretch sm:items-center justify-between mb-4">
        <div className="relative flex-1">
          <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search by role, company, or keyword (e.g. 'Senior SRE', 'Platform', 'GitLab')..."
            className="w-full pl-10 pr-4 py-2.5 bg-slate-950 border border-slate-700/80 rounded-lg text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500"
          />
          {searchQuery && (
            <button
              onClick={() => setSearchQuery('')}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300"
            >
              <X className="h-4 w-4" />
            </button>
          )}
        </div>

        {/* Work mode & India quick filters */}
        <div className="flex items-center gap-2">
          {/* Work mode selector */}
          <select
            value={remoteType}
            onChange={(e) => setRemoteType(e.target.value)}
            className="bg-slate-950 border border-slate-700/80 text-slate-300 text-xs rounded-lg px-3 py-2.5 focus:outline-none focus:border-sky-500"
          >
            <option value="">All Work Modes</option>
            <option value="remote">Remote Only</option>
            <option value="hybrid">Hybrid</option>
            <option value="onsite">On-Site</option>
          </select>

          {/* India toggle */}
          <button
            onClick={() => setIndiaOnly(!indiaOnly)}
            className={`px-3 py-2 rounded-lg text-xs font-medium flex items-center space-x-1.5 transition-colors border ${
              indiaOnly
                ? 'bg-amber-500/10 border-amber-500/40 text-amber-300 font-semibold'
                : 'bg-slate-950 border-slate-700/80 text-slate-400 hover:text-slate-200'
            }`}
          >
            <MapPin className="h-3.5 w-3.5 text-amber-400" />
            <span>India / Remote</span>
          </button>

          {hasActiveFilters && (
            <button
              onClick={onClearAll}
              className="px-2.5 py-2 text-xs text-rose-400 hover:text-rose-300 hover:bg-rose-500/10 rounded-lg transition-colors flex items-center space-x-1"
              title="Reset all filters"
            >
              <X className="h-3.5 w-3.5" />
              <span>Clear</span>
            </button>
          )}
        </div>
      </div>

      {/* Role Selector Pills */}
      <div className="mb-3">
        <div className="text-[11px] uppercase tracking-wider font-semibold text-slate-400 mb-2">Target Roles</div>
        <div className="flex flex-wrap gap-1.5">
          {COMMON_ROLES.map((role) => {
            const isSelected = selectedRoles.includes(role);
            return (
              <button
                key={role}
                onClick={() => toggleRole(role)}
                className={`text-xs px-2.5 py-1 rounded-md font-medium transition-all ${
                  isSelected
                    ? 'bg-sky-500/20 border border-sky-400 text-sky-200'
                    : 'bg-slate-950/60 border border-slate-800 text-slate-400 hover:text-slate-200 hover:border-slate-700'
                }`}
              >
                {role}
              </button>
            );
          })}
        </div>
      </div>

      {/* Skills Filter Pills */}
      <div>
        <div className="text-[11px] uppercase tracking-wider font-semibold text-slate-400 mb-2">Required Skills</div>
        <div className="flex flex-wrap gap-1.5">
          {COMMON_SKILLS.map((skill) => {
            const isSelected = selectedSkills.includes(skill);
            return (
              <button
                key={skill}
                onClick={() => toggleSkill(skill)}
                className={`text-xs px-2.5 py-1 rounded-md font-medium transition-all ${
                  isSelected
                    ? 'bg-emerald-500/20 border border-emerald-400 text-emerald-200'
                    : 'bg-slate-950/60 border border-slate-800 text-slate-400 hover:text-slate-200 hover:border-slate-700'
                }`}
              >
                {skill}
              </button>
            );
          })}
        </div>
      </div>
    </div>
  );
};
