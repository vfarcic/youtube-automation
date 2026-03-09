import { useSearchParams, useNavigate } from 'react-router-dom';
import { useSearchVideos } from '../api/hooks';
import { ProgressBar } from '../components/ProgressBar';
import { PHASE_NAMES } from '../lib/constants';
import { formatDate, parseVideoId } from '../lib/utils';

export function SearchResults() {
  const [searchParams] = useSearchParams();
  const query = searchParams.get('q') ?? '';
  const { data: videos, isLoading } = useSearchVideos(query);
  const navigate = useNavigate();

  return (
    <div>
      <h2 className="text-xl font-bold text-gray-100 mb-4">
        Search results for &ldquo;{query}&rdquo;
      </h2>
      {isLoading ? (
        <p className="text-gray-400">Searching...</p>
      ) : !videos || videos.length === 0 ? (
        <p className="text-gray-500">No videos found.</p>
      ) : (
        <div className="overflow-x-auto">
          <p className="text-sm text-gray-400 mb-2">
            {videos.length} {videos.length === 1 ? 'result' : 'results'}
          </p>
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-700 text-left text-gray-400">
                <th className="py-2 pr-4 font-medium">Name</th>
                <th className="py-2 pr-4 font-medium">Category</th>
                <th className="py-2 pr-4 font-medium">Phase</th>
                <th className="py-2 pr-4 font-medium">Date</th>
                <th className="py-2 pr-4 font-medium w-48">Progress</th>
              </tr>
            </thead>
            <tbody>
              {videos.map((v) => {
                const { category, name } = parseVideoId(v.id);
                return (
                  <tr
                    key={v.id}
                    onClick={() =>
                      navigate(
                        `/videos/${encodeURIComponent(category)}/${encodeURIComponent(name)}`,
                      )
                    }
                    className="border-b border-gray-700 hover:bg-gray-800 cursor-pointer"
                  >
                    <td className="py-2 pr-4 text-gray-100">
                      {v.title || v.name}
                    </td>
                    <td className="py-2 pr-4 text-gray-400">{v.category}</td>
                    <td className="py-2 pr-4 text-gray-400">
                      {PHASE_NAMES[v.phase] ?? `Phase ${v.phase}`}
                    </td>
                    <td className="py-2 pr-4 text-gray-400">
                      {formatDate(v.date ?? '')}
                    </td>
                    <td className="py-2 pr-4">
                      <ProgressBar progress={v.progress} />
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
