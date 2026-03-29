import { defineStore } from 'pinia'
import {
  GetSuiteStatus as GetSuiteStatusApi
} from '../../wailsjs/go/bindings/AutoBenchBinding'

const ENABLE_AUTOBENCH_BACKEND = typeof window !== 'undefined' && !!window.go?.bindings?.AutoBenchBinding

export const useAutoBenchStore = defineStore('autobench', {
  state: () => ({
    createdSuiteId: '',
    suiteStatus: null,
    isPolling: false
  }),

  getters: {
    hasActiveSuite: (state) => !!state.createdSuiteId,
    isSuiteRunning: (state) => state.suiteStatus?.status === 'running',
    isSuiteDone: (state) => !!state.suiteStatus && ['completed', 'failed', 'cancelled'].includes(state.suiteStatus.status)
  },

  actions: {
    setSuiteId(id) {
      this.createdSuiteId = id
    },

    setSuiteStatus(status) {
      this.suiteStatus = status
    },

    setPolling(val) {
      this.isPolling = val
    },

    async fetchSuiteStatus() {
      if (!this.createdSuiteId) return null

      try {
        if (ENABLE_AUTOBENCH_BACKEND) {
          const result = await GetSuiteStatusApi(this.createdSuiteId)
          if (result.error) {
            console.warn('fetchSuiteStatus error:', result.error)
            return null
          }
          this.suiteStatus = result
          return result
        }
        return null
      } catch (err) {
        console.warn('fetchSuiteStatus exception:', err)
        return null
      }
    },

    resetSuite() {
      this.createdSuiteId = ''
      this.suiteStatus = null
      this.isPolling = false
    }
  }
})
