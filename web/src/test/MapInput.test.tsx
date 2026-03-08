import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi } from 'vitest';
import { MapInput } from '../components/forms/MapInput';
import type { ItemField } from '../api/types';

const metadataItemFields: ItemField[] = [
  { name: 'ID', fieldName: 'id', type: 'string', order: 1 },
  { name: 'Value', fieldName: 'value', type: 'string', order: 2 },
];

const sampleEntries = {
  key1: { id: 'id-1', value: 'val-1' },
  key2: { id: 'id-2', value: 'val-2' },
};

describe('MapInput', () => {
  it('renders entries by key', () => {
    render(
      <MapInput
        name="Metadata"
        fieldName="metadata"
        value={sampleEntries}
        onChange={() => {}}
        itemFields={metadataItemFields}
        mapKeyLabel="Key"
      />,
    );
    expect(screen.getByText('Metadata')).toBeInTheDocument();
    expect(screen.getByText(/Key: key1/)).toBeInTheDocument();
    expect(screen.getByText(/Key: key2/)).toBeInTheDocument();
    expect(screen.getByDisplayValue('val-1')).toBeInTheDocument();
    expect(screen.getByDisplayValue('val-2')).toBeInTheDocument();
  });

  it('renders empty state with add controls', () => {
    render(
      <MapInput
        name="Metadata"
        fieldName="metadata"
        value={{}}
        onChange={() => {}}
        itemFields={metadataItemFields}
      />,
    );
    expect(screen.getByText('+ Add Entry')).toBeInTheDocument();
    expect(screen.getByLabelText('New Key')).toBeInTheDocument();
  });

  it('adds entry with key', async () => {
    const onChange = vi.fn();
    render(
      <MapInput
        name="Metadata"
        fieldName="metadata"
        value={{}}
        onChange={onChange}
        itemFields={metadataItemFields}
      />,
    );
    const keyInput = screen.getByLabelText('New Key');
    await userEvent.type(keyInput, 'key3');
    await userEvent.click(screen.getByText('+ Add Entry'));
    expect(onChange).toHaveBeenCalledWith('metadata', {
      key3: { id: '', value: '' },
    });
  });

  it('removes entry by key', async () => {
    const onChange = vi.fn();
    render(
      <MapInput
        name="Metadata"
        fieldName="metadata"
        value={sampleEntries}
        onChange={onChange}
        itemFields={metadataItemFields}
      />,
    );
    await userEvent.click(screen.getByLabelText('Remove entry key1'));
    expect(onChange).toHaveBeenCalledWith('metadata', { key2: sampleEntries.key2 });
  });

  it('updates sub-field value', async () => {
    const onChange = vi.fn();
    render(
      <MapInput
        name="Metadata"
        fieldName="metadata"
        value={sampleEntries}
        onChange={onChange}
        itemFields={metadataItemFields}
      />,
    );
    const valueInput = screen.getByDisplayValue('val-1');
    await userEvent.type(valueInput, '!');
    const lastCall = onChange.mock.calls[onChange.mock.calls.length - 1];
    expect(lastCall[0]).toBe('metadata');
    expect(lastCall[1].key1.value).toBe('val-1!');
  });

  it('does not add entry with empty key', async () => {
    const onChange = vi.fn();
    render(
      <MapInput
        name="Metadata"
        fieldName="metadata"
        value={{}}
        onChange={onChange}
        itemFields={metadataItemFields}
      />,
    );
    const addButton = screen.getByText('+ Add Entry');
    expect(addButton).toBeDisabled();
  });
});
