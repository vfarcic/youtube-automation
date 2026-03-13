import { useState } from 'react';
import { useAMAGenerate, useAMAApply } from '../api/hooks';

export function AskMeAnything() {
  const [videoId, setVideoId] = useState('');
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [tags, setTags] = useState('');
  const [timecodes, setTimecodes] = useState('');
  const [status, setStatus] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  const generateMut = useAMAGenerate();
  const applyMut = useAMAApply();

  const handleGenerate = () => {
    if (!videoId.trim()) return;
    setStatus(null);
    generateMut.mutate(
      { videoId: videoId.trim() },
      {
        onSuccess: (data) => {
          setTitle(data.title);
          setDescription(data.description);
          setTags(data.tags);
          setTimecodes(data.timecodes);
          setStatus({ type: 'success', text: 'Content generated. Review and edit, then apply to YouTube.' });
        },
        onError: (err) => setStatus({ type: 'error', text: err.message || 'Failed to generate content.' }),
      },
    );
  };

  const handleApply = () => {
    if (!videoId.trim()) return;
    setStatus(null);
    applyMut.mutate(
      { videoId: videoId.trim(), title, description, tags, timecodes },
      {
        onSuccess: () => setStatus({ type: 'success', text: 'Video updated on YouTube!' }),
        onError: (err) => setStatus({ type: 'error', text: err.message || 'Failed to apply to YouTube.' }),
      },
    );
  };

  const hasContent = title || description || tags || timecodes;

  return (
    <div>
      <h2 className="text-xl font-bold text-gray-100 mb-4">Ask Me Anything</h2>
      <p className="text-sm text-gray-400 mb-6">
        Enter a YouTube Video ID, generate AI content from the transcript, then apply it to the video.
      </p>

      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1">YouTube Video ID</label>
          <div className="flex gap-2">
            <input
              type="text"
              value={videoId}
              onChange={(e) => setVideoId(e.target.value)}
              placeholder="e.g., dQw4w9WgXcQ"
              className="flex-1 px-3 py-1.5 text-sm bg-gray-900 border border-gray-700 rounded text-gray-200 placeholder-gray-500"
            />
            <button
              type="button"
              onClick={handleGenerate}
              disabled={!videoId.trim() || generateMut.isPending}
              className="px-4 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
            >
              {generateMut.isPending ? 'Generating...' : 'Generate with AI'}
            </button>
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1">Title</label>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Video title"
            className="w-full px-3 py-1.5 text-sm bg-gray-900 border border-gray-700 rounded text-gray-200 placeholder-gray-500"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1">Description</label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Video description"
            rows={5}
            className="w-full px-3 py-1.5 text-sm bg-gray-900 border border-gray-700 rounded text-gray-200 placeholder-gray-500 resize-y"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1">Tags</label>
          <textarea
            value={tags}
            onChange={(e) => setTags(e.target.value)}
            placeholder="Comma-separated tags (max 450 characters)"
            rows={2}
            maxLength={450}
            className="w-full px-3 py-1.5 text-sm bg-gray-900 border border-gray-700 rounded text-gray-200 placeholder-gray-500 resize-y"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1">Timecodes</label>
          <textarea
            value={timecodes}
            onChange={(e) => setTimecodes(e.target.value)}
            placeholder="Timestamped Q&A segments"
            rows={10}
            className="w-full px-3 py-1.5 text-sm bg-gray-900 border border-gray-700 rounded text-gray-200 placeholder-gray-500 resize-y"
          />
        </div>

        <div className="pt-2">
          <button
            type="button"
            onClick={handleApply}
            disabled={!videoId.trim() || !hasContent || applyMut.isPending}
            className="px-4 py-1.5 text-sm bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50"
          >
            {applyMut.isPending ? 'Applying...' : 'Apply to YouTube'}
          </button>
        </div>

        {status && (
          <p className={`text-sm ${status.type === 'success' ? 'text-green-400' : 'text-red-400'}`}>
            {status.text}
          </p>
        )}
      </div>
    </div>
  );
}
