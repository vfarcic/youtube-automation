import { useState } from 'react';
import { useAITranslate } from '../../api/hooks';

type TranslatedFields = Partial<
  Record<'title' | 'description' | 'tags' | 'timecodes' | 'shortTitles', string>
>;

interface TranslationPanelProps {
  category: string;
  videoName: string;
  onApply: (translatedFields: TranslatedFields) => void;
}

export type { TranslatedFields };

export function TranslationPanel({ category, videoName, onApply }: TranslationPanelProps) {
  const [expanded, setExpanded] = useState(false);
  const [targetLanguage, setTargetLanguage] = useState('');
  const translateMut = useAITranslate();

  const handleTranslate = () => {
    if (!targetLanguage.trim()) return;
    translateMut.mutate({ category, name: videoName, targetLanguage: targetLanguage.trim() });
  };

  const handleApplyAll = () => {
    if (!translateMut.data) return;
    const fields: TranslatedFields = {};
    if (translateMut.data.title) fields.title = translateMut.data.title;
    if (translateMut.data.description) fields.description = translateMut.data.description;
    if (translateMut.data.tags) fields.tags = translateMut.data.tags;
    if (translateMut.data.timecodes) fields.timecodes = translateMut.data.timecodes;
    if (translateMut.data.shortTitles && translateMut.data.shortTitles.length > 0) {
      fields.shortTitles = translateMut.data.shortTitles.join('\n');
    }
    if (Object.keys(fields).length === 0) return;
    onApply(fields);
    translateMut.reset();
    setExpanded(false);
    setTargetLanguage('');
  };

  if (!expanded) {
    return (
      <button
        type="button"
        onClick={() => setExpanded(true)}
        className="px-4 py-1.5 text-sm border border-gray-600 text-gray-300 rounded hover:bg-gray-800"
      >
        Translate Video
      </button>
    );
  }

  return (
    <div className="p-4 bg-gray-800 rounded space-y-3">
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-semibold text-gray-200">Translate Video</h4>
        <button
          type="button"
          onClick={() => { setExpanded(false); translateMut.reset(); }}
          className="text-xs text-gray-500 hover:text-gray-300"
        >
          Close
        </button>
      </div>

      <div className="flex gap-2">
        <label htmlFor="translation-target-language" className="sr-only">
          Target language
        </label>
        <input
          id="translation-target-language"
          type="text"
          value={targetLanguage}
          onChange={(e) => setTargetLanguage(e.target.value)}
          placeholder="Target language (e.g. Spanish)"
          className="flex-1 px-3 py-1.5 text-sm bg-gray-900 border border-gray-700 rounded text-gray-200 placeholder-gray-500"
        />
        <button
          type="button"
          onClick={handleTranslate}
          disabled={!targetLanguage.trim() || translateMut.isPending}
          className="px-4 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
        >
          {translateMut.isPending ? 'Translating...' : 'Translate'}
        </button>
      </div>

      {translateMut.error && (
        <p className="text-xs text-red-400">{translateMut.error.message}</p>
      )}

      {translateMut.data && (
        <div className="space-y-2">
          {translateMut.data.title && (
            <div>
              <p className="text-xs text-gray-500">Title</p>
              <p className="text-sm text-gray-300">{translateMut.data.title}</p>
            </div>
          )}
          {translateMut.data.description && (
            <div>
              <p className="text-xs text-gray-500">Description</p>
              <pre className="text-sm text-gray-300 whitespace-pre-wrap">{translateMut.data.description}</pre>
            </div>
          )}
          {translateMut.data.tags && (
            <div>
              <p className="text-xs text-gray-500">Tags</p>
              <p className="text-sm text-gray-300">{translateMut.data.tags}</p>
            </div>
          )}
          {translateMut.data.timecodes && (
            <div>
              <p className="text-xs text-gray-500">Timecodes</p>
              <pre className="text-sm text-gray-300 whitespace-pre-wrap">{translateMut.data.timecodes}</pre>
            </div>
          )}
          {translateMut.data.shortTitles && translateMut.data.shortTitles.length > 0 && (
            <div>
              <p className="text-xs text-gray-500">Short Titles</p>
              <ul className="text-sm text-gray-300 list-disc list-inside">
                {translateMut.data.shortTitles.map((t, i) => <li key={i}>{t}</li>)}
              </ul>
            </div>
          )}
          <button
            type="button"
            onClick={handleApplyAll}
            className="px-4 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Apply All
          </button>
        </div>
      )}
    </div>
  );
}
