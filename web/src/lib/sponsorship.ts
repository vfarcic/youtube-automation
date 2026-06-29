import type { VideoResponse } from '../api/types';

/**
 * Returns true when the video has real sponsorship data.
 *
 * Mirrors the backend `videoHasSponsorship` check (raw string, no trim) so field
 * visibility and progress counting stay in lockstep: a non-empty amount that
 * isn't the "N/A" / "-" placeholder.
 */
export function isVideoSponsored(video: VideoResponse): boolean {
  const amount = video.sponsorship?.amount ?? '';
  return amount !== '' && amount !== 'N/A' && amount !== '-';
}

/**
 * Field names whose presence is conditional on the video being sponsored. These
 * are hidden when not sponsored and, on the backend, excluded from the progress
 * count entirely (see CompletionService.isFieldApplicable).
 */
export const SPONSORSHIP_CONDITIONAL_FIELDS = new Set<string>([
  'youTubeComment',
  'notifiedSponsors',
  'sponsorship.emails',
]);
