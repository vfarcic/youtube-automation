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

  it('shows Download link when driveFileId exists', () => {
    renderWithClient(
      <FileUploadInput
        videoName="test-video"
        category="devops"
        variantIndex={0}
        currentDriveFileId="existing-id"
      />,
    );
    const downloadLink = screen.getByText('Download');
    expect(downloadLink).toBeInTheDocument();
    expect(downloadLink).toHaveAttribute('href', 'https://drive.google.com/uc?id=existing-id&export=download');
  });

  it('does not show Download link when no driveFileId', () => {
    renderWithClient(
      <FileUploadInput videoName="test-video" category="devops" variantIndex={0} />,
    );
    expect(screen.queryByText('Download')).not.toBeInTheDocument();
  });

  it('shows Uploaded and Download after successful upload', async () => {
    renderWithClient(
      <FileUploadInput videoName="test-video" category="devops" variantIndex={0} />,
    );

    const fileInput = screen.getByTestId('thumbnail-file-input');
    const file = new File(['fake-image'], 'thumb.png', { type: 'image/png' });
    await userEvent.upload(fileInput, file);

    // Wait for success
    const uploaded = await screen.findByText('Uploaded');
    expect(uploaded).toBeInTheDocument();

    // Download link should appear
    const downloadLink = await screen.findByText('Download');
    expect(downloadLink).toBeInTheDocument();
  });
});
