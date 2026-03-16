import { useState } from 'react';
import { usePublishShort } from '../../api/hooks';

interface ShortPublishActionProps {
  videoName: string;
  category: string;
  shortId: string;
  driveFileId?: string;
  filePath?: string;
  scheduledDate?: string;
  youtubeId?: string;
}

export function ShortPublishAction({
  videoName,
  category,
  shortId,
  driveFileId,
  filePath,
  scheduledDate,
  youtubeId,
}: ShortPublishActionProps) {
  const publishMutation = usePublishShort();
  const [publishError, setPublishError] = useState<string | null>(null);

  const hasFile = Boolean(driveFileId || filePath);
  const hasScheduledDate = Boolean(scheduledDate);
  const publishedId = publishMutation.data?.youtubeId ?? youtubeId;

  const handlePublish = () => {
    setPublishError(null);
    publishMutation.mutate(
      { name: videoName, category, shortId },
      { onError: (err) => setPublishError(err.message) },
    );
  };

  return (
    <div className="flex items-center gap-2">
      {publishedId ? (
        <span className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-green-400 bg-green-900/30 border border-green-700 rounded">
          Published: <code className="text-xs">{publishedId}</code>
        </span>
      ) : hasFile && hasScheduledDate ? (
        <button
          type="button"
          onClick={handlePublish}
          disabled={publishMutation.isPending}
          className="px-4 py-1.5 text-sm font-medium bg-purple-600 text-white rounded hover:bg-purple-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {publishMutation.isPending ? 'Publishing...' : 'Publish to YouTube'}
        </button>
      ) : (
        <button
          type="button"
          disabled
          className="px-4 py-1.5 text-sm font-medium bg-purple-600 text-white rounded opacity-50 cursor-not-allowed"
        >
          Publish to YouTube
        </button>
      )}
      {!publishedId && !hasFile && (
        <p className="text-xs text-yellow-400">Upload a file first</p>
      )}
      {!publishedId && hasFile && !hasScheduledDate && (
        <p className="text-xs text-yellow-400">Set scheduled date first</p>
      )}
      {publishError && <p className="text-xs text-red-400">{publishError}</p>}
    </div>
  );
}
