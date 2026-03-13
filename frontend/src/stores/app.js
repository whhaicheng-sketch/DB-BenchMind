import { defineStore } from 'pinia'

export const useAppStore = defineStore('app', {
  state: () => ({
    activeTab: 'connections',
    pendingTaskTemplate: null
  }),

  actions: {
    setActiveTab(tabId) {
      this.activeTab = tabId
    },

    queueTemplateForTask(payload) {
      this.pendingTaskTemplate = payload
      this.activeTab = 'tasks'
    },

    clearPendingTaskTemplate() {
      this.pendingTaskTemplate = null
    }
  }
})
