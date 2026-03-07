import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect } from 'vitest';
import { SocialPostButton, isSocialField } from '../components/forms/SocialPostButton';

function renderSocialButton(fieldName: string, value: boolean) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return render(
    <QueryClientProvider client={qc}>
      <SocialPostButton
        fieldName={fieldName}
        value={value}
        category="devops"
        videoName="test-video"
      />
    </QueryClientProvider>,
  );
}

describe('isSocialField', () => {
  it('returns true for social fields', () => {
    expect(isSocialField('blueSkyPosted')).toBe(true);
    expect(isSocialField('slackPosted')).toBe(true);
    expect(isSocialField('linkedInPosted')).toBe(true);
    expect(isSocialField('hnPosted')).toBe(true);
    expect(isSocialField('dotPosted')).toBe(true);
  });

  it('returns false for non-social fields', () => {
    expect(isSocialField('delayed')).toBe(false);
    expect(isSocialField('code')).toBe(false);
  });
});

describe('SocialPostButton', () => {
  it('shows done state when value is true for BlueSky', () => {
    renderSocialButton('blueSkyPosted', true);
    expect(screen.getByText('Posted to BlueSky')).toBeInTheDocument();
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it('shows done state when value is true for Slack', () => {
    renderSocialButton('slackPosted', true);
    expect(screen.getByText('Posted to Slack')).toBeInTheDocument();
  });

  it('shows done state when value is true for LinkedIn', () => {
    renderSocialButton('linkedInPosted', true);
    expect(screen.getByText('Posted to LinkedIn')).toBeInTheDocument();
  });

  it('renders post button when value is false', () => {
    renderSocialButton('blueSkyPosted', false);
    expect(screen.getByRole('button', { name: 'Post to BlueSky' })).toBeInTheDocument();
  });

  it('renders post button for LinkedIn when not posted', () => {
    renderSocialButton('linkedInPosted', false);
    expect(screen.getByRole('button', { name: 'Post to LinkedIn' })).toBeInTheDocument();
  });

  it('calls automated post endpoint for BlueSky', async () => {
    renderSocialButton('blueSkyPosted', false);
    const btn = screen.getByRole('button', { name: 'Post to BlueSky' });
    await userEvent.click(btn);
    // Automated post - no copy dialog should appear
    await waitFor(() => {
      expect(screen.queryByText('Copy to Clipboard')).not.toBeInTheDocument();
    });
  });

  it('shows copy dialog for manual platforms like LinkedIn', async () => {
    renderSocialButton('linkedInPosted', false);
    const btn = screen.getByRole('button', { name: 'Post to LinkedIn' });
    await userEvent.click(btn);
    await waitFor(() => {
      expect(screen.getByText('Copy to Clipboard')).toBeInTheDocument();
      expect(screen.getByText('Copy this text to post manually.')).toBeInTheDocument();
    });
  });

  it('closes copy dialog when Close button is clicked', async () => {
    renderSocialButton('linkedInPosted', false);
    await userEvent.click(screen.getByRole('button', { name: 'Post to LinkedIn' }));
    await waitFor(() => {
      expect(screen.getByText('Copy to Clipboard')).toBeInTheDocument();
    });
    const closeButtons = screen.getAllByRole('button', { name: /Close/ });
    // Click the text "Close" button (not the × icon)
    await userEvent.click(closeButtons[closeButtons.length - 1]);
    await waitFor(() => {
      expect(screen.queryByText('Copy to Clipboard')).not.toBeInTheDocument();
    });
  });
});
