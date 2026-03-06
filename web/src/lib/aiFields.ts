export const AI_FIELD_CONFIG: Record<string, {
  hook: 'titles' | 'description' | 'tags' | 'descriptionTags' | 'tweets' | 'shorts';
  label: string;
}> = {
  titles: { hook: 'titles', label: 'Generate Titles' },
  description: { hook: 'description', label: 'Generate Description' },
  tags: { hook: 'tags', label: 'Generate Tags' },
  descriptionTags: { hook: 'descriptionTags', label: 'Generate Tags' },
  tweet: { hook: 'tweets', label: 'Generate Tweets' },
  shorts: { hook: 'shorts', label: 'Generate Shorts' },
};
