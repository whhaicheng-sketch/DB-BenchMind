import { defineStore } from 'pinia'
import {
  ListReports as ListReportsApi,
  GetReport as GetReportApi,
  GetReportMetrics as GetReportMetricsApi,
  ListSuites as ListSuitesApi,
  GetSuite as GetSuiteApi,
  ExportReportJSON as ExportReportJSONApi,
  ExportReportHTML as ExportReportHTMLApi,
  GetExportFilePaths as GetExportFilePathsApi
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
    exporting: false,
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
        this.error = err.message || '加载报告列表失败'
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
        this.error = err.message || '加载报告详情失败'
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
          return result.metrics
        } else {
          this.selectedReportMetrics = null
          return null
        }
      } catch (err) {
        this.error = err.message || '加载报告指标失败'
        this.showNotice(this.error, 'warning')
        return null
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
        this.error = err.message || '加载套件列表失败'
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
        this.error = err.message || '加载套件详情失败'
        this.showNotice(this.error, 'warning')
        return null
      }
    },

    // 导出功能
    async exportReportJSON(id) {
      this.exporting = true
      this.error = null

      try {
        if (ENABLE_REPORT_BACKEND) {
          const result = await ExportReportJSONApi(id)
          if (result.error) {
            throw new Error(result.error)
          }
          // 返回 JSON 字符串
          return result.data
        } else {
          // Mock 导出
          const mockData = {
            schema_version: 'v1',
            report_id: id,
            exported_at: new Date().toISOString(),
            report: this.selectedReport,
            metrics: this.selectedReportMetrics
          }
          return JSON.stringify(mockData, null, 2)
        }
      } catch (err) {
        this.error = err.message || '导出 JSON 失败'
        this.showNotice(this.error, 'warning')
        return null
      } finally {
        this.exporting = false
      }
    },

    async exportReportHTML(id) {
      this.exporting = true
      this.error = null

      try {
        if (ENABLE_REPORT_BACKEND) {
          const result = await ExportReportHTMLApi(id)
          if (result.error) {
            throw new Error(result.error)
          }
          return result.html
        } else {
          // Mock HTML 导出
          return this.generateMockHTML()
        }
      } catch (err) {
        this.error = err.message || '导出 HTML 失败'
        this.showNotice(this.error, 'warning')
        return null
      } finally {
        this.exporting = false
      }
    },

    async getExportFilePaths(id) {
      try {
        if (ENABLE_REPORT_BACKEND) {
          const result = await GetExportFilePathsApi(id)
          if (result.error) {
            throw new Error(result.error)
          }
          return {
            metrics: result.metrics,
            monitoring: result.monitoring,
            raw: result.raw,
            html: result.html
          }
        }
        return null
      } catch (err) {
        this.error = err.message || '获取导出路径失败'
        this.showNotice(this.error, 'warning')
        return null
      }
    },

    // 下载 JSON 文件
    async downloadJSON(id, filename) {
      const jsonData = await this.exportReportJSON(id)
      if (!jsonData) return false

      try {
        const blob = new Blob([jsonData], { type: 'application/json' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = filename || `report-${id}.json`
        document.body.appendChild(a)
        a.click()
        document.body.removeChild(a)
        URL.revokeObjectURL(url)

        this.showNotice('JSON 导出成功', 'success')
        return true
      } catch (err) {
        this.error = err.message || '下载 JSON 失败'
        this.showNotice(this.error, 'warning')
        return false
      }
    },

    // 下载 HTML 文件
    async downloadHTML(id, filename) {
      const htmlContent = await this.exportReportHTML(id)
      if (!htmlContent) return false

      try {
        const blob = new Blob([htmlContent], { type: 'text/html' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = filename || `report-${id}.html`
        document.body.appendChild(a)
        a.click()
        document.body.removeChild(a)
        URL.revokeObjectURL(url)

        this.showNotice('HTML 导出成功', 'success')
        return true
      } catch (err) {
        this.error = err.message || '下载 HTML 失败'
        this.showNotice(this.error, 'warning')
        return false
      }
    },

    // 复制 JSON 到剪贴板
    async copyJSONToClipboard(id) {
      const jsonData = await this.exportReportJSON(id)
      if (!jsonData) return false

      try {
        await navigator.clipboard.writeText(jsonData)
        this.showNotice('已复制到剪贴板', 'success')
        return true
      } catch (err) {
        this.error = err.message || '复制失败'
        this.showNotice(this.error, 'warning')
        return false
      }
    },

    // 生成 Mock HTML（用于开发测试）
    generateMockHTML() {
      const report = this.selectedReport || {}
      const metrics = this.selectedReportMetrics || {}

      return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>压测报告 - ${report.id || 'Unknown'}</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; }
    h1 { color: #333; }
    .metric { display: inline-block; margin: 10px 20px 10px 0; padding: 15px; background: #f5f5f5; border-radius: 8px; }
    .metric-label { font-size: 12px; color: #666; text-transform: uppercase; }
    .metric-value { font-size: 24px; font-weight: bold; color: #333; }
    .section { margin: 20px 0; padding: 20px; border: 1px solid #eee; border-radius: 8px; }
  </style>
</head>
<body>
  <h1>压测报告</h1>
  <div class="section">
    <h2>基本信息</h2>
    <p><strong>报告 ID:</strong> ${report.id || 'N/A'}</p>
    <p><strong>数据库:</strong> ${report.database_type || 'N/A'}</p>
    <p><strong>状态:</strong> ${report.status || 'N/A'}</p>
    <p><strong>开始时间:</strong> ${report.started_at || 'N/A'}</p>
  </div>
  <div class="section">
    <h2>性能指标</h2>
    <div class="metric">
      <div class="metric-label">TPM</div>
      <div class="metric-value">${report.tpm?.toFixed(2) || 'N/A'}</div>
    </div>
    <div class="metric">
      <div class="metric-label">TPS</div>
      <div class="metric-value">${report.tps?.toFixed(2) || 'N/A'}</div>
    </div>
    <div class="metric">
      <div class="metric-label">平均延迟</div>
      <div class="metric-value">${report.latency_avg_ms?.toFixed(2) || 'N/A'} ms</div>
    </div>
  </div>
  <footer style="margin-top: 40px; padding-top: 20px; border-top: 1px solid #eee; color: #666; font-size: 12px;">
    Generated by DB-BenchMind at ${new Date().toISOString()}
  </footer>
</body>
</html>`
    },

    clearNotice() {
      this.notice = null
    },

    showNotice(message, tone = 'info') {
      this.notice = { message, tone, timestamp: Date.now() }
    }
  }
})
