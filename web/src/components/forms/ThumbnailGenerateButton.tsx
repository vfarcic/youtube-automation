import { useState, useEffect } from 'react';
import { useSuggestTaglineAndIllustrations, useSaveThumbnailConfig, useGenerateThumbnails, useSelectGeneratedThumbnail } from '../../api/hooks';
import { getBlob } from '../../api/client';
import type { ThumbnailGenerateMeta, VideoResponse } from '../../api/types';

function AuthImage({ src, alt, className }: { src: string; alt: string; className?: string }) {
  const [blobUrl, setBlobUrl] = useState<string | null>(null);

  useEffect(() => {
    let revoked = false;
    let url: string | undefined;
    getBlob(src).then((blob) => {
      if (revoked) return;
      url = URL.createObjectURL(blob);
      setBlobUrl(url);
    }).catch(() => {
      // Image load failed — leave blank
    });
    return () => {
      revoked = true;
      if (url) URL.revokeObjectURL(url);
    };
  }, [src]);

  if (!blobUrl) return <div className={className} />;
  return <img src={blobUrl} alt={alt} className={className} loading="lazy" />;
}

interface ThumbnailGenerateButtonProps {
  category: string;
  videoName: string;
  video: VideoResponse;
}

export function ThumbnailGenerateButton({ category, videoName, video }: ThumbnailGenerateButtonProps) {
  const [taglines, setTaglines] = useState<string[]>([]);
  const [illustrations, setIllustrations] = useState<string[]>([]);
  const [selectedTagline, setSelectedTagline] = useState<string | null>(null);
  const [selectedIllustration, setSelectedIllustration] = useState<string | null>(null);
  const [generatedThumbnails, setGeneratedThumbnails] = useState<ThumbnailGenerateMeta[]>([]);
  const [generationErrors, setGenerationErrors] = useState<string[]>([]);
  const [selectingId, setSelectingId] = useState<string | null>(null);
  const [selectSuccess, setSelectSuccess] = useState<string | null>(null);

  const suggestMutation = useSuggestTaglineAndIllustrations();
  const saveMutation = useSaveThumbnailConfig();
  const generateMutation = useGenerateThumbnails();
  const selectMutation = useSelectGeneratedThumbnail();

  const handleSuggest = () => {
    setTaglines([]);
    setIllustrations([]);
    setSelectedTagline(null);
    setSelectedIllustration(null);
    setGeneratedThumbnails([]);
    setGenerationErrors([]);
    setSelectSuccess(null);
    suggestMutation.mutate(
      { category, name: videoName },
      {
        onSuccess: (data) => {
          setTaglines(data.taglines ?? []);
          setIllustrations(data.illustrations ?? []);
        },
      },
    );
  };

  const handleSaveSelection = () => {
    if (!selectedTagline) return;
    saveMutation.mutate(
      {
        videoName,
        category,
        tagline: selectedTagline,
        illustration: selectedIllustration === '__none__' ? '' : (selectedIllustration ?? ''),
      },
      {
        onSuccess: () => {
          setGeneratedThumbnails([]);
          setGenerationErrors([]);
        },
      },
    );
  };

  const handleGenerateThumbnails = () => {
    setGeneratedThumbnails([]);
    setGenerationErrors([]);
    setSelectSuccess(null);
    generateMutation.mutate(
      { category, name: videoName },
      {
        onSuccess: (data) => {
          setGeneratedThumbnails(data.thumbnails ?? []);
          setGenerationErrors(data.errors ?? []);
        },
      },
    );
  };

  const handleSelect = (thumb: ThumbnailGenerateMeta) => {
    const variantIndex = video.thumbnailVariants?.length ?? 0;
    setSelectingId(thumb.id);
    setSelectSuccess(null);
    selectMutation.mutate(
      {
        id: thumb.id,
        category,
        name: videoName,
        variantIndex,
      },
      {
        onSuccess: () => {
          setSelectSuccess(thumb.id);
          setSelectingId(null);
          setGeneratedThumbnails((prev) => prev.filter((t) => t.id !== thumb.id));
        },
        onError: () => {
          setSelectingId(null);
        },
      },
    );
  };

  const showSelection = taglines.length > 0 || illustrations.length > 0 || suggestMutation.isSuccess;
  const canSave = selectedTagline !== null && selectedIllustration !== null;
  const hasStoredTagline = !!(video.tagline);
  const canGenerate = hasStoredTagline || saveMutation.isSuccess;

  // Group thumbnails by provider
  const grouped: Record<string, ThumbnailGenerateMeta[]> = {};
  for (const thumb of generatedThumbnails) {
    if (!grouped[thumb.provider]) grouped[thumb.provider] = [];
    grouped[thumb.provider].push(thumb);
  }

  return (
    <div className="mt-3 p-3 bg-gray-800 rounded border border-gray-700">
      <p className="text-xs text-gray-400 mb-2">AI Thumbnail Generation</p>

      {/* Show stored selections */}
      {video.tagline && (
        <p className="text-xs text-gray-400 mb-1">Tagline: <span className="text-gray-200 font-medium">{video.tagline}</span></p>
      )}
      {video.illustration && (
        <p className="text-xs text-gray-400 mb-1">Illustration: <span className="text-gray-200 font-medium">{video.illustration}</span></p>
      )}

      {/* Step 1: Suggest Tagline & Illustrations */}
      <button
        type="button"
        onClick={handleSuggest}
        disabled={suggestMutation.isPending}
        className="px-3 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {suggestMutation.isPending ? 'Suggesting...' : 'Suggest Tagline & Illustrations'}
      </button>

      {suggestMutation.isError && (
        <p className="text-xs text-red-400 mt-1">{suggestMutation.error.message}</p>
      )}

      {/* Tagline & Illustration selection */}
      {showSelection && (
        <div className="mt-2 space-y-3">
          {/* Tagline selection */}
          {taglines.length > 0 && (
            <div className="space-y-1">
              <p className="text-xs text-gray-400">Select a tagline:</p>
              {taglines.map((tl, i) => (
                <label key={i} className="flex items-start gap-2 text-sm text-gray-200 cursor-pointer">
                  <input
                    type="radio"
                    name="tagline"
                    checked={selectedTagline === tl}
                    onChange={() => setSelectedTagline(tl)}
                    className="mt-0.5 shrink-0"
                  />
                  <span>{tl}</span>
                </label>
              ))}
            </div>
          )}

          {/* Illustration selection */}
          {illustrations.length > 0 && (
            <div className="space-y-1">
              <p className="text-xs text-gray-400">Select an illustration idea:</p>
              {illustrations.map((ill, i) => (
                <label key={i} className="flex items-start gap-2 text-sm text-gray-200 cursor-pointer">
                  <input
                    type="radio"
                    name="illustration"
                    checked={selectedIllustration === ill}
                    onChange={() => setSelectedIllustration(ill)}
                    className="mt-0.5 shrink-0"
                  />
                  <span>{ill}</span>
                </label>
              ))}
              <label className="flex items-start gap-2 text-sm text-gray-400 cursor-pointer">
                <input
                  type="radio"
                  name="illustration"
                  checked={selectedIllustration === '__none__'}
                  onChange={() => setSelectedIllustration('__none__')}
                  className="mt-0.5 shrink-0"
                />
                <span>None (text only)</span>
              </label>
            </div>
          )}

          {/* Save Selection button */}
          <button
            type="button"
            onClick={handleSaveSelection}
            disabled={!canSave || saveMutation.isPending}
            className="px-3 py-1 text-xs bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {saveMutation.isPending ? 'Saving...' : 'Save Selection'}
          </button>
          {saveMutation.isSuccess && (
            <p className="text-xs text-green-400">Selection saved.</p>
          )}
          {saveMutation.isError && (
            <p className="text-xs text-red-400">{saveMutation.error.message}</p>
          )}
        </div>
      )}

      {/* Step 2: Generate Thumbnails */}
      {canGenerate && (
        <div className="mt-2">
          <button
            type="button"
            onClick={handleGenerateThumbnails}
            disabled={generateMutation.isPending}
            className="px-3 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {generateMutation.isPending ? 'Generating...' : 'Generate Thumbnails'}
          </button>
          {generateMutation.isPending && (
            <p className="text-xs text-yellow-400 mt-1">Generating thumbnails... this may take up to 2 minutes.</p>
          )}
          {generateMutation.isError && (
            <p className="text-xs text-red-400 mt-1">{generateMutation.error.message}</p>
          )}
        </div>
      )}

      {/* Generation errors (partial failures) */}
      {generationErrors.length > 0 && (
        <div className="mt-2">
          {generationErrors.map((err, i) => (
            <p key={i} className="text-xs text-yellow-400">{err}</p>
          ))}
        </div>
      )}

      {/* Step 3: Image grid */}
      {generatedThumbnails.length > 0 && (
        <div className="mt-3 space-y-3">
          {Object.entries(grouped).map(([provider, thumbs]) => (
            <div key={provider}>
              <p className="text-xs font-medium text-gray-300 mb-1">{provider}</p>
              <div className="grid grid-cols-2 gap-2">
                {thumbs.map((thumb) => (
                  <div key={thumb.id} className="border border-gray-600 rounded p-2">
                    <AuthImage
                      src={`/api/thumbnails/generated/${encodeURIComponent(thumb.id)}`}
                      alt={`${thumb.provider} - ${thumb.style}`}
                      className="w-full rounded mb-1"
                    />
                    <p className="text-xs text-gray-400 mb-1">{thumb.style}</p>
                    <button
                      type="button"
                      onClick={() => handleSelect(thumb)}
                      disabled={selectingId !== null}
                      className="px-2 py-0.5 text-xs bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      {selectingId === thumb.id ? 'Uploading...' : 'Use This'}
                    </button>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Select success */}
      {selectSuccess && (
        <p className="text-xs text-green-400 mt-2">Thumbnail uploaded to Drive and saved as variant.</p>
      )}

      {/* Select error */}
      {selectMutation.isError && (
        <p className="text-xs text-red-400 mt-2">{selectMutation.error.message}</p>
      )}
    </div>
  );
}
