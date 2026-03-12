import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from './server';
import { useApplyRandomTiming } from '../api/hooks';
import type { ReactNode } from 'react';

function createWrapper() {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={qc}>{children}</QueryClientProvider>
  );
}

describe('useApplyRandomTiming', () => {
  it('returns new date and recommendation on success', async () => {
    const { result } = renderHook(() => useApplyRandomTiming(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ name: 'test-video', category: 'devops' });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual({
      newDate: '2026-01-14T14:30:00Z',
      originalDate: '2026-01-15',
      day: 'Wednesday',
      time: '14:30',
      reasoning: 'Mid-week afternoon uploads show 20% higher initial engagement',
    });
  });

  it('returns error when video has no date', async () => {
    server.use(
      http.post('/api/videos/:videoName/apply-random-timing', () =>
        HttpResponse.json(
          { error: 'No date set', detail: 'Video must have a date before applying random timing' },
          { status: 400 },
        ),
      ),
    );

    const { result } = renderHook(() => useApplyRandomTiming(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ name: 'no-date-video', category: 'devops' });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('returns error when no timing recommendations exist', async () => {
    server.use(
      http.post('/api/videos/:videoName/apply-random-timing', () =>
        HttpResponse.json(
          { error: 'No timing recommendations', detail: 'No timing recommendations found in settings.yaml' },
          { status: 400 },
        ),
      ),
    );

    const { result } = renderHook(() => useApplyRandomTiming(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ name: 'test-video', category: 'devops' });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('returns error when video not found', async () => {
    server.use(
      http.post('/api/videos/:videoName/apply-random-timing', () =>
        HttpResponse.json(
          { error: 'Video not found', detail: 'not found' },
          { status: 404 },
        ),
      ),
    );

    const { result } = renderHook(() => useApplyRandomTiming(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ name: 'nonexistent', category: 'devops' });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('includes syncWarning when present', async () => {
    server.use(
      http.post('/api/videos/:videoName/apply-random-timing', () =>
        HttpResponse.json({
          newDate: '2026-01-14T14:30:00Z',
          originalDate: '2026-01-15',
          day: 'Wednesday',
          time: '14:30',
          reasoning: 'Mid-week uploads perform well',
          syncWarning: 'git sync not configured — changes saved locally only',
        }),
      ),
    );

    const { result } = renderHook(() => useApplyRandomTiming(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ name: 'test-video', category: 'devops' });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.syncWarning).toBe(
      'git sync not configured — changes saved locally only',
    );
  });
});
