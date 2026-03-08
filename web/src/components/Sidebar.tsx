import { NavLink } from 'react-router-dom';
import { usePhases } from '../api/hooks';
import { PHASE_NAMES, PHASE_COLORS } from '../lib/constants';

export function Sidebar() {
  const { data: phases } = usePhases();

  return (
    <aside className="w-60 shrink-0 border-r border-gray-700 bg-gray-800 h-screen overflow-y-auto">
      <div className="p-4">
        <h1 className="text-lg font-bold text-gray-100">YT Automation</h1>
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
      </nav>
    </aside>
  );
}
