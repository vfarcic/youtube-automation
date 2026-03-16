import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ShortUploadAction } from '../components/forms/ShortUploadAction';
import { server } from './server';
import { http, HttpResponse } from 'msw';
import { describe, it, expect } from 'vitest';

function renderWithClient(ui: React.ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

describe('ShortUploadAction', () => {
  it('renders "Upload to Drive" when no driveFileId', () => {
    renderWithClient(
      <ShortUploadAction videoName="test-video" category="devops" shortId="short1" />,
    );
    expect(screen.getByText('Upload to Drive')).toBeInTheDocument();
  });

  it('shows "Replace" and Download when driveFileId exists', () => {
    renderWithClient(
      <ShortUploadAction
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
      <ShortUploadAction videoName="test-video" category="devops" shortId="short1" />,
    );

    const fileInput = screen.getByTestId('short-file-input-short1');
    const file = new File(['fake-short'], 'short.mp4', { type: 'video/mp4' });
    await userEvent.upload(fileInput, file);

    expect(screen.getByText('Uploading...')).toBeInTheDocument();
    await screen.findByText('Uploaded');
  });

  it('shows "Uploaded" on success', async () => {
    renderWithClient(
      <ShortUploadAction videoName="test-video" category="devops" shortId="short1" />,
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
      <ShortUploadAction videoName="test-video" category="devops" shortId="short1" />,
    );

    const fileInput = screen.getByTestId('short-file-input-short1');
    const file = new File(['fake-short'], 'short.mp4', { type: 'video/mp4' });
    await userEvent.upload(fileInput, file);

    const errorEl = await screen.findByTestId('short-upload-error-short1');
    expect(errorEl).toBeInTheDocument();
  });
});
