import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useCreateVideo, useCategories } from '../api/hooks';

interface CreateVideoDialogProps {
  open: boolean;
  onClose: () => void;
}

export function CreateVideoDialog({ open, onClose }: CreateVideoDialogProps) {
  const navigate = useNavigate();
  const createVideo = useCreateVideo();
  const { data: categories = [] } = useCategories();
  const [name, setName] = useState('');
  const [category, setCategory] = useState('');
  const [date, setDate] = useState('');
  const [error, setError] = useState('');

  if (!open) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || !category.trim()) {
      setError('Name and category are required.');
      return;
    }
    setError('');
    createVideo.mutate(
      { name: name.trim(), category: category.trim(), date: date || undefined },
      {
        onSuccess: (data) => {
          onClose();
          navigate(`/videos/${encodeURIComponent(data.category)}/${encodeURIComponent(data.name)}`);
        },
        onError: (err) => setError(err.message || 'Failed to create video.'),
      },
    );
  };

  const handleClose = () => {
    setName('');
    setCategory('');
    setDate('');
    setError('');
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
      <div className="bg-gray-800 rounded-lg shadow-lg w-full max-w-md p-6">
        <h3 className="text-lg font-semibold text-gray-100 mb-4">Create Video</h3>
        <form onSubmit={handleSubmit}>
          <div className="space-y-3">
            <div>
              <label htmlFor="cv-name" className="block text-sm font-medium text-gray-300 mb-1">
                Name <span className="text-red-400">*</span>
              </label>
              <input
                id="cv-name"
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="my-video-name"
                className="w-full border border-gray-600 bg-gray-700 text-gray-100 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500 placeholder-gray-500"
              />
            </div>
            <div>
              <label htmlFor="cv-category" className="block text-sm font-medium text-gray-300 mb-1">
                Category <span className="text-red-400">*</span>
              </label>
              <select
                id="cv-category"
                value={category}
                onChange={(e) => setCategory(e.target.value)}
                className="w-full border border-gray-600 bg-gray-700 text-gray-100 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option value="">Select a category</option>
                {categories.map((cat) => (
                  <option key={cat.path} value={cat.name}>
                    {cat.name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label htmlFor="cv-date" className="block text-sm font-medium text-gray-300 mb-1">
                Date <span className="text-xs text-gray-500">(optional)</span>
              </label>
              <input
                id="cv-date"
                type="datetime-local"
                value={date}
                onChange={(e) => setDate(e.target.value)}
                className="w-full border border-gray-600 bg-gray-700 text-gray-100 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
          </div>

          {error && <p className="mt-3 text-sm text-red-400">{error}</p>}

          <div className="flex justify-end gap-3 mt-6">
            <button
              type="button"
              onClick={handleClose}
              className="px-4 py-1.5 text-sm border border-gray-600 text-gray-300 rounded hover:bg-gray-700"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={createVideo.isPending}
              className="px-4 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
            >
              {createVideo.isPending ? 'Creating...' : 'Create'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
