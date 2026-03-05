import type { ProgressInfo } from '../api/types';
import { progressPercent } from '../lib/utils';

interface ProgressBarProps {
  progress: ProgressInfo;
  color?: string;
  showLabel?: boolean;
}

export function ProgressBar({
  progress,
  color = 'bg-blue-500',
  showLabel = true,
}: ProgressBarProps) {
  const pct = progressPercent(progress);
  return (
    <div className="flex items-center gap-2">
      <div className="h-2 flex-1 rounded-full bg-gray-200">
        <div
          className={`h-2 rounded-full ${color}`}
          style={{ width: `${pct}%` }}
        />
      </div>
      {showLabel && (
        <span className="text-xs text-gray-500 w-12 text-right">
          {progress.completed}/{progress.total}
        </span>
      )}
    </div>
  );
}
