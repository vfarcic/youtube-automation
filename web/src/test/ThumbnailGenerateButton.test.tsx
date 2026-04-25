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
  const video = { ...mockVideo, tagline: 'Test tagline for thumbnails', ...videoOverrides };
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
  it('renders "Suggest Illustrations" button', () => {
    renderButton();
    expect(screen.getByRole('button', { name: 'Suggest Illustrations' })).toBeInTheDocument();
  });

  it('does not show "Generate Thumbnails" button initially', () => {
    renderButton();
    expect(screen.queryByRole('button', { name: 'Generate Thumbnails' })).not.toBeInTheDocument();
  });

  it('shows loading state when suggesting illustrations', async () => {
    server.use(
      http.post('/api/ai/illustrations/:category/:name', async () => {
        await new Promise((r) => setTimeout(r, 200));
        return HttpResponse.json({ illustrations: ['idea 1'] });
      }),
    );
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Illustrations' }));
    expect(screen.getByRole('button', { name: 'Suggesting...' })).toBeDisabled();
  });

  it('displays illustration options after suggestion', async () => {
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Illustrations' }));
    await waitFor(() => {
      expect(screen.getByText('A robot assembling containers')).toBeInTheDocument();
      expect(screen.getByText('A developer at a whiteboard')).toBeInTheDocument();
      expect(screen.getByText('Kubernetes pods floating in clouds')).toBeInTheDocument();
      expect(screen.getByText('None (text only)')).toBeInTheDocument();
    });
  });

  it('shows "Generate Thumbnails" button after illustration selection', async () => {
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Illustrations' }));
    await waitFor(() => {
      expect(screen.getByText('A robot assembling containers')).toBeInTheDocument();
    });
    // Generate button should exist but be disabled until selection
    expect(screen.getByRole('button', { name: 'Generate Thumbnails' })).toBeDisabled();

    // Select an illustration
    const radios = screen.getAllByRole('radio');
    await userEvent.click(radios[0]);

    expect(screen.getByRole('button', { name: 'Generate Thumbnails' })).toBeEnabled();
  });

  it('allows selecting "None" illustration', async () => {
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Illustrations' }));
    await waitFor(() => {
      expect(screen.getByText('None (text only)')).toBeInTheDocument();
    });
    // Click the "None" radio
    const noneLabel = screen.getByText('None (text only)');
    const noneRadio = noneLabel.closest('label')!.querySelector('input[type="radio"]')!;
    await userEvent.click(noneRadio);

    expect(screen.getByRole('button', { name: 'Generate Thumbnails' })).toBeEnabled();
  });

  it('shows error on illustration suggestion failure', async () => {
    server.use(
      http.post('/api/ai/illustrations/:category/:name', () =>
        new HttpResponse('AI generation failed', { status: 500 }),
      ),
    );
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Illustrations' }));
    await waitFor(() => {
      expect(screen.getByText('AI generation failed')).toBeInTheDocument();
    });
  });

  it('shows loading state during thumbnail generation', async () => {
    server.use(
      http.post('/api/thumbnails/generate', async () => {
        await new Promise((r) => setTimeout(r, 200));
        return HttpResponse.json({ thumbnails: [], errors: [] });
      }),
    );
    renderButton();
    // First suggest illustrations
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Illustrations' }));
    await waitFor(() => {
      expect(screen.getByText('A robot assembling containers')).toBeInTheDocument();
    });
    // Select an illustration
    const radios = screen.getAllByRole('radio');
    await userEvent.click(radios[0]);
    // Generate thumbnails
    await userEvent.click(screen.getByRole('button', { name: 'Generate Thumbnails' }));
    expect(screen.getByRole('button', { name: 'Generating...' })).toBeDisabled();
    expect(screen.getByText(/may take up to 2 minutes/)).toBeInTheDocument();
  });

  it('displays generated thumbnails in a grid grouped by provider', async () => {
    renderButton();
    // Suggest illustrations
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Illustrations' }));
    await waitFor(() => {
      expect(screen.getByText('A robot assembling containers')).toBeInTheDocument();
    });
    // Select an illustration
    const radios = screen.getAllByRole('radio');
    await userEvent.click(radios[0]);
    // Generate thumbnails
    await userEvent.click(screen.getByRole('button', { name: 'Generate Thumbnails' }));
    await waitFor(() => {
      expect(screen.getByText('gemini')).toBeInTheDocument();
      expect(screen.getByText('gpt-image')).toBeInTheDocument();
    });
    // Check style labels
    expect(screen.getAllByText('with illustration')).toHaveLength(2);
    expect(screen.getAllByText('without illustration')).toHaveLength(2);
    // Check "Use This" buttons
    expect(screen.getAllByRole('button', { name: 'Use This' })).toHaveLength(4);
  });

  it('handles "Use This" selection and shows success message', async () => {
    renderButton();
    // Suggest illustrations
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Illustrations' }));
    await waitFor(() => {
      expect(screen.getByText('A robot assembling containers')).toBeInTheDocument();
    });
    // Select an illustration
    const radios = screen.getAllByRole('radio');
    await userEvent.click(radios[0]);
    // Generate thumbnails
    await userEvent.click(screen.getByRole('button', { name: 'Generate Thumbnails' }));
    await waitFor(() => {
      expect(screen.getAllByRole('button', { name: 'Use This' })).toHaveLength(4);
    });
    // Click "Use This" on the first thumbnail
    const useButtons = screen.getAllByRole('button', { name: 'Use This' });
    await userEvent.click(useButtons[0]);
    await waitFor(() => {
      expect(screen.getByText(/uploaded to Drive/)).toBeInTheDocument();
    });
    // The selected thumbnail should be removed from the grid
    expect(screen.getAllByRole('button', { name: 'Use This' })).toHaveLength(3);
  });

  it('shows error on generation failure', async () => {
    server.use(
      http.post('/api/thumbnails/generate', () =>
        new HttpResponse('All providers failed', { status: 500 }),
      ),
    );
    renderButton();
    // Suggest illustrations
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Illustrations' }));
    await waitFor(() => {
      expect(screen.getByText('A robot assembling containers')).toBeInTheDocument();
    });
    // Select "None"
    const noneLabel = screen.getByText('None (text only)');
    const noneRadio = noneLabel.closest('label')!.querySelector('input[type="radio"]')!;
    await userEvent.click(noneRadio);
    // Generate
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
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Illustrations' }));
    await waitFor(() => {
      expect(screen.getByText('A robot assembling containers')).toBeInTheDocument();
    });
    const radios = screen.getAllByRole('radio');
    await userEvent.click(radios[0]);
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
    renderButton();
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Illustrations' }));
    await waitFor(() => {
      expect(screen.getByText('A robot assembling containers')).toBeInTheDocument();
    });
    const radios = screen.getAllByRole('radio');
    await userEvent.click(radios[0]);
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

  it('uses existing thumbnail variants length for variantIndex', async () => {
    renderButton({
      thumbnailVariants: [
        { index: 1, path: '', driveFileId: 'existing-id', clickShare: 0 },
      ],
    });
    // The component should use variantIndex = 1 (length of existing variants)
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Illustrations' }));
    await waitFor(() => {
      expect(screen.getByText('A robot assembling containers')).toBeInTheDocument();
    });
    const radios = screen.getAllByRole('radio');
    await userEvent.click(radios[0]);
    await userEvent.click(screen.getByRole('button', { name: 'Generate Thumbnails' }));
    await waitFor(() => {
      expect(screen.getAllByRole('button', { name: 'Use This' })).toHaveLength(4);
    });
  });
});
