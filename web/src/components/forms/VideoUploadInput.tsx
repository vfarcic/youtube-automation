import { useRef, useState } from 'react';
import { useUploadVideoToDrive } from '../../api/hooks';

interface VideoUploadInputProps {
  videoName: string;
  category: string;
  currentDriveFileId?: string;
}

export function VideoUploadInput({
  videoName,
  category,
  currentDriveFileId,
}: VideoUploadInputProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const uploadMutation = useUploadVideoToDrive();
  const [progress, setProgress] = useState(0);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setProgress(0);
    uploadMutation.mutate(
      {
        name: videoName,
        category,
        file,
        onProgress: (percent) => setProgress(percent),
      },
      {
        onSettled: () => setProgress(0),
      },
    );
    if (fileInputRef.current) fileInputRef.current.value = '';
  };

  const driveFileId = uploadMutation.data?.driveFileId ?? currentDriveFileId;

  const downloadUrl = `/api/drive/download/video/${encodeURIComponent(videoName)}?category=${encodeURIComponent(category)}&token=${encodeURIComponent(localStorage.getItem('api_token') || '')}`;

  return (
    <div className="mt-1">
      <div className="flex items-center gap-2">
        <input
          ref={fileInputRef}
          type="file"
          accept="video/*"
          onChange={handleFileChange}
          className="hidden"
          data-testid="video-file-input"
        />
        <button
          type="button"
          onClick={() => fileInputRef.current?.click()}
          disabled={uploadMutation.isPending}
          className="px-2 py-1 text-xs border border-gray-600 text-gray-300 rounded hover:border-blue-400 hover:text-blue-400 disabled:opacity-50"
        >
          {uploadMutation.isPending ? 'Uploading...' : driveFileId ? 'Replace' : 'Upload to Drive'}
        </button>
        {driveFileId && (
          <a
            href={downloadUrl}
            download
            className="px-2 py-1 text-xs border border-gray-600 text-gray-300 rounded hover:border-blue-400 hover:text-blue-400"
            data-testid="video-download-link"
          >
            Download
          </a>
        )}
        {uploadMutation.isPending && (
          <div className="flex items-center gap-2" data-testid="upload-progress">
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
          <span className="text-xs text-red-400" data-testid="video-upload-error">
            {uploadMutation.error?.message || 'Upload failed'}
          </span>
        )}
      </div>
    </div>
  );
}
