import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi } from 'vitest';
import { AIGenerateButton } from '../components/forms/AIGenerateButton';

function renderButton(fieldName: string, onApply = vi.fn()) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return {
    onApply,
    ...render(
      <QueryClientProvider client={qc}>
        <AIGenerateButton
          fieldName={fieldName}
          category="devops"
          videoName="test-video"
          onApply={onApply}
        />
      </QueryClientProvider>,
    ),
  };
}

describe('AIGenerateButton', () => {
  it('renders nothing for non-AI fields', () => {
    const { container } = renderButton('projectName');
    expect(container.innerHTML).toBe('');
  });

  it('renders Generate button for AI-eligible fields', () => {
    renderButton('description');
    expect(screen.getByRole('button', { name: 'Generate Description' })).toBeInTheDocument();
  });

  it('generates description and applies it', async () => {
    const onApply = vi.fn();
    renderButton('description', onApply);
    await userEvent.click(screen.getByRole('button', { name: 'Generate Description' }));
    await waitFor(() => {
      expect(screen.getByText('AI generated description')).toBeInTheDocument();
    });
    await userEvent.click(screen.getByRole('button', { name: 'Apply' }));
    expect(onApply).toHaveBeenCalledWith('AI generated description');
    // Results panel should be dismissed
    expect(screen.queryByText('AI generated description')).not.toBeInTheDocument();
  });

  it('generates tags and applies them', async () => {
    const onApply = vi.fn();
    renderButton('tags', onApply);
    await userEvent.click(screen.getByRole('button', { name: 'Generate Tags' }));
    await waitFor(() => {
      expect(screen.getByText('ai,generated,tags')).toBeInTheDocument();
    });
    await userEvent.click(screen.getByRole('button', { name: 'Apply' }));
    expect(onApply).toHaveBeenCalledWith('ai,generated,tags');
    expect(screen.queryByText('ai,generated,tags')).not.toBeInTheDocument();
  });

  it('generates titles with checkbox selection', async () => {
    const onApply = vi.fn();
    renderButton('titles', onApply);
    await userEvent.click(screen.getByRole('button', { name: 'Generate Titles' }));
    await waitFor(() => {
      expect(screen.getByText('AI Title 1')).toBeInTheDocument();
    });
    // Select first two titles
    const checkboxes = screen.getAllByRole('checkbox');
    await userEvent.click(checkboxes[0]);
    await userEvent.click(checkboxes[1]);
    await userEvent.click(screen.getByRole('button', { name: 'Apply Selected' }));
    expect(onApply).toHaveBeenCalledWith([
      { index: 1, text: 'AI Title 1', watchTimeShare: 0 },
      { index: 2, text: 'AI Title 2', watchTimeShare: 0 },
    ]);
  });

  it('generates tweets with radio selection and appends [YOUTUBE]', async () => {
    const onApply = vi.fn();
    renderButton('tweet', onApply);
    await userEvent.click(screen.getByRole('button', { name: 'Generate Tweets' }));
    await waitFor(() => {
      expect(screen.getByText('Tweet 1')).toBeInTheDocument();
    });
    const radios = screen.getAllByRole('radio');
    await userEvent.click(radios[0]);
    await userEvent.click(screen.getByRole('button', { name: 'Apply Selected' }));
    expect(onApply).toHaveBeenCalledWith('Tweet 1\n\n[YOUTUBE]');
  });

  it('generates shorts with checkbox selection', async () => {
    const onApply = vi.fn();
    renderButton('shorts', onApply);
    await userEvent.click(screen.getByRole('button', { name: 'Generate Shorts' }));
    await waitFor(() => {
      expect(screen.getByText('Short One')).toBeInTheDocument();
    });
    const checkboxes = screen.getAllByRole('checkbox');
    await userEvent.click(checkboxes[0]);
    await userEvent.click(screen.getByRole('button', { name: 'Apply Selected' }));
    expect(onApply).toHaveBeenCalledWith([
      { id: 'short1', title: 'Short One', text: 'text', filePath: '', scheduledDate: '', youtubeId: '' },
    ]);
  });
});
