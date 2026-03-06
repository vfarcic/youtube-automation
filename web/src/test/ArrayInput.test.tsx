import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi } from 'vitest';
import { ArrayInput } from '../components/forms/ArrayInput';
import type { ItemField } from '../api/types';

const titleItemFields: ItemField[] = [
  { name: 'Index', fieldName: 'index', type: 'number', order: 1 },
  { name: 'Text', fieldName: 'text', type: 'string', order: 2 },
  { name: 'Share', fieldName: 'share', type: 'number', order: 3 },
];

const sampleItems = [
  { index: 1, text: 'First Title', share: 0.6 },
  { index: 2, text: 'Second Title', share: 0.4 },
];

describe('ArrayInput', () => {
  it('renders items from value prop', () => {
    render(
      <ArrayInput
        name="Titles"
        fieldName="titles"
        value={sampleItems}
        onChange={() => {}}
        itemFields={titleItemFields}
      />,
    );
    expect(screen.getByText('Titles')).toBeInTheDocument();
    expect(screen.getByText('Item 1')).toBeInTheDocument();
    expect(screen.getByText('Item 2')).toBeInTheDocument();
    expect(screen.getByDisplayValue('First Title')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Second Title')).toBeInTheDocument();
  });

  it('renders empty state with add button', () => {
    render(
      <ArrayInput
        name="Titles"
        fieldName="titles"
        value={[]}
        onChange={() => {}}
        itemFields={titleItemFields}
      />,
    );
    expect(screen.getByText('+ Add Item')).toBeInTheDocument();
    expect(screen.queryByText('Item 1')).not.toBeInTheDocument();
  });

  it('calls onChange with new item on add', async () => {
    const onChange = vi.fn();
    render(
      <ArrayInput
        name="Titles"
        fieldName="titles"
        value={[]}
        onChange={onChange}
        itemFields={titleItemFields}
      />,
    );
    await userEvent.click(screen.getByText('+ Add Item'));
    expect(onChange).toHaveBeenCalledWith('titles', [{ index: 0, text: '', share: 0 }]);
  });

  it('calls onChange without removed item on remove', async () => {
    const onChange = vi.fn();
    render(
      <ArrayInput
        name="Titles"
        fieldName="titles"
        value={sampleItems}
        onChange={onChange}
        itemFields={titleItemFields}
      />,
    );
    await userEvent.click(screen.getByLabelText('Remove item 1'));
    expect(onChange).toHaveBeenCalledWith('titles', [sampleItems[1]]);
  });

  it('calls onChange with updated sub-field value', async () => {
    const onChange = vi.fn();
    render(
      <ArrayInput
        name="Titles"
        fieldName="titles"
        value={sampleItems}
        onChange={onChange}
        itemFields={titleItemFields}
      />,
    );
    const textInput = screen.getByDisplayValue('First Title');
    await userEvent.type(textInput, '!');
    // The last call should have the appended text
    const lastCall = onChange.mock.calls[onChange.mock.calls.length - 1];
    expect(lastCall[0]).toBe('titles');
    expect(lastCall[1][0].text).toBe('First Title!');
  });
});
