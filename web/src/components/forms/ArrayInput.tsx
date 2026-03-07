import { FieldLabel } from './FieldLabel';
import { FileUploadInput } from './FileUploadInput';
import type { ItemField } from '../../api/types';

interface ArrayInputProps {
  name: string;
  fieldName: string;
  value: Record<string, unknown>[];
  onChange: (fieldName: string, value: Record<string, unknown>[]) => void;
  itemFields: ItemField[];
  helpText?: string;
  complete?: boolean;
  category?: string;
  videoName?: string;
}

export function ArrayInput({
  name,
  fieldName,
  value,
  onChange,
  itemFields,
  helpText,
  complete,
  category,
  videoName,
}: ArrayInputProps) {
  const items = Array.isArray(value) ? value : [];
  const isSingleField = itemFields.length === 1;
  const isThumbnailVariants = fieldName === 'thumbnailVariants';
  const isUploadOnly = isThumbnailVariants && itemFields.length === 0;

  const handleItemChange = (index: number, subField: string, subValue: unknown) => {
    const updated = items.map((item, i) =>
      i === index ? { ...item, [subField]: subValue } : item,
    );
    onChange(fieldName, updated);
  };

  const handleAdd = () => {
    if (isUploadOnly) {
      onChange(fieldName, [...items, {}]);
      return;
    }
    const empty: Record<string, unknown> = {};
    for (const f of itemFields) {
      empty[f.fieldName] = f.type === 'number' ? 0 : f.type === 'boolean' ? false : '';
    }
    onChange(fieldName, [...items, empty]);
  };

  const handleRemove = (index: number) => {
    onChange(fieldName, items.filter((_, i) => i !== index));
  };

  return (
    <div>
      <FieldLabel name={name} helpText={helpText} complete={complete} />
      <div className="space-y-2">
        {items.map((item, index) =>
          isUploadOnly && category && videoName ? (
            <div key={index} className="flex items-center gap-2">
              <span className="text-xs text-gray-400 shrink-0">{item.index ? `#${item.index}` : `Variant ${index + 1}`}</span>
              {item.driveFileId != null && item.driveFileId !== '' && (
                <code className="text-xs text-gray-300 bg-gray-800 px-1 rounded">{String(item.driveFileId)}</code>
              )}
              <FileUploadInput
                videoName={videoName}
                category={category}
                variantIndex={index}
                currentDriveFileId={item.driveFileId as string | undefined}
              />
              <button
                type="button"
                onClick={() => handleRemove(index)}
                className="text-xs text-red-500 hover:text-red-700 shrink-0"
                aria-label={`Remove item ${index + 1}`}
              >
                Remove
              </button>
            </div>
          ) : isSingleField ? (
            <div key={index}>
              <div className="flex gap-2 items-center">
                {itemFields[0].type === 'number' ? (
                  <input
                    type="number"
                    value={Number(item[itemFields[0].fieldName] ?? 0)}
                    onChange={(e) =>
                      handleItemChange(index, itemFields[0].fieldName, Number(e.target.value))
                    }
                    className="flex-1 border border-gray-600 bg-gray-800 text-gray-100 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
                  />
                ) : (
                  <input
                    type="text"
                    value={String(item[itemFields[0].fieldName] ?? '')}
                    onChange={(e) =>
                      handleItemChange(index, itemFields[0].fieldName, e.target.value)
                    }
                    className="flex-1 border border-gray-600 bg-gray-800 text-gray-100 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
                  />
                )}
                <button
                  type="button"
                  onClick={() => handleRemove(index)}
                  className="text-xs text-red-500 hover:text-red-700 shrink-0"
                  aria-label={`Remove item ${index + 1}`}
                >
                  Remove
                </button>
              </div>
              {isThumbnailVariants && category && videoName && (
                <FileUploadInput
                  videoName={videoName}
                  category={category}
                  variantIndex={index}
                  currentDriveFileId={item.driveFileId as string | undefined}
                />
              )}
            </div>
          ) : (
            <div key={index} className="border border-gray-700 rounded p-3">
              <div className="flex justify-between items-center mb-2">
                <span className="text-xs font-medium text-gray-400">Item {index + 1}</span>
                <button
                  type="button"
                  onClick={() => handleRemove(index)}
                  className="text-xs text-red-500 hover:text-red-700"
                  aria-label={`Remove item ${index + 1}`}
                >
                  Remove
                </button>
              </div>
              <div className="space-y-2">
                {itemFields.map((subField) => (
                  <div key={subField.fieldName}>
                    <label className="block text-xs text-gray-400 mb-0.5">{subField.name}</label>
                    {subField.type === 'label' ? (
                      <code className="block text-sm text-gray-300 bg-gray-800 px-2 py-1 rounded">{String(item[subField.fieldName] ?? '')}</code>
                    ) : subField.type === 'number' ? (
                      <input
                        type="number"
                        value={Number(item[subField.fieldName] ?? 0)}
                        onChange={(e) =>
                          handleItemChange(index, subField.fieldName, Number(e.target.value))
                        }
                        className="w-full border border-gray-600 bg-gray-800 text-gray-100 rounded px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
                      />
                    ) : (
                      <input
                        type="text"
                        value={String(item[subField.fieldName] ?? '')}
                        onChange={(e) =>
                          handleItemChange(index, subField.fieldName, e.target.value)
                        }
                        className="w-full border border-gray-600 bg-gray-800 text-gray-100 rounded px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
                      />
                    )}
                    {isThumbnailVariants && subField.fieldName === 'path' && category && videoName && (
                      <FileUploadInput
                        videoName={videoName}
                        category={category}
                        variantIndex={index}
                        currentDriveFileId={item.driveFileId as string | undefined}
                      />
                    )}
                  </div>
                ))}
              </div>
            </div>
          ),
        )}
      </div>
      <button
        type="button"
        onClick={handleAdd}
        className="mt-2 px-3 py-1 text-xs border border-dashed border-gray-600 text-gray-400 rounded hover:border-blue-400 hover:text-blue-600"
      >
        + Add Item
      </button>
    </div>
  );
}
