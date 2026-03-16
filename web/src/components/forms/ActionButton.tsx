import { useState } from 'react';
import { useRequestThumbnail, useRequestEdit, useNotifySponsors } from '../../api/hooks';

/** Field names that should render as action buttons instead of toggles. */
const ACTION_FIELDS: Record<string, { label: string; sentLabel: string }> = {
  requestThumbnail: { label: 'Request Thumbnail', sentLabel: 'Thumbnail Requested' },
  requestEdit: { label: 'Request Edit', sentLabel: 'Edit Requested' },
  notifiedSponsors: { label: 'Notify Sponsors', sentLabel: 'Sponsors Notified' },
};

/** Returns true if a field name should be rendered as an ActionButton. */
export function isActionField(fieldName: string): boolean {
  return fieldName in ACTION_FIELDS;
}

interface ActionButtonProps {
  fieldName: keyof typeof ACTION_FIELDS;
  value: boolean;
  category: string;
  videoName: string;
}

export function ActionButton({ fieldName, value, category, videoName }: ActionButtonProps) {
  const config = ACTION_FIELDS[fieldName];
  const mutationMap = {
    requestThumbnail: useRequestThumbnail(),
    requestEdit: useRequestEdit(),
    notifiedSponsors: useNotifySponsors(),
  } as const;
  const [error, setError] = useState<string | null>(null);

  const mutation = mutationMap[fieldName];
  const isLoading = mutation.isPending;

  const handleClick = () => {
    setError(null);
    mutation.mutate(
      { name: videoName, category },
      {
        onError: (err) => setError(err.message),
        onSuccess: (data) => {
          if (data.emailError) {
            setError(data.emailError);
          }
        },
      },
    );
  };

  if (value) {
    return (
      <div className="flex items-center gap-2 py-2">
        <span className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-green-400 bg-green-900/30 border border-green-700 rounded">
          {config.sentLabel}
        </span>
      </div>
    );
  }

  return (
    <div className="py-2">
      <button
        type="button"
        onClick={handleClick}
        disabled={isLoading}
        className="px-4 py-1.5 text-sm font-medium bg-amber-600 text-white rounded hover:bg-amber-700 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {isLoading ? 'Sending...' : config.label}
      </button>
      {error && (
        <p className="mt-1 text-xs text-red-400">{error}</p>
      )}
    </div>
  );
}
