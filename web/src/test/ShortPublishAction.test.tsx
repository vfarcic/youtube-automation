import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ShortPublishAction } from '../components/forms/ShortPublishAction';
import { describe, it, expect } from 'vitest';

function renderWithClient(ui: React.ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

describe('ShortPublishAction', () => {
  it('shows "Publish to YouTube" when prerequisites met', () => {
    renderWithClient(
      <ShortPublishAction
        videoName="test-video"
        category="devops"
        shortId="short1"
        driveFileId="some-id"
        scheduledDate="2026-01-15T10:00"
      />,
    );
    const btn = screen.getByText('Publish to YouTube');
    expect(btn).toBeInTheDocument();
    expect(btn).not.toBeDisabled();
  });

  it('disables publish with "Upload a file first" when no file', () => {
    renderWithClient(
      <ShortPublishAction
        videoName="test-video"
        category="devops"
        shortId="short1"
        scheduledDate="2026-01-15T10:00"
      />,
    );
    expect(screen.getByText('Publish to YouTube')).toBeDisabled();
    expect(screen.getByText('Upload a file first')).toBeInTheDocument();
  });

  it('disables publish with "Set scheduled date first" when no date', () => {
    renderWithClient(
      <ShortPublishAction
        videoName="test-video"
        category="devops"
        shortId="short1"
        driveFileId="some-id"
      />,
    );
    expect(screen.getByText('Publish to YouTube')).toBeDisabled();
    expect(screen.getByText('Set scheduled date first')).toBeInTheDocument();
  });

  it('shows published badge when youtubeId exists', () => {
    renderWithClient(
      <ShortPublishAction
        videoName="test-video"
        category="devops"
        shortId="short1"
        youtubeId="yt-existing-456"
      />,
    );
    expect(screen.getByText('Published:')).toBeInTheDocument();
    expect(screen.getByText('yt-existing-456')).toBeInTheDocument();
  });

  it('publishes short and shows published badge', async () => {
    renderWithClient(
      <ShortPublishAction
        videoName="test-video"
        category="devops"
        shortId="short1"
        driveFileId="some-id"
        scheduledDate="2026-01-15T10:00"
      />,
    );

    await userEvent.click(screen.getByText('Publish to YouTube'));
    await screen.findByText('Published:');
    expect(screen.getByText('yt-short-456')).toBeInTheDocument();
  });
});
