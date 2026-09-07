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
    expect(onChange).toHaveBeenCalledWith('titles', [{ index: 1, text: '', share: 0 }]);
  });

  // In production `index` is tagged ui:"auto" on the Go struct, so the server
  // strips it from itemFields entirely. The added item must still carry one:
  // the upload title is the one with index 1, and titles left at index 0 make
  // the video read as having no title at all.
  it('numbers added items even when index is not an editable field', async () => {
    const onChange = vi.fn();
    const withoutIndex: ItemField[] = [{ name: 'Text', fieldName: 'text', type: 'string', order: 1 }];
    render(
      <ArrayInput
        name="Titles"
        fieldName="titles"
        value={[]}
        onChange={onChange}
        itemFields={withoutIndex}
      />,
    );
    await userEvent.click(screen.getByText('+ Add Item'));
    expect(onChange).toHaveBeenCalledWith('titles', [{ index: 1, text: '' }]);
  });

  it('gives an added item the next free index', async () => {
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
    await userEvent.click(screen.getByText('+ Add Item'));
    expect(onChange).toHaveBeenCalledWith('titles', [
      ...sampleItems,
      { index: 3, text: '', share: 0 },
    ]);
  });

  it('reuses a freed index rather than colliding with a surviving sibling', async () => {
    const onChange = vi.fn();
    render(
      <ArrayInput
        name="Titles"
        fieldName="titles"
        value={[{ index: 2, text: 'Second Title', share: 0.4 }]}
        onChange={onChange}
        itemFields={titleItemFields}
      />,
    );
    await userEvent.click(screen.getByText('+ Add Item'));
    expect(onChange).toHaveBeenCalledWith('titles', [
      { index: 2, text: 'Second Title', share: 0.4 },
      { index: 1, text: '', share: 0 },
    ]);
  });

  it('leaves non-indexed arrays without an index field', async () => {
    const onChange = vi.fn();
    const shortFields: ItemField[] = [{ name: 'Title', fieldName: 'title', type: 'string', order: 1 }];
    render(
      <ArrayInput
        name="Shorts"
        fieldName="shorts"
        value={[]}
        onChange={onChange}
        itemFields={shortFields}
      />,
    );
    await userEvent.click(screen.getByText('+ Add Item'));
    expect(onChange).toHaveBeenCalledWith('shorts', [{ title: '' }]);
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

  it('does NOT render upload/publish buttons for shorts in ArrayInput', () => {
    const shortFields: ItemField[] = [
      { name: 'Title', fieldName: 'title', type: 'string', order: 1 },
      { name: 'Text', fieldName: 'text', type: 'string', order: 2 },
    ];
    const shortItems = [
      { id: 'short1', title: 'My Short', text: 'Some text', scheduledDate: '2026-01-15' },
    ];
    render(
      <ArrayInput
        name="Shorts"
        fieldName="shorts"
        value={shortItems}
        onChange={() => {}}
        itemFields={shortFields}
        category="devops"
        videoName="test-video"
      />,
    );
    expect(screen.queryByText('Upload to Drive')).not.toBeInTheDocument();
    expect(screen.queryByText('Publish to YouTube')).not.toBeInTheDocument();
  });
});
