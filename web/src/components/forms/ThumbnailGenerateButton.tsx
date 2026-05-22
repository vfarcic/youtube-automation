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
  const [photoRealisticSubject, setPhotoRealisticSubject] = useState<string>(video.photoRealisticSubject ?? '');
  const [generatedThumbnails, setGeneratedThumbnails] = useState<ThumbnailGenerateMeta[]>([]);
  const [generationErrors, setGenerationErrors] = useState<string[]>([]);
  const [selectingId, setSelectingId] = useState<string | null>(null);
  const [selectSuccess, setSelectSuccess] = useState<string | null>(null);
  const [photoRealisticVariantPresent, setPhotoRealisticVariantPresent] = useState<boolean>(false);

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
        photoRealisticSubject,
      },
      {
        onSuccess: () => {
          setGeneratedThumbnails([]);
          setGenerationErrors([]);
          setPhotoRealisticVariantPresent(false);
        },
      },
    );
  };

  // latestSavedTagline / latestSavedIllustration / latestSavedSubject expose
  // the most recently persisted values, preferring the response from the
  // last save over the parent's video prop. The parent refetches via
  // react-query invalidation after a save, but until that refetch resolves
  // the video prop is stale. Reading saveMutation.data first guarantees the
  // auto-save chain (Save Selection → type subject → Generate) sees the
  // values the user actually saved, not the empty values from the original
  // video document.
  const latestSavedTagline = saveMutation.data?.tagline ?? video.tagline ?? '';
  const latestSavedIllustration = saveMutation.data?.illustration ?? video.illustration ?? '';
  const latestSavedSubject = saveMutation.data?.photoRealisticSubject ?? video.photoRealisticSubject ?? '';

  const handleSavePhotoRealisticSubject = () => {
    // Persist the subject without resetting the tagline/illustration. The
    // backend requires tagline; use whatever the video already has (this
    // button is only enabled when the video has a stored tagline).
    if (!latestSavedTagline) return;
    saveMutation.mutate(
      {
        videoName,
        category,
        tagline: latestSavedTagline,
        illustration: latestSavedIllustration,
        photoRealisticSubject,
      },
    );
  };

  const handleGenerateThumbnails = async () => {
    setGeneratedThumbnails([]);
    setGenerationErrors([]);
    setSelectSuccess(null);
    setPhotoRealisticVariantPresent(false);

    // PRD 401 must-have: "Manual subject override is honored over AI inference."
    // The backend reads only from the stored video state, so a value typed
    // into the input but never saved would be silently discarded and the AI
    // fallback would run instead. Persist the typed value first when it
    // differs from the stored value, and abort generation on save failure
    // so the user sees the error rather than getting an AI fallback they
    // didn't ask for.
    //
    // Use latestSavedTagline/Illustration (not the video props directly) so
    // an unsaved-tagline → Save Selection → type-subject → Generate flow
    // sends the just-saved tagline rather than the parent's stale empty one.
    if (photoRealisticSubject !== latestSavedSubject) {
      if (!latestSavedTagline) {
        // Save requires a tagline; without one we cannot persist the subject.
        // Fall through to generate (the backend will treat the subject as
        // empty and decide on AI fallback). This path is unreachable from
        // the UI because canGenerate gates the Generate button.
      } else {
        try {
          await saveMutation.mutateAsync({
            videoName,
            category,
            tagline: latestSavedTagline,
            illustration: latestSavedIllustration,
            photoRealisticSubject,
          });
        } catch {
          // saveMutation.error is surfaced inline by the existing JSX block.
          return;
        }
      }
    }

    try {
      const data = await generateMutation.mutateAsync({ category, name: videoName });
      const thumbs = data.thumbnails ?? [];
      setGeneratedThumbnails(thumbs);
      setGenerationErrors(data.errors ?? []);
      setPhotoRealisticVariantPresent(thumbs.some((t) => t.style === 'photorealistic'));
    } catch {
      // generateMutation.error is surfaced inline by the existing JSX block.
    }
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
      {video.photoRealisticSubject && (
        <p className="text-xs text-gray-400 mb-1">Photo-realistic subject: <span className="text-gray-200 font-medium">{video.photoRealisticSubject}</span></p>
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

      {/* Photo-realistic subject input (PRD 401 M4): a third subject input
          alongside tagline/illustration. Empty input → backend AI-suggests
          on generate. User input takes precedence over AI inference. */}
      {canGenerate && (
        <div className="mt-3 space-y-1">
          <label htmlFor="photoRealisticSubject" className="block text-xs text-gray-400">
            Photo-realistic subject (optional)
          </label>
          <input
            id="photoRealisticSubject"
            type="text"
            value={photoRealisticSubject}
            onChange={(e) => setPhotoRealisticSubject(e.target.value)}
            placeholder="e.g. a small white rabbit holding a checklist"
            maxLength={200}
            className="w-full px-2 py-1 text-sm bg-gray-900 text-gray-200 border border-gray-600 rounded focus:outline-none focus:border-blue-500"
          />
          <p className="text-xs text-gray-500">
            Leave empty for an AI-suggested subject. Anything you type takes precedence.
          </p>
          {photoRealisticSubject !== latestSavedSubject && (
            <button
              type="button"
              onClick={handleSavePhotoRealisticSubject}
              disabled={saveMutation.isPending}
              className="mt-1 px-3 py-1 text-xs bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {saveMutation.isPending ? 'Saving...' : 'Save subject'}
            </button>
          )}
          {/* Surface auto-save failure from the Generate-with-unsaved-subject
              flow (PRD must-have: manual override is honored — if we cannot
              persist the typed value we abort generate and show why, rather
              than silently falling back to AI inference). */}
          {saveMutation.isError && !showSelection && (
            <p className="text-xs text-red-400 mt-1">{saveMutation.error.message}</p>
          )}
        </div>
      )}

      {/* Step 2: Generate Thumbnails */}
      {canGenerate && (
        <div className="mt-2">
          <button
            type="button"
            onClick={handleGenerateThumbnails}
            // Disable during BOTH the auto-save and the generate phase so
            // double-clicks cannot kick off duplicate save/generate chains.
            disabled={generateMutation.isPending || saveMutation.isPending}
            className="px-3 py-1 text-xs bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {generateMutation.isPending
              ? 'Generating...'
              : saveMutation.isPending
                ? 'Saving...'
                : 'Generate Thumbnails'}
          </button>
          {generateMutation.isPending && (
            <p className="text-xs text-yellow-400 mt-1">Generating thumbnails... this may take up to 5 minutes.</p>
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

      {/* Photo-realistic variant skipped notice (PRD 401 M4): non-blocking
          inline message when generation produced only the two B&W variants
          (e.g., subject empty AND AI suggestion unavailable). */}
      {generatedThumbnails.length > 0 && !photoRealisticVariantPresent && (
        <p className="text-xs text-gray-400 mt-2">
          Photo-realistic variant skipped — provide a subject to enable.
        </p>
      )}

      {/* Step 3: Image grid */}
      {generatedThumbnails.length > 0 && (
        <div className="mt-3 space-y-3">
          {Object.entries(grouped).map(([provider, thumbs]) => (
            <div key={provider}>
              <p className="text-xs font-medium text-gray-300 mb-1">{provider}</p>
              <div className={`grid gap-2 ${thumbs.length >= 3 ? 'grid-cols-3' : 'grid-cols-2'}`}>
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
