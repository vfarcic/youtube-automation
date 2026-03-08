import { useState } from 'react';
import { AI_FIELD_CONFIG } from '../../lib/aiFields';
import type { ShortCandidate } from '../../api/types';
import {
  useAITitles,
  useAIDescription,
  useAITags,
  useAIDescriptionTags,
  useAITweets,
  useAIShorts,
} from '../../api/hooks';

interface AIGenerateButtonProps {
  fieldName: string;
  category: string;
  videoName: string;
  onApply: (value: unknown) => void;
}

export function AIGenerateButton({ fieldName, category, videoName, onApply }: AIGenerateButtonProps) {
  const config = AI_FIELD_CONFIG[fieldName];
  if (!config) return null;

  return (
    <AIGenerateButtonInner
      hookType={config.hook}
      label={config.label}
      category={category}
      videoName={videoName}
      onApply={onApply}
    />
  );
}

interface InnerProps {
  hookType: 'titles' | 'description' | 'tags' | 'descriptionTags' | 'tweets' | 'shorts';
  label: string;
  category: string;
  videoName: string;
  onApply: (value: unknown) => void;
}

function AIGenerateButtonInner({ hookType, label, category, videoName, onApply }: InnerProps) {
  const params = { category, name: videoName };

  const titlesMut = useAITitles();
  const descMut = useAIDescription();
  const tagsMut = useAITags();
  const descTagsMut = useAIDescriptionTags();
  const tweetsMut = useAITweets();
  const shortsMut = useAIShorts();

  const [selectedTitles, setSelectedTitles] = useState<Set<number>>(new Set());
  const [selectedTweet, setSelectedTweet] = useState<number | null>(null);
  const [selectedShorts, setSelectedShorts] = useState<Set<number>>(new Set());

  const getMutation = () => {
    switch (hookType) {
      case 'titles': return titlesMut;
      case 'description': return descMut;
      case 'tags': return tagsMut;
      case 'descriptionTags': return descTagsMut;
      case 'tweets': return tweetsMut;
      case 'shorts': return shortsMut;
    }
  };

  const mutation = getMutation();

  const handleGenerate = () => {
    setSelectedTitles(new Set());
    setSelectedTweet(null);
    setSelectedShorts(new Set());
    mutation.mutate(params as never);
  };

  const toggleTitle = (idx: number) => {
    setSelectedTitles((prev) => {
      const next = new Set(prev);
      if (next.has(idx)) next.delete(idx);
      else if (next.size < 3) next.add(idx);
      return next;
    });
  };

  const toggleShort = (idx: number) => {
    setSelectedShorts((prev) => {
      const next = new Set(prev);
      if (next.has(idx)) next.delete(idx);
      else next.add(idx);
      return next;
    });
  };

  const renderResults = () => {
    if (!mutation.data) return null;

    switch (hookType) {
      case 'titles': {
        const data = titlesMut.data!;
        return (
          <div className="mt-2 p-3 bg-gray-800 rounded space-y-2">
            <p className="text-xs text-gray-500">Select up to 3 titles:</p>
            {data.titles.map((title, idx) => (
              <label key={idx} className="flex items-start gap-2 text-sm cursor-pointer">
                <input
                  type="checkbox"
                  checked={selectedTitles.has(idx)}
                  onChange={() => toggleTitle(idx)}
                  className="mt-0.5"
                />
                <span className="text-gray-300">{title}</span>
              </label>
            ))}
            <button
              onClick={() => {
                const titles = [...selectedTitles].sort().map((idx, i) => ({
                  index: i + 1,
                  text: data.titles[idx],
                  watchTimeShare: 0,
                }));
                onApply(titles);
                mutation.reset();
              }}
              disabled={selectedTitles.size === 0}
              className="px-3 py-1 text-sm bg-blue-600 text-white rounded disabled:opacity-50"
            >
              Apply Selected
            </button>
          </div>
        );
      }
      case 'description': {
        const data = descMut.data!;
        return (
          <div className="mt-2 p-3 bg-gray-800 rounded">
            <pre className="text-sm whitespace-pre-wrap text-gray-300 mb-2">{data.description}</pre>
            <button
              onClick={() => { onApply(data.description); mutation.reset(); }}
              className="px-3 py-1 text-sm bg-blue-600 text-white rounded"
            >
              Apply
            </button>
          </div>
        );
      }
      case 'tags': {
        const data = tagsMut.data!;
        return (
          <div className="mt-2 p-3 bg-gray-800 rounded">
            <p className="text-sm text-gray-300 mb-2">{data.tags}</p>
            <button
              onClick={() => { onApply(data.tags); mutation.reset(); }}
              className="px-3 py-1 text-sm bg-blue-600 text-white rounded"
            >
              Apply
            </button>
          </div>
        );
      }
      case 'descriptionTags': {
        const data = descTagsMut.data!;
        return (
          <div className="mt-2 p-3 bg-gray-800 rounded">
            <p className="text-sm text-gray-300 mb-2">{data.descriptionTags}</p>
            <button
              onClick={() => { onApply(data.descriptionTags); mutation.reset(); }}
              className="px-3 py-1 text-sm bg-blue-600 text-white rounded"
            >
              Apply
            </button>
          </div>
        );
      }
      case 'tweets': {
        const data = tweetsMut.data!;
        return (
          <div className="mt-2 p-3 bg-gray-800 rounded space-y-2">
            <p className="text-xs text-gray-500">Select one tweet:</p>
            {data.tweets.map((tweet, idx) => (
              <label key={idx} className="flex items-start gap-2 text-sm cursor-pointer">
                <input
                  type="radio"
                  name={`tweet-select-${fieldNameForRadio}`}
                  checked={selectedTweet === idx}
                  onChange={() => setSelectedTweet(idx)}
                  className="mt-0.5"
                />
                <span className="text-gray-300">{tweet}</span>
              </label>
            ))}
            <button
              onClick={() => {
                if (selectedTweet !== null) {
                  onApply(data.tweets[selectedTweet] + '\n\n[YOUTUBE]');
                  mutation.reset();
                }
              }}
              disabled={selectedTweet === null}
              className="px-3 py-1 text-sm bg-blue-600 text-white rounded disabled:opacity-50"
            >
              Apply Selected
            </button>
          </div>
        );
      }
      case 'shorts': {
        const data = shortsMut.data!;
        return (
          <div className="mt-2 p-3 bg-gray-800 rounded space-y-2">
            <p className="text-xs text-gray-500">Select shorts to keep:</p>
            {data.candidates.map((candidate: ShortCandidate, idx: number) => (
              <label key={candidate.id} className="block p-2 border border-gray-700 rounded cursor-pointer hover:bg-gray-700">
                <div className="flex items-start gap-2">
                  <input
                    type="checkbox"
                    checked={selectedShorts.has(idx)}
                    onChange={() => toggleShort(idx)}
                    className="mt-1"
                  />
                  <div>
                    <p className="font-medium text-sm text-white">{candidate.title}</p>
                    <p className="text-sm text-gray-200 mt-1 whitespace-pre-wrap">{candidate.text}</p>
                    <p className="text-xs text-gray-500 mt-1 italic">{candidate.rationale}</p>
                  </div>
                </div>
              </label>
            ))}
            <button
              onClick={() => {
                const shorts = [...selectedShorts].sort().map((idx) => {
                  const c = data.candidates[idx];
                  return { id: c.id, title: c.title, text: c.text, filePath: '', scheduledDate: '', youtubeId: '' };
                });
                onApply(shorts);
                mutation.reset();
              }}
              disabled={selectedShorts.size === 0}
              className="px-3 py-1 text-sm bg-blue-600 text-white rounded disabled:opacity-50"
            >
              Apply Selected
            </button>
          </div>
        );
      }
    }
  };

  // Unique name for radio group to avoid collisions
  const fieldNameForRadio = `inline-${hookType}`;

  return (
    <div>
      <button
        type="button"
        onClick={handleGenerate}
        disabled={mutation.isPending}
        className="ml-2 px-2 py-0.5 text-xs bg-gray-700 text-gray-300 rounded hover:bg-gray-600 disabled:opacity-50"
      >
        {mutation.isPending ? 'Generating...' : label}
      </button>
      {mutation.error && (
        <p className="text-xs text-red-400 mt-1">{mutation.error.message}</p>
      )}
      {renderResults()}
    </div>
  );
}
