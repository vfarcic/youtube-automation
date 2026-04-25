import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { http, HttpResponse } from 'msw';
import { describe, it, expect } from 'vitest';
import { server } from './server';
import { mockVideo } from './handlers';
import { ThumbnailGenerateButton } from '../components/forms/ThumbnailGenerateButton';
import type { VideoResponse } from '../api/types';

function renderButton(videoOverrides: Partial<VideoResponse> = {}) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  const video = { ...mockVideo, ...videoOverrides };
  return render(
    <QueryClientProvider client={qc}>
      <ThumbnailGenerateButton
        category="devops"
        videoName="test-video"
        video={video}
      />
    </QueryClientProvider>,
  );
}

describe('ThumbnailGenerateButton', () => {
  it('renders "Suggest Tagline & Illustrations" button', () => {
    renderButton();
    expect(screen.getByRole('button', { name: 'Suggest Tagline & Illustrations' })).toBeInTheDocument();
  });

  it('does not show "Generate Thumbnails" button when no tagline stored', () => {
    renderButton();
    expect(screen.queryByRole('button', { name: 'Generate Thumbnails' })).not.toBeInTheDocument();
  });

  it('shows "Generate Thumbnails" button when video has stored tagline', () => {
    renderButton({ tagline: 'Existing Tagline' });
    expect(screen.getByRole('button', { name: 'Generate Thumbnails' })).toBeInTheDocument();
  });

  it('displays stored tagline and illustration', () => {
    renderButton({ tagline: 'My Tagline', illustration: 'A robot building things' });
    expect(screen.getByText(/My Tagline/)).toBeInTheDocument();
    expect(screen.getByText(/A robot building things/)).toBeInTheDocument();
  });

  it('shows loading state when suggesting', async () => {
    server.use(
      http.post('/api/ai/tagline-and-illustrations/:category/:name', async () => {
        await new Promise((r) => setTimeout(r, 200));
        return HttpResponse.json({ taglines: ['Tag'], illustrations: ['idea 1'] });
      }),
    );
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Tagline & Illustrations' }));
    expect(screen.getByRole('button', { name: 'Suggesting...' })).toBeDisabled();
  });

  it('displays tagline and illustration options after suggestion', async () => {
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Tagline & Illustrations' }));
    await waitFor(() => {
      // Taglines
      expect(screen.getByText('Contain Everything')).toBeInTheDocument();
      expect(screen.getByText('Ship Faster')).toBeInTheDocument();
      expect(screen.getByText('Deploy Smart')).toBeInTheDocument();
      // Illustrations
      expect(screen.getByText('A robot assembling containers')).toBeInTheDocument();
      expect(screen.getByText('A developer at a whiteboard')).toBeInTheDocument();
      expect(screen.getByText('None (text only)')).toBeInTheDocument();
    });
  });

  it('shows "Save Selection" button after selecting tagline and illustration', async () => {
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Tagline & Illustrations' }));
    await waitFor(() => {
      expect(screen.getByText('Contain Everything')).toBeInTheDocument();
    });

    // Save button should be disabled initially
    expect(screen.getByRole('button', { name: 'Save Selection' })).toBeDisabled();

    // Select a tagline and illustration
    const radios = screen.getAllByRole('radio');
    await userEvent.click(radios[0]); // first tagline
    // Still disabled — need illustration too
    expect(screen.getByRole('button', { name: 'Save Selection' })).toBeDisabled();

    await userEvent.click(radios[3]); // first illustration (after 3 taglines)
    expect(screen.getByRole('button', { name: 'Save Selection' })).toBeEnabled();
  });

  it('saves selection and shows success', async () => {
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Tagline & Illustrations' }));
    await waitFor(() => {
      expect(screen.getByText('Contain Everything')).toBeInTheDocument();
    });

    const radios = screen.getAllByRole('radio');
    await userEvent.click(radios[0]); // tagline
    await userEvent.click(radios[3]); // illustration
    await userEvent.click(screen.getByRole('button', { name: 'Save Selection' }));
    await waitFor(() => {
      expect(screen.getByText('Selection saved.')).toBeInTheDocument();
    });
  });

  it('allows selecting "None" illustration', async () => {
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Tagline & Illustrations' }));
    await waitFor(() => {
      expect(screen.getByText('None (text only)')).toBeInTheDocument();
    });
    // Select a tagline
    const radios = screen.getAllByRole('radio');
    await userEvent.click(radios[0]);
    // Click the "None" radio
    const noneLabel = screen.getByText('None (text only)');
    const noneRadio = noneLabel.closest('label')!.querySelector('input[type="radio"]')!;
    await userEvent.click(noneRadio);

    expect(screen.getByRole('button', { name: 'Save Selection' })).toBeEnabled();
  });

  it('shows error on suggestion failure', async () => {
    server.use(
      http.post('/api/ai/tagline-and-illustrations/:category/:name', () =>
        new HttpResponse('AI generation failed', { status: 500 }),
      ),
    );
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Tagline & Illustrations' }));
    await waitFor(() => {
      expect(screen.getByText('AI generation failed')).toBeInTheDocument();
    });
  });

  it('generates thumbnails with stored tagline', async () => {
    renderButton({ tagline: 'Existing Tagline' });
    await userEvent.click(screen.getByRole('button', { name: 'Generate Thumbnails' }));
    await waitFor(() => {
      expect(screen.getByText('gemini')).toBeInTheDocument();
      expect(screen.getByText('gpt-image')).toBeInTheDocument();
    });
    expect(screen.getAllByText('with illustration')).toHaveLength(2);
    expect(screen.getAllByText('without illustration')).toHaveLength(2);
    expect(screen.getAllByRole('button', { name: 'Use This' })).toHaveLength(4);
  });

  it('handles "Use This" selection and shows success message', async () => {
    renderButton({ tagline: 'Test Tagline' });
    await userEvent.click(screen.getByRole('button', { name: 'Generate Thumbnails' }));
    await waitFor(() => {
      expect(screen.getAllByRole('button', { name: 'Use This' })).toHaveLength(4);
    });
    const useButtons = screen.getAllByRole('button', { name: 'Use This' });
    await userEvent.click(useButtons[0]);
    await waitFor(() => {
      expect(screen.getByText(/uploaded to Drive/)).toBeInTheDocument();
    });
    expect(screen.getAllByRole('button', { name: 'Use This' })).toHaveLength(3);
  });

  it('shows error on generation failure', async () => {
    server.use(
      http.post('/api/thumbnails/generate', () =>
        new HttpResponse('All providers failed', { status: 500 }),
      ),
    );
    renderButton({ tagline: 'Test Tagline' });
    await userEvent.click(screen.getByRole('button', { name: 'Generate Thumbnails' }));
    await waitFor(() => {
      expect(screen.getByText('All providers failed')).toBeInTheDocument();
    });
  });

  it('shows partial failure warnings', async () => {
    server.use(
      http.post('/api/thumbnails/generate', () =>
        HttpResponse.json({
          thumbnails: [{ id: 'thumb-1', provider: 'gemini', style: 'with illustration' }],
          errors: ['a provider failed to generate an image'],
        }),
      ),
    );
    renderButton({ tagline: 'Test Tagline' });
    await userEvent.click(screen.getByRole('button', { name: 'Generate Thumbnails' }));
    await waitFor(() => {
      expect(screen.getByText('a provider failed to generate an image')).toBeInTheDocument();
      expect(screen.getByText('gemini')).toBeInTheDocument();
    });
  });

  it('shows error on select failure', async () => {
    server.use(
      http.post('/api/thumbnails/generated/:id/select', () =>
        new HttpResponse('Drive upload failed', { status: 500 }),
      ),
    );
    renderButton({ tagline: 'Test Tagline' });
    await userEvent.click(screen.getByRole('button', { name: 'Generate Thumbnails' }));
    await waitFor(() => {
      expect(screen.getAllByRole('button', { name: 'Use This' })).toHaveLength(4);
    });
    const useButtons = screen.getAllByRole('button', { name: 'Use This' });
    await userEvent.click(useButtons[0]);
    await waitFor(() => {
      expect(screen.getByText('Drive upload failed')).toBeInTheDocument();
    });
  });
});
