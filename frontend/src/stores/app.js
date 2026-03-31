import { defineStore } from 'pinia'
import { clearPendingTaskTemplateState, queueTemplateForTaskState } from './appState.mjs'

export const useAppStore = defineStore('app', {
  state: () => ({
    activeTab: 'connections',
    pendingTaskTemplate: null,
    pendingReportId: ''
  }),

  actions: {
    setActiveTab(tabId) {
      this.activeTab = tabId
    },

    queueTemplateForTask(payload) {
      Object.assign(this, queueTemplateForTaskState(this.$state, payload))
    },

    clearPendingTaskTemplate() {
      Object.assign(this, clearPendingTaskTemplateState(this.$state))
    },

    setPendingReportId(id) {
      this.pendingReportId = id
    },

    consumePendingReportId() {
      const id = this.pendingReportId
      this.pendingReportId = ''
      return id
    }
  }
})
