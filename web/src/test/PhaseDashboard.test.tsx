import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect } from 'vitest';
import { PhaseDashboard } from '../pages/PhaseDashboard';

function renderWithProviders(ui: React.ReactElement) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>{ui}</MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('PhaseDashboard', () => {
  it('renders 8 phase cards with correct counts', async () => {
    renderWithProviders(<PhaseDashboard />);

    expect(await screen.findByText('Ideas')).toBeInTheDocument();
    expect(screen.getByText('Started')).toBeInTheDocument();
    expect(screen.getByText('Material Done')).toBeInTheDocument();
    expect(screen.getByText('Edit Requested')).toBeInTheDocument();
    expect(screen.getByText('Publish Pending')).toBeInTheDocument();
    expect(screen.getByText('Published')).toBeInTheDocument();
    expect(screen.getByText('Delayed')).toBeInTheDocument();
    expect(screen.getByText('Sponsored/Blocked')).toBeInTheDocument();

    expect(screen.getByText('10')).toBeInTheDocument();
  });

  it('shows loading state', () => {
    renderWithProviders(<PhaseDashboard />);
    expect(screen.getByText('Loading phases...')).toBeInTheDocument();
  });
});
