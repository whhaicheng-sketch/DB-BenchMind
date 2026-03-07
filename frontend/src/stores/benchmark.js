/**
 * Benchmark Pinia Store
 * Manages benchmark execution state for DB-BenchMind Wails frontend.
 */
import { defineStore } from 'pinia'
import {
  StartBenchmark,
  PrepareOnly,
  RunBenchmark,
  CleanupOnly,
  StopBenchmark,
  GetBenchmarkStatus,
  ListBenchmarks
} from '../../wailsjs/go/bindings/BenchmarkBinding'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

export const useBenchmarkStore = defineStore('benchmark', {
  state: () => ({
    // Current running benchmark
    currentRun: null,
    currentRunId: null,
    // Running state
    isRunning: false,
    isPreparing: false,
    isCleaning: false,
    // Logs
    logLines: [],
    // Status polling
    statusPollingInterval: null,
    // History
    runHistory: [],
    // Loading states
    loading: false,
    error: null
  }),

  getters: {
    // Get current state
    currentState: (state) => {
      return state.currentRun?.state || 'idle'
    },

    // Check if can start
    canStart: (state) => {
      return !state.isRunning && !state.isPreparing && !state.isCleaning
    },

    // Check if can stop
    canStop: (state) => {
      return state.isRunning || state.isPreparing || state.isCleaning
    },

    // Get result if completed
    result: (state) => {
      return state.currentRun?.result || null
    },

    // State labels
    stateLabels: () => ({
      'pending': 'Pending',
      'preparing': 'Preparing',
      'prepared': 'Prepared',
      'warming_up': 'Warming Up',
      'running': 'Running',
      'completed': 'Completed',
      'failed': 'Failed',
      'cancelled': 'Cancelled',
      'timeout': 'Timeout',
      'force_stopped': 'Force Stopped',
      'idle': 'Idle'
    })
  },

  actions: {
    /**
     * Start a full benchmark (prepare + run + cleanup)
     */
    async startBenchmark(connectionId, templateId, parameters, options = {}) {
      this.loading = true
      this.error = null
      this.logLines = ['Starting benchmark...']

      try {
        const result = await StartBenchmark({
          connection_id: connectionId,
          template_id: templateId,
          parameters: parameters,
          options: {
            skip_prepare: options.skipPrepare || false,
            skip_cleanup: options.skipCleanup || false,
            warmup_time: options.warmupTime || 0,
            dry_run: options.dryRun || false
          }
        })

        if (result.error) {
          this.error = result.error
          this.logLines.push(`Error: ${result.error}`)
          return null
        }

        this.currentRunId = result.run_id
        this.isRunning = true
        this.logLines.push(`Benchmark started with ID: ${result.run_id}`)

        // Start status polling
        this.startStatusPolling()

        return result.run_id
      } catch (err) {
        this.error = err.message || 'Failed to start benchmark'
        this.logLines.push(`Error: ${this.error}`)
        return null
      } finally {
        this.loading = false
      }
    },

    /**
     * Run prepare phase only
     */
    async prepareOnly(connectionId, templateId, parameters, options = {}) {
      this.loading = true
      this.error = null
      this.isPreparing = true
      this.logLines = ['Preparing data...']

      try {
        const result = await PrepareOnly({
          connection_id: connectionId,
          template_id: templateId,
          parameters: parameters,
          options: {
            skip_prepare: false,
            skip_cleanup: true,
            warmup_time: 0,
            dry_run: options.dryRun || false
          }
        })

        if (result.error) {
          this.error = result.error
          this.logLines.push(`Error: ${result.error}`)
          this.isPreparing = false
          return null
        }

        this.currentRunId = result.run_id
        this.logLines.push(`Prepare started with ID: ${result.run_id}`)

        // Start status polling
        this.startStatusPolling()

        return result.run_id
      } catch (err) {
        this.error = err.message || 'Failed to prepare'
        this.logLines.push(`Error: ${this.error}`)
        this.isPreparing = false
        return null
      } finally {
        this.loading = false
      }
    },

    /**
     * Run benchmark only (skip prepare/cleanup)
     */
    async runBenchmark(connectionId, templateId, parameters, options = {}) {
      this.loading = true
      this.error = null
      this.isRunning = true
      this.logLines = ['Running benchmark...']

      try {
        const result = await RunBenchmark({
          connection_id: connectionId,
          template_id: templateId,
          parameters: parameters,
          options: {
            skip_prepare: true,
            skip_cleanup: true,
            warmup_time: options.warmupTime || 0,
            dry_run: options.dryRun || false
          }
        })

        if (result.error) {
          this.error = result.error
          this.logLines.push(`Error: ${result.error}`)
          this.isRunning = false
          return null
        }

        this.currentRunId = result.run_id
        this.logLines.push(`Benchmark started with ID: ${result.run_id}`)

        // Start status polling
        this.startStatusPolling()

        return result.run_id
      } catch (err) {
        this.error = err.message || 'Failed to run benchmark'
        this.logLines.push(`Error: ${this.error}`)
        this.isRunning = false
        return null
      } finally {
        this.loading = false
      }
    },

    /**
     * Run cleanup phase only
     */
    async cleanupOnly(connectionId, templateId, parameters, options = {}) {
      this.loading = true
      this.error = null
      this.isCleaning = true
      this.logLines = ['Cleaning up...']

      try {
        const result = await CleanupOnly({
          connection_id: connectionId,
          template_id: templateId,
          parameters: parameters,
          options: {
            skip_prepare: true,
            skip_cleanup: false,
            warmup_time: 0,
            dry_run: options.dryRun || false
          }
        })

        if (result.error) {
          this.error = result.error
          this.logLines.push(`Error: ${result.error}`)
          this.isCleaning = false
          return null
        }

        this.currentRunId = result.run_id
        this.logLines.push(`Cleanup started with ID: ${result.run_id}`)

        // Start status polling
        this.startStatusPolling()

        return result.run_id
      } catch (err) {
        this.error = err.message || 'Failed to cleanup'
        this.logLines.push(`Error: ${this.error}`)
        this.isCleaning = false
        return null
      } finally {
        this.loading = false
      }
    },

    /**
     * Stop running benchmark
     */
    async stopBenchmark(force = false) {
      if (!this.currentRunId) return false

      this.logLines.push(force ? 'Force stopping...' : 'Stopping...')

      try {
        const result = await StopBenchmark(this.currentRunId, force)
        if (result.success) {
          this.logLines.push('Benchmark stopped')
          this.stopStatusPolling()
          this.isRunning = false
          this.isPreparing = false
          this.isCleaning = false
          return true
        } else {
          this.error = result.error
          this.logLines.push(`Error: ${result.error}`)
          return false
        }
      } catch (err) {
        this.error = err.message || 'Failed to stop benchmark'
        this.logLines.push(`Error: ${this.error}`)
        return false
      }
    },

    /**
     * Get benchmark status
     */
    async fetchStatus() {
      if (!this.currentRunId) return null

      try {
        const result = await GetBenchmarkStatus(this.currentRunId)
        if (result.error) {
          console.error('Failed to fetch status:', result.error)
          return null
        }

        this.currentRun = result.run

        // Check if completed
        if (result.run) {
          const terminalStates = ['completed', 'failed', 'cancelled', 'timeout', 'force_stopped']
          if (terminalStates.includes(result.run.state)) {
            this.onBenchmarkComplete()
          }
        }

        return result.run
      } catch (err) {
        console.error('fetchStatus error:', err)
        return null
      }
    },

    /**
     * Start polling for status updates
     */
    startStatusPolling() {
      this.stopStatusPolling() // Clear any existing interval

      this.statusPollingInterval = setInterval(async () => {
        await this.fetchStatus()
      }, 1000) // Poll every second
    },

    /**
     * Stop polling
     */
    stopStatusPolling() {
      if (this.statusPollingInterval) {
        clearInterval(this.statusPollingInterval)
        this.statusPollingInterval = null
      }
    },

    /**
     * Handle benchmark completion
     */
    onBenchmarkComplete() {
      this.stopStatusPolling()
      this.isRunning = false
      this.isPreparing = false
      this.isCleaning = false

      if (this.currentRun) {
        if (this.currentRun.state === 'completed') {
          this.logLines.push('✓ Benchmark completed successfully')
          if (this.currentRun.result) {
            this.logLines.push(`TPS: ${this.currentRun.result.tps?.toFixed(2) || 'N/A'}`)
            this.logLines.push(`TPM: ${this.currentRun.result.tpm?.toFixed(2) || 'N/A'}`)
            this.logLines.push(`Latency Avg: ${this.currentRun.result.latency_avg_ms?.toFixed(2) || 'N/A'} ms`)
          }
        } else if (this.currentRun.state === 'failed') {
          this.logLines.push(`✗ Benchmark failed: ${this.currentRun.error_message || 'Unknown error'}`)
        } else {
          this.logLines.push(`Benchmark ended with state: ${this.currentRun.state}`)
        }
      }
    },

    /**
     * Fetch run history
     */
    async fetchHistory(limit = 20) {
      this.loading = true

      try {
        const result = await ListBenchmarks(limit)
        if (result.error) {
          this.error = result.error
          return []
        }

        this.runHistory = result.runs || []
        return this.runHistory
      } catch (err) {
        this.error = err.message || 'Failed to fetch history'
        return []
      } finally {
        this.loading = false
      }
    },

    /**
     * Add log line
     */
    addLog(line) {
      this.logLines.push(line)
      // Keep only last 500 lines
      if (this.logLines.length > 500) {
        this.logLines = this.logLines.slice(-500)
      }
    },

    /**
     * Clear logs
     */
    clearLogs() {
      this.logLines = []
    },

    /**
     * Reset store state
     */
    reset() {
      this.stopStatusPolling()
      this.currentRun = null
      this.currentRunId = null
      this.isRunning = false
      this.isPreparing = false
      this.isCleaning = false
      this.logLines = []
      this.error = null
    },

    /**
     * Clear error
     */
    clearError() {
      this.error = null
    },

    /**
     * Subscribe to Wails events
     */
    subscribeToEvents() {
      // Log events
      EventsOn('log:append', (data) => {
        this.addLog(data.content)
      })

      // Status events
      EventsOn('benchmark:status', (data) => {
        if (this.currentRunId === data.run_id) {
          // Status updates are handled by polling
        }
      })

      // Metric events
      EventsOn('benchmark:metric', (data) => {
        if (this.currentRunId === data.run_id) {
          // Update current run with metrics
          if (this.currentRun && this.currentRun.result) {
            this.currentRun.result.tps = data.tps || this.currentRun.result.tps
            this.currentRun.result.tpm = data.tpm || this.currentRun.result.tpm
            this.currentRun.result.latency_avg_ms = data.latency_avg || this.currentRun.result.latency_avg_ms
            this.currentRun.result.latency_p95_ms = data.latency_p95 || this.currentRun.result.latency_p95_ms
            this.currentRun.result.latency_p99_ms = data.latency_p99 || this.currentRun.result.latency_p99_ms
          }
        }
      })

      // Progress events
      EventsOn('benchmark:progress', (data) => {
        if (this.currentRunId === data.run_id) {
          this.addLog(`Progress: ${data.percentage.toFixed(1)}% (${data.phase || 'running'})`)
        }
      })
    },

    /**
     * Unsubscribe from Wails events
     */
    unsubscribeFromEvents() {
      EventsOff('log:append')
      EventsOff('benchmark:status')
      EventsOff('benchmark:metric')
      EventsOff('benchmark:progress')
    }
  }
})
