import { defineStore } from 'pinia'
import {
  ListReports as ListReportsApi,
  GetReport as GetReportApi,
  GetReportMetrics as GetReportMetricsApi,
  ListSuites as ListSuitesApi,
  GetSuite as GetSuiteApi
} from '../../wailsjs/go/bindings/ReportBinding'

const ENABLE_REPORT_BACKEND = typeof window !== 'undefined' && !!window.go?.bindings?.ReportBinding

export const useReportStore = defineStore('report', {
  state: () => ({
    reports: [],
    suites: [],
    selectedReportId: '',
    selectedReport: null,
    selectedReportMetrics: null,
    isDetailOpen: false,
    loading: false,
    error: null,
    notice: null,
    pagination: {
      page: 1,
      pageSize: 20,
      total: 0
    },
    filters: {
      suiteId: '',
      status: '',
      connectionId: ''
    }
  }),

  getters: {
    hasReports: (state) => state.reports.length > 0,
    isSelected: (state) => !!state.selectedReportId,
    hasMetrics: (state) => !!state.selectedReportMetrics,
    reportById: (state) => (id) => state.reports.find((r) => r.id === id),
    reportsBySuiteId: (state) => (suiteId) => state.reports.filter((r) => r.suite_id === suiteId),
    reportsByStatus: (state) => (status) => state.reports.filter((r) => r.status === status)
  },

  actions: {
    async fetchReports(opts = {}) {
      this.loading = true
      this.error = null

      const options = {
        page: opts.page || this.pagination.page,
        page_size: opts.pageSize || this.pagination.pageSize,
        suite_id: opts.suiteId || this.filters.suiteId || '',
        status: opts.status || this.filters.status || '',
        connection_id: opts.connectionId || this.filters.connectionId || ''
      }

      try {
        if (ENABLE_REPORT_BACKEND) {
          const result = await ListReportsApi(options)
          if (result.error) {
            throw new Error(result.error)
          }
          this.reports = result.reports || []
          this.pagination.total = result.total || 0
        } else {
          // Mock data for development
          this.reports = []
          this.pagination.total = 0
        }
      } catch (err) {
        this.error = err.message || 'Failed to load reports'
        this.reports = []
        this.showNotice(this.error, 'warning')
      } finally {
        this.loading = false
      }
    },

    async fetchReport(id) {
      this.loading = true
      this.error = null

      try {
        if (ENABLE_REPORT_BACKEND) {
          const result = await GetReportApi(id)
          if (result.error) {
            throw new Error(result.error)
          }
          this.selectedReport = result.report
          this.selectedReportId = id
        } else {
          this.selectedReport = null
          this.selectedReportId = id
        }
      } catch (err) {
        this.error = err.message || 'Failed to load report'
        this.showNotice(this.error, 'warning')
      } finally {
        this.loading = false
      }
    },

    async fetchReportMetrics(id) {
      this.loading = true
      this.error = null

      try {
        if (ENABLE_REPORT_BACKEND) {
          const result = await GetReportMetricsApi(id)
          if (result.error) {
            throw new Error(result.error)
          }
          this.selectedReportMetrics = result.metrics
        } else {
          this.selectedReportMetrics = null
        }
      } catch (err) {
        this.error = err.message || 'Failed to load report metrics'
        this.showNotice(this.error, 'warning')
      } finally {
        this.loading = false
      }
    },

    async openReportDetail(id) {
      this.selectedReportId = id
      this.isDetailOpen = true

      await Promise.all([
        this.fetchReport(id),
        this.fetchReportMetrics(id)
      ])
    },

    closeReportDetail() {
      this.isDetailOpen = false
      this.selectedReportId = ''
      this.selectedReport = null
      this.selectedReportMetrics = null
    },

    setFilter(key, value) {
      if (key in this.filters) {
        this.filters[key] = value
      }
    },

    resetFilters() {
      this.filters = {
        suiteId: '',
        status: '',
        connectionId: ''
      }
    },

    setPage(page) {
      this.pagination.page = page
    },

    setPageSize(pageSize) {
      this.pagination.pageSize = pageSize
      this.pagination.page = 1
    },

    async fetchSuites(opts = {}) {
      const options = {
        page: opts.page || 1,
        page_size: opts.pageSize || 20,
        status: opts.status || ''
      }

      try {
        if (ENABLE_REPORT_BACKEND) {
          const result = await ListSuitesApi(options)
          if (result.error) {
            throw new Error(result.error)
          }
          this.suites = result.suites || []
        } else {
          this.suites = []
        }
      } catch (err) {
        this.error = err.message || 'Failed to load suites'
        this.showNotice(this.error, 'warning')
      }
    },

    async fetchSuite(id) {
      try {
        if (ENABLE_REPORT_BACKEND) {
          const result = await GetSuiteApi(id)
          if (result.error) {
            throw new Error(result.error)
          }
          return result.suite
        }
        return null
      } catch (err) {
        this.error = err.message || 'Failed to load suite'
        this.showNotice(this.error, 'warning')
        return null
      }
    },

    clearNotice() {
      this.notice = null
    },

    showNotice(message, tone = 'info') {
      this.notice = { message, tone, timestamp: Date.now() }
    }
  }
})
