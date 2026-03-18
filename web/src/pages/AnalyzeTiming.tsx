import { useTimingRecommendations, useGenerateTimingRecommendations } from '../api/hooks';
import type { TimingRecommendation } from '../api/types';

function RecommendationsTable({ recommendations }: { recommendations: TimingRecommendation[] }) {
  if (recommendations.length === 0) {
    return (
      <p className="text-gray-500 text-sm italic">No timing recommendations configured yet.</p>
    );
  }

  return (
    <table className="w-full text-sm border border-gray-700 rounded overflow-hidden">
      <thead className="bg-gray-800">
        <tr>
          <th className="px-4 py-2 text-left text-gray-300 font-medium">Day</th>
          <th className="px-4 py-2 text-left text-gray-300 font-medium">Time (UTC)</th>
          <th className="px-4 py-2 text-left text-gray-300 font-medium">Reasoning</th>
        </tr>
      </thead>
      <tbody>
        {recommendations.map((rec) => (
          <tr key={`${rec.day}-${rec.time}`} className="border-t border-gray-700">
            <td className="px-4 py-2 text-gray-100">{rec.day}</td>
            <td className="px-4 py-2 text-gray-100 font-mono">{rec.time}</td>
            <td className="px-4 py-2 text-gray-400">{rec.reasoning}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

export function AnalyzeTiming() {
  const { data: currentData, isLoading, isError: isLoadError, error: loadError } = useTimingRecommendations();
  const generateMutation = useGenerateTimingRecommendations();

  const handleGenerate = () => {
    generateMutation.mutate();
  };

  return (
    <div className="p-6 max-w-4xl">
      <h1 className="text-2xl font-bold text-gray-100 mb-2">Timing Recommendations</h1>
      <p className="text-gray-400 text-sm mb-6">
        Analyze YouTube analytics to find optimal publishing days and times.
        Recommendations are auto-saved to settings.yaml.
      </p>

      <button
        onClick={handleGenerate}
        disabled={generateMutation.isPending}
        className="px-4 py-2 bg-teal-600 hover:bg-teal-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded text-sm font-medium"
      >
        {generateMutation.isPending ? 'Generating...' : 'Generate New Recommendations'}
      </button>

      {generateMutation.isPending && (
        <div className="mt-4 text-gray-400 text-sm">
          <p>Fetching YouTube analytics and generating AI recommendations...</p>
          <p className="mt-1">This may take a minute.</p>
        </div>
      )}

      {generateMutation.isError && (
        <div className="mt-4 p-3 bg-red-900/30 border border-red-700 rounded text-red-300 text-sm">
          {generateMutation.error.message}
        </div>
      )}

      {generateMutation.isSuccess && (
        <div className="mt-4 space-y-3">
          <p className="text-gray-300 text-sm">
            Generated from <span className="font-semibold text-gray-100" data-testid="video-count">{generateMutation.data.videoCount}</span> videos.
            Recommendations saved automatically.
          </p>
          {generateMutation.data.syncWarning && (
            <div className="p-3 bg-yellow-900/30 border border-yellow-700 rounded text-yellow-300 text-sm">
              {generateMutation.data.syncWarning}
            </div>
          )}
        </div>
      )}

      <div className="mt-8">
        <h2 className="text-lg font-semibold text-gray-100 mb-3">Current Recommendations</h2>
        {isLoading ? (
          <p className="text-gray-500 text-sm">Loading...</p>
        ) : isLoadError ? (
          <div className="p-3 bg-red-900/30 border border-red-700 rounded text-red-300 text-sm">
            {loadError instanceof Error ? loadError.message : 'Failed to load timing recommendations.'}
          </div>
        ) : (
          <RecommendationsTable recommendations={currentData?.recommendations ?? []} />
        )}
      </div>
    </div>
  );
}
