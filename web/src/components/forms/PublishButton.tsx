import { useState } from 'react';
import { usePublishYouTube, usePublishHugo } from '../../api/hooks';
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
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<string | null>(null);

  const mutation = fieldName === 'videoId' ? publishYouTube : publishHugo;
  const isLoading = mutation.isPending;

  const handleClick = () => {
    setError(null);
    setResult(null);
    mutation.mutate(
      { name: videoName, category },
      {
        onError: (err) => setError(err.message),
        onSuccess: (data) => {
          if ('videoId' in data) setResult(data.videoId);
          else if ('hugoPath' in data) setResult(data.hugoPath);
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
          <span className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-green-400 bg-green-900/30 border border-green-700 rounded">
            {config.doneLabel}: <code className="text-xs">{displayValue}</code>
          </span>
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
      </div>
      {!hasPrerequisites && !displayValue && (
        <p className="mt-1 text-xs text-yellow-400">
          {fieldName === 'videoId' ? 'Upload a video file first' : 'Publish to YouTube first'}
        </p>
      )}
      {error && <p className="mt-1 text-xs text-red-400">{error}</p>}
    </div>
  );
}
