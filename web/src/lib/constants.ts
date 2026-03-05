export const PHASE_NAMES: Record<number, string> = {
  0: 'Ideas',
  1: 'Started',
  2: 'Material Done',
  3: 'Edit Requested',
  4: 'Publish Pending',
  5: 'Published',
  6: 'Delayed',
  7: 'Sponsored/Blocked',
};

export const PHASE_COLORS: Record<number, string> = {
  0: 'bg-gray-500',
  1: 'bg-blue-500',
  2: 'bg-indigo-500',
  3: 'bg-yellow-500',
  4: 'bg-orange-500',
  5: 'bg-green-500',
  6: 'bg-red-500',
  7: 'bg-purple-500',
};

export const PHASE_ACCENT_COLORS: Record<number, string> = {
  0: 'border-gray-500',
  1: 'border-blue-500',
  2: 'border-indigo-500',
  3: 'border-yellow-500',
  4: 'border-orange-500',
  5: 'border-green-500',
  6: 'border-red-500',
  7: 'border-purple-500',
};

export const ASPECT_LABELS: Record<string, string> = {
  initialDetails: 'Initial Details',
  workProgress: 'Work Progress',
  definition: 'Definition',
  postProduction: 'Post Production',
  publishing: 'Publishing',
  postPublish: 'Post Publish',
  analysis: 'Analysis',
};
