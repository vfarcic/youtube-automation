import type { Short } from '../../api/types';

interface ShortsSectionProps {
  shorts: Short[];
  renderAction: (short: Short) => React.ReactNode;
  title: string;
}

export function ShortsSection({ shorts, renderAction, title }: ShortsSectionProps) {
  const validShorts = shorts.filter((s) => typeof s.id === 'string' && s.id.trim() !== '');
  if (validShorts.length === 0) return null;

  return (
    <div className="mt-6 pt-4 border-t border-gray-700">
      <h3 className="text-sm font-semibold text-gray-400 mb-3">{title}</h3>
      <div className="space-y-3">
        {validShorts.map((short) => (
          <div key={short.id} className="border border-gray-700 rounded p-3">
            <div className="flex items-center gap-2 mb-2">
              <span className="text-sm font-medium text-gray-300">{short.title || short.id}</span>
              {short.scheduledDate && (
                <span className="text-xs text-gray-500">{short.scheduledDate}</span>
              )}
            </div>
            {renderAction(short)}
          </div>
        ))}
      </div>
    </div>
  );
}
