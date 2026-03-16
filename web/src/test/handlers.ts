import { http, HttpResponse } from 'msw';
import type {
  PhaseInfo,
  VideoListItem,
  VideoResponse,
  OverallProgressResponse,
  AspectsResponse,
} from '../api/types';

export const mockPhases: PhaseInfo[] = [
  { id: 0, name: 'Published', count: 10 },
  { id: 1, name: 'Publish Pending', count: 0 },
  { id: 2, name: 'Edit Requested', count: 4 },
  { id: 3, name: 'Material Done', count: 1 },
  { id: 4, name: 'Started', count: 2 },
  { id: 5, name: 'Delayed', count: 1 },
  { id: 6, name: 'Sponsored', count: 2 },
  { id: 7, name: 'Ideas', count: 3 },
];

export const mockVideoList: VideoListItem[] = [
  {
    id: 'devops/test-video',
    name: 'test-video',
    category: 'devops',
    date: '2026-01-15',
    title: 'Test Video Title',
    phase: 1,
    progress: { completed: 5, total: 20 },
    sponsored: true,
    isFarFuture: false,
  },
  {
    id: 'devops/another-video',
    name: 'another-video',
    category: 'devops',
    title: 'Another Video',
    phase: 4,
    progress: { completed: 10, total: 20 },
    sponsored: false,
    isFarFuture: true,
  },
];

export const mockVideo: VideoResponse = {
  id: 'devops/test-video',
  name: 'test-video',
  path: 'manuscript/devops/test-video.yaml',
  category: 'devops',
  phase: 1,
  init: { completed: 2, total: 4 },
  work: { completed: 1, total: 7 },
  define: { completed: 0, total: 5 },
  edit: { completed: 0, total: 3 },
  publish: { completed: 0, total: 4 },
  postPublish: { completed: 0, total: 6 },
  projectName: 'Test Project',
  projectURL: 'https://example.com',
  date: '2026-01-15',
  delayed: false,
  screen: true,
  head: false,
  thumbnails: false,
  diagrams: false,
  screenshots: false,
  requestThumbnail: false,
  requestEdit: false,
  movie: false,
  animations: '',
  members: '',
  description: '',
  tags: '',
  descriptionTags: '',
  location: '',
  tagline: '',
  taglineIdeas: '',
  otherLogos: '',
  language: '',
  timecodes: '',
  relatedVideos: '',
  hugoPath: '',
  gist: '',
  titles: [],
  thumbnail: '',
  thumbnailVariants: [],
  uploadVideo: '',
  tweet: '',
  appliedLanguage: '',
  appliedAudioLanguage: '',
  audioLanguage: '',
  sponsorship: { amount: '', emails: '', blocked: '', name: '', url: '' },
  notifiedSponsors: false,
  videoId: '',
  linkedInPosted: false,
  slackPosted: false,
  hnPosted: false,
  dotPosted: false,
  blueSkyPosted: false,
  youTubeHighlight: false,
  youTubeComment: false,
  youTubeCommentReply: false,
  slides: false,
  gde: false,
  code: true,
  repo: '',
  shorts: [],
};

export const mockProgress: OverallProgressResponse = {
  overall: { completed: 3, total: 29 },
  aspects: [
    { aspectKey: 'initial-details', title: 'Initial Details', completed: 2, total: 4 },
    { aspectKey: 'work-progress', title: 'Work Progress', completed: 1, total: 7 },
    { aspectKey: 'definition', title: 'Definition', completed: 0, total: 5 },
    { aspectKey: 'post-production', title: 'Post Production', completed: 0, total: 3 },
    { aspectKey: 'publishing', title: 'Publishing', completed: 0, total: 4 },
    { aspectKey: 'post-publish', title: 'Post Publish', completed: 0, total: 6 },
  ],
};

export const mockAspects: AspectsResponse = {
  aspects: [
    {
      key: 'initial-details',
      title: 'Initial Details',
      description: 'Basic video information',
      endpoint: '/api/videos/{videoName}/initial-details',
      icon: 'info',
      order: 1,
      fields: [
        {
          name: 'Project Name',
          fieldName: 'projectName',
          type: 'string',
          required: true,
          order: 1,
          description: 'Name of the project',
          completionCriteria: 'filled_only',
          uiHints: { inputType: 'text', placeholder: 'Enter project name', helpText: 'The project name', multiline: false },
        },
        {
          name: 'Date',
          fieldName: 'date',
          type: 'date',
          required: false,
          order: 2,
          description: 'Scheduled date',
          completionCriteria: 'filled_only',
          uiHints: { inputType: 'date', placeholder: '', helpText: '', multiline: false },
        },
        {
          name: 'Delayed',
          fieldName: 'delayed',
          type: 'boolean',
          required: false,
          order: 3,
          description: 'Whether the video is delayed',
          completionCriteria: 'false_only',
          uiHints: { inputType: 'checkbox', placeholder: '', helpText: '', multiline: false },
        },
      ],
    },
    {
      key: 'work-progress',
      title: 'Work Progress',
      description: 'Track work items',
      endpoint: '/api/videos/{videoName}/work-progress',
      icon: 'hammer',
      order: 2,
      fields: [
        {
          name: 'Screen Recording',
          fieldName: 'screen',
          type: 'boolean',
          required: false,
          order: 1,
          description: 'Screen recording done',
          completionCriteria: 'true_only',
          uiHints: { inputType: 'checkbox', placeholder: '', helpText: '', multiline: false },
        },
        {
          name: 'Description',
          fieldName: 'description',
          type: 'text',
          required: false,
          order: 2,
          description: 'Video description',
          completionCriteria: 'filled_only',
          uiHints: { inputType: 'textarea', placeholder: 'Enter description', helpText: 'Full video description', multiline: true, rows: 5 },
        },
      ],
    },
  ],
};

// AI mock responses
const mockAITitles = { titles: ['AI Title 1', 'AI Title 2', 'AI Title 3'] };
const mockAIDescription = { description: 'AI generated description' };
const mockAITags = { tags: 'ai,generated,tags' };
const mockAITweets = { tweets: ['Tweet 1', 'Tweet 2'] };
const mockAIDescriptionTags = { descriptionTags: '#ai #gen #tags' };
const mockAIShorts = { candidates: [{ id: 'short1', title: 'Short One', text: 'text', rationale: 'good' }] };

// Publishing mock responses
const mockPublishYouTube = { videoId: 'yt-abc123' };
const mockPublishThumbnail = { success: true };
const mockPublishShort = { youtubeId: 'yt-short-456' };
const mockPublishHugo = { hugoPath: '/content/devops/test-video.md' };
const mockTranscript = { transcript: 'This is the transcript text.' };
const mockMetadata = { title: 'Test Title', description: 'Test Desc', tags: ['tag1', 'tag2'], publishedAt: '2026-01-15T00:00:00Z' };
const mockSocialPostAutomated = { posted: true };
const mockSocialPostManual = { posted: false, message: 'Copy this text to post manually.' };

export const handlers = [
  http.get('/api/videos/phases', () => HttpResponse.json(mockPhases)),
  http.get('/api/videos/list', () => HttpResponse.json(mockVideoList)),
  http.get('/api/videos/:videoName/progress', () => HttpResponse.json(mockProgress)),
  http.get('/api/videos/:videoName', () => HttpResponse.json(mockVideo)),
  http.get('/api/aspects', () => HttpResponse.json(mockAspects)),
  http.patch('/api/videos/:videoName', async ({ request }) => {
    const body = await request.json();
    return HttpResponse.json({ ...mockVideo, ...(body as object) });
  }),
  http.post('/api/videos', async ({ request }) => {
    const body = (await request.json()) as { name: string; category: string; date?: string };
    return HttpResponse.json(
      { ...mockVideo, name: body.name, category: body.category, id: `${body.category}/${body.name}` },
      { status: 201 },
    );
  }),
  http.delete('/api/videos/:videoName', () => new HttpResponse(null, { status: 204 })),
  // AI endpoints
  http.post('/api/ai/titles/:category/:name', () => HttpResponse.json(mockAITitles)),
  http.post('/api/ai/description/:category/:name', () => HttpResponse.json(mockAIDescription)),
  http.post('/api/ai/tags/:category/:name', () => HttpResponse.json(mockAITags)),
  http.post('/api/ai/tweets/:category/:name', () => HttpResponse.json(mockAITweets)),
  http.post('/api/ai/description-tags/:category/:name', () => HttpResponse.json(mockAIDescriptionTags)),
  http.post('/api/ai/shorts/:category/:name', () => HttpResponse.json(mockAIShorts)),
  http.post('/api/ai/thumbnails', () => HttpResponse.json({ subtle: 'subtle prompt', bold: 'bold prompt' })),
  http.post('/api/ai/translate', () => HttpResponse.json({ title: 'Titulo', description: 'Desc', tags: 'tags', timecodes: '' })),
  http.post('/api/ai/ama/content', () => HttpResponse.json({ title: 'AMA', timecodes: '00:00', description: 'Desc', tags: 'tags' })),
  http.post('/api/ai/ama/title', () => HttpResponse.json({ title: 'AMA Title' })),
  http.post('/api/ai/ama/description', () => HttpResponse.json({ description: 'AMA Desc' })),
  http.post('/api/ai/ama/timecodes', () => HttpResponse.json({ timecodes: '00:00 Intro' })),
  http.post('/api/drive/upload/thumbnail/:videoName', () =>
    HttpResponse.json({ driveFileId: 'mock-drive-id-123', variantIndex: 0 }),
  ),
  http.post('/api/drive/upload/short/:videoName/:shortId', () =>
    HttpResponse.json({ driveFileId: 'short-drive-id-789', filePath: 'drive://short-drive-id-789' }),
  ),
  // Action endpoints
  http.post('/api/actions/request-thumbnail/:videoName', () =>
    HttpResponse.json({
      alreadyRequested: false,
      emailSent: true,
      video: { ...mockVideo, requestThumbnail: true },
    }),
  ),
  http.post('/api/actions/request-edit/:videoName', () =>
    HttpResponse.json({
      alreadyRequested: false,
      emailSent: true,
      video: { ...mockVideo, requestEdit: true },
    }),
  ),
  // Publishing endpoints
  http.post('/api/publish/youtube/:videoName/thumbnail', () => HttpResponse.json(mockPublishThumbnail)),
  http.post('/api/publish/youtube/:videoName/shorts/:shortId', () => HttpResponse.json(mockPublishShort)),
  http.post('/api/publish/youtube/:videoName', () => HttpResponse.json(mockPublishYouTube)),
  http.post('/api/publish/hugo/:videoName', () => HttpResponse.json(mockPublishHugo)),
  http.get('/api/publish/transcript/:videoId', () => HttpResponse.json(mockTranscript)),
  http.get('/api/publish/metadata/:videoId', () => HttpResponse.json(mockMetadata)),
  // Social endpoints
  http.post('/api/social/bluesky/:videoName', () => HttpResponse.json(mockSocialPostAutomated)),
  http.post('/api/social/slack/:videoName', () => HttpResponse.json(mockSocialPostAutomated)),
  http.post('/api/social/linkedin/:videoName', () => HttpResponse.json(mockSocialPostManual)),
  http.post('/api/social/hackernews/:videoName', () => HttpResponse.json(mockSocialPostManual)),
  http.post('/api/social/dot/:videoName', () => HttpResponse.json(mockSocialPostManual)),
  // Analyze endpoints
  http.post('/api/analyze/titles', () =>
    HttpResponse.json({
      videoCount: 5,
      highPerformingPatterns: [
        { pattern: 'Provocative', description: 'Challenge assumptions', impact: 'high', examples: ['Stop Using X!'] },
      ],
      lowPerformingPatterns: [
        { pattern: 'Listicle', description: 'Generic lists', impact: 'low', examples: ['Top 10 Tools'] },
      ],
      recommendations: [
        { recommendation: 'Use provocative titles', evidence: '55% A/B share', example: 'Why X Is Dead' },
      ],
      titlesMdContent: '# Title Patterns\n\n1. Provocative opinions work best',
    }),
  ),
  http.post('/api/analyze/titles/apply', () =>
    HttpResponse.json({ applied: true }),
  ),
  // AMA endpoints
  http.post('/api/ama/generate', () => HttpResponse.json({
    title: 'Generated AMA Title',
    description: 'Generated AMA Description',
    tags: 'ama,generated,tags',
    timecodes: '00:00 Intro\n01:00 First Question',
    transcript: 'Hello welcome to the AMA',
  })),
  http.post('/api/ama/apply', () => HttpResponse.json({ success: true })),
  // Animations endpoint
  http.get('/api/videos/:videoName/animations', () =>
    HttpResponse.json({
      animations: ['Add fade transition', 'Section: Main Demo', 'Show terminal output'],
      sections: ['Section: Main Demo'],
    }),
  ),
  // Random timing endpoint
  http.post('/api/videos/:videoName/apply-random-timing', () =>
    HttpResponse.json({
      newDate: '2026-01-14T14:30',
      originalDate: '2026-01-15',
      day: 'Wednesday',
      time: '14:30',
      reasoning: 'Mid-week afternoon uploads show 20% higher initial engagement',
    }),
  ),
];
