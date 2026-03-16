import { useState, useCallback, useMemo, useEffect, useRef } from 'react';
import type { AspectField, VideoResponse } from '../../api/types';
import { TextInput } from './TextInput';
import { TextArea } from './TextArea';
import { Toggle } from './Toggle';
import { DateInput } from './DateInput';
import { NumberInput } from './NumberInput';
import { SelectInput } from './SelectInput';
import { ArrayInput } from './ArrayInput';
import { MapInput } from './MapInput';
import { ActionButton, isActionField } from './ActionButton';
import { PublishButton, isPublishField } from './PublishButton';
import { SocialPostButton, isSocialField } from './SocialPostButton';
import { AIGenerateButton } from './AIGenerateButton';
import { RandomTimingButton } from './RandomTimingButton';
import { GenerateAnimationsButton } from './GenerateAnimationsButton';
import { VideoUploadInput } from './VideoUploadInput';
import { FieldLabel } from './FieldLabel';
import { AI_FIELD_CONFIG } from '../../lib/aiFields';
import { ShortsSection } from './ShortsSection';
import { ShortUploadAction } from './ShortUploadAction';
import { ShortPublishAction } from './ShortPublishAction';
import type { Short } from '../../api/types';

interface DynamicFormProps {
  fields: AspectField[];
  video: VideoResponse;
  onSave: (changedFields: Record<string, unknown>) => void;
  onSaveCrossAspect?: (aspect: string, fields: Record<string, unknown>) => void;
  saving?: boolean;
  category?: string;
  videoName?: string;
  aspectKey?: string;
}

/** Resolve a possibly dot-notated path like "sponsorship.amount" from a video object. */
function getFieldValue(video: VideoResponse, fieldName: string): unknown {
  const parts = fieldName.split('.');
  let current: unknown = video;
  for (const part of parts) {
    if (current == null || typeof current !== 'object') return undefined;
    current = (current as Record<string, unknown>)[part];
  }
  return current;
}

export function DynamicForm({ fields, video, onSave, onSaveCrossAspect, saving, category, videoName, aspectKey }: DynamicFormProps) {
  const initialValues = useMemo(() => {
    const vals: Record<string, unknown> = {};
    for (const field of fields) {
      vals[field.fieldName] = getFieldValue(video, field.fieldName) ?? fieldDefault(field);
    }
    return vals;
  }, [fields, video]);

  const [values, setValues] = useState<Record<string, unknown>>(initialValues);

  // Sync form state when server data changes (e.g. after thumbnail upload),
  // but preserve any fields the user has manually edited.
  const prevInitialRef = useRef(initialValues);
  useEffect(() => {
    setValues((prev) => {
      const next = { ...prev };
      let changed = false;
      for (const key of Object.keys(initialValues)) {
        const prevInit = prevInitialRef.current[key];
        const currVal = prev[key];
        // Only update fields the user hasn't manually changed
        const matches =
          typeof currVal === 'object' && currVal !== null
            ? JSON.stringify(currVal) === JSON.stringify(prevInit)
            : currVal === prevInit;
        if (matches && initialValues[key] !== prevInit) {
          next[key] = initialValues[key];
          changed = true;
        }
      }
      prevInitialRef.current = initialValues;
      return changed ? next : prev;
    });
  }, [initialValues]);

  const dirtyFields = useMemo(() => {
    const dirty: Record<string, unknown> = {};
    for (const key of Object.keys(values)) {
      const curr = values[key];
      const init = initialValues[key];
      // Use deep comparison for objects/arrays, reference equality for primitives
      if (typeof curr === 'object' && curr !== null) {
        if (JSON.stringify(curr) !== JSON.stringify(init)) {
          dirty[key] = curr;
        }
      } else if (curr !== init) {
        dirty[key] = curr;
      }
    }
    return dirty;
  }, [values, initialValues]);

  const isDirty = Object.keys(dirtyFields).length > 0;

  const handleChange = useCallback((fieldName: string, value: unknown) => {
    setValues((prev) => ({ ...prev, [fieldName]: value }));
  }, []);

  const handleSave = () => {
    if (!isDirty) return;
    // Only send fields that belong to the current aspect
    const aspectFieldNames = new Set(fields.map((f) => f.fieldName));
    const filtered: Record<string, unknown> = {};
    for (const [key, val] of Object.entries(dirtyFields)) {
      if (aspectFieldNames.has(key)) {
        filtered[key] = val;
      }
    }
    if (Object.keys(filtered).length > 0) onSave(filtered);
  };

  const handleReset = () => {
    setValues(initialValues);
  };

  const sorted = useMemo(
    () => [...fields].sort((a, b) => a.order - b.order),
    [fields],
  );

  return (
    <div>
      <div className="space-y-4">
        {sorted.map((field) => (
          <div key={field.fieldName}>
            {renderField(field, values[field.fieldName], handleChange, isFieldComplete(field, values[field.fieldName]), category, videoName, video)}
            {category && videoName && AI_FIELD_CONFIG[field.fieldName] && (
              <AIGenerateButton
                fieldName={field.fieldName}
                category={category}
                videoName={videoName}
                onApply={(value) => handleChange(field.fieldName, value)}
              />
            )}
            {category && videoName && field.fieldName === 'animations' && (
              <GenerateAnimationsButton
                category={category}
                videoName={videoName}
                onApply={(animations, timecodes) => {
                  handleChange('animations', animations);
                  if (timecodes && onSaveCrossAspect) {
                    onSaveCrossAspect('post-production', { timecodes });
                  }
                }}
              />
            )}
            {category && videoName && field.type === 'date' && field.fieldName === 'date' && (
              <RandomTimingButton
                category={category}
                videoName={videoName}
                onApply={(newDate) => handleChange(field.fieldName, newDate)}
              />
            )}
          </div>
        ))}
      </div>

      {category && videoName && aspectKey === 'post-production' && video.shorts?.length > 0 && (
        <ShortsSection
          shorts={video.shorts as Short[]}
          title="Shorts Upload"
          renderAction={(short) => (
            <ShortUploadAction
              videoName={videoName}
              category={category}
              shortId={short.id}
              driveFileId={short.driveFileId}
            />
          )}
        />
      )}

      {category && videoName && aspectKey === 'publishing' && video.shorts?.length > 0 && (
        <ShortsSection
          shorts={video.shorts as Short[]}
          title="Shorts Publish"
          renderAction={(short) => (
            <ShortPublishAction
              videoName={videoName}
              category={category}
              shortId={short.id}
              driveFileId={short.driveFileId}
              filePath={short.filePath}
              scheduledDate={short.scheduledDate}
              youtubeId={short.youtubeId}
            />
          )}
        />
      )}

      <div className="flex gap-3 mt-6 pt-4 border-t border-gray-700">
        <button
          type="button"
          onClick={handleSave}
          disabled={!isDirty || saving}
          className="px-4 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {saving ? 'Saving...' : 'Save'}
        </button>
        <button
          type="button"
          onClick={handleReset}
          disabled={!isDirty || saving}
          className="px-4 py-1.5 text-sm border border-gray-600 text-gray-300 rounded hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Reset
        </button>
      </div>
    </div>
  );
}

function isFieldComplete(field: AspectField, value: unknown): boolean {
  switch (field.completionCriteria) {
    case 'filled_only':
      if (Array.isArray(value)) return value.length > 0;
      if (value != null && typeof value === 'object') return Object.keys(value).length > 0;
      return typeof value === 'string' ? value.trim().length > 0 : value != null;
    case 'true_only':
      return value === true;
    case 'false_only':
      return value === false;
    case 'no_fixme':
      return typeof value === 'string' && !value.toLowerCase().includes('fixme');
    case 'empty_or_filled':
      return true;
    default:
      if (Array.isArray(value)) return value.length > 0;
      if (value != null && typeof value === 'object') return Object.keys(value).length > 0;
      return typeof value === 'string' ? value.trim().length > 0 : Boolean(value);
  }
}

function toStringValue(value: unknown): string {
  if (value == null) return '';
  if (typeof value === 'string') return value;
  if (typeof value === 'object') return JSON.stringify(value, null, 2);
  return String(value);
}

function fieldDefault(field: AspectField): unknown {
  switch (field.type) {
    case 'boolean':
      return false;
    case 'number':
      return 0;
    case 'array':
      return [];
    case 'map':
      return {};
    default:
      return '';
  }
}

function renderField(
  field: AspectField,
  value: unknown,
  onChange: (fieldName: string, value: unknown) => void,
  complete: boolean,
  category?: string,
  videoName?: string,
  video?: VideoResponse,
) {
  const { fieldName, name, required, uiHints } = field;
  const helpText = uiHints?.helpText;
  const placeholder = uiHints?.placeholder;

  switch (field.type) {
    case 'label':
      if (isPublishField(fieldName) && category && videoName && video) {
        return (
          <PublishButton
            fieldName={fieldName}
            value={toStringValue(value)}
            category={category}
            videoName={videoName}
            video={video}
          />
        );
      }
      return (
        <div>
          <FieldLabel name={name} helpText={helpText} complete={complete} />
          <div className="flex items-center gap-2">
            {value != null && value !== '' && (
              <code className="text-xs text-gray-300 bg-gray-800 px-1 rounded">{toStringValue(value)}</code>
            )}
            {category && videoName && fieldName === 'videoFile' && (
              <VideoUploadInput
                videoName={videoName}
                category={category}
                currentDriveFileId={video?.videoDriveFileId as string | undefined}
              />
            )}
          </div>
        </div>
      );
    case 'boolean':
      if (isActionField(fieldName) && category && videoName) {
        return (
          <ActionButton
            fieldName={fieldName}
            value={Boolean(value)}
            category={category}
            videoName={videoName}
          />
        );
      }
      if (isSocialField(fieldName) && category && videoName) {
        return (
          <SocialPostButton
            fieldName={fieldName}
            value={Boolean(value)}
            category={category}
            videoName={videoName}
          />
        );
      }
      return (
        <Toggle
          name={name}
          fieldName={fieldName}
          value={Boolean(value)}
          onChange={onChange}
          helpText={helpText}
          complete={complete}
        />
      );
    case 'text':
      return (
        <TextArea
          name={name}
          fieldName={fieldName}
          value={toStringValue(value)}
          onChange={onChange}
          placeholder={placeholder}
          required={required}
          helpText={helpText}
          rows={uiHints?.rows}
          complete={complete}
        />
      );
    case 'date':
      return (
        <DateInput
          name={name}
          fieldName={fieldName}
          value={toStringValue(value)}
          onChange={onChange}
          required={required}
          helpText={helpText}
          complete={complete}
        />
      );
    case 'number':
      return (
        <NumberInput
          name={name}
          fieldName={fieldName}
          value={Number(value ?? 0)}
          onChange={onChange}
          required={required}
          helpText={helpText}
          min={field.validationHints?.min}
          max={field.validationHints?.max}
          complete={complete}
        />
      );
    case 'select':
      return (
        <SelectInput
          name={name}
          fieldName={fieldName}
          value={toStringValue(value)}
          onChange={onChange}
          options={uiHints?.options ?? []}
          required={required}
          helpText={helpText}
          placeholder={placeholder}
          complete={complete}
        />
      );
    case 'array':
      return (
        <ArrayInput
          name={name}
          fieldName={fieldName}
          value={(Array.isArray(value) ? value : []) as Record<string, unknown>[]}
          onChange={onChange}
          itemFields={field.itemFields ?? []}
          helpText={helpText}
          complete={complete}
          category={category}
          videoName={videoName}
        />
      );
    case 'map':
      return (
        <MapInput
          name={name}
          fieldName={fieldName}
          value={(value != null && typeof value === 'object' && !Array.isArray(value) ? value : {}) as Record<string, Record<string, unknown>>}
          onChange={onChange}
          itemFields={field.itemFields ?? []}
          mapKeyLabel={field.mapKeyLabel}
          helpText={helpText}
          complete={complete}
        />
      );
    default:
      if (isPublishField(fieldName) && category && videoName && video) {
        return (
          <PublishButton
            fieldName={fieldName}
            value={toStringValue(value)}
            category={category}
            videoName={videoName}
            video={video}
          />
        );
      }
      return (
        <TextInput
          name={name}
          fieldName={fieldName}
          value={toStringValue(value)}
          onChange={onChange}
          placeholder={placeholder}
          required={required}
          helpText={helpText}
          complete={complete}
        />
      );
  }
}
