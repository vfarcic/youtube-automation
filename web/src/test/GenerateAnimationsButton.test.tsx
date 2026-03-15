import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { http, HttpResponse } from 'msw';
import { describe, it, expect, vi } from 'vitest';
import { server } from './server';
import { GenerateAnimationsButton } from '../components/forms/GenerateAnimationsButton';

function renderButton(onApply = vi.fn()) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return {
    onApply,
    ...render(
      <QueryClientProvider client={qc}>
        <GenerateAnimationsButton
          category="devops"
          videoName="test-video"
          onApply={onApply}
        />
      </QueryClientProvider>,
    ),
  };
}

describe('GenerateAnimationsButton', () => {
  it('renders "Generate from Gist" button', () => {
    renderButton();
    expect(screen.getByRole('button', { name: 'Generate from Gist' })).toBeInTheDocument();
  });

  it('shows "Generating..." loading state on click', async () => {
    server.use(
      http.get('/api/videos/:videoName/animations', async () => {
        await new Promise((r) => setTimeout(r, 200));
        return HttpResponse.json({
          animations: ['item'],
          sections: [],
        });
      }),
    );
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Generate from Gist' }));
    expect(screen.getByRole('button', { name: 'Generating...' })).toBeDisabled();
  });

  it('calls onApply with formatted animations and timecodes', async () => {
    const onApply = vi.fn();
    renderButton(onApply);
    await userEvent.click(screen.getByRole('button', { name: 'Generate from Gist' }));
    await waitFor(() => {
      expect(onApply).toHaveBeenCalledWith(
        '- Add fade transition\n- Section: Main Demo\n- Show terminal output',
        '00:00 FIXME:\nFIXME:FIXME Main Demo',
      );
    });
  });

  it('calls onApply with empty timecodes when no sections', async () => {
    server.use(
      http.get('/api/videos/:videoName/animations', () =>
        HttpResponse.json({
          animations: ['First cue', 'Second cue'],
          sections: [],
        }),
      ),
    );
    const onApply = vi.fn();
    renderButton(onApply);
    await userEvent.click(screen.getByRole('button', { name: 'Generate from Gist' }));
    await waitFor(() => {
      expect(onApply).toHaveBeenCalledWith(
        '- First cue\n- Second cue',
        '',
      );
    });
  });

  it('shows error message on failure', async () => {
    server.use(
      http.get('/api/videos/:videoName/animations', () =>
        new HttpResponse('no gist path set for video', { status: 404 }),
      ),
    );
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Generate from Gist' }));
    await waitFor(() => {
      expect(screen.getByText('no gist path set for video')).toBeInTheDocument();
    });
  });
});
