import { useState } from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
import { usePhases } from '../api/hooks';
import { PHASE_NAMES, PHASE_COLORS } from '../lib/constants';

export function Sidebar() {
  const { data: phases } = usePhases();
  const [search, setSearch] = useState('');
  const navigate = useNavigate();

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    const q = search.trim();
    if (q) {
      navigate(`/search?q=${encodeURIComponent(q)}`);
    }
  };

  return (
    <aside className="w-60 shrink-0 border-r border-gray-700 bg-gray-800 h-screen overflow-y-auto">
      <div className="p-4">
        <h1 className="text-lg font-bold text-gray-100">YT Automation</h1>
      </div>
      <div className="px-3 mb-2">
        <form onSubmit={handleSearch}>
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search videos..."
            className="w-full px-2 py-1.5 text-sm bg-gray-900 border border-gray-600 rounded text-gray-100 placeholder-gray-500 focus:outline-none focus:border-blue-500"
          />
        </form>
      </div>
      <nav className="px-2 pb-4">
        <NavLink
          to="/"
          end
          className={({ isActive }) =>
            `block px-3 py-2 rounded text-sm font-medium ${isActive ? 'bg-gray-700 text-gray-100' : 'text-gray-400 hover:bg-gray-700'}`
          }
        >
          Dashboard
        </NavLink>
        {phases && (
          <div className="mt-4">
            <div className="px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1">
              Phases
            </div>
            {phases.map((p) => (
              <NavLink
                key={p.id}
                to={`/phases/${p.id}`}
                className={({ isActive }) =>
                  `flex items-center gap-2 px-3 py-1.5 rounded text-sm ${isActive ? 'bg-gray-700 text-gray-100' : 'text-gray-400 hover:bg-gray-700'}`
                }
              >
                <span
                  className={`w-2 h-2 rounded-full ${PHASE_COLORS[p.id] ?? 'bg-gray-500'}`}
                />
                <span className="flex-1">
                  {PHASE_NAMES[p.id] ?? `Phase ${p.id}`}
                </span>
                <span className="text-xs text-gray-500">{p.count}</span>
              </NavLink>
            ))}
          </div>
        )}
        <div className="mt-4">
          <div className="px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1">
            Analyze
          </div>
          <NavLink
            to="/analyze/titles"
            className={({ isActive }) =>
              `flex items-center gap-2 px-3 py-1.5 rounded text-sm ${isActive ? 'bg-gray-700 text-gray-100' : 'text-gray-400 hover:bg-gray-700'}`
            }
          >
            <span className="w-2 h-2 rounded-full bg-purple-500" />
            <span className="flex-1">Titles</span>
          </NavLink>
          <NavLink
            to="/analyze/timing"
            className={({ isActive }) =>
              `flex items-center gap-2 px-3 py-1.5 rounded text-sm ${isActive ? 'bg-gray-700 text-gray-100' : 'text-gray-400 hover:bg-gray-700'}`
            }
          >
            <span className="w-2 h-2 rounded-full bg-teal-500" />
            <span className="flex-1">Timing</span>
          </NavLink>
        </div>
        <div className="mt-4">
          <div className="px-3 text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1">
            Tools
          </div>
          <NavLink
            to="/ama"
            className={({ isActive }) =>
              `flex items-center gap-2 px-3 py-1.5 rounded text-sm ${isActive ? 'bg-gray-700 text-gray-100' : 'text-gray-400 hover:bg-gray-700'}`
            }
          >
            <span className="w-2 h-2 rounded-full bg-orange-500" />
            <span className="flex-1">Ask Me Anything</span>
          </NavLink>
        </div>
      </nav>
    </aside>
  );
}
