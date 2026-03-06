import { useRef } from 'react';
import { useUploadThumbnailToDrive } from '../../api/hooks';

interface FileUploadInputProps {
  videoName: string;
  category: string;
  variantIndex: number;
  currentDriveFileId?: string;
}

export function FileUploadInput({
  videoName,
  category,
  variantIndex,
  currentDriveFileId,
}: FileUploadInputProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const uploadMutation = useUploadThumbnailToDrive();

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    uploadMutation.mutate({
      name: videoName,
      category,
      variantIndex,
      file,
    });
    if (fileInputRef.current) fileInputRef.current.value = '';
  };

  // Show the most recent Drive file ID (from upload response or existing data)
  const driveFileId = uploadMutation.data?.driveFileId ?? currentDriveFileId;

  return (
    <div className="mt-1">
      <div className="flex items-center gap-2">
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          onChange={handleFileChange}
          className="hidden"
          data-testid="thumbnail-file-input"
        />
        <button
          type="button"
          onClick={() => fileInputRef.current?.click()}
          disabled={uploadMutation.isPending}
          className="px-2 py-1 text-xs border border-gray-600 text-gray-300 rounded hover:border-blue-400 hover:text-blue-400 disabled:opacity-50"
        >
          {uploadMutation.isPending ? 'Uploading...' : driveFileId ? 'Replace' : 'Upload to Drive'}
        </button>
        {uploadMutation.isSuccess && !uploadMutation.data?.syncWarning && (
          <span className="text-xs text-green-400">Uploaded</span>
        )}
        {uploadMutation.isSuccess && uploadMutation.data?.syncWarning && (
          <span className="text-xs text-yellow-400">Uploaded (git sync failed: {uploadMutation.data.syncWarning})</span>
        )}
        {uploadMutation.isError && (
          <span className="text-xs text-red-400" data-testid="drive-upload-error">
            {uploadMutation.error?.message || 'Upload failed'}
          </span>
        )}
      </div>
      {driveFileId && (
        <div className="mt-1 flex items-center gap-1" data-testid="drive-uploaded-badge">
          <span className="text-xs text-gray-400">Drive ID:</span>
          <code className="text-xs text-gray-300 bg-gray-800 px-1 rounded">{driveFileId}</code>
        </div>
      )}
    </div>
  );
}
