import { usePhases } from '../api/hooks';
import { PhaseCard } from '../components/PhaseCard';

export function PhaseDashboard() {
  const { data: phases, isLoading, error } = usePhases();

  if (isLoading) {
    return <p className="text-gray-400">Loading phases...</p>;
  }

  if (error) {
    return <p className="text-red-400">Failed to load phases.</p>;
  }

  return (
    <div>
      <h2 className="text-xl font-bold text-gray-100 mb-4">Dashboard</h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {phases?.map((p) => (
          <PhaseCard key={p.id} phase={p} />
        ))}
      </div>
    </div>
  );
}
