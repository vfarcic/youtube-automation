import { useRef, useState } from 'react';
import { useUploadShortToDrive, usePublishShort } from '../../api/hooks';

interface ShortItemActionsProps {
  videoName: string;
  category: string;
  shortId: string;
  driveFileId?: string;
  filePath?: string;
  scheduledDate?: string;
  youtubeId?: string;
}

export function ShortItemActions({
  videoName,
  category,
  shortId,
  driveFileId,
  filePath,
  scheduledDate,
  youtubeId,
}: ShortItemActionsProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const uploadMutation = useUploadShortToDrive();
  const publishMutation = usePublishShort();
  const [progress, setProgress] = useState(0);
  const [publishError, setPublishError] = useState<string | null>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setProgress(0);
    uploadMutation.mutate(
      {
        name: videoName,
        category,
        shortId,
        file,
        onProgress: (percent) => setProgress(percent),
      },
      {
        onSettled: () => setProgress(0),
      },
    );
    if (fileInputRef.current) fileInputRef.current.value = '';
  };

  const effectiveDriveFileId = uploadMutation.data?.driveFileId ?? driveFileId;
  const hasFile = Boolean(effectiveDriveFileId || filePath);
  const hasScheduledDate = Boolean(scheduledDate);
  const publishedId = publishMutation.data?.youtubeId ?? youtubeId;

  const downloadUrl = `/api/drive/download/short/${encodeURIComponent(videoName)}/${encodeURIComponent(shortId)}?category=${encodeURIComponent(category)}&token=${encodeURIComponent(localStorage.getItem('api_token') || '')}`;

  const handlePublish = () => {
    setPublishError(null);
    publishMutation.mutate(
      { name: videoName, category, shortId },
      { onError: (err) => setPublishError(err.message) },
    );
  };

  return (
    <div className="mt-2 pt-2 border-t border-gray-700 space-y-2">
      {/* Upload section */}
      <div className="flex items-center gap-2 flex-wrap">
        <input
          ref={fileInputRef}
          type="file"
          accept="video/*"
          onChange={handleFileChange}
          className="hidden"
          data-testid={`short-file-input-${shortId}`}
        />
        <button
          type="button"
          onClick={() => fileInputRef.current?.click()}
          disabled={uploadMutation.isPending}
          className="px-2 py-1 text-xs border border-gray-600 text-gray-300 rounded hover:border-blue-400 hover:text-blue-400 disabled:opacity-50"
        >
          {uploadMutation.isPending ? 'Uploading...' : effectiveDriveFileId ? 'Replace' : 'Upload to Drive'}
        </button>
        {effectiveDriveFileId && (
          <a
            href={downloadUrl}
            download
            className="px-2 py-1 text-xs border border-gray-600 text-gray-300 rounded hover:border-blue-400 hover:text-blue-400"
            data-testid={`short-download-link-${shortId}`}
          >
            Download
          </a>
        )}
        {uploadMutation.isPending && (
          <div className="flex items-center gap-2" data-testid={`short-upload-progress-${shortId}`}>
            <div className="w-32 h-2 bg-gray-700 rounded overflow-hidden">
              <div
                className="h-full bg-blue-500 transition-all"
                style={{ width: `${progress}%` }}
              />
            </div>
            <span className="text-xs text-gray-400">
              {progress >= 50 ? 'Uploading to Drive...' : `${progress}%`}
            </span>
          </div>
        )}
        {uploadMutation.isSuccess && !uploadMutation.data?.syncWarning && (
          <span className="text-xs text-green-400">Uploaded</span>
        )}
        {uploadMutation.isSuccess && uploadMutation.data?.syncWarning && (
          <span className="text-xs text-yellow-400">Uploaded (git sync failed: {uploadMutation.data.syncWarning})</span>
        )}
        {uploadMutation.isError && (
          <span className="text-xs text-red-400" data-testid={`short-upload-error-${shortId}`}>
            {uploadMutation.error?.message || 'Upload failed'}
          </span>
        )}
      </div>

      {/* Publish section */}
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
    </div>
  );
}
