import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi } from 'vitest';
import {
  PinnedCommentToggle,
  isVideoSponsored,
  buildSponsorCommentText,
} from '../components/forms/PinnedCommentToggle';
import type { VideoResponse } from '../api/types';

function makeVideo(sponsorship: Partial<VideoResponse['sponsorship']>): VideoResponse {
  return {
    sponsorship: { amount: '', emails: '', blocked: '', name: '', url: '', adFile: '', ...sponsorship },
  } as VideoResponse;
}

describe('isVideoSponsored', () => {
  it('is true when amount is a real value', () => {
    expect(isVideoSponsored(makeVideo({ amount: '1000' }))).toBe(true);
  });

  it('is false for empty, N/A, or dash amounts', () => {
    expect(isVideoSponsored(makeVideo({ amount: '' }))).toBe(false);
    expect(isVideoSponsored(makeVideo({ amount: 'N/A' }))).toBe(false);
    expect(isVideoSponsored(makeVideo({ amount: '-' }))).toBe(false);
  });
});

describe('buildSponsorCommentText', () => {
  it('uses sponsor name and url', () => {
    const text = buildSponsorCommentText(makeVideo({ name: 'Acme Corp', url: 'https://acme.example' }));
    expect(text).toBe('This video is sponsored by Acme Corp. Please visit https://acme.example.');
  });

  it('falls back to placeholders when name/url are missing', () => {
    const text = buildSponsorCommentText(makeVideo({ amount: '1000' }));
    expect(text).toBe('This video is sponsored by [SPONSOR NAME]. Please visit [SPONSOR URL].');
  });
});

describe('PinnedCommentToggle', () => {
  it('renders the exact copy/paste comment text', () => {
    render(
      <PinnedCommentToggle
        name="YouTube Pinned Comment Added (manual)"
        fieldName="youTubeComment"
        value={false}
        onChange={() => {}}
        video={makeVideo({ amount: '1000', name: 'Acme Corp', url: 'https://acme.example' })}
      />,
    );
    expect(
      screen.getByText('This video is sponsored by Acme Corp. Please visit https://acme.example.'),
    ).toBeInTheDocument();
  });

  it('toggles the value on click', async () => {
    const onChange = vi.fn();
    render(
      <PinnedCommentToggle
        name="YouTube Pinned Comment Added (manual)"
        fieldName="youTubeComment"
        value={false}
        onChange={onChange}
        video={makeVideo({ amount: '1000', name: 'Acme Corp', url: 'https://acme.example' })}
      />,
    );
    await userEvent.click(screen.getByRole('switch'));
    expect(onChange).toHaveBeenCalledWith('youTubeComment', true);
  });
});
