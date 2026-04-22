import { useState } from 'react';
import { usePublishYouTube, usePublishHugo, useReuploadYouTube } from '../../api/hooks';
import type { VideoResponse } from '../../api/types';
import { FieldLabel } from './FieldLabel';

/** Field names that render as publish buttons. */
const PUBLISH_FIELDS: Record<string, { label: string; doneLabel: string; valueKey: keyof VideoResponse }> = {
  videoId: { label: 'Publish to YouTube', doneLabel: 'Published', valueKey: 'videoId' },
  hugoPath: { label: 'Publish Hugo Post', doneLabel: 'Published', valueKey: 'hugoPath' },
};

/** Returns true if a field should render as a PublishButton. */
export function isPublishField(fieldName: string): boolean {
  return fieldName in PUBLISH_FIELDS;
}

interface PublishButtonProps {
  fieldName: string;
  value: string;
  category: string;
  videoName: string;
  video: VideoResponse;
}

export function PublishButton({ fieldName, value, category, videoName, video }: PublishButtonProps) {
  const config = PUBLISH_FIELDS[fieldName];
  const publishYouTube = usePublishYouTube();
  const publishHugo = usePublishHugo();
  const reuploadYouTube = useReuploadYouTube();
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<string | null>(null);
  const [warning, setWarning] = useState<string | null>(null);
  const [confirmReupload, setConfirmReupload] = useState(false);

  const mutation = fieldName === 'videoId' ? publishYouTube : publishHugo;
  const isLoading = mutation.isPending;
  const isReuploading = reuploadYouTube.isPending;

  const handleClick = () => {
    setError(null);
    setResult(null);
    setWarning(null);
    mutation.mutate(
      { name: videoName, category },
      {
        onError: (err) => setError(err.message),
        onSuccess: (data) => {
          if ('videoId' in data) {
            setResult(data.videoId);
            if ('thumbnailWarning' in data && data.thumbnailWarning) {
              setWarning(data.thumbnailWarning);
            }
          } else if ('hugoPath' in data) setResult(data.hugoPath);
        },
      },
    );
  };

  const handleReupload = () => {
    setError(null);
    setResult(null);
    setWarning(null);
    setConfirmReupload(false);
    reuploadYouTube.mutate(
      { name: videoName, category },
      {
        onError: (err) => setError(err.message),
        onSuccess: (data) => {
          setResult(data.videoId);
          if (data.thumbnailWarning) {
            setWarning(data.thumbnailWarning);
          }
        },
      },
    );
  };

  const hasPrerequisites = fieldName === 'videoId'
    ? Boolean(video.uploadVideo || video.videoDriveFileId || video.videoFile)
    : Boolean(video.videoId);

  const displayValue = result || value;

  return (
    <div>
      <FieldLabel name={config.label} helpText="" complete={Boolean(displayValue)} />
      <div className="flex items-center gap-2 py-1">
        {displayValue ? (
          <>
            <span className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-green-400 bg-green-900/30 border border-green-700 rounded">
              {config.doneLabel}: <code className="text-xs">{displayValue}</code>
            </span>
            {fieldName === 'videoId' && hasPrerequisites && !confirmReupload && (
              <button
                type="button"
                onClick={() => setConfirmReupload(true)}
                disabled={isReuploading}
                className="px-3 py-1.5 text-sm font-medium text-orange-400 border border-orange-700 rounded hover:bg-orange-900/30 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isReuploading ? 'Re-uploading...' : 'Re-upload'}
              </button>
            )}
            {confirmReupload && (
              <div className="flex items-center gap-2">
                <span className="text-xs text-orange-400">This will delete the YouTube video and upload again. Continue?</span>
                <button
                  type="button"
                  onClick={handleReupload}
                  disabled={isReuploading}
                  className="px-3 py-1.5 text-sm font-medium bg-orange-600 text-white rounded hover:bg-orange-700 disabled:opacity-50"
                >
                  {isReuploading ? 'Re-uploading...' : 'Confirm'}
                </button>
                <button
                  type="button"
                  onClick={() => setConfirmReupload(false)}
                  disabled={isReuploading}
                  className="px-3 py-1.5 text-sm font-medium text-gray-400 border border-gray-600 rounded hover:bg-gray-800"
                >
                  Cancel
                </button>
              </div>
            )}
          </>
        ) : (
          <button
            type="button"
            onClick={handleClick}
            disabled={isLoading || !hasPrerequisites}
            className="px-4 py-1.5 text-sm font-medium bg-purple-600 text-white rounded hover:bg-purple-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isLoading ? 'Publishing...' : config.label}
          </button>
        )}
        {!hasPrerequisites && !displayValue && (
          <p className="text-xs text-yellow-400">
            {fieldName === 'videoId' ? 'Upload a video file first' : 'Publish to YouTube first'}
          </p>
        )}
      </div>
      {warning && <p className="mt-1 text-xs text-yellow-400">{warning}</p>}
      {error && <p className="mt-1 text-xs text-red-400">{error}</p>}
    </div>
  );
}
