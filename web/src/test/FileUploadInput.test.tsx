import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { FileUploadInput } from '../components/forms/FileUploadInput';
import { setupServer } from 'msw/node';
import { http, HttpResponse } from 'msw';
import { beforeAll, afterAll, afterEach, describe, it, expect } from 'vitest';

const server = setupServer(
  http.post('/api/drive/upload/thumbnail/:videoName', () =>
    HttpResponse.json({ driveFileId: 'new-drive-id', variantIndex: 0 }),
  ),
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

function renderWithClient(ui: React.ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

describe('FileUploadInput', () => {
  it('renders Upload to Drive button', () => {
    renderWithClient(
      <FileUploadInput videoName="test-video" category="devops" variantIndex={0} />,
    );
    expect(screen.getByText('Upload to Drive')).toBeInTheDocument();
  });

  it('shows Uploaded badge when driveFileId exists', () => {
    renderWithClient(
      <FileUploadInput
        videoName="test-video"
        category="devops"
        variantIndex={0}
        currentDriveFileId="existing-id"
      />,
    );
    expect(screen.getByTestId('drive-uploaded-badge')).toBeInTheDocument();
  });

  it('does not show Uploaded badge when no driveFileId', () => {
    renderWithClient(
      <FileUploadInput videoName="test-video" category="devops" variantIndex={0} />,
    );
    expect(screen.queryByTestId('drive-uploaded-badge')).not.toBeInTheDocument();
  });

  it('shows Uploaded after successful upload', async () => {
    renderWithClient(
      <FileUploadInput videoName="test-video" category="devops" variantIndex={0} />,
    );

    const fileInput = screen.getByTestId('thumbnail-file-input');
    const file = new File(['fake-image'], 'thumb.png', { type: 'image/png' });
    await userEvent.upload(fileInput, file);

    // Wait for success
    const uploaded = await screen.findByText('Uploaded');
    expect(uploaded).toBeInTheDocument();

    // Drive file ID should be displayed
    const badge = await screen.findByTestId('drive-uploaded-badge');
    expect(badge).toBeInTheDocument();
  });
});
