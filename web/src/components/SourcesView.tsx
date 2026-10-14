'use client';

import React from 'react';
import { Building2, CheckCircle2, AlertCircle, ExternalLink, RefreshCw } from 'lucide-react';
import { Source } from '../types/job';

interface SourcesViewProps {
  sources: Source[];
  loading?: boolean;
}

export const SourcesView: React.FC<SourcesViewProps> = ({ sources, loading = false }) => {
  return (
    <div className="mb-12">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-4 border-b border-slate-800 gap-4 mb-6">
        <div>
          <h2 className="text-xl font-bold text-white flex items-center space-x-2">
            <Building2 className="h-5 w-5 text-sky-400" />
            <span>Monitored Company Boards</span>
          </h2>
          <p className="text-xs text-slate-400 mt-0.5">
            Active ATS public API integrations polling Greenhouse, Lever, and Ashby job boards.
          </p>
        </div>

        <div className="text-xs text-slate-400 bg-slate-900 border border-slate-800 px-3 py-1.5 rounded-lg">
          <span className="font-semibold text-white">{sources.length}</span> Active Boards Polled
        </div>
      </div>

      <div className="bg-slate-900/60 border border-slate-800 rounded-xl overflow-hidden shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-slate-950/80 border-b border-slate-800 text-slate-400 font-semibold uppercase tracking-wider">
              <tr>
                <th className="py-3 px-4">Company</th>
                <th className="py-3 px-4">ATS Provider</th>
                <th className="py-3 px-4">Board Identifier</th>
                <th className="py-3 px-4">Health Status</th>
                <th className="py-3 px-4 text-right">Career Board</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60 text-slate-300">
              {sources.map((s) => (
                <tr key={s.id} className="hover:bg-slate-800/30 transition-colors">
                  <td className="py-3 px-4 font-semibold text-white flex items-center space-x-2">
                    <span>{s.company_name}</span>
                  </td>
                  <td className="py-3 px-4">
                    <span
                      className={`px-2 py-0.5 rounded text-[11px] font-medium uppercase tracking-wide border ${
                        s.source_type === 'greenhouse'
                          ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30'
                          : s.source_type === 'lever'
                          ? 'bg-sky-500/10 text-sky-400 border-sky-500/30'
                          : 'bg-purple-500/10 text-purple-400 border-purple-500/30'
                      }`}
                    >
                      {s.source_type}
                    </span>
                  </td>
                  <td className="py-3 px-4 font-mono text-slate-400">{s.board_identifier}</td>
                  <td className="py-3 px-4">
                    {s.last_error ? (
                      <span className="flex items-center space-x-1 text-rose-400">
                        <AlertCircle className="h-3.5 w-3.5" />
                        <span>Error</span>
                      </span>
                    ) : (
                      <span className="flex items-center space-x-1 text-emerald-400">
                        <CheckCircle2 className="h-3.5 w-3.5" />
                        <span>Healthy</span>
                      </span>
                    )}
                  </td>
                  <td className="py-3 px-4 text-right">
                    <a
                      href={s.source_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sky-400 hover:text-sky-300 inline-flex items-center space-x-1"
                    >
                      <span>Visit</span>
                      <ExternalLink className="h-3 w-3" />
                    </a>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
