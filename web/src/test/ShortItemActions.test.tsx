import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ShortItemActions } from '../components/forms/ShortItemActions';
import { server } from './server';
import { http, HttpResponse } from 'msw';
import { describe, it, expect } from 'vitest';

function renderWithClient(ui: React.ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

describe('ShortItemActions', () => {
  it('renders "Upload to Drive" when no driveFileId', () => {
    renderWithClient(
      <ShortItemActions videoName="test-video" category="devops" shortId="short1" />,
    );
    expect(screen.getByText('Upload to Drive')).toBeInTheDocument();
  });

  it('shows "Replace" and Download when driveFileId exists', () => {
    renderWithClient(
      <ShortItemActions
        videoName="test-video"
        category="devops"
        shortId="short1"
        driveFileId="existing-id"
      />,
    );
    expect(screen.getByText('Replace')).toBeInTheDocument();
    expect(screen.getByTestId('short-download-link-short1')).toBeInTheDocument();
  });

  it('shows progress bar during upload', async () => {
    server.use(
      http.post('/api/drive/upload/short/:videoName/:shortId', async () => {
        await new Promise((r) => setTimeout(r, 200));
        return HttpResponse.json({ driveFileId: 'short-drive-id', filePath: 'drive://short-drive-id' });
      }),
    );

    renderWithClient(
      <ShortItemActions videoName="test-video" category="devops" shortId="short1" />,
    );

    const fileInput = screen.getByTestId('short-file-input-short1');
    const file = new File(['fake-short'], 'short.mp4', { type: 'video/mp4' });
    await userEvent.upload(fileInput, file);

    expect(screen.getByText('Uploading...')).toBeInTheDocument();
    await screen.findByText('Uploaded');
  });

  it('shows "Uploaded" on success', async () => {
    renderWithClient(
      <ShortItemActions videoName="test-video" category="devops" shortId="short1" />,
    );

    const fileInput = screen.getByTestId('short-file-input-short1');
    const file = new File(['fake-short'], 'short.mp4', { type: 'video/mp4' });
    await userEvent.upload(fileInput, file);

    await screen.findByText('Uploaded');
    expect(screen.getByText('Uploaded')).toBeInTheDocument();
    expect(screen.getByTestId('short-download-link-short1')).toBeInTheDocument();
  });

  it('shows error on upload failure', async () => {
    server.use(
      http.post('/api/drive/upload/short/:videoName/:shortId', () =>
        new HttpResponse('Drive quota exceeded', { status: 500 }),
      ),
    );

    renderWithClient(
      <ShortItemActions videoName="test-video" category="devops" shortId="short1" />,
    );

    const fileInput = screen.getByTestId('short-file-input-short1');
    const file = new File(['fake-short'], 'short.mp4', { type: 'video/mp4' });
    await userEvent.upload(fileInput, file);

    const errorEl = await screen.findByTestId('short-upload-error-short1');
    expect(errorEl).toBeInTheDocument();
  });

  it('shows "Publish to YouTube" when prerequisites met', () => {
    renderWithClient(
      <ShortItemActions
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
      <ShortItemActions
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
      <ShortItemActions
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
      <ShortItemActions
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
      <ShortItemActions
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
