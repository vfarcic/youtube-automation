import { useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useVideo, useVideoProgress, useAspects, usePatchVideo, useDeleteVideo } from '../api/hooks';
import { ProgressBar } from '../components/ProgressBar';
import { DynamicForm } from '../components/forms';
import { ASPECT_LABELS } from '../lib/constants';

export function VideoDetail() {
  const { category, videoName } = useParams<{
    category: string;
    videoName: string;
  }>();
  const navigate = useNavigate();
  const { data: video, isLoading, error } = useVideo(videoName, category);
  const { data: progress } = useVideoProgress(videoName, category);
  const { data: aspectsData } = useAspects();
  const patchVideo = usePatchVideo();
  const deleteVideo = useDeleteVideo();

  const [activeTab, setActiveTab] = useState<number>(0);
  const [saveMsg, setSaveMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);

  if (isLoading) {
    return <p className="text-gray-400">Loading video...</p>;
  }

  if (error) {
    return <p className="text-red-400">Failed to load video.</p>;
  }

  if (!video) {
    return <p className="text-gray-500">Video not found.</p>;
  }

  const aspects = aspectsData?.aspects ?? [];
  const currentAspect = aspects[activeTab];

  const handleSave = (changedFields: Record<string, unknown>) => {
    if (!currentAspect || !videoName || !category) return;
    setSaveMsg(null);
    patchVideo.mutate(
      { name: videoName, category, aspect: currentAspect.key, fields: changedFields },
      {
        onSuccess: () => setSaveMsg({ type: 'success', text: 'Saved successfully.' }),
        onError: (err) => setSaveMsg({ type: 'error', text: err.message || 'Save failed.' }),
      },
    );
  };

  const handleDelete = () => {
    if (!videoName || !category) return;
    deleteVideo.mutate(
      { name: videoName, category },
      {
        onSuccess: () => navigate(`/phases/${video.phase}`),
        onError: (err) => setSaveMsg({ type: 'error', text: err.message || 'Delete failed.' }),
      },
    );
  };

  // Find per-aspect progress from the progress response
  const getAspectProgress = (aspectKey: string) => {
    return progress?.aspects.find((a) => a.aspectKey === aspectKey);
  };

  return (
    <div>
      <Link
        to={`/phases/${video.phase}`}
        className="text-sm text-blue-400 hover:underline"
      >
        Back to phase
      </Link>
      <h2 className="text-xl font-bold text-gray-100 mt-2 mb-4">
        {video.name}
      </h2>

      {progress && (
        <div className="mb-6">
          <h3 className="text-sm font-semibold text-gray-400 mb-2">
            Overall Progress
          </h3>
          <ProgressBar progress={progress.overall} color="bg-green-500" />
        </div>
      )}

      {aspects.length > 0 && (
        <>
          <div className="flex flex-wrap gap-1 border-b border-gray-700 mb-4" role="tablist">
            {aspects.map((aspect, idx) => {
              const ap = getAspectProgress(aspect.key);
              const label = ASPECT_LABELS[aspect.key] ?? aspect.title;
              const isComplete = ap && ap.total > 0 && ap.completed === ap.total;
              const isPartial = ap && ap.completed > 0 && ap.completed < ap.total;
              const dotColor = isComplete
                ? 'bg-green-500'
                : isPartial
                  ? 'bg-yellow-500'
                  : 'bg-gray-600';
              return (
                <button
                  key={aspect.key}
                  role="tab"
                  aria-selected={idx === activeTab}
                  onClick={() => { setActiveTab(idx); setSaveMsg(null); }}
                  className={`px-3 py-2 text-sm font-medium border-b-2 transition-colors flex items-center gap-1.5 ${
                    idx === activeTab
                      ? 'border-blue-500 text-blue-400'
                      : 'border-transparent text-gray-400 hover:text-gray-300'
                  }`}
                >
                  {ap && (
                    <span
                      className={`inline-block w-2 h-2 rounded-full shrink-0 ${dotColor}`}
                      title={`${ap.completed}/${ap.total} complete`}
                    />
                  )}
                  {label}
                  {ap && (
                    <span className="text-xs text-gray-500">
                      {ap.completed}/{ap.total}
                    </span>
                  )}
                </button>
              );
            })}
          </div>

          {currentAspect && (
            <DynamicForm
              key={currentAspect.key}
              fields={currentAspect.fields}
              video={video}
              onSave={handleSave}
              saving={patchVideo.isPending}
              category={category}
              videoName={videoName}
            />
          )}

          {saveMsg && (
            <p className={`mt-3 text-sm ${saveMsg.type === 'success' ? 'text-green-400' : 'text-red-400'}`}>
              {saveMsg.text}
            </p>
          )}
        </>
      )}

      <div className="mt-8 pt-4 border-t border-gray-700">
        {!confirmDelete ? (
          <button
            type="button"
            onClick={() => setConfirmDelete(true)}
            className="px-4 py-1.5 text-sm text-red-400 border border-red-800 rounded hover:bg-red-900/30"
          >
            Delete Video
          </button>
        ) : (
          <div className="flex items-center gap-3">
            <span className="text-sm text-red-400">Are you sure?</span>
            <button
              type="button"
              onClick={handleDelete}
              disabled={deleteVideo.isPending}
              className="px-4 py-1.5 text-sm bg-red-600 text-white rounded hover:bg-red-700 disabled:opacity-50"
            >
              {deleteVideo.isPending ? 'Deleting...' : 'Confirm Delete'}
            </button>
            <button
              type="button"
              onClick={() => setConfirmDelete(false)}
              className="px-4 py-1.5 text-sm border border-gray-600 text-gray-300 rounded hover:bg-gray-800"
            >
              Cancel
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
