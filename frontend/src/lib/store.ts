import { create } from 'zustand'

export type AppPanel = 'chat' | 'tasks' | 'runs' | 'agents' | 'commands' | 'settings'
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
  setActivePanel: (panel) =>
    set((state) => (state.activePanel === panel ? state : { activePanel: panel })),
  setActiveConversationId: (id) =>
    set((state) => (state.activeConversationId === id ? state : { activeConversationId: id })),
  setActiveTaskId: (id) =>
    set((state) => (state.activeTaskId === id ? state : { activeTaskId: id })),
  setActiveRunId: (id) =>
    set((state) => (state.activeRunId === id ? state : { activeRunId: id })),
  setTheme: (theme) => set((state) => (state.theme === theme ? state : { theme })),
  toggleTheme: () =>
    set((state) => ({
      theme: state.theme === 'dark' ? 'light' : 'dark',
    })),
  setProviderModalOpen: (open) =>
    set((state) => (state.providerModalOpen === open ? state : { providerModalOpen: open })),
  setTaskWizardOpen: (open, conversationId) =>
    set((state) => {
      const nextConversationId = conversationId ?? null
      if (state.taskWizardOpen === open && state.taskWizardConversationId === nextConversationId) {
        return state
      }
      return {
        taskWizardOpen: open,
        taskWizardConversationId: nextConversationId,
      }
    }),
}))
