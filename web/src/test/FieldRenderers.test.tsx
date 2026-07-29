import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi } from 'vitest';
import { TextInput } from '../components/forms/TextInput';
import { TextArea } from '../components/forms/TextArea';
import { Toggle } from '../components/forms/Toggle';
import { DateInput } from '../components/forms/DateInput';
import { NumberInput } from '../components/forms/NumberInput';
import { SelectInput } from '../components/forms/SelectInput';

describe('TextInput', () => {
  it('renders with value and placeholder', () => {
    render(
      <TextInput name="Title" fieldName="title" value="hello" onChange={() => {}} placeholder="Enter title" />,
    );
    const input = screen.getByRole('textbox');
    expect(input).toHaveValue('hello');
    expect(input).toHaveAttribute('placeholder', 'Enter title');
    expect(screen.getByText('Title')).toBeInTheDocument();
  });

  it('calls onChange with field name', async () => {
    const onChange = vi.fn();
    render(<TextInput name="Title" fieldName="title" value="" onChange={onChange} />);
    await userEvent.type(screen.getByRole('textbox'), 'a');
    expect(onChange).toHaveBeenCalledWith('title', 'a');
  });

  it('shows required indicator', () => {
    render(<TextInput name="Title" fieldName="title" value="" onChange={() => {}} required />);
    expect(screen.getByText('*')).toBeInTheDocument();
  });
});

describe('TextArea', () => {
  it('renders with rows from uiHints', () => {
    render(
      <TextArea name="Desc" fieldName="desc" value="content" onChange={() => {}} rows={8} />,
    );
    const textarea = screen.getByRole('textbox');
    expect(textarea).toHaveValue('content');
    expect(textarea).toHaveAttribute('rows', '8');
  });
});

describe('Toggle', () => {
  it('renders as switch with correct state', () => {
    render(<Toggle name="Active" fieldName="active" value={true} onChange={() => {}} />);
    const sw = screen.getByRole('switch');
    expect(sw).toHaveAttribute('aria-checked', 'true');
  });

  it('calls onChange with toggled value', async () => {
    const onChange = vi.fn();
    render(<Toggle name="Active" fieldName="active" value={false} onChange={onChange} />);
    await userEvent.click(screen.getByRole('switch'));
    expect(onChange).toHaveBeenCalledWith('active', true);
  });
});

describe('DateInput', () => {
  it('renders date input', () => {
    render(<DateInput name="Date" fieldName="date" value="2026-01-15T10:00" onChange={() => {}} />);
    expect(screen.getByText('Date')).toBeInTheDocument();
  });
});

describe('NumberInput', () => {
  it('renders with min and max', () => {
    render(
      <NumberInput name="Count" fieldName="count" value={5} onChange={() => {}} min={0} max={100} />,
    );
    const input = screen.getByRole('spinbutton');
    expect(input).toHaveValue(5);
    expect(input).toHaveAttribute('min', '0');
    expect(input).toHaveAttribute('max', '100');
  });
});

describe('SelectInput', () => {
  it('renders options', () => {
    const options = [
      { label: 'Public', value: 'public' },
      { label: 'Private', value: 'private' },
    ];
    render(
      <SelectInput
        name="Visibility"
        fieldName="visibility"
        value="public"
        onChange={() => {}}
        options={options}
        placeholder="Pick one"
      />,
    );
    const select = screen.getByRole('combobox');
    expect(select).toHaveValue('public');
    expect(screen.getByText('Public')).toBeInTheDocument();
    expect(screen.getByText('Private')).toBeInTheDocument();
    expect(screen.getByText('Pick one')).toBeInTheDocument();
  });
});
