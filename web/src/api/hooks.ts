import { useQuery } from '@tanstack/react-query';
import { get } from './client';
import type {
  PhaseInfo,
  VideoListItem,
  VideoResponse,
  OverallProgressResponse,
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
