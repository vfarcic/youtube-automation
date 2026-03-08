import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useVideosList } from '../api/hooks';
import { ProgressBar } from '../components/ProgressBar';
import { CreateVideoDialog } from '../components/CreateVideoDialog';
import { PHASE_NAMES } from '../lib/constants';
import { formatDate, parseVideoId } from '../lib/utils';

export function VideoList() {
  const { phaseId } = useParams<{ phaseId: string }>();
  const phase = phaseId !== undefined ? Number(phaseId) : undefined;
  const { data: videos, isLoading, error } = useVideosList(phase);
  const navigate = useNavigate();
  const [showCreate, setShowCreate] = useState(false);

  const phaseName = phase !== undefined
    ? (PHASE_NAMES[phase] ?? `Phase ${phase}`)
    : 'All Videos';

  if (isLoading) {
    return <p className="text-gray-400">Loading videos...</p>;
  }

  if (error) {
    return <p className="text-red-400">Failed to load videos.</p>;
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-bold text-gray-100">{phaseName}</h2>
        <button
          type="button"
          onClick={() => setShowCreate(true)}
          className="px-4 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Create Video
        </button>
      </div>
      {!videos || videos.length === 0 ? (
        <p className="text-gray-500">No videos in this phase.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-700 text-left text-gray-400">
                <th className="py-2 pr-4 font-medium">Name</th>
                <th className="py-2 pr-4 font-medium">Category</th>
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
      <CreateVideoDialog open={showCreate} onClose={() => setShowCreate(false)} />
    </div>
  );
}
