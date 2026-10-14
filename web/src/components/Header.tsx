'use client';

import React from 'react';
import { Compass, BookmarkCheck, Building2, Terminal } from 'lucide-react';

interface HeaderProps {
  activeTab: 'explore' | 'tracker' | 'sources';
  setActiveTab: (tab: 'explore' | 'tracker' | 'sources') => void;
  savedCount: number;
}

export const Header: React.FC<HeaderProps> = ({ activeTab, setActiveTab, savedCount }) => {
  return (
    <header className="sticky top-0 z-50 border-b border-slate-800 bg-slate-950/80 backdrop-blur-md">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        {/* Brand */}
        <div className="flex items-center space-x-3 cursor-pointer" onClick={() => setActiveTab('explore')}>
          <div className="h-9 w-9 rounded-lg bg-sky-500/10 border border-sky-500/30 flex items-center justify-center text-sky-400">
            <Terminal className="h-5 w-5" />
          </div>
          <div>
            <div className="flex items-center space-x-2">
              <span className="font-bold text-lg text-white tracking-tight">KubeJobs</span>
              <span className="text-[10px] uppercase font-semibold px-1.5 py-0.5 rounded bg-sky-500/20 text-sky-300 border border-sky-500/30">
                getfetch
              </span>
            </div>
            <p className="text-xs text-slate-400">Kubernetes & Platform Engineering Intelligence</p>
          </div>
        </div>

        {/* Navigation Tabs */}
        <nav className="flex items-center space-x-1 sm:space-x-2">
          <button
            onClick={() => setActiveTab('explore')}
            className={`px-3 py-1.5 rounded-md text-sm font-medium flex items-center space-x-2 transition-colors ${
              activeTab === 'explore'
                ? 'bg-slate-800 text-sky-400 border border-slate-700'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-900'
            }`}
          >
            <Compass className="h-4 w-4" />
            <span>Explore Roles</span>
          </button>

          <button
            onClick={() => setActiveTab('tracker')}
            className={`px-3 py-1.5 rounded-md text-sm font-medium flex items-center space-x-2 transition-colors relative ${
              activeTab === 'tracker'
                ? 'bg-slate-800 text-sky-400 border border-slate-700'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-900'
            }`}
          >
            <BookmarkCheck className="h-4 w-4" />
            <span>Tracker</span>
            {savedCount > 0 && (
              <span className="ml-1.5 px-1.5 py-0.2 rounded-full text-xs font-semibold bg-sky-500 text-slate-950">
                {savedCount}
              </span>
            )}
          </button>

          <button
            onClick={() => setActiveTab('sources')}
            className={`px-3 py-1.5 rounded-md text-sm font-medium flex items-center space-x-2 transition-colors ${
              activeTab === 'sources'
                ? 'bg-slate-800 text-sky-400 border border-slate-700'
                : 'text-slate-400 hover:text-slate-200 hover:bg-slate-900'
            }`}
          >
            <Building2 className="h-4 w-4" />
            <span>Monitored Sources</span>
          </button>
        </nav>

        {/* External links */}
        <div className="flex items-center space-x-3">
          <a
            href="https://github.com/sarkarbikram90/getfetch"
            target="_blank"
            rel="noopener noreferrer"
            className="text-slate-400 hover:text-white p-2 rounded-lg hover:bg-slate-900 transition-colors"
            title="GitHub Repository"
          >
            <svg className="h-5 w-5 fill-current" viewBox="0 0 24 24">
              <path fillRule="evenodd" clipRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
            </svg>
          </a>
        </div>
      </div>
    </header>
  );
};
