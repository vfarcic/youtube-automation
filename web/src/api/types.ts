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
}

export interface Sponsorship {
  amount: string;
  emails: string;
  blocked: string;
  name: string;
  url: string;
}

export interface TitleVariant {
  index: number;
  text: string;
  watchTimeShare: number;
}

export interface ThumbnailVariant {
  index: number;
  path: string;
  clickShare: number;
}

export interface Short {
  id: string;
  title: string;
  text: string;
  filePath: string;
  scheduledDate: string;
  youtubeId: string;
}

export interface DubbingInfo {
  dubbingId: string;
  dubbedVideoPath: string;
  title: string;
  description: string;
  tags: string;
  timecodes: string;
  uploadedVideoId: string;
  dubbingStatus: string;
  dubbingError: string;
  thumbnailPath: string;
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
  dubbing: Record<string, DubbingInfo>;
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
