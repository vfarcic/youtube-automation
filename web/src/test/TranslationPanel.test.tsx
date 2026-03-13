import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi } from 'vitest';
import { TranslationPanel } from '../components/forms/TranslationPanel';

function renderPanel(onApply = vi.fn()) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return {
    onApply,
    ...render(
      <QueryClientProvider client={qc}>
        <TranslationPanel
          category="devops"
          videoName="test-video"
          onApply={onApply}
        />
      </QueryClientProvider>,
    ),
  };
}

describe('TranslationPanel', () => {
  it('renders collapsed Translate Video button', () => {
    renderPanel();
    expect(screen.getByRole('button', { name: 'Translate Video' })).toBeInTheDocument();
  });

  it('expands panel on click', async () => {
    renderPanel();
    await userEvent.click(screen.getByRole('button', { name: 'Translate Video' }));
    expect(screen.getByPlaceholderText('Target language (e.g. Spanish)')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Translate' })).toBeInTheDocument();
  });

  it('disables Translate button when language is empty', async () => {
    renderPanel();
    await userEvent.click(screen.getByRole('button', { name: 'Translate Video' }));
    expect(screen.getByRole('button', { name: 'Translate' })).toBeDisabled();
  });

  it('translates and shows results', async () => {
    renderPanel();
    await userEvent.click(screen.getByRole('button', { name: 'Translate Video' }));
    await userEvent.type(screen.getByPlaceholderText('Target language (e.g. Spanish)'), 'Spanish');
    await userEvent.click(screen.getByRole('button', { name: 'Translate' }));
    await waitFor(() => {
      expect(screen.getByText('Titulo')).toBeInTheDocument();
    });
    expect(screen.getByText('Desc')).toBeInTheDocument();
    expect(screen.getByText('tags')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Apply All' })).toBeInTheDocument();
  });

  it('calls onApply with translated fields when Apply All is clicked', async () => {
    const onApply = vi.fn();
    renderPanel(onApply);
    await userEvent.click(screen.getByRole('button', { name: 'Translate Video' }));
    await userEvent.type(screen.getByPlaceholderText('Target language (e.g. Spanish)'), 'Spanish');
    await userEvent.click(screen.getByRole('button', { name: 'Translate' }));
    await waitFor(() => {
      expect(screen.getByText('Titulo')).toBeInTheDocument();
    });
    await userEvent.click(screen.getByRole('button', { name: 'Apply All' }));
    expect(onApply).toHaveBeenCalledWith({
      title: 'Titulo',
      description: 'Desc',
      tags: 'tags',
    });
  });

  it('collapses panel after Apply All', async () => {
    const onApply = vi.fn();
    renderPanel(onApply);
    await userEvent.click(screen.getByRole('button', { name: 'Translate Video' }));
    await userEvent.type(screen.getByPlaceholderText('Target language (e.g. Spanish)'), 'Spanish');
    await userEvent.click(screen.getByRole('button', { name: 'Translate' }));
    await waitFor(() => {
      expect(screen.getByText('Titulo')).toBeInTheDocument();
    });
    await userEvent.click(screen.getByRole('button', { name: 'Apply All' }));
    // Panel should collapse back to button
    expect(screen.getByRole('button', { name: 'Translate Video' })).toBeInTheDocument();
  });

  it('closes panel on Close click', async () => {
    renderPanel();
    await userEvent.click(screen.getByRole('button', { name: 'Translate Video' }));
    await userEvent.click(screen.getByRole('button', { name: 'Close' }));
    expect(screen.getByRole('button', { name: 'Translate Video' })).toBeInTheDocument();
  });
});
