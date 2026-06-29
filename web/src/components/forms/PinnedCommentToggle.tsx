import { Toggle } from './Toggle';
import type { VideoResponse } from '../../api/types';
import { isVideoSponsored } from '../../lib/sponsorship';

// Re-exported for existing callers/tests; the canonical definition lives in lib/sponsorship.
export { isVideoSponsored };

/**
 * Builds the exact pinned-comment text to copy/paste onto a sponsored video.
 * Falls back to bracketed placeholders when name/url are missing so it's still
 * obvious what to fill in.
 */
export function buildSponsorCommentText(video: VideoResponse): string {
  const name = (video.sponsorship?.name ?? '').trim() || '[SPONSOR NAME]';
  const url = (video.sponsorship?.url ?? '').trim() || '[SPONSOR URL]';
  return `This video is sponsored by ${name}. Please visit ${url}.`;
}

interface PinnedCommentToggleProps {
  name: string;
  fieldName: string;
  value: boolean;
  onChange: (fieldName: string, value: boolean) => void;
  complete?: boolean;
  video: VideoResponse;
}

/**
 * Post-Publish checkbox reminding the user to post a pinned sponsor comment.
 * Renders the exact comment text in a selectable block for easy copy/paste.
 * Only meant to be rendered for sponsored videos (see isVideoSponsored).
 */
export function PinnedCommentToggle({ name, fieldName, value, onChange, complete, video }: PinnedCommentToggleProps) {
  const commentText = buildSponsorCommentText(video);
  return (
    <div className="py-1">
      <Toggle
        name={name}
        fieldName={fieldName}
        value={value}
        onChange={onChange}
        helpText="Post this as a pinned comment on the video, then check this off:"
        complete={complete}
      />
      <code className="mt-1 block select-all whitespace-pre-wrap rounded bg-gray-800 px-2 py-1 text-xs text-gray-200">
        {commentText}
      </code>
    </div>
  );
}
