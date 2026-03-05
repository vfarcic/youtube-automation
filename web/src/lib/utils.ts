import type { ProgressInfo } from '../api/types';

export function progressPercent(p: ProgressInfo): number {
  if (p.total === 0) return 0;
  return Math.round((p.completed / p.total) * 100);
}

export function formatDate(dateStr: string): string {
  if (!dateStr) return '';
  try {
    return new Date(dateStr).toLocaleDateString();
  } catch {
    return dateStr;
  }
}

export function parseVideoId(id: string): { category: string; name: string } {
  const slashIdx = id.indexOf('/');
  if (slashIdx === -1) return { category: '', name: id };
  return { category: id.slice(0, slashIdx), name: id.slice(slashIdx + 1) };
}
