import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi } from 'vitest';
import { MapInput } from '../components/forms/MapInput';
import type { ItemField } from '../api/types';

const dubbingItemFields: ItemField[] = [
  { name: 'Dubbing ID', fieldName: 'dubbingId', type: 'string', order: 1 },
  { name: 'Title', fieldName: 'title', type: 'string', order: 2 },
];

const sampleEntries = {
  es: { dubbingId: 'dub-123', title: 'Título' },
  fr: { dubbingId: 'dub-456', title: 'Titre' },
};

describe('MapInput', () => {
  it('renders entries by key', () => {
    render(
      <MapInput
        name="Dubbing"
        fieldName="dubbing"
        value={sampleEntries}
        onChange={() => {}}
        itemFields={dubbingItemFields}
        mapKeyLabel="Language Code"
      />,
    );
    expect(screen.getByText('Dubbing')).toBeInTheDocument();
    expect(screen.getByText(/Language Code: es/)).toBeInTheDocument();
    expect(screen.getByText(/Language Code: fr/)).toBeInTheDocument();
    expect(screen.getByDisplayValue('Título')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Titre')).toBeInTheDocument();
  });

  it('renders empty state with add controls', () => {
    render(
      <MapInput
        name="Dubbing"
        fieldName="dubbing"
        value={{}}
        onChange={() => {}}
        itemFields={dubbingItemFields}
      />,
    );
    expect(screen.getByText('+ Add Entry')).toBeInTheDocument();
    expect(screen.getByLabelText('New Key')).toBeInTheDocument();
  });

  it('adds entry with key', async () => {
    const onChange = vi.fn();
    render(
      <MapInput
        name="Dubbing"
        fieldName="dubbing"
        value={{}}
        onChange={onChange}
        itemFields={dubbingItemFields}
      />,
    );
    const keyInput = screen.getByLabelText('New Key');
    await userEvent.type(keyInput, 'de');
    await userEvent.click(screen.getByText('+ Add Entry'));
    expect(onChange).toHaveBeenCalledWith('dubbing', {
      de: { dubbingId: '', title: '' },
    });
  });

  it('removes entry by key', async () => {
    const onChange = vi.fn();
    render(
      <MapInput
        name="Dubbing"
        fieldName="dubbing"
        value={sampleEntries}
        onChange={onChange}
        itemFields={dubbingItemFields}
      />,
    );
    await userEvent.click(screen.getByLabelText('Remove entry es'));
    expect(onChange).toHaveBeenCalledWith('dubbing', { fr: sampleEntries.fr });
  });

  it('updates sub-field value', async () => {
    const onChange = vi.fn();
    render(
      <MapInput
        name="Dubbing"
        fieldName="dubbing"
        value={sampleEntries}
        onChange={onChange}
        itemFields={dubbingItemFields}
      />,
    );
    const titleInput = screen.getByDisplayValue('Título');
    await userEvent.type(titleInput, '!');
    const lastCall = onChange.mock.calls[onChange.mock.calls.length - 1];
    expect(lastCall[0]).toBe('dubbing');
    expect(lastCall[1].es.title).toBe('Título!');
  });

  it('does not add entry with empty key', async () => {
    const onChange = vi.fn();
    render(
      <MapInput
        name="Dubbing"
        fieldName="dubbing"
        value={{}}
        onChange={onChange}
        itemFields={dubbingItemFields}
      />,
    );
    const addButton = screen.getByText('+ Add Entry');
    expect(addButton).toBeDisabled();
  });
});
