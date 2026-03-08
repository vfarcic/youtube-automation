import { create } from 'zustand';

interface UIState {
  sidebarOpen: boolean;
  selectedPhase: number | null;
  setSidebarOpen: (open: boolean) => void;
  toggleSidebar: () => void;
  setSelectedPhase: (phase: number | null) => void;
}

export const useUIStore = create<UIState>((set) => ({
  sidebarOpen: true,
  selectedPhase: null,
  setSidebarOpen: (open) => set({ sidebarOpen: open }),
  toggleSidebar: () => set((s) => ({ sidebarOpen: !s.sidebarOpen })),
  setSelectedPhase: (phase) => set({ selectedPhase: phase }),
}));
