import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { get, patch, post, del, uploadFile } from './client';
import type {
  PhaseInfo,
  VideoListItem,
  VideoResponse,
  OverallProgressResponse,
  AspectsResponse,
  CreateVideoRequest,
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
