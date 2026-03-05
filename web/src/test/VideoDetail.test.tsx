import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect } from 'vitest';
import { VideoDetail } from '../pages/VideoDetail';

function renderWithRoute() {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={['/videos/devops/test-video']}>
        <Routes>
          <Route
            path="/videos/:category/:videoName"
            element={<VideoDetail />}
          />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('VideoDetail', () => {
  it('displays video name', async () => {
    renderWithRoute();
    expect(await screen.findByText('test-video')).toBeInTheDocument();
  });

  it('displays fields grouped by aspect', async () => {
    renderWithRoute();
    await screen.findByText('test-video');

    // Each aspect label appears in both the progress section and the fields section
    expect(screen.getAllByText('Initial Details').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('Work Progress').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('Definition').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('Post Production').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('Publishing').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('Post Publish').length).toBeGreaterThanOrEqual(1);
  });

  it('shows overall progress', async () => {
    renderWithRoute();
    await screen.findByText('test-video');
    expect(screen.getByText('Overall Progress')).toBeInTheDocument();
  });
});
