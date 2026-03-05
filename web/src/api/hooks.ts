import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { get, patch, post, del } from './client';
import type {
  PhaseInfo,
  VideoListItem,
  VideoResponse,
  OverallProgressResponse,
  AspectsResponse,
  CreateVideoRequest,
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
