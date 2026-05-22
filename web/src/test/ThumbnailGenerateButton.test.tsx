import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { http, HttpResponse } from 'msw';
import { describe, it, expect } from 'vitest';
import { server } from './server';
import { mockVideo } from './handlers';
import { ThumbnailGenerateButton } from '../components/forms/ThumbnailGenerateButton';
import type { VideoResponse } from '../api/types';

// Shape of the JSON body POSTed to /api/videos/:videoName/thumbnail-config.
// Mirrors the input to useSaveThumbnailConfig in src/api/hooks.ts.
interface ThumbnailConfigSaveBody {
  tagline?: string;
  illustration?: string;
  photoRealisticSubject?: string;
}

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

  // ---------------------------------------------------------------------------
  // PRD 401 M4: photo-realistic subject input + variant rendering
  // ---------------------------------------------------------------------------

  it('shows photo-realistic subject input when a tagline is already stored', () => {
    renderButton({ tagline: 'Test Tagline' });
    expect(screen.getByLabelText(/Photo-realistic subject/i)).toBeInTheDocument();
  });

  it('persists user-typed subject via save endpoint', async () => {
    // Wrap captured value in an object so the closure-set body survives
    // TypeScript's literal-null narrowing of a plain `let x: T | null = null`.
    // tsc -b otherwise narrows the variable to `never` at the assertion site
    // because it doesn't track the msw handler reassignment.
    const captured: { body?: ThumbnailConfigSaveBody } = {};
    server.use(
      http.post('/api/videos/:videoName/thumbnail-config', async ({ request }) => {
        captured.body = (await request.json()) as ThumbnailConfigSaveBody;
        return HttpResponse.json({
          tagline: 'Test Tagline',
          illustration: '',
          photoRealisticSubject: captured.body?.photoRealisticSubject ?? '',
        });
      }),
    );

    renderButton({ tagline: 'Test Tagline' });
    const input = screen.getByLabelText(/Photo-realistic subject/i);
    await userEvent.type(input, 'a small white rabbit holding a checklist');

    // A 'Save subject' button appears once the input differs from the stored value.
    const saveBtn = await screen.findByRole('button', { name: 'Save subject' });
    await userEvent.click(saveBtn);

    await waitFor(() => {
      expect(captured.body).toBeDefined();
    });
    expect(captured.body?.photoRealisticSubject).toBe('a small white rabbit holding a checklist');
  });

  it('does not show "Save subject" button when input matches the stored subject', () => {
    renderButton({ tagline: 'Test Tagline', photoRealisticSubject: 'a vintage typewriter' });
    expect(screen.queryByRole('button', { name: 'Save subject' })).not.toBeInTheDocument();
  });

  it('renders 3 tiles per provider when the photo-realistic variant is returned', async () => {
    server.use(
      http.post('/api/thumbnails/generate', () =>
        HttpResponse.json({
          thumbnails: [
            { id: 'thumb-1', provider: 'gemini', style: 'with illustration' },
            { id: 'thumb-2', provider: 'gemini', style: 'without illustration' },
            { id: 'thumb-3', provider: 'gemini', style: 'photorealistic' },
            { id: 'thumb-4', provider: 'gpt-image', style: 'with illustration' },
            { id: 'thumb-5', provider: 'gpt-image', style: 'without illustration' },
            { id: 'thumb-6', provider: 'gpt-image', style: 'photorealistic' },
          ],
          errors: [],
        }),
      ),
    );
    renderButton({ tagline: 'Test Tagline', photoRealisticSubject: 'a robot' });
    await userEvent.click(screen.getByRole('button', { name: 'Generate Thumbnails' }));
    await waitFor(() => {
      expect(screen.getAllByText('photorealistic')).toHaveLength(2);
    });
    expect(screen.getAllByText('with illustration')).toHaveLength(2);
    expect(screen.getAllByText('without illustration')).toHaveLength(2);
    expect(screen.getAllByRole('button', { name: 'Use This' })).toHaveLength(6);
  });

  it('renders only 2 tiles per provider when only 2 variants returned (backwards compat)', async () => {
    // Default handler returns 2 styles per provider — no photorealistic.
    renderButton({ tagline: 'Test Tagline' });
    await userEvent.click(screen.getByRole('button', { name: 'Generate Thumbnails' }));
    await waitFor(() => {
      expect(screen.getAllByRole('button', { name: 'Use This' })).toHaveLength(4);
    });
    expect(screen.queryByText('photorealistic')).not.toBeInTheDocument();
  });

  it('selects the photo-realistic tile via the existing select endpoint', async () => {
    let selectedId: string | null = null;
    server.use(
      http.post('/api/thumbnails/generate', () =>
        HttpResponse.json({
          thumbnails: [
            { id: 'bw-1', provider: 'gemini', style: 'with illustration' },
            { id: 'bw-2', provider: 'gemini', style: 'without illustration' },
            { id: 'photoreal-id', provider: 'gemini', style: 'photorealistic' },
          ],
          errors: [],
        }),
      ),
      http.post('/api/thumbnails/generated/:id/select', ({ params }) => {
        selectedId = params.id as string;
        return HttpResponse.json({ driveFileId: 'drive-x', variantIndex: 0 });
      }),
    );
    renderButton({ tagline: 'Test Tagline', photoRealisticSubject: 'a robot' });
    await userEvent.click(screen.getByRole('button', { name: 'Generate Thumbnails' }));
    await waitFor(() => {
      expect(screen.getByText('photorealistic')).toBeInTheDocument();
    });

    // Find the 'Use This' button next to the photorealistic tile.
    const photorealLabel = screen.getByText('photorealistic');
    const tile = photorealLabel.closest('div');
    expect(tile).not.toBeNull();
    const useBtn = tile!.querySelector('button')!;
    await userEvent.click(useBtn);

    await waitFor(() => {
      expect(selectedId).toBe('photoreal-id');
    });
  });

  it('shows skipped-variant notice when generation returns only 2 variants', async () => {
    // Default handler returns 4 variants (2 providers × 2 styles, no photoreal).
    renderButton({ tagline: 'Test Tagline' });
    await userEvent.click(screen.getByRole('button', { name: 'Generate Thumbnails' }));
    await waitFor(() => {
      expect(screen.getByText(/Photo-realistic variant skipped/i)).toBeInTheDocument();
    });
    // It is non-blocking: tiles are still rendered.
    expect(screen.getAllByRole('button', { name: 'Use This' })).toHaveLength(4);
  });

  it('does NOT show skipped-variant notice when photo-realistic variant is present', async () => {
    server.use(
      http.post('/api/thumbnails/generate', () =>
        HttpResponse.json({
          thumbnails: [
            { id: 't-1', provider: 'gemini', style: 'with illustration' },
            { id: 't-2', provider: 'gemini', style: 'without illustration' },
            { id: 't-3', provider: 'gemini', style: 'photorealistic' },
          ],
          errors: [],
        }),
      ),
    );
    renderButton({ tagline: 'Test Tagline', photoRealisticSubject: 'a robot' });
    await userEvent.click(screen.getByRole('button', { name: 'Generate Thumbnails' }));
    await waitFor(() => {
      expect(screen.getByText('photorealistic')).toBeInTheDocument();
    });
    expect(screen.queryByText(/Photo-realistic variant skipped/i)).not.toBeInTheDocument();
  });

  it('displays stored photo-realistic subject', () => {
    renderButton({ tagline: 'Test Tagline', photoRealisticSubject: 'a small white rabbit' });
    expect(screen.getByText(/a small white rabbit/)).toBeInTheDocument();
  });

  it('subject input pre-fills from the video state', () => {
    renderButton({ tagline: 'Test Tagline', photoRealisticSubject: 'a vintage typewriter' });
    const input = screen.getByLabelText(/Photo-realistic subject/i) as HTMLInputElement;
    expect(input.value).toBe('a vintage typewriter');
  });

  // ---------------------------------------------------------------------------
  // PRD 401 M4 fix: chain save-before-generate when user types a subject and
  // clicks Generate without clicking Save subject first. The previous
  // implementation discarded the typed value and the backend silently fell
  // through to AI inference.
  // ---------------------------------------------------------------------------

  it('auto-saves typed subject before generate when input differs from stored', async () => {
    const callOrder: string[] = [];
    // See note above (persists user-typed subject test) on why we wrap.
    const captured: { body?: ThumbnailConfigSaveBody } = {};

    server.use(
      http.post('/api/videos/:videoName/thumbnail-config', async ({ request }) => {
        captured.body = (await request.json()) as ThumbnailConfigSaveBody;
        callOrder.push('save');
        return HttpResponse.json({
          tagline: 'Test Tagline',
          illustration: '',
          photoRealisticSubject: captured.body?.photoRealisticSubject ?? '',
        });
      }),
      http.post('/api/thumbnails/generate', () => {
        callOrder.push('generate');
        return HttpResponse.json({
          thumbnails: [
            { id: 'thumb-1', provider: 'gemini', style: 'with illustration' },
            { id: 'thumb-2', provider: 'gemini', style: 'without illustration' },
            { id: 'thumb-3', provider: 'gemini', style: 'photorealistic' },
          ],
          errors: [],
        });
      }),
    );

    renderButton({ tagline: 'Test Tagline' });

    const input = screen.getByLabelText(/Photo-realistic subject/i);
    await userEvent.type(input, 'a small white rabbit');

    // User clicks Generate WITHOUT clicking 'Save subject' first.
    await userEvent.click(screen.getByRole('button', { name: 'Generate Thumbnails' }));

    await waitFor(() => {
      expect(callOrder).toEqual(['save', 'generate']);
    });
    expect(captured.body?.photoRealisticSubject).toBe('a small white rabbit');
  });

  it('does NOT auto-save when input matches the stored subject', async () => {
    let saveCalled = false;
    server.use(
      http.post('/api/videos/:videoName/thumbnail-config', () => {
        saveCalled = true;
        return HttpResponse.json({
          tagline: 'Test Tagline',
          illustration: '',
          photoRealisticSubject: 'a robot',
        });
      }),
    );

    renderButton({ tagline: 'Test Tagline', photoRealisticSubject: 'a robot' });

    await userEvent.click(screen.getByRole('button', { name: 'Generate Thumbnails' }));

    await waitFor(() => {
      expect(screen.getAllByRole('button', { name: 'Use This' }).length).toBeGreaterThan(0);
    });
    expect(saveCalled).toBe(false);
  });

  it('does NOT call generate when the pre-generate save fails', async () => {
    let generateCalled = false;
    server.use(
      http.post('/api/videos/:videoName/thumbnail-config', () =>
        new HttpResponse('save failed: storage error', { status: 500 }),
      ),
      http.post('/api/thumbnails/generate', () => {
        generateCalled = true;
        return HttpResponse.json({ thumbnails: [], errors: [] });
      }),
    );

    renderButton({ tagline: 'Test Tagline' });

    const input = screen.getByLabelText(/Photo-realistic subject/i);
    await userEvent.type(input, 'a robot waving');

    await userEvent.click(screen.getByRole('button', { name: 'Generate Thumbnails' }));

    // Save error must surface in the UI (the existing inline error block
    // for saveMutation renders saveMutation.error.message).
    await waitFor(() => {
      expect(screen.getByText(/save failed: storage error/i)).toBeInTheDocument();
    });
    expect(generateCalled).toBe(false);
  });

  it('input enforces a 200-character maxLength', async () => {
    renderButton({ tagline: 'Test Tagline' });
    const input = screen.getByLabelText(/Photo-realistic subject/i) as HTMLInputElement;
    expect(input.maxLength).toBe(200);

    // Pasting a 300-char string must be clamped to 200 by the browser /
    // jsdom-equivalent maxLength enforcement.
    const longInput = 'x'.repeat(300);
    await userEvent.click(input);
    await userEvent.paste(longInput);

    expect(input.value.length).toBe(200);
  });

  // ---------------------------------------------------------------------------
  // PRD 401 M4 follow-up fixes
  // ---------------------------------------------------------------------------

  // ISSUE A regression: after Save Selection, video props are stale (parent
  // refetch hasn't completed). The auto-save chain before Generate must
  // still send the just-saved tagline + the typed photoRealisticSubject —
  // not the empty stale video.tagline.
  it('auto-save after Save Selection uses the just-saved tagline (not stale video props)', async () => {
    const savePayloads: Array<Record<string, string>> = [];
    server.use(
      http.post('/api/videos/:videoName/thumbnail-config', async ({ request }) => {
        const body = (await request.json()) as Record<string, string>;
        savePayloads.push(body);
        return HttpResponse.json({
          tagline: body.tagline ?? '',
          illustration: body.illustration ?? '',
          photoRealisticSubject: body.photoRealisticSubject ?? '',
        });
      }),
    );

    // Step 1: render with empty video (no stored tagline → canGenerate=false).
    renderButton(); // mockVideo has empty tagline/illustration/photoRealisticSubject

    // Step 2: user clicks Suggest, picks tagline+illustration, clicks Save Selection.
    await userEvent.click(screen.getByRole('button', { name: 'Suggest Tagline & Illustrations' }));
    await waitFor(() => {
      expect(screen.getByText('Contain Everything')).toBeInTheDocument();
    });
    const radios = screen.getAllByRole('radio');
    await userEvent.click(radios[0]); // first tagline = 'Contain Everything'
    await userEvent.click(radios[3]); // first illustration = 'A robot assembling containers'
    await userEvent.click(screen.getByRole('button', { name: 'Save Selection' }));
    await waitFor(() => {
      expect(screen.getByText('Selection saved.')).toBeInTheDocument();
    });

    // Step 3: type a subject (parent's video prop is still stale here).
    const input = screen.getByLabelText(/Photo-realistic subject/i);
    await userEvent.type(input, 'a small white rabbit');

    // Step 4: click Generate. Auto-save must fire with the JUST-SAVED
    // tagline 'Contain Everything', NOT the empty stale video.tagline.
    await userEvent.click(screen.getByRole('button', { name: 'Generate Thumbnails' }));

    await waitFor(() => {
      // Two save calls total: the explicit Save Selection, then the auto-save.
      expect(savePayloads.length).toBe(2);
    });
    const autoSave = savePayloads[1];
    expect(autoSave.tagline).toBe('Contain Everything');
    expect(autoSave.illustration).toBe('A robot assembling containers');
    expect(autoSave.photoRealisticSubject).toBe('a small white rabbit');
  });

  // ISSUE B regression: rapid double-click on Generate must not produce
  // duplicate save+generate chains. The button must disable while either
  // saveMutation or generateMutation is in flight.
  it('rapid double-click on Generate does not duplicate save/generate calls', async () => {
    let saveCount = 0;
    let generateCount = 0;
    server.use(
      http.post('/api/videos/:videoName/thumbnail-config', async ({ request }) => {
        saveCount++;
        // Add a small delay so the save is visibly in flight when the
        // second click would land.
        await new Promise((r) => setTimeout(r, 30));
        const body = (await request.json()) as Record<string, string>;
        return HttpResponse.json({
          tagline: body.tagline ?? '',
          illustration: body.illustration ?? '',
          photoRealisticSubject: body.photoRealisticSubject ?? '',
        });
      }),
      http.post('/api/thumbnails/generate', async () => {
        generateCount++;
        return HttpResponse.json({
          thumbnails: [
            { id: 'thumb-1', provider: 'gemini', style: 'with illustration' },
            { id: 'thumb-2', provider: 'gemini', style: 'without illustration' },
            { id: 'thumb-3', provider: 'gemini', style: 'photorealistic' },
          ],
          errors: [],
        });
      }),
    );

    renderButton({ tagline: 'Test Tagline' });

    const input = screen.getByLabelText(/Photo-realistic subject/i);
    await userEvent.type(input, 'a robot');

    const generateBtn = screen.getByRole('button', { name: 'Generate Thumbnails' });
    // First click triggers handleGenerateThumbnails → save starts.
    // Then immediately attempt a second click: by the time React re-renders
    // with saveMutation.isPending=true the button is disabled; userEvent
    // will fail to click a disabled button.
    await userEvent.click(generateBtn);
    await userEvent.click(generateBtn);

    // Wait for the chain to finish so the counts are stable.
    await waitFor(() => {
      expect(screen.getByText('photorealistic')).toBeInTheDocument();
    });

    expect(saveCount).toBe(1);
    expect(generateCount).toBe(1);
  });
});
