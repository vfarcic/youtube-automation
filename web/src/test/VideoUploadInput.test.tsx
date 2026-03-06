import { render, screen, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { VideoUploadInput } from '../components/forms/VideoUploadInput';
import { setupServer } from 'msw/node';
import { http, HttpResponse } from 'msw';
import { beforeAll, afterAll, afterEach, describe, it, expect, vi } from 'vitest';

const server = setupServer(
  http.post('/api/drive/upload/video/:videoName', () =>
    HttpResponse.json({ driveFileId: 'video-drive-id', videoFile: 'drive://video-drive-id' }),
  ),
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

function renderWithClient(ui: React.ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

describe('VideoUploadInput', () => {
  it('renders Upload to Drive button', () => {
    renderWithClient(
      <VideoUploadInput videoName="test-video" category="devops" />,
    );
    expect(screen.getByText('Upload to Drive')).toBeInTheDocument();
  });

  it('shows Replace button and Download link when currentDriveFileId exists', () => {
    renderWithClient(
      <VideoUploadInput
        videoName="test-video"
        category="devops"
        currentDriveFileId="existing-video-id"
      />,
    );
    expect(screen.getByText('Replace')).toBeInTheDocument();
    expect(screen.getByTestId('video-download-link')).toBeInTheDocument();
  });

  it('does not show Download link when no driveFileId', () => {
    renderWithClient(
      <VideoUploadInput videoName="test-video" category="devops" />,
    );
    expect(screen.queryByTestId('video-download-link')).not.toBeInTheDocument();
  });

  it('shows Uploaded after successful upload', async () => {
    renderWithClient(
      <VideoUploadInput videoName="test-video" category="devops" />,
    );

    const fileInput = screen.getByTestId('video-file-input');
    const file = new File(['fake-video'], 'recording.mp4', { type: 'video/mp4' });
    await userEvent.upload(fileInput, file);

    const uploaded = await screen.findByText('Uploaded');
    expect(uploaded).toBeInTheDocument();

    const downloadLink = await screen.findByTestId('video-download-link');
    expect(downloadLink).toBeInTheDocument();
  });

  it('shows error on upload failure', async () => {
    server.use(
      http.post('/api/drive/upload/video/:videoName', () =>
        new HttpResponse('Drive quota exceeded', { status: 500 }),
      ),
    );

    renderWithClient(
      <VideoUploadInput videoName="test-video" category="devops" />,
    );

    const fileInput = screen.getByTestId('video-file-input');
    const file = new File(['fake-video'], 'recording.mp4', { type: 'video/mp4' });
    await userEvent.upload(fileInput, file);

    const errorEl = await screen.findByTestId('video-upload-error');
    expect(errorEl).toBeInTheDocument();
  });

  it('shows download link when driveFileId exists', () => {
    renderWithClient(
      <VideoUploadInput
        videoName="test-video"
        category="devops"
        currentDriveFileId="existing-video-id"
      />,
    );
    const link = screen.getByTestId('video-download-link');
    expect(link).toBeInTheDocument();
    expect(link.tagName).toBe('A');
    expect(link).toHaveAttribute('download');
    expect(link).toHaveAttribute('href', expect.stringContaining('/api/drive/download/video/test-video'));
  });

  it('does not show download link when no driveFileId', () => {
    renderWithClient(
      <VideoUploadInput videoName="test-video" category="devops" />,
    );
    expect(screen.queryByTestId('video-download-link')).not.toBeInTheDocument();
  });

  it('shows progress bar during upload', async () => {
    // Use a handler that delays response to keep mutation pending
    server.use(
      http.post('/api/drive/upload/video/:videoName', async () => {
        await new Promise((r) => setTimeout(r, 200));
        return HttpResponse.json({ driveFileId: 'video-drive-id', videoFile: 'drive://video-drive-id' });
      }),
    );

    renderWithClient(
      <VideoUploadInput videoName="test-video" category="devops" />,
    );

    const fileInput = screen.getByTestId('video-file-input');
    const file = new File(['fake-video'], 'recording.mp4', { type: 'video/mp4' });
    await userEvent.upload(fileInput, file);

    // The button should show "Uploading..." while pending
    expect(screen.getByText('Uploading...')).toBeInTheDocument();

    // Wait for completion
    await screen.findByText('Uploaded');
  });

  it('shows "Uploading to Drive..." label when progress >= 50', async () => {
    // Mock uploadFileWithProgress to capture the onProgress callback
    // and control when the promise resolves
    let capturedOnProgress: ((percent: number) => void) | undefined;
    let resolveUpload: ((value: any) => void) | undefined;

    const clientModule = await import('../api/client');
    const origFn = clientModule.uploadFileWithProgress;
    vi.spyOn(clientModule, 'uploadFileWithProgress').mockImplementation(
      (_path, _file, _field, onProgress) => {
        capturedOnProgress = onProgress;
        return new Promise((resolve) => { resolveUpload = resolve; });
      },
    );

    renderWithClient(
      <VideoUploadInput videoName="test-video" category="devops" />,
    );

    const fileInput = screen.getByTestId('video-file-input');
    const file = new File(['fake-video'], 'recording.mp4', { type: 'video/mp4' });
    await userEvent.upload(fileInput, file);

    // Mutation is pending
    expect(screen.getByText('Uploading...')).toBeInTheDocument();

    // Simulate progress at 30% (XHR send phase) — should show percentage
    act(() => capturedOnProgress!(30));
    expect(screen.getByText('30%')).toBeInTheDocument();

    // Simulate progress at 50% (server processing phase) — should show Drive label
    act(() => capturedOnProgress!(50));
    expect(screen.getByText('Uploading to Drive...')).toBeInTheDocument();
    expect(screen.getByTestId('upload-progress')).toBeInTheDocument();

    // Simulate progress at 75% — should still show Drive label
    act(() => capturedOnProgress!(75));
    expect(screen.getByText('Uploading to Drive...')).toBeInTheDocument();

    // Resolve the upload
    await act(async () => resolveUpload!({ driveFileId: 'id', videoFile: 'drive://id' }));
    await screen.findByText('Uploaded');

    vi.restoreAllMocks();
  });
});
