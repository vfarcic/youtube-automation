import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { get, patch, post, del, uploadFile, uploadFileWithProgress } from './client';
import type {
  PhaseInfo,
  VideoListItem,
  VideoResponse,
  OverallProgressResponse,
  AspectsResponse,
  CreateVideoRequest,
  Category,
  AITitlesResponse,
  AIDescriptionResponse,
  AITagsResponse,
  AITweetsResponse,
  AIDescriptionTagsResponse,
  AIShortsResponse,
  AIThumbnailsResponse,
  AITranslateResponse,
  AIAMAContentResponse,
  ActionResponse,
  PublishYouTubeResponse,
  PublishThumbnailResponse,
  PublishShortResponse,
  PublishHugoResponse,
  TranscriptResponse,
  MetadataResponse,
  SocialPostResponse,
  AnalyzeTitlesResponse,
  ApplyTitlesResponse,
  ApplyRandomTimingResponse,
  AnimationsResponse,
  AMAGenerateResponse,
  AMAApplyResponse,
} from './types';

export function usePhases() {
  return useQuery<PhaseInfo[]>({
    queryKey: ['phases'],
    queryFn: () => get<PhaseInfo[]>('/api/videos/phases'),
  });
}

export function useVideosList(phase?: number) {
  return useQuery<VideoListItem[]>({
    queryKey: ['videosList', phase],
    queryFn: () =>
      get<VideoListItem[]>(
        `/api/videos/list${phase !== undefined ? `?phase=${phase}` : ''}`,
      ),
    enabled: phase !== undefined,
  });
}

export function useSearchVideos(query: string) {
  return useQuery<VideoListItem[]>({
    queryKey: ['searchVideos', query],
    queryFn: () =>
      get<VideoListItem[]>(`/api/videos/search?q=${encodeURIComponent(query)}`),
    enabled: query.length > 0,
  });
}

export function useVideo(name?: string, category?: string) {
  return useQuery<VideoResponse>({
    queryKey: ['video', name, category],
    queryFn: () =>
      get<VideoResponse>(
        `/api/videos/${encodeURIComponent(name!)}?category=${encodeURIComponent(category!)}`,
      ),
    enabled: !!name && !!category,
  });
}

export function useVideoProgress(name?: string, category?: string) {
  return useQuery<OverallProgressResponse>({
    queryKey: ['videoProgress', name, category],
    queryFn: () =>
      get<OverallProgressResponse>(
        `/api/videos/${encodeURIComponent(name!)}/progress?category=${encodeURIComponent(category!)}`,
      ),
    enabled: !!name && !!category,
  });
}

export function useAspects() {
  return useQuery<AspectsResponse>({
    queryKey: ['aspects'],
    queryFn: () => get<AspectsResponse>('/api/aspects'),
    staleTime: 5 * 60 * 1000,
  });
}

export function usePatchVideo() {
  const qc = useQueryClient();
  return useMutation<
    VideoResponse,
    Error,
    { name: string; category: string; aspect: string; fields: Record<string, unknown> }
  >({
    mutationFn: ({ name, category, aspect, fields }) =>
      patch<VideoResponse>(
        `/api/videos/${encodeURIComponent(name)}?category=${encodeURIComponent(category)}&aspect=${encodeURIComponent(aspect)}`,
        fields,
      ),
    onSuccess: (_data, { name, category }) => {
      qc.invalidateQueries({ queryKey: ['video', name, category] });
      qc.invalidateQueries({ queryKey: ['videoProgress', name, category] });
      qc.invalidateQueries({ queryKey: ['videosList'] });
      qc.invalidateQueries({ queryKey: ['phases'] });
    },
  });
}

export function useCategories() {
  return useQuery<Category[]>({
    queryKey: ['categories'],
    queryFn: () => get<Category[]>('/api/categories'),
  });
}

export function useCreateVideo() {
  const qc = useQueryClient();
  return useMutation<VideoResponse, Error, CreateVideoRequest>({
    mutationFn: (body) => post<VideoResponse>('/api/videos', body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['videosList'] });
      qc.invalidateQueries({ queryKey: ['phases'] });
    },
  });
}

export function useDeleteVideo() {
  const qc = useQueryClient();
  return useMutation<void, Error, { name: string; category: string }>({
    mutationFn: ({ name, category }) =>
      del(
        `/api/videos/${encodeURIComponent(name)}?category=${encodeURIComponent(category)}`,
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['videosList'] });
      qc.invalidateQueries({ queryKey: ['phases'] });
    },
  });
}

// --- AI Mutation Hooks ---

export function useAITitles() {
  return useMutation<AITitlesResponse, Error, { category: string; name: string }>({
    mutationFn: ({ category, name }) =>
      post<AITitlesResponse>(`/api/ai/titles/${encodeURIComponent(category)}/${encodeURIComponent(name)}`, {}),
  });
}

export function useAIDescription() {
  return useMutation<AIDescriptionResponse, Error, { category: string; name: string }>({
    mutationFn: ({ category, name }) =>
      post<AIDescriptionResponse>(`/api/ai/description/${encodeURIComponent(category)}/${encodeURIComponent(name)}`, {}),
  });
}

export function useAITags() {
  return useMutation<AITagsResponse, Error, { category: string; name: string }>({
    mutationFn: ({ category, name }) =>
      post<AITagsResponse>(`/api/ai/tags/${encodeURIComponent(category)}/${encodeURIComponent(name)}`, {}),
  });
}

export function useAITweets() {
  return useMutation<AITweetsResponse, Error, { category: string; name: string }>({
    mutationFn: ({ category, name }) =>
      post<AITweetsResponse>(`/api/ai/tweets/${encodeURIComponent(category)}/${encodeURIComponent(name)}`, {}),
  });
}

export function useAIDescriptionTags() {
  return useMutation<AIDescriptionTagsResponse, Error, { category: string; name: string }>({
    mutationFn: ({ category, name }) =>
      post<AIDescriptionTagsResponse>(`/api/ai/description-tags/${encodeURIComponent(category)}/${encodeURIComponent(name)}`, {}),
  });
}

export function useAIShorts() {
  return useMutation<AIShortsResponse, Error, { category: string; name: string }>({
    mutationFn: ({ category, name }) =>
      post<AIShortsResponse>(`/api/ai/shorts/${encodeURIComponent(category)}/${encodeURIComponent(name)}`, {}),
  });
}

export function useAIThumbnails() {
  return useMutation<AIThumbnailsResponse, Error, { imagePath: string }>({
    mutationFn: (body) => post<AIThumbnailsResponse>('/api/ai/thumbnails', body),
  });
}

export function useAITranslate() {
  return useMutation<AITranslateResponse, Error, { category: string; name: string; targetLanguage: string }>({
    mutationFn: (body) => post<AITranslateResponse>('/api/ai/translate', body),
  });
}

export function useAIAMAContent() {
  return useMutation<AIAMAContentResponse, Error, { category: string; name: string }>({
    mutationFn: (body) => post<AIAMAContentResponse>('/api/ai/ama/content', body),
  });
}

export function useAIAMATitle() {
  return useMutation<{ title: string }, Error, { category: string; name: string }>({
    mutationFn: (body) => post<{ title: string }>('/api/ai/ama/title', body),
  });
}

export function useAIAMADescription() {
  return useMutation<{ description: string }, Error, { category: string; name: string }>({
    mutationFn: (body) => post<{ description: string }>('/api/ai/ama/description', body),
  });
}

export function useAIAMATimecodes() {
  return useMutation<{ timecodes: string }, Error, { category: string; name: string }>({
    mutationFn: (body) => post<{ timecodes: string }>('/api/ai/ama/timecodes', body),
  });
}

// --- Action Button Hooks ---

export function useRequestThumbnail() {
  const qc = useQueryClient();
  return useMutation<ActionResponse, Error, { name: string; category: string }>({
    mutationFn: ({ name, category }) =>
      post<ActionResponse>(
        `/api/actions/request-thumbnail/${encodeURIComponent(name)}?category=${encodeURIComponent(category)}`,
        {},
      ),
    onSuccess: (_data, { name, category }) => {
      qc.invalidateQueries({ queryKey: ['video', name, category] });
      qc.invalidateQueries({ queryKey: ['videoProgress', name, category] });
      qc.invalidateQueries({ queryKey: ['videosList'] });
      qc.invalidateQueries({ queryKey: ['phases'] });
    },
  });
}

export function useRequestEdit() {
  const qc = useQueryClient();
  return useMutation<ActionResponse, Error, { name: string; category: string }>({
    mutationFn: ({ name, category }) =>
      post<ActionResponse>(
        `/api/actions/request-edit/${encodeURIComponent(name)}?category=${encodeURIComponent(category)}`,
        {},
      ),
    onSuccess: (_data, { name, category }) => {
      qc.invalidateQueries({ queryKey: ['video', name, category] });
      qc.invalidateQueries({ queryKey: ['videoProgress', name, category] });
      qc.invalidateQueries({ queryKey: ['videosList'] });
      qc.invalidateQueries({ queryKey: ['phases'] });
    },
  });
}

export function useUploadVideoToDrive() {
  const qc = useQueryClient();
  return useMutation<
    { driveFileId: string; videoFile: string; syncWarning?: string },
    Error,
    { name: string; category: string; file: File; onProgress?: (percent: number) => void }
  >({
    mutationFn: ({ name, category, file, onProgress }) =>
      uploadFileWithProgress<{ driveFileId: string; videoFile: string; syncWarning?: string }>(
        `/api/drive/upload/video/${encodeURIComponent(name)}?category=${encodeURIComponent(category)}`,
        file,
        'video',
        onProgress,
      ),
    onSuccess: (_data, { name, category }) => {
      qc.invalidateQueries({ queryKey: ['video', name, category] });
      qc.invalidateQueries({ queryKey: ['videoProgress', name, category] });
      qc.invalidateQueries({ queryKey: ['videosList'] });
      qc.invalidateQueries({ queryKey: ['phases'] });
    },
  });
}

export function useUploadThumbnailToDrive() {
  const qc = useQueryClient();
  return useMutation<
    { driveFileId: string; variantIndex: number; syncWarning?: string },
    Error,
    { name: string; category: string; variantIndex: number; file: File }
  >({
    mutationFn: ({ name, category, variantIndex, file }) =>
      uploadFile<{ driveFileId: string; variantIndex: number; syncWarning?: string }>(
        `/api/drive/upload/thumbnail/${encodeURIComponent(name)}?category=${encodeURIComponent(category)}&variantIndex=${variantIndex}`,
        file,
      ),
    onSuccess: (_data, { name, category }) => {
      qc.invalidateQueries({ queryKey: ['video', name, category] });
    },
  });
}

// --- Publishing Hooks ---

function invalidateVideoQueries(qc: ReturnType<typeof useQueryClient>, name: string, category: string) {
  qc.invalidateQueries({ queryKey: ['video', name, category] });
  qc.invalidateQueries({ queryKey: ['videoProgress', name, category] });
  qc.invalidateQueries({ queryKey: ['videosList'] });
  qc.invalidateQueries({ queryKey: ['phases'] });
}

export function usePublishYouTube() {
  const qc = useQueryClient();
  return useMutation<PublishYouTubeResponse, Error, { name: string; category: string }>({
    mutationFn: ({ name, category }) =>
      post<PublishYouTubeResponse>(
        `/api/publish/youtube/${encodeURIComponent(name)}?category=${encodeURIComponent(category)}`,
        {},
      ),
    onSuccess: (_data, { name, category }) => invalidateVideoQueries(qc, name, category),
  });
}

export function usePublishThumbnail() {
  const qc = useQueryClient();
  return useMutation<PublishThumbnailResponse, Error, { name: string; category: string }>({
    mutationFn: ({ name, category }) =>
      post<PublishThumbnailResponse>(
        `/api/publish/youtube/${encodeURIComponent(name)}/thumbnail?category=${encodeURIComponent(category)}`,
        {},
      ),
    onSuccess: (_data, { name, category }) => invalidateVideoQueries(qc, name, category),
  });
}

export function usePublishShort() {
  const qc = useQueryClient();
  return useMutation<PublishShortResponse, Error, { name: string; category: string; shortId: string }>({
    mutationFn: ({ name, category, shortId }) =>
      post<PublishShortResponse>(
        `/api/publish/youtube/${encodeURIComponent(name)}/shorts/${encodeURIComponent(shortId)}?category=${encodeURIComponent(category)}`,
        {},
      ),
    onSuccess: (_data, { name, category }) => invalidateVideoQueries(qc, name, category),
  });
}

export function usePublishHugo() {
  const qc = useQueryClient();
  return useMutation<PublishHugoResponse, Error, { name: string; category: string }>({
    mutationFn: ({ name, category }) =>
      post<PublishHugoResponse>(
        `/api/publish/hugo/${encodeURIComponent(name)}?category=${encodeURIComponent(category)}`,
        {},
      ),
    onSuccess: (_data, { name, category }) => invalidateVideoQueries(qc, name, category),
  });
}

export function useTranscript(videoId?: string) {
  return useQuery<TranscriptResponse>({
    queryKey: ['transcript', videoId],
    queryFn: () => get<TranscriptResponse>(`/api/publish/transcript/${encodeURIComponent(videoId!)}`),
    enabled: !!videoId,
  });
}

export function useVideoMetadata(videoId?: string) {
  return useQuery<MetadataResponse>({
    queryKey: ['videoMetadata', videoId],
    queryFn: () => get<MetadataResponse>(`/api/publish/metadata/${encodeURIComponent(videoId!)}`),
    enabled: !!videoId,
  });
}

export function useSocialPost() {
  const qc = useQueryClient();
  return useMutation<SocialPostResponse, Error, { platform: string; name: string; category: string }>({
    mutationFn: ({ platform, name, category }) =>
      post<SocialPostResponse>(
        `/api/social/${encodeURIComponent(platform)}/${encodeURIComponent(name)}?category=${encodeURIComponent(category)}`,
        {},
      ),
    onSuccess: (_data, { name, category }) => invalidateVideoQueries(qc, name, category),
  });
}

// --- Analyze Hooks ---

export function useAnalyzeTitles() {
  return useMutation<AnalyzeTitlesResponse, Error, void>({
    mutationFn: () => post<AnalyzeTitlesResponse>('/api/analyze/titles', {}),
  });
}

export function useApplyTitlesTemplate() {
  return useMutation<ApplyTitlesResponse, Error, { content: string }>({
    mutationFn: ({ content }) =>
      post<ApplyTitlesResponse>('/api/analyze/titles/apply', { content }),
  });
}

// --- Random Timing Hooks ---

export function useApplyRandomTiming() {
  const qc = useQueryClient();
  return useMutation<ApplyRandomTimingResponse, Error, { name: string; category: string }>({
    mutationFn: ({ name, category }) =>
      post<ApplyRandomTimingResponse>(
        `/api/videos/${encodeURIComponent(name)}/apply-random-timing?category=${encodeURIComponent(category)}`,
        {},
      ),
    onSuccess: (_data, { name, category }) => {
      qc.invalidateQueries({ queryKey: ['video', name, category] });
      qc.invalidateQueries({ queryKey: ['videoProgress', name, category] });
      qc.invalidateQueries({ queryKey: ['videosList'] });
      qc.invalidateQueries({ queryKey: ['phases'] });
    },
  });
}

// --- Animations Hooks ---

export function useGenerateAnimations() {
  return useMutation<AnimationsResponse, Error, { name: string; category: string }>({
    mutationFn: ({ name, category }) =>
      get<AnimationsResponse>(
        `/api/videos/${encodeURIComponent(name)}/animations?category=${encodeURIComponent(category)}`,
      ),
  });
}

// --- AMA Hooks ---

export function useAMAGenerate() {
  return useMutation<AMAGenerateResponse, Error, { videoId: string }>({
    mutationFn: (body) => post<AMAGenerateResponse>('/api/ama/generate', body),
  });
}

export function useAMAApply() {
  return useMutation<AMAApplyResponse, Error, { videoId: string; title: string; description: string; tags: string; timecodes: string }>({
    mutationFn: (body) => post<AMAApplyResponse>('/api/ama/apply', body),
  });
}
