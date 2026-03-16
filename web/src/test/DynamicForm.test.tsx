import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi } from 'vitest';
import { DynamicForm } from '../components/forms/DynamicForm';
import type { AspectField } from '../api/types';
import { mockVideo } from './handlers';

function createWrapper() {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={qc}>{children}</QueryClientProvider>
  );
}

const sampleFields: AspectField[] = [
  {
    name: 'Project Name',
    fieldName: 'projectName',
    type: 'string',
    required: true,
    order: 1,
    description: '',
    completionCriteria: 'filled_only',
    uiHints: { inputType: 'text', placeholder: 'Enter name', helpText: '', multiline: false },
  },
  {
    name: 'Screen Recording',
    fieldName: 'screen',
    type: 'boolean',
    required: false,
    order: 2,
    description: '',
    completionCriteria: 'true_only',
    uiHints: { inputType: 'checkbox', placeholder: '', helpText: '', multiline: false },
  },
  {
    name: 'Description',
    fieldName: 'description',
    type: 'text',
    required: false,
    order: 3,
    description: '',
    completionCriteria: 'filled_only',
    uiHints: { inputType: 'textarea', placeholder: '', helpText: '', multiline: true, rows: 4 },
  },
];

const dateField: AspectField = {
  name: 'Date',
  fieldName: 'date',
  type: 'date',
  required: false,
  order: 4,
  description: 'Scheduled date',
  completionCriteria: 'filled_only',
  uiHints: { inputType: 'date', placeholder: '', helpText: '', multiline: false },
};

const arrayField: AspectField = {
  name: 'Titles',
  fieldName: 'titles',
  type: 'array',
  required: false,
  order: 4,
  description: '',
  completionCriteria: 'filled_only',
  itemFields: [
    { name: 'Index', fieldName: 'index', type: 'number', order: 1 },
    { name: 'Text', fieldName: 'text', type: 'string', order: 2 },
    { name: 'Share', fieldName: 'share', type: 'number', order: 3 },
  ],
};

const mapField: AspectField = {
  name: 'Metadata',
  fieldName: 'metadata',
  type: 'map',
  required: false,
  order: 5,
  description: '',
  completionCriteria: 'filled_only',
  mapKeyLabel: 'Key',
  itemFields: [
    { name: 'ID', fieldName: 'id', type: 'string', order: 1 },
    { name: 'Value', fieldName: 'value', type: 'string', order: 2 },
  ],
};

const dotNotationFields: AspectField[] = [
  {
    name: 'Sponsorship Amount',
    fieldName: 'sponsorship.amount',
    type: 'string',
    required: false,
    order: 1,
    description: '',
    completionCriteria: 'filled_only',
    uiHints: { inputType: 'text', placeholder: '', helpText: '', multiline: false },
  },
];

describe('DynamicForm', () => {
  it('renders fields from aspect metadata', () => {
    render(<DynamicForm fields={sampleFields} video={mockVideo} onSave={() => {}} />, { wrapper: createWrapper() });
    expect(screen.getByText('Project Name')).toBeInTheDocument();
    expect(screen.getByText('Screen Recording')).toBeInTheDocument();
    expect(screen.getByText('Description')).toBeInTheDocument();
  });

  it('shows save and reset buttons', () => {
    render(<DynamicForm fields={sampleFields} video={mockVideo} onSave={() => {}} />, { wrapper: createWrapper() });
    expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Reset' })).toBeDisabled();
  });

  it('enables save when form is dirty', async () => {
    render(<DynamicForm fields={sampleFields} video={mockVideo} onSave={() => {}} />, { wrapper: createWrapper() });
    const input = screen.getByDisplayValue('Test Project');
    await userEvent.clear(input);
    await userEvent.type(input, 'New Name');
    expect(screen.getByRole('button', { name: 'Save' })).toBeEnabled();
  });

  it('sends only changed fields on save', async () => {
    const onSave = vi.fn();
    render(<DynamicForm fields={sampleFields} video={mockVideo} onSave={onSave} />, { wrapper: createWrapper() });
    const input = screen.getByDisplayValue('Test Project');
    await userEvent.clear(input);
    await userEvent.type(input, 'Changed');
    await userEvent.click(screen.getByRole('button', { name: 'Save' }));
    expect(onSave).toHaveBeenCalledWith({ projectName: 'Changed' });
  });

  it('resets to initial values', async () => {
    render(<DynamicForm fields={sampleFields} video={mockVideo} onSave={() => {}} />, { wrapper: createWrapper() });
    const input = screen.getByDisplayValue('Test Project');
    await userEvent.clear(input);
    await userEvent.type(input, 'Changed');
    await userEvent.click(screen.getByRole('button', { name: 'Reset' }));
    expect(screen.getByDisplayValue('Test Project')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled();
  });

  it('renders array fields with ArrayInput component', () => {
    const video = { ...mockVideo, titles: [{ index: 1, text: 'My Title', watchTimeShare: 0.5 }] };
    render(<DynamicForm fields={[arrayField]} video={video} onSave={() => {}} />, { wrapper: createWrapper() });
    expect(screen.getByText('Titles')).toBeInTheDocument();
    expect(screen.getByText('Item 1')).toBeInTheDocument();
    expect(screen.getByDisplayValue('My Title')).toBeInTheDocument();
  });

  it('renders map fields with MapInput component', () => {
    const video = { ...mockVideo, metadata: { key1: { id: 'id-1', value: 'val-1' } } } as any;
    render(<DynamicForm fields={[mapField]} video={video} onSave={() => {}} />, { wrapper: createWrapper() });
    expect(screen.getByText('Metadata')).toBeInTheDocument();
    expect(screen.getByText(/Key: key1/)).toBeInTheDocument();
  });

  it('resolves dot-notation fields', () => {
    const video = { ...mockVideo, sponsorship: { ...mockVideo.sponsorship, amount: '500' } };
    render(<DynamicForm fields={dotNotationFields} video={video} onSave={() => {}} />, { wrapper: createWrapper() });
    expect(screen.getByDisplayValue('500')).toBeInTheDocument();
  });

  it('syncs form state when video prop changes for untouched fields', async () => {
    const onSave = vi.fn();
    const video1 = { ...mockVideo, projectName: 'Original' };
    const wrapper = createWrapper();
    const { rerender } = render(
      <DynamicForm fields={sampleFields} video={video1} onSave={onSave} />,
      { wrapper },
    );
    expect(screen.getByDisplayValue('Original')).toBeInTheDocument();

    // Simulate server-side update (e.g. thumbnail upload changed the video)
    const video2 = { ...mockVideo, projectName: 'Updated By Server' };
    rerender(<DynamicForm fields={sampleFields} video={video2} onSave={onSave} />);

    // The untouched field should sync to the new server value
    expect(screen.getByDisplayValue('Updated By Server')).toBeInTheDocument();
    // Save button should NOT be enabled since the form matches the new server state
    expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled();
  });

  it('preserves user edits when video prop changes', async () => {
    const onSave = vi.fn();
    const video1 = { ...mockVideo, projectName: 'Original', description: 'Desc' };
    const descField: AspectField = {
      ...sampleFields[2],
      fieldName: 'description',
    };
    const fields = [sampleFields[0], descField];
    const wrapper = createWrapper();
    const { rerender } = render(
      <DynamicForm fields={fields} video={video1} onSave={onSave} />,
      { wrapper },
    );

    // User edits projectName
    const input = screen.getByDisplayValue('Original');
    await userEvent.clear(input);
    await userEvent.type(input, 'User Edit');

    // Server updates description (simulating thumbnail upload changing another field)
    const video2 = { ...mockVideo, projectName: 'Original', description: 'Server Updated Desc' };
    rerender(<DynamicForm fields={fields} video={video2} onSave={onSave} />);

    // User edit should be preserved
    expect(screen.getByDisplayValue('User Edit')).toBeInTheDocument();
    // Server update to untouched field should be synced
    expect(screen.getByDisplayValue('Server Updated Desc')).toBeInTheDocument();
  });

  it('renders RandomTimingButton when category and videoName provided with date field', () => {
    render(
      <DynamicForm
        fields={[dateField]}
        video={mockVideo}
        onSave={() => {}}
        category="devops"
        videoName="test-video"
      />,
      { wrapper: createWrapper() },
    );
    expect(screen.getByRole('button', { name: 'Apply Random Timing' })).toBeInTheDocument();
  });

  it('does not render RandomTimingButton when category or videoName missing', () => {
    render(
      <DynamicForm fields={[dateField]} video={mockVideo} onSave={() => {}} />,
      { wrapper: createWrapper() },
    );
    expect(screen.queryByRole('button', { name: 'Apply Random Timing' })).not.toBeInTheDocument();
  });

  it('renders Shorts Upload section when aspectKey is post-production', () => {
    const videoWithShorts = {
      ...mockVideo,
      shorts: [
        { id: 'short1', title: 'My Short', text: 'text', filePath: '', scheduledDate: '2026-01-15', youtubeId: '', driveFileId: '' },
      ],
    };
    render(
      <DynamicForm
        fields={sampleFields}
        video={videoWithShorts}
        onSave={() => {}}
        category="devops"
        videoName="test-video"
        aspectKey="post-production"
      />,
      { wrapper: createWrapper() },
    );
    expect(screen.getByText('Shorts Upload')).toBeInTheDocument();
    expect(screen.getByText('Upload to Drive')).toBeInTheDocument();
    expect(screen.queryByText('Publish to YouTube')).not.toBeInTheDocument();
  });

  it('renders Shorts Publish section when aspectKey is publishing', () => {
    const videoWithShorts = {
      ...mockVideo,
      shorts: [
        { id: 'short1', title: 'My Short', text: 'text', filePath: '', scheduledDate: '2026-01-15', youtubeId: '', driveFileId: 'some-id' },
      ],
    };
    render(
      <DynamicForm
        fields={sampleFields}
        video={videoWithShorts}
        onSave={() => {}}
        category="devops"
        videoName="test-video"
        aspectKey="publishing"
      />,
      { wrapper: createWrapper() },
    );
    expect(screen.getByText('Shorts Publish')).toBeInTheDocument();
    expect(screen.getByText('Publish to YouTube')).toBeInTheDocument();
    expect(screen.queryByText('Upload to Drive')).not.toBeInTheDocument();
  });

  it('does not render shorts sections when aspectKey is definition', () => {
    const videoWithShorts = {
      ...mockVideo,
      shorts: [
        { id: 'short1', title: 'My Short', text: 'text', filePath: '', scheduledDate: '2026-01-15', youtubeId: '', driveFileId: '' },
      ],
    };
    render(
      <DynamicForm
        fields={sampleFields}
        video={videoWithShorts}
        onSave={() => {}}
        category="devops"
        videoName="test-video"
        aspectKey="definition"
      />,
      { wrapper: createWrapper() },
    );
    expect(screen.queryByText('Shorts Upload')).not.toBeInTheDocument();
    expect(screen.queryByText('Shorts Publish')).not.toBeInTheDocument();
  });

  it('does not render shorts sections when video has no shorts', () => {
    render(
      <DynamicForm
        fields={sampleFields}
        video={mockVideo}
        onSave={() => {}}
        category="devops"
        videoName="test-video"
        aspectKey="post-production"
      />,
      { wrapper: createWrapper() },
    );
    expect(screen.queryByText('Shorts Upload')).not.toBeInTheDocument();
  });
});
