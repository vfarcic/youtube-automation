import { useRef, useState } from 'react';
import { useUploadShortToDrive } from '../../api/hooks';

interface ShortUploadActionProps {
  videoName: string;
  category: string;
  shortId: string;
  driveFileId?: string;
}

export function ShortUploadAction({
  videoName,
  category,
  shortId,
  driveFileId,
}: ShortUploadActionProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const uploadMutation = useUploadShortToDrive();
  const [progress, setProgress] = useState(0);

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

  const downloadUrl = `/api/drive/download/short/${encodeURIComponent(videoName)}/${encodeURIComponent(shortId)}?category=${encodeURIComponent(category)}&token=${encodeURIComponent(localStorage.getItem('api_token') || '')}`;

  return (
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
  );
}
