import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect } from 'vitest';
import { VideoList } from '../pages/VideoList';

function renderWithRoute(phaseId: string) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[`/phases/${phaseId}`]}>
        <Routes>
          <Route path="/phases/:phaseId" element={<VideoList />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('VideoList', () => {
  it('renders video rows', async () => {
    renderWithRoute('1');
    expect(await screen.findByText('Test Video Title')).toBeInTheDocument();
    expect(screen.getByText('Another Video')).toBeInTheDocument();
  });

  it('shows loading state', () => {
    renderWithRoute('1');
    expect(screen.getByText('Loading videos...')).toBeInTheDocument();
  });

  it('displays category column', async () => {
    renderWithRoute('1');
    const cells = await screen.findAllByText('devops');
    expect(cells.length).toBeGreaterThan(0);
  });
});
