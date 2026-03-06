import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect } from 'vitest';
import { VideoDetail } from '../pages/VideoDetail';

function renderWithRoute() {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={['/videos/devops/test-video']}>
        <Routes>
          <Route
            path="/videos/:category/:videoName"
            element={<VideoDetail />}
          />
          <Route path="/phases/:phaseId" element={<div>Phase Page</div>} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('VideoDetail', () => {
  it('displays video name', async () => {
    renderWithRoute();
    expect(await screen.findByText('test-video')).toBeInTheDocument();
  });

  it('shows overall progress', async () => {
    renderWithRoute();
    await screen.findByText('test-video');
    expect(screen.getByText('Overall Progress')).toBeInTheDocument();
  });

  it('renders tabs from aspect metadata', async () => {
    renderWithRoute();
    await screen.findByText('test-video');
    // Tabs come from mock aspects
    const tabs = await screen.findAllByRole('tab');
    expect(tabs.length).toBe(2);
    expect(tabs[0]).toHaveTextContent('Initial Details');
    expect(tabs[1]).toHaveTextContent('Work Progress');
  });

  it('renders first tab fields by default', async () => {
    renderWithRoute();
    await screen.findByText('test-video');
    // First aspect fields should be visible
    expect(await screen.findByText('Project Name')).toBeInTheDocument();
  });

  it('switches tabs when clicked', async () => {
    renderWithRoute();
    await screen.findByText('test-video');
    const tabs = await screen.findAllByRole('tab');
    await userEvent.click(tabs[1]);
    // Work Progress fields
    expect(await screen.findByText('Screen Recording')).toBeInTheDocument();
  });

  it('shows delete button with confirmation', async () => {
    renderWithRoute();
    await screen.findByText('test-video');
    const deleteBtn = screen.getByRole('button', { name: 'Delete Video' });
    await userEvent.click(deleteBtn);
    expect(screen.getByText('Are you sure?')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Confirm Delete' })).toBeInTheDocument();
  });

  it('cancels delete confirmation', async () => {
    renderWithRoute();
    await screen.findByText('test-video');
    await userEvent.click(screen.getByRole('button', { name: 'Delete Video' }));
    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(screen.getByRole('button', { name: 'Delete Video' })).toBeInTheDocument();
  });

  it('saves triggers PATCH and shows success', async () => {
    renderWithRoute();
    await screen.findByText('test-video');
    // Edit the project name field
    const input = await screen.findByDisplayValue('Test Project');
    await userEvent.clear(input);
    await userEvent.type(input, 'Updated');
    await userEvent.click(screen.getByRole('button', { name: 'Save' }));
    await waitFor(() => {
      expect(screen.getByText('Saved and synced.')).toBeInTheDocument();
    });
  });

  it('confirms delete navigates away', async () => {
    renderWithRoute();
    await screen.findByText('test-video');
    await userEvent.click(screen.getByRole('button', { name: 'Delete Video' }));
    await userEvent.click(screen.getByRole('button', { name: 'Confirm Delete' }));
    await waitFor(() => {
      expect(screen.getByText('Phase Page')).toBeInTheDocument();
    });
  });
});
