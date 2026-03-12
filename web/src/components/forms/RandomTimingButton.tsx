import { useApplyRandomTiming } from '../../api/hooks';

interface RandomTimingButtonProps {
  category: string;
  videoName: string;
  onApply: (newDate: string) => void;
}

export function RandomTimingButton({ category, videoName, onApply }: RandomTimingButtonProps) {
  const mutation = useApplyRandomTiming();

  const handleClick = () => {
    mutation.mutate(
      { name: videoName, category },
      {
        onSuccess: (data) => {
          // datetime-local needs "YYYY-MM-DDTHH:mm" — strip seconds and Z
          const formatted = data.newDate.replace(/:\d{2}Z$/, '').replace(/Z$/, '');
          onApply(formatted);
        },
      },
    );
  };

  return (
    <div>
      <button
        type="button"
        onClick={handleClick}
        disabled={mutation.isPending}
        className="ml-2 px-2 py-0.5 text-xs bg-gray-700 text-gray-300 rounded hover:bg-gray-600 disabled:opacity-50"
      >
        {mutation.isPending ? 'Applying...' : 'Apply Random Timing'}
      </button>
      {mutation.error && (
        <p className="text-xs text-red-400 mt-1">{mutation.error.message}</p>
      )}
      {mutation.data && (
        <div className="mt-2 p-3 bg-gray-800 rounded space-y-1">
          <p className="text-sm text-gray-300">
            <span className="text-gray-500">Day:</span> {mutation.data.day}
          </p>
          <p className="text-sm text-gray-300">
            <span className="text-gray-500">Time:</span> {mutation.data.time}
          </p>
          <p className="text-sm text-gray-300">
            <span className="text-gray-500">Reasoning:</span> {mutation.data.reasoning}
          </p>
          {mutation.data.syncWarning && (
            <p className="text-xs text-yellow-400 mt-1">{mutation.data.syncWarning}</p>
          )}
        </div>
      )}
    </div>
  );
}
