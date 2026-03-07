/**
 * Monitor Pinia Store
 * Manages real-time monitoring state for DB-BenchMind Wails frontend.
 */
import { defineStore } from 'pinia'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import {
  StartMonitoring,
  StopMonitoring,
  GetMonitorState,
  GetMonitorData,
  ClearData,
  StartSystemMonitoring,
  StopSystemMonitoring,
  GetSystemMetrics,
  IsSystemMonitoring
} from '../../wailsjs/go/bindings/MonitorBinding'

/**
 * Get fluctuation status based on CV value
 * @param {number} cv - Coefficient of Variation
 * @returns {string} - 'stable', 'fluctuating', or 'sawtooth'
 */
function getFluctuationStatus(cv) {
  if (cv < 0.05) return 'stable'
  if (cv < 0.10) return 'fluctuating'
  return 'sawtooth'
}

export const useMonitorStore = defineStore('monitor', {
  state: () => ({
    // Benchmark monitoring state
    isMonitoring: false,
    currentRunId: null,

    // Current benchmark metrics
    currentTPM: 0,
    currentTPS: 0,
    currentLatency: 0,
    currentErrors: 0,

    // Benchmark statistics
    tpmStats: {
      avg: 0,
      max: 0,
      min: 0,
      stddev: 0,
      cv: 0,
      directionChanges: 0
    },
    tpsStats: {
      avg: 0,
      max: 0,
      min: 0,
      stddev: 0,
      cv: 0,
      directionChanges: 0
    },

    // Fluctuation status (stable/fluctuating/sawtooth)
    tpmStatus: 'stable',
    tpsStatus: 'stable',

    // Benchmark history data for charts
    tpmHistory: [],
    tpsHistory: [],

    // System monitoring state
    systemMonitoring: false,

    // System metrics
    cpuPercent: 0,
    diskReadBps: 0,
    diskWriteBps: 0,
    diskUsedPercent: 0,
    diskUsedGB: 0,
    diskTotalGB: 0,

    // System history for charts
    cpuHistory: [],
    diskIOHistory: [],
    diskSpaceHistory: [],

    // Loading state
    loading: false,
    error: null
  }),

  getters: {
    // Get formatted TPM for display
    formattedTPM: (state) => {
      return state.currentTPM.toLocaleString('en-US', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      })
    },

    // Get formatted TPS for display
    formattedTPS: (state) => {
      return state.currentTPS.toLocaleString('en-US', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      })
    },

    // Get formatted latency for display
    formattedLatency: (state) => {
      return `${state.currentLatency.toFixed(2)} ms`
    },

    // Check if has benchmark data
    hasData: (state) => {
      return state.tpmHistory.length > 0 || state.tpsHistory.length > 0
    },

    // Get latest N TPM points for sparkline
    getLatestTPMPoints: (state) => (count) => {
      if (count <= 0) return state.tpmHistory
      return state.tpmHistory.slice(-count)
    },

    // Get latest N TPS points for sparkline
    getLatestTPSPoints: (state) => (count) => {
      if (count <= 0) return state.tpsHistory
      return state.tpsHistory.slice(-count)
    },

    // Get latest N CPU points for sparkline
    getLatestCPUPoints: (state) => (count) => {
      if (count <= 0) return state.cpuHistory
      return state.cpuHistory.slice(-count)
    },

    // Get TPM fluctuation color based on CV
    tpmColor: (state) => {
      if (!state.isMonitoring || state.tpmStats.cv === 0) return '#DC5050' // Default red
      if (state.tpmStats.cv < 0.05) return '#4CAF50' // Green - stable
      if (state.tpmStats.cv < 0.10) return '#FFC107' // Yellow - fluctuating
      return '#F44336' // Red - sawtooth
    },

    // Get TPS fluctuation color based on CV
    tpsColor: (state) => {
      if (!state.isMonitoring || state.tpsStats.cv === 0) return '#FFA500' // Default orange
      if (state.tpsStats.cv < 0.05) return '#4CAF50' // Green - stable
      if (state.tpsStats.cv < 0.10) return '#FFC107' // Yellow - fluctuating
      return '#F44336' // Red - sawtooth
    },

    // Get formatted TPM CV
    formattedTPMCV: (state) => {
      return (state.tpmStats.cv * 100).toFixed(2) + '%'
    },

    // Get formatted TPS CV
    formattedTPSCV: (state) => {
      return (state.tpsStats.cv * 100).toFixed(2) + '%'
    }
  },

  actions: {
    // ==================== Benchmark Monitoring ==================

    /**
     * Start benchmark monitoring
     */
    async startMonitoring(runId, config = {}) {
      try {
        const result = await StartMonitoring(runId, {
          buffer_size: config.bufferSize || 60,
          sample_rate_ms: config.sampleRateMs || 1000
        })
        if (result.success) {
          this.isMonitoring = true
          this.currentRunId = runId
          this.subscribeToBenchmarkEvents()
        } else {
          this.error = result.error || 'Failed to start monitoring'
        }
        return result
      } catch (err) {
        this.error = err.message || 'Failed to start monitoring'
        return { success: false, error: this.error }
      }
    },

    /**
     * Stop benchmark monitoring
     */
    async stopMonitoring() {
      try {
        const result = await StopMonitoring()
        this.isMonitoring = false
        this.currentRunId = null
        this.unsubscribeFromBenchmarkEvents()
        return result
      } catch (err) {
        this.error = err.message || 'Failed to stop monitoring'
        return { success: false, error: this.error }
      }
    },

    /**
     * Subscribe to benchmark monitoring events
     */
    subscribeToBenchmarkEvents() {
      // Monitor started event
      EventsOn('monitor:started', (data) => {
        console.log('Monitoring started:', data)
        this.isMonitoring = true
      })

      // Monitor stopped event
      EventsOn('monitor:stopped', (data) => {
        console.log('Monitoring stopped:', data)
        this.isMonitoring = false
      })

      // Metrics update event
      EventsOn('monitor:metrics_update', (data) => {
        this.currentTPM = data.tpm_current || 0
        this.currentTPS = data.tps_current || 0
        this.currentLatency = data.latency || 0

        // Update stats with CV and fluctuation data
        if (data.tpm_stats) {
          this.tpmStats = {
            avg: data.tpm_stats.tpm_avg || 0,
            max: data.tpm_stats.tpm_max || 0,
            min: data.tpm_stats.tpm_min || 0,
            stddev: data.tpm_stats.tpm_stddev || 0,
            cv: data.tpm_cv || data.tpm_stats.tpm_cv || 0,
            directionChanges: data.tpm_direction_changes || 0
          }
        }
        if (data.tps_stats) {
          this.tpsStats = {
            avg: data.tps_stats.tps_avg || 0,
            max: data.tps_stats.tps_max || 0,
            min: data.tps_stats.tps_min || 0,
            stddev: data.tps_stats.tps_stddev || 0,
            cv: data.tps_cv || data.tps_stats.tps_cv || 0,
            directionChanges: data.tps_direction_changes || 0
          }
        }

        // Update status
        if (data.tpm_cv !== undefined) {
          this.tpmStatus = getFluctuationStatus(data.tpm_cv)
        }
        if (data.tps_cv !== undefined) {
          this.tpsStatus = getFluctuationStatus(data.tps_cv)
        }

        // Update history
        if (data.tpm_current !== undefined) {
          this.tpmHistory.push({
            timestamp: Date.now(),
            tpm: data.tpm_current,
            tps: data.tps_current || 0
          })
          if (this.tpmHistory.length > 60) {
            this.tpmHistory = this.tpmHistory.slice(-60)
          }
        }
        if (data.tps_current !== undefined) {
          this.tpsHistory.push({
            timestamp: Date.now(),
            tpm: data.tpm_current || 0,
            tps: data.tps_current
          })
          if (this.tpsHistory.length > 60) {
            this.tpsHistory = this.tpsHistory.slice(-60)
          }
        }
      })
    },

    /**
     * Unsubscribe from benchmark monitoring events
     */
    unsubscribeFromBenchmarkEvents() {
      EventsOff('monitor:started')
      EventsOff('monitor:stopped')
      EventsOff('monitor:metrics_update')
    },

    /**
     * Fetch current monitor state
     */
    async fetchState() {
      try {
        const state = await GetMonitorState()
        this.isMonitoring = state.is_running || false
        this.currentRunId = state.run_id || null
        this.systemMonitoring = state.system_running || false
        return state
      } catch (err) {
        this.error = err.message || 'Failed to fetch state'
        return null
      }
    },

    /**
     * Fetch current monitor data
     */
    async fetchData() {
      try {
        const data = await GetMonitorData()
        this.currentTPM = data.current_tpm || 0
        this.currentTPS = data.current_tps || 0
        this.tpmHistory = data.tpm_points || []
        this.tpsHistory = data.tps_points || []
        if (data.stats) {
          this.tpmStats = {
            avg: data.stats.tpm_avg || 0,
            max: data.stats.tpm_max || 0,
            min: data.stats.tpm_min || 0
          }
          this.tpsStats = {
            avg: data.stats.tps_avg || 0,
            max: data.stats.tps_max || 0,
            min: data.stats.tps_min || 0
          }
        }
        return data
      } catch (err) {
        this.error = err.message || 'Failed to fetch data'
        return null
      }
    },

    /**
     * Clear all monitoring data
     */
    async clearData() {
      try {
        await ClearData()
        this.currentTPM = 0
        this.currentTPS = 0
        this.currentLatency = 0
        this.tpmHistory = []
        this.tpsHistory = []
        this.tpmStats = { avg: 0, max: 0, min: 0, stddev: 0, cv: 0, directionChanges: 0 }
        this.tpsStats = { avg: 0, max: 0, min: 0, stddev: 0, cv: 0, directionChanges: 0 }
        this.tpmStatus = 'stable'
        this.tpsStatus = 'stable'
      } catch (err) {
        this.error = err.message || 'Failed to clear data'
      }
    },

    // ==================== System Monitoring ==================

    /**
     * Start system monitoring
     */
    async startSystemMonitoring() {
      try {
        const result = await StartSystemMonitoring()
        if (result.success) {
          this.systemMonitoring = true
          this.subscribeToSystemEvents()
        } else {
          this.error = result.error || 'Failed to start system monitoring'
        }
        return result
      } catch (err) {
        this.error = err.message || 'Failed to start system monitoring'
        return { success: false, error: this.error }
      }
    },

    /**
     * Stop system monitoring
     */
    async stopSystemMonitoringAction() {
      try {
        const result = await StopSystemMonitoring()
        this.systemMonitoring = false
        this.unsubscribeFromSystemEvents()
        return result
      } catch (err) {
        this.error = err.message || 'Failed to stop system monitoring'
        return { success: false, error: this.error }
      }
    },

    /**
     * Subscribe to system monitoring events
     */
    subscribeToSystemEvents() {
      // System started event
      EventsOn('system:started', (data) => {
        console.log('System monitoring started:', data)
        this.systemMonitoring = true
      })

      // System stopped event
      EventsOn('system:stopped', (data) => {
        console.log('System monitoring stopped:', data)
        this.systemMonitoring = false
      })

      // System metrics update event
      EventsOn('system:metrics_update', (data) => {
        this.cpuPercent = data.cpu_percent || 0
        this.diskReadBps = data.disk_read_bps || 0
        this.diskWriteBps = data.disk_write_bps || 0
        this.diskUsedPercent = data.disk_used_percent || 0
        this.diskUsedGB = data.disk_used_gb || 0
        this.diskTotalGB = data.disk_total_gb || 0

        // Update CPU history
        this.cpuHistory.push({
          timestamp: Date.now(),
          value: data.cpu_percent
        })
        if (this.cpuHistory.length > 60) {
          this.cpuHistory = this.cpuHistory.slice(-60)
        }

        // Update Disk IO history
        this.diskIOHistory.push({
          timestamp: Date.now(),
          readBps: data.disk_read_bps,
          writeBps: data.disk_write_bps
        })
        if (this.diskIOHistory.length > 60) {
          this.diskIOHistory = this.diskIOHistory.slice(-60)
        }

        // Update Disk Space history
        this.diskSpaceHistory.push({
          timestamp: Date.now(),
          value: data.disk_used_percent
        })
        if (this.diskSpaceHistory.length > 60) {
          this.diskSpaceHistory = this.diskSpaceHistory.slice(-60)
        }
      })
    },

    /**
     * Unsubscribe from system monitoring events
     */
    unsubscribeFromSystemEvents() {
      EventsOff('system:started')
      EventsOff('system:stopped')
      EventsOff('system:metrics_update')
    },

    /**
     * Fetch current system metrics
     */
    async fetchSystemMetrics() {
      try {
        const data = await GetSystemMetrics()
        this.cpuPercent = data.cpu_percent || 0
        this.diskReadBps = data.disk_read_bps || 0
        this.diskWriteBps = data.disk_write_bps || 0
        this.diskUsedPercent = data.disk_used_percent || 0
        this.diskUsedGB = data.disk_used_gb || 0
        this.diskTotalGB = data.disk_total_gb || 0
        return data
      } catch (err) {
        this.error = err.message || 'Failed to fetch system metrics'
        return null
      }
    },

    /**
     * Check if system monitoring is running
     */
    async isSystemMonitoringRunning() {
      try {
        return await IsSystemMonitoring()
      } catch (err) {
        console.error('Failed to check system monitoring:', err)
        return false
      }
    },

    /**
     * Start system monitoring on app init
     * This should be called when the app starts
     */
    async initSystemMonitoring() {
      try {
        const result = await StartSystemMonitoring()
        if (result.success) {
          this.systemMonitoring = true
          this.subscribeToSystemEvents()
          console.log('System monitoring initialized')
        } else {
          console.error('Failed to initialize system monitoring:', result.error)
        }
        return result
      } catch (err) {
        console.error('Failed to initialize system monitoring:', err)
        return { success: false, error: err.message }
      }
    }
  }
})
