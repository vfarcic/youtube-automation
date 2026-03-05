import { useParams, Link } from 'react-router-dom';
import { useVideo, useVideoProgress } from '../api/hooks';
import { ProgressBar } from '../components/ProgressBar';
import { ASPECT_LABELS } from '../lib/constants';
import type { VideoResponse } from '../api/types';

type AspectKey =
  | 'init'
  | 'work'
  | 'define'
  | 'edit'
  | 'publish'
  | 'postPublish';

const ASPECT_FIELD_GROUPS: { key: AspectKey; fields: (keyof VideoResponse)[] }[] = [
  {
    key: 'init',
    fields: ['projectName', 'projectURL', 'date', 'category'],
  },
  {
    key: 'work',
    fields: ['screen', 'head', 'thumbnails', 'diagrams', 'screenshots', 'code', 'slides'],
  },
  {
    key: 'define',
    fields: ['description', 'tags', 'descriptionTags', 'tagline', 'taglineIdeas', 'timecodes', 'relatedVideos'],
  },
  {
    key: 'edit',
    fields: ['requestEdit', 'requestThumbnail', 'movie', 'animations', 'members'],
  },
  {
    key: 'publish',
    fields: ['uploadVideo', 'videoId', 'hugoPath', 'gist', 'tweet', 'language'],
  },
  {
    key: 'postPublish',
    fields: ['linkedInPosted', 'slackPosted', 'blueSkyPosted', 'hnPosted', 'dotPosted', 'youTubeHighlight', 'youTubeComment', 'youTubeCommentReply'],
  },
];

const ASPECT_DISPLAY_NAMES: Record<AspectKey, string> = {
  init: 'Initial Details',
  work: 'Work Progress',
  define: 'Definition',
  edit: 'Post Production',
  publish: 'Publishing',
  postPublish: 'Post Publish',
};

function FieldValue({ value }: { value: unknown }) {
  if (typeof value === 'boolean') {
    return (
      <span className={value ? 'text-green-600' : 'text-gray-400'}>
        {value ? 'Yes' : 'No'}
      </span>
    );
  }
  if (typeof value === 'string') {
    return (
      <span className={value ? 'text-gray-900' : 'text-gray-400'}>
        {value || '-'}
      </span>
    );
  }
  return <span className="text-gray-400">-</span>;
}

function formatFieldName(name: string): string {
  return name
    .replace(/([A-Z])/g, ' $1')
    .replace(/^./, (c) => c.toUpperCase())
    .trim();
}

export function VideoDetail() {
  const { category, videoName } = useParams<{
    category: string;
    videoName: string;
  }>();
  const { data: video, isLoading, error } = useVideo(videoName, category);
  const { data: progress } = useVideoProgress(videoName, category);

  if (isLoading) {
    return <p className="text-gray-500">Loading video...</p>;
  }

  if (error) {
    return <p className="text-red-500">Failed to load video.</p>;
  }

  if (!video) {
    return <p className="text-gray-400">Video not found.</p>;
  }

  return (
    <div>
      <Link
        to={`/phases/${video.phase}`}
        className="text-sm text-blue-500 hover:underline"
      >
        Back to phase
      </Link>
      <h2 className="text-xl font-bold text-gray-900 mt-2 mb-4">
        {video.name}
      </h2>

      {progress && (
        <div className="mb-6">
          <h3 className="text-sm font-semibold text-gray-500 mb-2">
            Overall Progress
          </h3>
          <ProgressBar progress={progress.overall} color="bg-green-500" />
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 mt-3">
            {progress.aspects.map((a) => (
              <div key={a.aspectKey} className="text-sm">
                <div className="text-gray-600 mb-1">
                  {ASPECT_LABELS[a.aspectKey] ?? a.title}
                </div>
                <ProgressBar progress={a} />
              </div>
            ))}
          </div>
        </div>
      )}

      {ASPECT_FIELD_GROUPS.map(({ key, fields }) => (
        <div key={key} className="mb-6">
          <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wider mb-2 border-b pb-1">
            {ASPECT_DISPLAY_NAMES[key]}
          </h3>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-2">
            {fields.map((f) => (
              <div key={f} className="flex justify-between py-1 text-sm">
                <span className="text-gray-500">{formatFieldName(f)}</span>
                <FieldValue value={video[f]} />
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}
