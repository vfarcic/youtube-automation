import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import { describe, it, expect } from 'vitest';
import { AskMeAnything } from '../pages/AskMeAnything';

function renderPage() {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <AskMeAnything />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('AskMeAnything', () => {
  it('renders the page with input fields', () => {
    renderPage();
    expect(screen.getByText('Ask Me Anything')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('e.g., dQw4w9WgXcQ')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Generate with AI' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Apply to YouTube' })).toBeInTheDocument();
  });

  it('disables Generate button when video ID is empty', () => {
    renderPage();
    expect(screen.getByRole('button', { name: 'Generate with AI' })).toBeDisabled();
  });

  it('disables Apply button when no content', () => {
    renderPage();
    expect(screen.getByRole('button', { name: 'Apply to YouTube' })).toBeDisabled();
  });

  it('generates content and populates fields', async () => {
    renderPage();
    await userEvent.type(screen.getByPlaceholderText('e.g., dQw4w9WgXcQ'), 'abc123');
    await userEvent.click(screen.getByRole('button', { name: 'Generate with AI' }));

    await waitFor(() => {
      expect(screen.getByDisplayValue('Generated AMA Title')).toBeInTheDocument();
    });
    expect(screen.getByDisplayValue('Generated AMA Description')).toBeInTheDocument();
    expect(screen.getByDisplayValue('ama,generated,tags')).toBeInTheDocument();
    expect(screen.getByText('Content generated. Review and edit, then apply to YouTube.')).toBeInTheDocument();
  });

  it('applies content to YouTube', async () => {
    renderPage();
    await userEvent.type(screen.getByPlaceholderText('e.g., dQw4w9WgXcQ'), 'abc123');
    await userEvent.click(screen.getByRole('button', { name: 'Generate with AI' }));

    await waitFor(() => {
      expect(screen.getByDisplayValue('Generated AMA Title')).toBeInTheDocument();
    });

    await userEvent.click(screen.getByRole('button', { name: 'Apply to YouTube' }));

    await waitFor(() => {
      expect(screen.getByText('Video updated on YouTube!')).toBeInTheDocument();
    });
  });

  it('allows editing fields before applying', async () => {
    renderPage();
    await userEvent.type(screen.getByPlaceholderText('e.g., dQw4w9WgXcQ'), 'abc123');
    await userEvent.click(screen.getByRole('button', { name: 'Generate with AI' }));

    await waitFor(() => {
      expect(screen.getByDisplayValue('Generated AMA Title')).toBeInTheDocument();
    });

    const titleInput = screen.getByDisplayValue('Generated AMA Title');
    await userEvent.clear(titleInput);
    await userEvent.type(titleInput, 'Custom Title');
    expect(screen.getByDisplayValue('Custom Title')).toBeInTheDocument();
  });
});
