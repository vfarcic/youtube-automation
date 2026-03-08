import { useState } from 'react';
import { useSocialPost } from '../../api/hooks';
import { CopyTextDialog } from './CopyTextDialog';
import { FieldLabel } from './FieldLabel';

/** Maps boolean field names to social platform config. */
const SOCIAL_FIELDS: Record<string, { platform: string; label: string; doneLabel: string; automated: boolean }> = {
  blueSkyPosted: { platform: 'bluesky', label: 'Post to BlueSky', doneLabel: 'Posted to BlueSky', automated: true },
  slackPosted: { platform: 'slack', label: 'Post to Slack', doneLabel: 'Posted to Slack', automated: true },
  linkedInPosted: { platform: 'linkedin', label: 'Post to LinkedIn', doneLabel: 'Posted to LinkedIn', automated: false },
  hnPosted: { platform: 'hackernews', label: 'Post to Hacker News', doneLabel: 'Posted to HN', automated: false },
  dotPosted: { platform: 'dot', label: 'Post to DOT', doneLabel: 'Posted to DOT', automated: false },
};

/** Returns true if a field should render as a SocialPostButton. */
export function isSocialField(fieldName: string): boolean {
  return fieldName in SOCIAL_FIELDS;
}

interface SocialPostButtonProps {
  fieldName: string;
  value: boolean;
  category: string;
  videoName: string;
}

export function SocialPostButton({ fieldName, value, category, videoName }: SocialPostButtonProps) {
  const config = SOCIAL_FIELDS[fieldName];
  const socialPost = useSocialPost();
  const [error, setError] = useState<string | null>(null);
  const [copyText, setCopyText] = useState<string | null>(null);

  const isLoading = socialPost.isPending;

  const handleClick = () => {
    setError(null);
    setCopyText(null);
    socialPost.mutate(
      { platform: config.platform, name: videoName, category },
      {
        onError: (err) => setError(err.message),
        onSuccess: (data) => {
          if (!config.automated && data.message) {
            setCopyText(data.message);
          }
        },
      },
    );
  };

  if (value) {
    return (
      <div>
        <FieldLabel name={config.label} helpText="" complete={true} />
        <div className="flex items-center gap-2 py-2">
          <span className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-green-400 bg-green-900/30 border border-green-700 rounded">
            {config.doneLabel}
          </span>
        </div>
      </div>
    );
  }

  return (
    <div>
      <FieldLabel name={config.label} helpText="" complete={false} />
      <div className="py-2">
        <button
          type="button"
          onClick={handleClick}
          disabled={isLoading}
          className="px-4 py-1.5 text-sm font-medium bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isLoading ? 'Posting...' : config.label}
        </button>
        {error && <p className="mt-1 text-xs text-red-400">{error}</p>}
      </div>
      {copyText && (
        <CopyTextDialog text={copyText} onClose={() => setCopyText(null)} />
      )}
    </div>
  );
}
