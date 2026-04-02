export interface PhaseInfo {
  id: number;
  name: string;
  count: number;
}

export interface ProgressInfo {
  completed: number;
  total: number;
}

export interface VideoListItem {
  id: string;
  name: string;
  category: string;
  date?: string;
  title?: string;
  phase: number;
  progress: ProgressInfo;
  sponsored?: boolean;
  isFarFuture?: boolean;
}

export interface Sponsorship {
  amount: string;
  emails: string;
  blocked: string;
  name: string;
  url: string;
  adFile: string;
}

export interface TitleVariant {
  index: number;
  text: string;
  watchTimeShare: number;
}

export interface ThumbnailVariant {
  index: number;
  path: string;
  driveFileId?: string;
  clickShare: number;
}

export interface Short {
  id: string;
  title: string;
  text: string;
  filePath: string;
  driveFileId?: string;
  scheduledDate: string;
  youtubeId: string;
}

export interface VideoResponse {
  id: string;
  name: string;
  path: string;
  category: string;
  phase: number;
  init: ProgressInfo;
  work: ProgressInfo;
  define: ProgressInfo;
  edit: ProgressInfo;
  publish: ProgressInfo;
  postPublish: ProgressInfo;
  // Video fields
  projectName: string;
  projectURL: string;
  date: string;
  delayed: boolean;
  screen: boolean;
  head: boolean;
  thumbnails: boolean;
  diagrams: boolean;
  screenshots: boolean;
  requestThumbnail: boolean;
  requestEdit: boolean;
  movie: boolean;
  animations: string;
  members: string;
  description: string;
  tags: string;
  descriptionTags: string;
  location: string;
  tagline: string;
  taglineIdeas: string;
  otherLogos: string;
  language: string;
  timecodes: string;
  relatedVideos: string;
  hugoPath: string;
  gist: string;
  titles: TitleVariant[];
  thumbnail: string;
  thumbnailVariants: ThumbnailVariant[];
  uploadVideo: string;
  videoFile?: string;
  videoDriveFileId?: string;
  tweet: string;
  appliedLanguage: string;
  appliedAudioLanguage: string;
  audioLanguage: string;
  sponsorship: Sponsorship;
  notifiedSponsors: boolean;
  videoId: string;
  linkedInPosted: boolean;
  slackPosted: boolean;
  hnPosted: boolean;
  dotPosted: boolean;
  blueSkyPosted: boolean;
  youTubeHighlight: boolean;
  youTubeComment: boolean;
  youTubeCommentReply: boolean;
  slides: boolean;
  gde: boolean;
  code: boolean;
  repo: string;
  shorts: Short[];
  syncWarning?: string;
}

export interface AspectProgressInfo {
  aspectKey: string;
  title: string;
  completed: number;
  total: number;
}

export interface OverallProgressResponse {
  overall: ProgressInfo;
  aspects: AspectProgressInfo[];
}

export interface SelectOption {
  label: string;
  value: unknown;
}

export interface AspectFieldUIHints {
  inputType: string;
  placeholder: string;
  helpText: string;
  rows?: number;
  charLimit?: number;
  multiline: boolean;
  options?: SelectOption[];
  attributes?: Record<string, unknown>;
}

export interface AspectFieldValidationHints {
  required: boolean;
  minLength?: number;
  maxLength?: number;
  pattern?: string;
  patternDesc?: string;
  min?: number;
  max?: number;
}

export interface FieldOptions {
  values?: string[];
}

export interface ItemField {
  name: string;
  fieldName: string;
  type: string;
  required?: boolean;
  order: number;
  description?: string;
}

export interface AspectField {
  name: string;
  fieldName: string;
  type: string;
  required: boolean;
  order: number;
  description: string;
  options?: FieldOptions;
  uiHints?: AspectFieldUIHints;
  validationHints?: AspectFieldValidationHints;
  defaultValue?: unknown;
  completionCriteria: string;
  itemFields?: ItemField[];
  mapKeyLabel?: string;
}

export interface AspectMetadata {
  key: string;
  title: string;
  description: string;
  endpoint: string;
  icon: string;
  order: number;
  fields: AspectField[];
}

export interface AspectsResponse {
  aspects: AspectMetadata[];
}

export interface CreateVideoRequest {
  name: string;
  category: string;
  date?: string;
}

export interface Category {
  name: string;
  path: string;
}

// --- AI Response Types ---

export interface AITitlesResponse {
  titles: string[];
}

export interface AIDescriptionResponse {
  description: string;
}

export interface AITagsResponse {
  tags: string;
}

export interface AITweetsResponse {
  tweets: string[];
}

export interface AIDescriptionTagsResponse {
  descriptionTags: string;
}

export interface ShortCandidate {
  id: string;
  title: string;
  text: string;
  rationale: string;
}

export interface AIShortsResponse {
  candidates: ShortCandidate[];
}

export interface AIThumbnailsResponse {
  subtle: string;
  bold: string;
}

export interface AITranslateResponse {
  title: string;
  description: string;
  tags: string;
  timecodes: string;
  shortTitles?: string[];
}

export interface AIAMAContentResponse {
  title: string;
  timecodes: string;
  description: string;
  tags: string;
}

// --- Action Button Types ---

export interface ActionResponse {
  alreadyRequested: boolean;
  emailSent: boolean;
  emailError?: string;
  video: VideoResponse;
  syncWarning?: string;
}

// --- Publishing Response Types ---

export interface PublishYouTubeResponse {
  videoId: string;
  syncWarning?: string;
  thumbnailWarning?: string;
}

export interface PublishThumbnailResponse {
  success: boolean;
  syncWarning?: string;
}

export interface PublishShortResponse {
  youtubeId: string;
  syncWarning?: string;
}

export interface PublishHugoResponse {
  hugoPath: string;
  syncWarning?: string;
}

export interface TranscriptResponse {
  transcript: string;
}

export interface MetadataResponse {
  title: string;
  description: string;
  tags: string[];
  publishedAt: string;
}

export interface SocialPostResponse {
  posted: boolean;
  message?: string;
  postUrl?: string;
  syncWarning?: string;
}

// --- Analyze Types ---

export interface TitlePattern {
  pattern: string;
  description: string;
  impact: string;
  examples: string[];
}

export interface TitleRecommendation {
  recommendation: string;
  evidence: string;
  example: string;
}

export interface AnalyzeTitlesResponse {
  videoCount: number;
  highPerformingPatterns: TitlePattern[];
  lowPerformingPatterns: TitlePattern[];
  recommendations: TitleRecommendation[];
  titlesMdContent: string;
}

export interface ApplyTitlesResponse {
  applied: boolean;
  syncWarning?: string;
}

// --- AMA Types ---

export interface AMAGenerateResponse {
  title: string;
  description: string;
  tags: string;
  timecodes: string;
  transcript: string;
}

export interface AMAApplyResponse {
  success: boolean;
}

// --- Animations Types ---

export interface AnimationsResponse {
  animations: string[];
  sections: string[];
}

// --- Random Timing Types ---

export interface ApplyRandomTimingResponse {
  newDate: string;
  originalDate: string;
  day: string;
  time: string;
  reasoning: string;
  syncWarning?: string;
}

// --- Timing Recommendation Types ---

export interface TimingRecommendation {
  day: string;
  time: string;
  reasoning: string;
}

export interface GetTimingResponse {
  recommendations: TimingRecommendation[];
}

export interface PutTimingResponse {
  saved: boolean;
  syncWarning?: string;
}

export interface GenerateTimingResponse {
  recommendations: TimingRecommendation[];
  videoCount: number;
}
