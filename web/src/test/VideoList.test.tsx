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

  it('shows sponsored indicator for sponsored videos', async () => {
    renderWithRoute('1');
    await screen.findByText('Test Video Title');
    const sponsoredIcons = screen.getAllByTitle('Sponsored');
    expect(sponsoredIcons).toHaveLength(1);
    expect(sponsoredIcons[0]).toHaveTextContent('$');
    expect(sponsoredIcons[0]).toHaveClass('text-orange-400');
  });

  it('applies far-future styling to far-future videos', async () => {
    renderWithRoute('1');
    const farFutureName = await screen.findByText('Another Video');
    expect(farFutureName.closest('td')).toHaveClass('text-cyan-400');
  });

  it('does not apply far-future styling to normal videos', async () => {
    renderWithRoute('1');
    const normalName = await screen.findByText('Test Video Title');
    expect(normalName.closest('td')).toHaveClass('text-gray-100');
    expect(normalName.closest('td')).not.toHaveClass('text-cyan-400');
  });
});
