export const PHASE_NAMES: Record<number, string> = {
  0: 'Published',
  1: 'Publish Pending',
  2: 'Edit Requested',
  3: 'Material Done',
  4: 'Started',
  5: 'Delayed',
  6: 'Sponsored/Blocked',
  7: 'Ideas',
};

export const PHASE_COLORS: Record<number, string> = {
  0: 'bg-green-500',
  1: 'bg-orange-500',
  2: 'bg-yellow-500',
  3: 'bg-indigo-500',
  4: 'bg-blue-500',
  5: 'bg-red-500',
  6: 'bg-purple-500',
  7: 'bg-gray-500',
};

export const PHASE_ACCENT_COLORS: Record<number, string> = {
  0: 'border-green-500',
  1: 'border-orange-500',
  2: 'border-yellow-500',
  3: 'border-indigo-500',
  4: 'border-blue-500',
  5: 'border-red-500',
  6: 'border-purple-500',
  7: 'border-gray-500',
};

export const ASPECT_LABELS: Record<string, string> = {
  'initial-details': 'Initial Details',
  'work-progress': 'Work Progress',
  'definition': 'Definition',
  'post-production': 'Post Production',
  'publishing': 'Publishing',
  'post-publish': 'Post Publish',
  'analysis': 'Analysis',
};
