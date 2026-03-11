import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect } from 'vitest';
import { AnalyzeTitles } from '../pages/AnalyzeTitles';

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

describe('AnalyzeTitles', () => {
  it('renders the page with Run Analysis button', () => {
    renderWithProviders(<AnalyzeTitles />);
    expect(screen.getByText('Title Analysis')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Run Analysis' })).toBeInTheDocument();
  });

  it('runs analysis and displays results', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AnalyzeTitles />);

    await user.click(screen.getByRole('button', { name: 'Run Analysis' }));

    expect(await screen.findByText(/Analyzed/)).toBeInTheDocument();
    expect(screen.getByText('5')).toBeInTheDocument(); // videoCount from mock handler in test/handlers.ts

    expect(screen.getByText('High-Performing Patterns')).toBeInTheDocument();
    expect(screen.getByText('Provocative')).toBeInTheDocument();

    expect(screen.getByText('Low-Performing Patterns')).toBeInTheDocument();
    expect(screen.getByText('Listicle')).toBeInTheDocument();

    expect(screen.getByText('Recommendations')).toBeInTheDocument();
    expect(screen.getByText('Use provocative titles')).toBeInTheDocument();

    expect(screen.getByText('Proposed titles.md')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Save & Push/ })).toBeInTheDocument();
  });

  it('applies titles.md and shows success', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AnalyzeTitles />);

    await user.click(screen.getByRole('button', { name: 'Run Analysis' }));
    expect(await screen.findByText('Proposed titles.md')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: /Save & Push/ }));

    await waitFor(() => {
      expect(screen.getByText(/updated and pushed/)).toBeInTheDocument();
    });
  });
});
