import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from './server';
import { AnalyzeTiming } from '../pages/AnalyzeTiming';

function renderWithProviders(ui: React.ReactElement) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>{ui}</MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('AnalyzeTiming', () => {
  it('renders page with current recommendations table', async () => {
    renderWithProviders(<AnalyzeTiming />);

    expect(screen.getByText('Timing Recommendations')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Generate New Recommendations' })).toBeInTheDocument();

    // Wait for current recommendations to load from MSW
    expect(await screen.findByText('Wednesday')).toBeInTheDocument();
    expect(screen.getByText('14:00')).toBeInTheDocument();
    expect(screen.getByText('Mid-week peak')).toBeInTheDocument();
    expect(screen.getByText('Monday')).toBeInTheDocument();
    expect(screen.getByText('09:00')).toBeInTheDocument();
  });

  it('shows empty state when no recommendations exist', async () => {
    server.use(
      http.get('/api/analyze/timing', () =>
        HttpResponse.json({ recommendations: [] }),
      ),
    );

    renderWithProviders(<AnalyzeTiming />);

    expect(await screen.findByText(/No timing recommendations configured yet/)).toBeInTheDocument();
  });

  it('generates new recommendations and shows preview with Save button', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AnalyzeTiming />);

    await user.click(screen.getByRole('button', { name: 'Generate New Recommendations' }));

    // Wait for generation results
    expect(await screen.findByTestId('video-count')).toHaveTextContent('42');
    expect(screen.getByText(/Review the recommendations below/)).toBeInTheDocument();

    // Preview table should show generated recommendations
    expect(screen.getByText('Thursday')).toBeInTheDocument();
    expect(screen.getByText('16:00')).toBeInTheDocument();

    // Save & Push button should be present
    expect(screen.getByRole('button', { name: 'Save & Push' })).toBeInTheDocument();
  });

  it('saves generated recommendations when Save & Push is clicked', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AnalyzeTiming />);

    await user.click(screen.getByRole('button', { name: 'Generate New Recommendations' }));

    // Wait for generation
    expect(await screen.findByRole('button', { name: 'Save & Push' })).toBeInTheDocument();

    // Click Save & Push
    await user.click(screen.getByRole('button', { name: 'Save & Push' }));

    // Should show success message
    expect(await screen.findByText('Recommendations saved.')).toBeInTheDocument();
  });

  it('shows sync warning after save when present', async () => {
    server.use(
      http.put('/api/analyze/timing', () =>
        HttpResponse.json({ saved: true, syncWarning: 'git sync not configured' }),
      ),
    );

    const user = userEvent.setup();
    renderWithProviders(<AnalyzeTiming />);

    await user.click(screen.getByRole('button', { name: 'Generate New Recommendations' }));
    expect(await screen.findByRole('button', { name: 'Save & Push' })).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'Save & Push' }));

    expect(await screen.findByText(/git sync not configured/)).toBeInTheDocument();
  });

  it('shows loading state during generation', async () => {
    // Make the generate endpoint slow
    server.use(
      http.post('/api/analyze/timing/generate', async () => {
        await new Promise((r) => setTimeout(r, 100));
        return HttpResponse.json({
          recommendations: [{ day: 'Wednesday', time: '14:00', reasoning: 'test' }],
          videoCount: 10,
        });
      }),
    );

    const user = userEvent.setup();
    renderWithProviders(<AnalyzeTiming />);

    await user.click(screen.getByRole('button', { name: 'Generate New Recommendations' }));

    expect(screen.getByText(/Fetching YouTube analytics/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Generating...' })).toBeDisabled();
  });

  it('shows error state on generation failure', async () => {
    server.use(
      http.post('/api/analyze/timing/generate', () =>
        new HttpResponse('AI provider failed', { status: 500 }),
      ),
    );

    const user = userEvent.setup();
    renderWithProviders(<AnalyzeTiming />);

    await user.click(screen.getByRole('button', { name: 'Generate New Recommendations' }));

    await waitFor(() => {
      expect(screen.getByText(/AI provider failed/)).toBeInTheDocument();
    });
  });

  it('shows error state when loading recommendations fails', async () => {
    server.use(
      http.get('/api/analyze/timing', () =>
        new HttpResponse('Internal server error', { status: 500 }),
      ),
    );

    renderWithProviders(<AnalyzeTiming />);

    await waitFor(() => {
      expect(screen.getByText(/Failed to load timing recommendations|Internal server error/)).toBeInTheDocument();
    });
  });
});
