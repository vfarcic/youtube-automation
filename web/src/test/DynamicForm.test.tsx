import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi } from 'vitest';
import { DynamicForm } from '../components/forms/DynamicForm';
import type { AspectField } from '../api/types';
import { mockVideo } from './handlers';

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
  name: 'Dubbing',
  fieldName: 'dubbing',
  type: 'map',
  required: false,
  order: 5,
  description: '',
  completionCriteria: 'filled_only',
  mapKeyLabel: 'Language Code',
  itemFields: [
    { name: 'Dubbing ID', fieldName: 'dubbingId', type: 'string', order: 1 },
    { name: 'Title', fieldName: 'title', type: 'string', order: 2 },
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
    render(<DynamicForm fields={sampleFields} video={mockVideo} onSave={() => {}} />);
    expect(screen.getByText('Project Name')).toBeInTheDocument();
    expect(screen.getByText('Screen Recording')).toBeInTheDocument();
    expect(screen.getByText('Description')).toBeInTheDocument();
  });

  it('shows save and reset buttons', () => {
    render(<DynamicForm fields={sampleFields} video={mockVideo} onSave={() => {}} />);
    expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Reset' })).toBeDisabled();
  });

  it('enables save when form is dirty', async () => {
    render(<DynamicForm fields={sampleFields} video={mockVideo} onSave={() => {}} />);
    const input = screen.getByDisplayValue('Test Project');
    await userEvent.clear(input);
    await userEvent.type(input, 'New Name');
    expect(screen.getByRole('button', { name: 'Save' })).toBeEnabled();
  });

  it('sends only changed fields on save', async () => {
    const onSave = vi.fn();
    render(<DynamicForm fields={sampleFields} video={mockVideo} onSave={onSave} />);
    const input = screen.getByDisplayValue('Test Project');
    await userEvent.clear(input);
    await userEvent.type(input, 'Changed');
    await userEvent.click(screen.getByRole('button', { name: 'Save' }));
    expect(onSave).toHaveBeenCalledWith({ projectName: 'Changed' });
  });

  it('resets to initial values', async () => {
    render(<DynamicForm fields={sampleFields} video={mockVideo} onSave={() => {}} />);
    const input = screen.getByDisplayValue('Test Project');
    await userEvent.clear(input);
    await userEvent.type(input, 'Changed');
    await userEvent.click(screen.getByRole('button', { name: 'Reset' }));
    expect(screen.getByDisplayValue('Test Project')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled();
  });

  it('renders array fields with ArrayInput component', () => {
    const video = { ...mockVideo, titles: [{ index: 1, text: 'My Title', watchTimeShare: 0.5 }] };
    render(<DynamicForm fields={[arrayField]} video={video} onSave={() => {}} />);
    expect(screen.getByText('Titles')).toBeInTheDocument();
    expect(screen.getByText('Item 1')).toBeInTheDocument();
    expect(screen.getByDisplayValue('My Title')).toBeInTheDocument();
  });

  it('renders map fields with MapInput component', () => {
    const video = { ...mockVideo, dubbing: { es: { dubbingId: 'dub-1', title: 'Título', description: '', tags: '', timecodes: '', dubbedVideoPath: '', uploadedVideoId: '', dubbingStatus: '', dubbingError: '', thumbnailPath: '' } } };
    render(<DynamicForm fields={[mapField]} video={video} onSave={() => {}} />);
    expect(screen.getByText('Dubbing')).toBeInTheDocument();
    expect(screen.getByText(/Language Code: es/)).toBeInTheDocument();
  });

  it('resolves dot-notation fields', () => {
    const video = { ...mockVideo, sponsorship: { ...mockVideo.sponsorship, amount: '500' } };
    render(<DynamicForm fields={dotNotationFields} video={video} onSave={() => {}} />);
    expect(screen.getByDisplayValue('500')).toBeInTheDocument();
  });
});
