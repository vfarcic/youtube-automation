import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { http, HttpResponse } from 'msw';
import { describe, it, expect, vi } from 'vitest';
import { server } from './server';
import { RandomTimingButton } from '../components/forms/RandomTimingButton';

function renderButton(onApply = vi.fn()) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return {
    onApply,
    ...render(
      <QueryClientProvider client={qc}>
        <RandomTimingButton
          category="devops"
          videoName="test-video"
          onApply={onApply}
        />
      </QueryClientProvider>,
    ),
  };
}

describe('RandomTimingButton', () => {
  it('renders "Apply Random Timing" button', () => {
    renderButton();
    expect(screen.getByRole('button', { name: 'Apply Random Timing' })).toBeInTheDocument();
  });

  it('shows "Applying..." loading state on click', async () => {
    // Use a delayed response so we can observe the loading state
    server.use(
      http.post('/api/videos/:videoName/apply-random-timing', async () => {
        await new Promise((r) => setTimeout(r, 200));
        return HttpResponse.json({
          newDate: '2026-01-14T14:30:00Z',
          originalDate: '2026-01-15',
          day: 'Wednesday',
          time: '14:30',
          reasoning: 'test',
        });
      }),
    );
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Apply Random Timing' }));
    expect(screen.getByRole('button', { name: 'Applying...' })).toBeDisabled();
  });

  it('calls onApply with formatted date and shows info panel on success', async () => {
    const onApply = vi.fn();
    renderButton(onApply);
    await userEvent.click(screen.getByRole('button', { name: 'Apply Random Timing' }));
    await waitFor(() => {
      expect(onApply).toHaveBeenCalledWith('2026-01-14T14:30');
    });
    expect(screen.getByText('Wednesday')).toBeInTheDocument();
    expect(screen.getByText('14:30')).toBeInTheDocument();
    expect(screen.getByText('Mid-week afternoon uploads show 20% higher initial engagement')).toBeInTheDocument();
  });

  it('shows sync warning when present', async () => {
    server.use(
      http.post('/api/videos/:videoName/apply-random-timing', () =>
        HttpResponse.json({
          newDate: '2026-01-14T14:30:00Z',
          originalDate: '2026-01-15',
          day: 'Wednesday',
          time: '14:30',
          reasoning: 'test reasoning',
          syncWarning: 'File was modified externally',
        }),
      ),
    );
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Apply Random Timing' }));
    await waitFor(() => {
      expect(screen.getByText('File was modified externally')).toBeInTheDocument();
    });
  });

  it('shows error message on 400 response', async () => {
    server.use(
      http.post('/api/videos/:videoName/apply-random-timing', () =>
        new HttpResponse('No timing recommendations available', { status: 400 }),
      ),
    );
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Apply Random Timing' }));
    await waitFor(() => {
      expect(screen.getByText('No timing recommendations available')).toBeInTheDocument();
    });
  });

  it('shows error message on 404 response', async () => {
    server.use(
      http.post('/api/videos/:videoName/apply-random-timing', () =>
        new HttpResponse('Video not found', { status: 404 }),
      ),
    );
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Apply Random Timing' }));
    await waitFor(() => {
      expect(screen.getByText('Video not found')).toBeInTheDocument();
    });
  });
});
