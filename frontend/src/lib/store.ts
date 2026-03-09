import { create } from 'zustand'

export type AppPanel = 'chat' | 'tasks' | 'runs' | 'settings'
export type AppTheme = 'light' | 'dark'

interface UIState {
  activePanel: AppPanel
  activeConversationId: string | null
  activeTaskId: string | null
  activeRunId: string | null
  theme: AppTheme
  providerModalOpen: boolean
  taskWizardOpen: boolean
  taskWizardConversationId: string | null
  setActivePanel: (panel: AppPanel) => void
  setActiveConversationId: (id: string | null) => void
  setActiveTaskId: (id: string | null) => void
  setActiveRunId: (id: string | null) => void
  setTheme: (theme: AppTheme) => void
  toggleTheme: () => void
  setProviderModalOpen: (open: boolean) => void
  setTaskWizardOpen: (open: boolean, conversationId?: string | null) => void
}

export const useUIStore = create<UIState>((set) => ({
  activePanel: 'chat',
  activeConversationId: null,
  activeTaskId: null,
  activeRunId: null,
  theme: 'dark',
  providerModalOpen: false,
  taskWizardOpen: false,
  taskWizardConversationId: null,
  setActivePanel: (panel) => set({ activePanel: panel }),
  setActiveConversationId: (id) => set({ activeConversationId: id }),
  setActiveTaskId: (id) => set({ activeTaskId: id }),
  setActiveRunId: (id) => set({ activeRunId: id }),
  setTheme: (theme) => set({ theme }),
  toggleTheme: () =>
    set((state) => ({
      theme: state.theme === 'dark' ? 'light' : 'dark',
    })),
  setProviderModalOpen: (open) => set({ providerModalOpen: open }),
  setTaskWizardOpen: (open, conversationId) =>
    set({
      taskWizardOpen: open,
      taskWizardConversationId: conversationId ?? null,
    }),
}))
