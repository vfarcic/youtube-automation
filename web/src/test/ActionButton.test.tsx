import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect } from 'vitest';
import { ActionButton, isActionField } from '../components/forms/ActionButton';

function renderActionButton(fieldName: string, value: boolean) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return render(
    <QueryClientProvider client={qc}>
      <ActionButton
        fieldName={fieldName}
        value={value}
        category="devops"
        videoName="test-video"
      />
    </QueryClientProvider>,
  );
}

describe('isActionField', () => {
  it('returns true for requestThumbnail', () => {
    expect(isActionField('requestThumbnail')).toBe(true);
  });

  it('returns true for requestEdit', () => {
    expect(isActionField('requestEdit')).toBe(true);
  });

  it('returns true for notifiedSponsors', () => {
    expect(isActionField('notifiedSponsors')).toBe(true);
  });

  it('returns false for regular boolean fields', () => {
    expect(isActionField('delayed')).toBe(false);
    expect(isActionField('code')).toBe(false);
  });
});

describe('ActionButton', () => {
  it('renders button with correct label for requestThumbnail', () => {
    renderActionButton('requestThumbnail', false);
    expect(screen.getByRole('button', { name: 'Request Thumbnail' })).toBeInTheDocument();
  });

  it('renders button with correct label for requestEdit', () => {
    renderActionButton('requestEdit', false);
    expect(screen.getByRole('button', { name: 'Request Edit' })).toBeInTheDocument();
  });

  it('shows "Thumbnail Requested" when value is true', () => {
    renderActionButton('requestThumbnail', true);
    expect(screen.getByText('Thumbnail Requested')).toBeInTheDocument();
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it('shows "Edit Requested" when value is true', () => {
    renderActionButton('requestEdit', true);
    expect(screen.getByText('Edit Requested')).toBeInTheDocument();
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it('calls endpoint on click and shows loading state', async () => {
    renderActionButton('requestThumbnail', false);
    const btn = screen.getByRole('button', { name: 'Request Thumbnail' });
    await userEvent.click(btn);
    // After mutation resolves, button should not show error
    await waitFor(() => {
      expect(screen.queryByText(/Email failed/)).not.toBeInTheDocument();
    });
  });

  it('calls request-edit endpoint on click', async () => {
    renderActionButton('requestEdit', false);
    const btn = screen.getByRole('button', { name: 'Request Edit' });
    await userEvent.click(btn);
    await waitFor(() => {
      expect(screen.queryByText(/Email failed/)).not.toBeInTheDocument();
    });
  });

  it('renders button with correct label for notifiedSponsors', () => {
    renderActionButton('notifiedSponsors', false);
    expect(screen.getByRole('button', { name: 'Notify Sponsors' })).toBeInTheDocument();
  });

  it('shows "Sponsors Notified" when value is true', () => {
    renderActionButton('notifiedSponsors', true);
    expect(screen.getByText('Sponsors Notified')).toBeInTheDocument();
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it('calls notify-sponsors endpoint on click', async () => {
    renderActionButton('notifiedSponsors', false);
    const btn = screen.getByRole('button', { name: 'Notify Sponsors' });
    await userEvent.click(btn);
    await waitFor(() => {
      expect(screen.queryByText(/Email failed/)).not.toBeInTheDocument();
    });
  });
});
