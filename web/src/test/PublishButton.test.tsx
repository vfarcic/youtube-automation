import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect } from 'vitest';
import { PublishButton, isPublishField } from '../components/forms/PublishButton';
import { mockVideo } from './handlers';

function renderPublishButton(
  fieldName: string,
  value: string,
  videoOverrides: Partial<typeof mockVideo> = {},
) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  const video = { ...mockVideo, ...videoOverrides };
  return render(
    <QueryClientProvider client={qc}>
      <PublishButton
        fieldName={fieldName}
        value={value}
        category="devops"
        videoName="test-video"
        video={video}
      />
    </QueryClientProvider>,
  );
}

describe('isPublishField', () => {
  it('returns true for videoId', () => {
    expect(isPublishField('videoId')).toBe(true);
  });

  it('returns true for hugoPath', () => {
    expect(isPublishField('hugoPath')).toBe(true);
  });

  it('returns false for regular fields', () => {
    expect(isPublishField('description')).toBe(false);
    expect(isPublishField('delayed')).toBe(false);
  });
});

describe('PublishButton', () => {
  it('shows published state with videoId when value is set', () => {
    renderPublishButton('videoId', 'abc123');
    expect(screen.getByText('Published:')).toBeInTheDocument();
    expect(screen.getByText('abc123')).toBeInTheDocument();
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it('shows published state with hugoPath when value is set', () => {
    renderPublishButton('hugoPath', '/content/test.md');
    expect(screen.getByText('Published:')).toBeInTheDocument();
    expect(screen.getByText('/content/test.md')).toBeInTheDocument();
  });

  it('renders YouTube publish button when no value and video file exists', () => {
    renderPublishButton('videoId', '', { uploadVideo: 'video.mp4' });
    expect(screen.getByRole('button', { name: 'Publish to YouTube' })).toBeInTheDocument();
  });

  it('disables YouTube publish button when no video file', () => {
    renderPublishButton('videoId', '', { uploadVideo: '', videoFile: undefined, videoDriveFileId: undefined });
    const btn = screen.getByRole('button', { name: 'Publish to YouTube' });
    expect(btn).toBeDisabled();
    expect(screen.getByText('Upload a video file first')).toBeInTheDocument();
  });

  it('enables YouTube publish button when video is on Drive', () => {
    renderPublishButton('videoId', '', { uploadVideo: '', videoDriveFileId: 'drive-abc' });
    const btn = screen.getByRole('button', { name: 'Publish to YouTube' });
    expect(btn).not.toBeDisabled();
  });

  it('enables YouTube publish button when videoFile is set', () => {
    renderPublishButton('videoId', '', { uploadVideo: '', videoFile: 'drive://abc123' });
    const btn = screen.getByRole('button', { name: 'Publish to YouTube' });
    expect(btn).not.toBeDisabled();
  });

  it('disables Hugo publish button when no videoId', () => {
    renderPublishButton('hugoPath', '', { videoId: '' });
    const btn = screen.getByRole('button', { name: 'Publish Hugo Post' });
    expect(btn).toBeDisabled();
    expect(screen.getByText('Publish to YouTube first')).toBeInTheDocument();
  });

  it('calls publish endpoint on click and shows result', async () => {
    renderPublishButton('videoId', '', { uploadVideo: 'video.mp4' });
    const btn = screen.getByRole('button', { name: 'Publish to YouTube' });
    await userEvent.click(btn);
    await waitFor(() => {
      expect(screen.getByText('yt-abc123')).toBeInTheDocument();
    });
  });

  it('calls Hugo publish endpoint on click', async () => {
    renderPublishButton('hugoPath', '', { videoId: 'yt-123' });
    const btn = screen.getByRole('button', { name: 'Publish Hugo Post' });
    await userEvent.click(btn);
    await waitFor(() => {
      expect(screen.getByText('/content/devops/test-video.md')).toBeInTheDocument();
    });
  });
});
