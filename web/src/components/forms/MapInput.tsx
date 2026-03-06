import { useState } from 'react';
import { FieldLabel } from './FieldLabel';
import type { ItemField } from '../../api/types';

interface MapInputProps {
  name: string;
  fieldName: string;
  value: Record<string, Record<string, unknown>>;
  onChange: (fieldName: string, value: Record<string, Record<string, unknown>>) => void;
  itemFields: ItemField[];
  mapKeyLabel?: string;
  helpText?: string;
  complete?: boolean;
}

export function MapInput({
  name,
  fieldName,
  value,
  onChange,
  itemFields,
  mapKeyLabel = 'Key',
  helpText,
  complete,
}: MapInputProps) {
  const entries = value && typeof value === 'object' ? value : {};
  const keys = Object.keys(entries);
  const [newKey, setNewKey] = useState('');

  const handleEntryChange = (key: string, subField: string, subValue: unknown) => {
    const updated = {
      ...entries,
      [key]: { ...entries[key], [subField]: subValue },
    };
    onChange(fieldName, updated);
  };

  const handleAdd = () => {
    const trimmed = newKey.trim();
    if (!trimmed || entries[trimmed]) return;
    const empty: Record<string, unknown> = {};
    for (const f of itemFields) {
      empty[f.fieldName] = f.type === 'number' ? 0 : f.type === 'boolean' ? false : '';
    }
    onChange(fieldName, { ...entries, [trimmed]: empty });
    setNewKey('');
  };

  const handleRemove = (key: string) => {
    const { [key]: _, ...rest } = entries;
    onChange(fieldName, rest);
  };

  return (
    <div>
      <FieldLabel name={name} helpText={helpText} complete={complete} />
      <div className="space-y-3">
        {keys.map((key) => (
          <div key={key} className="border border-gray-700 rounded p-3">
            <div className="flex justify-between items-center mb-2">
              <span className="text-xs font-medium text-gray-400">{mapKeyLabel}: {key}</span>
              <button
                type="button"
                onClick={() => handleRemove(key)}
                className="text-xs text-red-500 hover:text-red-700"
                aria-label={`Remove entry ${key}`}
              >
                Remove
              </button>
            </div>
            <div className="space-y-2">
              {itemFields.map((subField) => (
                <div key={subField.fieldName}>
                  <label className="block text-xs text-gray-400 mb-0.5">{subField.name}</label>
                  {subField.type === 'number' ? (
                    <input
                      type="number"
                      value={Number(entries[key]?.[subField.fieldName] ?? 0)}
                      onChange={(e) =>
                        handleEntryChange(key, subField.fieldName, Number(e.target.value))
                      }
                      className="w-full border border-gray-600 bg-gray-800 text-gray-100 rounded px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
                    />
                  ) : (
                    <input
                      type="text"
                      value={String(entries[key]?.[subField.fieldName] ?? '')}
                      onChange={(e) =>
                        handleEntryChange(key, subField.fieldName, e.target.value)
                      }
                      className="w-full border border-gray-600 bg-gray-800 text-gray-100 rounded px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
                    />
                  )}
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
      <div className="flex gap-2 mt-2">
        <input
          type="text"
          value={newKey}
          onChange={(e) => setNewKey(e.target.value)}
          placeholder={mapKeyLabel}
          className="border border-gray-600 bg-gray-800 text-gray-100 rounded px-2 py-1 text-xs focus:outline-none focus:ring-1 focus:ring-blue-500"
          aria-label={`New ${mapKeyLabel}`}
        />
        <button
          type="button"
          onClick={handleAdd}
          disabled={!newKey.trim()}
          className="px-3 py-1 text-xs border border-dashed border-gray-600 text-gray-400 rounded hover:border-blue-400 hover:text-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          + Add Entry
        </button>
      </div>
    </div>
  );
}
