import { describe, it, expect } from 'vitest';
import { useUIStore } from '../stores/ui';

describe('UI store', () => {
  it('starts with sidebar open', () => {
    const state = useUIStore.getState();
    expect(state.sidebarOpen).toBe(true);
  });

  it('toggles sidebar', () => {
    useUIStore.getState().toggleSidebar();
    expect(useUIStore.getState().sidebarOpen).toBe(false);
    useUIStore.getState().toggleSidebar();
    expect(useUIStore.getState().sidebarOpen).toBe(true);
  });

  it('sets selected phase', () => {
    useUIStore.getState().setSelectedPhase(3);
    expect(useUIStore.getState().selectedPhase).toBe(3);
    useUIStore.getState().setSelectedPhase(null);
    expect(useUIStore.getState().selectedPhase).toBeNull();
  });
});
