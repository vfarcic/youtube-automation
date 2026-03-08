import { useNavigate } from 'react-router-dom';
import type { PhaseInfo } from '../api/types';
import { PHASE_NAMES, PHASE_ACCENT_COLORS, PHASE_COLORS } from '../lib/constants';

interface PhaseCardProps {
  phase: PhaseInfo;
}

export function PhaseCard({ phase }: PhaseCardProps) {
  const navigate = useNavigate();
  const accent = PHASE_ACCENT_COLORS[phase.id] ?? 'border-gray-500';
  const bg = PHASE_COLORS[phase.id] ?? 'bg-gray-500';

  return (
    <button
      onClick={() => navigate(`/phases/${phase.id}`)}
      className={`border-l-4 ${accent} rounded-lg bg-gray-800 p-4 shadow-md shadow-black/20 hover:shadow-lg hover:shadow-black/30 transition-shadow text-left w-full`}
    >
      <div className="flex items-center gap-2 mb-2">
        <span className={`w-3 h-3 rounded-full ${bg}`} />
        <h3 className="font-semibold text-gray-100">
          {PHASE_NAMES[phase.id] ?? `Phase ${phase.id}`}
        </h3>
      </div>
      <p className="text-2xl font-bold text-gray-300">{phase.count}</p>
      <p className="text-xs text-gray-500">
        {phase.count === 1 ? 'video' : 'videos'}
      </p>
    </button>
  );
}
