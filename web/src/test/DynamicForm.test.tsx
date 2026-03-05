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

  it('resolves dot-notation fields', () => {
    const video = { ...mockVideo, sponsorship: { ...mockVideo.sponsorship, amount: '500' } };
    render(<DynamicForm fields={dotNotationFields} video={video} onSave={() => {}} />);
    expect(screen.getByDisplayValue('500')).toBeInTheDocument();
  });
});
