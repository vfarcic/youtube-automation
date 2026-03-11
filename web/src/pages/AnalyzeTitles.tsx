import { useState } from 'react';
import { useAnalyzeTitles, useApplyTitlesTemplate } from '../api/hooks';
import type { AnalyzeTitlesResponse } from '../api/types';

export function AnalyzeTitles() {
  const analyzeMutation = useAnalyzeTitles();
  const applyMutation = useApplyTitlesTemplate();
  const [result, setResult] = useState<AnalyzeTitlesResponse | null>(null);
  const [applied, setApplied] = useState(false);
  const [syncWarning, setSyncWarning] = useState('');

  const handleRunAnalysis = () => {
    setResult(null);
    setApplied(false);
    setSyncWarning('');
    analyzeMutation.mutate(undefined, {
      onSuccess: (data) => setResult(data),
    });
  };

  const handleApply = () => {
    if (!result?.titlesMdContent) return;
    applyMutation.mutate(
      { content: result.titlesMdContent },
      {
        onSuccess: (data) => {
          setApplied(true);
          if (data.syncWarning) setSyncWarning(data.syncWarning);
        },
      },
    );
  };

  return (
    <div className="p-6 max-w-4xl">
      <h1 className="text-2xl font-bold text-gray-100 mb-2">Title Analysis</h1>
      <p className="text-gray-400 text-sm mb-6">
        Analyze A/B test share data and first-week YouTube analytics to discover
        which title patterns work best on your channel.
      </p>

      <button
        onClick={handleRunAnalysis}
        disabled={analyzeMutation.isPending}
        className="px-4 py-2 bg-purple-600 hover:bg-purple-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded text-sm font-medium"
      >
        {analyzeMutation.isPending ? 'Analyzing...' : 'Run Analysis'}
      </button>

      {analyzeMutation.isPending && (
        <div className="mt-4 text-gray-400 text-sm">
          <p>Loading A/B data, fetching YouTube analytics, and running AI analysis...</p>
          <p className="mt-1">This may take a minute or two.</p>
        </div>
      )}

      {analyzeMutation.isError && (
        <div className="mt-4 p-3 bg-red-900/30 border border-red-700 rounded text-red-300 text-sm">
          {analyzeMutation.error.message}
        </div>
      )}

      {result && result.videoCount === 0 && (
        <div className="mt-4 p-3 bg-yellow-900/30 border border-yellow-700 rounded text-yellow-300 text-sm">
          No videos with A/B test data found. Videos need 2+ title variants with share data.
        </div>
      )}

      {result && result.videoCount > 0 && (
        <div className="mt-6 space-y-6">
          <p className="text-gray-300 text-sm">
            Analyzed <span className="font-semibold text-gray-100" data-testid="video-count">{result.videoCount}</span> videos with A/B test data.
          </p>

          {result.highPerformingPatterns?.length > 0 && (
            <section>
              <h2 className="text-lg font-semibold text-green-400 mb-3">High-Performing Patterns</h2>
              <div className="space-y-3">
                {result.highPerformingPatterns.map((p, i) => (
                  <div key={i} className="p-3 bg-gray-800 border border-gray-700 rounded">
                    <div className="font-medium text-gray-100">{p.pattern}</div>
                    <div className="text-sm text-gray-400 mt-1">{p.description}</div>
                    {p.examples?.length > 0 && (
                      <div className="mt-2 text-sm text-gray-500">
                        {p.examples.map((ex, j) => (
                          <span key={j} className="inline-block mr-2 px-2 py-0.5 bg-gray-700 rounded text-gray-300">
                            {ex}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </section>
          )}

          {result.lowPerformingPatterns?.length > 0 && (
            <section>
              <h2 className="text-lg font-semibold text-red-400 mb-3">Low-Performing Patterns</h2>
              <div className="space-y-3">
                {result.lowPerformingPatterns.map((p, i) => (
                  <div key={i} className="p-3 bg-gray-800 border border-gray-700 rounded">
                    <div className="font-medium text-gray-100">{p.pattern}</div>
                    <div className="text-sm text-gray-400 mt-1">{p.description}</div>
                    {p.examples?.length > 0 && (
                      <div className="mt-2 text-sm text-gray-500">
                        {p.examples.map((ex, j) => (
                          <span key={j} className="inline-block mr-2 px-2 py-0.5 bg-gray-700 rounded text-gray-300">
                            {ex}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </section>
          )}

          {result.recommendations?.length > 0 && (
            <section>
              <h2 className="text-lg font-semibold text-blue-400 mb-3">Recommendations</h2>
              <div className="space-y-3">
                {result.recommendations.map((r, i) => (
                  <div key={i} className="p-3 bg-gray-800 border border-gray-700 rounded">
                    <div className="font-medium text-gray-100">{r.recommendation}</div>
                    <div className="text-sm text-gray-400 mt-1">{r.evidence}</div>
                    {r.example && (
                      <div className="mt-1 text-sm text-gray-500 italic">Example: {r.example}</div>
                    )}
                  </div>
                ))}
              </div>
            </section>
          )}

          {result.titlesMdContent && (
            <section>
              <h2 className="text-lg font-semibold text-gray-100 mb-3">Proposed titles.md</h2>
              <pre className="p-4 bg-gray-900 border border-gray-700 rounded text-sm text-gray-300 overflow-x-auto whitespace-pre-wrap">
                {result.titlesMdContent}
              </pre>
              <div className="mt-3 flex items-center gap-3">
                <button
                  onClick={handleApply}
                  disabled={applyMutation.isPending || applied}
                  className="px-4 py-2 bg-green-600 hover:bg-green-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded text-sm font-medium"
                >
                  {applyMutation.isPending
                    ? 'Saving...'
                    : applied
                      ? 'Saved'
                      : 'Save titles.md'}
                </button>
                {applied && !syncWarning && (
                  <span className="text-green-400 text-sm">titles.md updated.</span>
                )}
                {syncWarning && (
                  <span className="text-yellow-400 text-sm">Saved locally, but push failed: {syncWarning}</span>
                )}
                {applyMutation.isError && (
                  <span className="text-red-400 text-sm">{applyMutation.error.message}</span>
                )}
              </div>
            </section>
          )}
        </div>
      )}
    </div>
  );
}
