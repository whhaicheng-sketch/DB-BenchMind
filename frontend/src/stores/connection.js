/**
 * Connection Pinia Store
 * Manages database connection state for DB-BenchMind Wails frontend.
 * Supports SSH Tunnel and WinRM configuration.
 */
import { defineStore } from 'pinia'
import {
  ListConnections,
  GetConnection,
  CreateConnection,
  UpdateConnection,
  DeleteConnection,
  TestConnection,
  TestSSHConnection,
  TestWinRMConnection
} from '../../wailsjs/go/bindings/ConnectionBinding'

export const useConnectionStore = defineStore('connection', {
  state: () => ({
    // Connection list
    connections: [],
    // Currently selected connection ID
    selectedConnectionId: null,
    // Loading state
    loading: false,
    // Error message
    error: null,
    // Test result
    testResult: null,
    // SSH test result
    sshTestResult: null,
    // WinRM test result
    winrmTestResult: null,
    // Per-connection test states (for list page feedback)
    testingById: {},
    testResultById: {},
    sshTestResultById: {}
  }),

  getters: {
    // Get selected connection object
    selectedConnection: (state) => {
      if (!state.selectedConnectionId) return null
      return state.connections.find(c => c.id === state.selectedConnectionId) || null
    },

    // Get testing state for a specific connection
    isTestingById: (state) => (id) => {
      return !!state.testingById[id]
    },
    // Get test result for a specific connection
    getTestResultById: (state) => (id) => {
      return state.testResultById[id] || null
    },
    getSSHTestResultById: (state) => (id) => {
      return state.sshTestResultById[id] || null
    },
    // Get connections grouped by type
    connectionsByType: (state) => {
      const grouped = {}
      for (const conn of state.connections) {
        if (!grouped[conn.type]) {
          grouped[conn.type] = []
        }
        grouped[conn.type].push(conn)
      }
      return grouped
    },

    // Connection type labels
    typeLabels: () => ({
      mysql: 'MySQL',
      postgresql: 'PostgreSQL',
      oracle: 'Oracle',
      sqlserver: 'SQL Server'
    }),

    // Database icons
    typeIcons: () => ({
      mysql: '🐬',
      postgresql: '🐘',
      oracle: '🔴',
      sqlserver: '🔷'
    })
  },

  actions: {
    /**
     * Fetch all connections from backend
     */
    async fetchConnections() {
      this.loading = true
      this.error = null

      try {
        const result = await ListConnections()
        if (result.error) {
          this.error = result.error
          console.error('Failed to fetch connections:', result.error)
        } else {
          this.connections = result.connections || []
        }
      } catch (err) {
        this.error = err.message || 'Failed to fetch connections'
        console.error('fetchConnections error:', err)
      } finally {
        this.loading = false
      }
    },

    /**
     * Select a connection by ID
     */
    selectConnection(id) {
      this.selectedConnectionId = id
    },

    /**
     * Clear connection selection
     */
    clearSelection() {
      this.selectedConnectionId = null
    },

    /**
     * Create a new connection
     */
    async createConnection(connectionData) {
      this.loading = true
      this.error = null

      try {
        const newConn = await CreateConnection({
          name: connectionData.name,
          type: connectionData.type,
          host: connectionData.host,
          port: connectionData.port,
          database: connectionData.database || '',
          username: connectionData.username,
          password: connectionData.password || '',
          // SSH Configuration
          ssh_enabled: connectionData.ssh_enabled || false,
          ssh_port: connectionData.ssh_port || 22,
          ssh_username: connectionData.ssh_username || '',
          ssh_password: connectionData.ssh_password || '',
          // WinRM Configuration
          winrm_enabled: connectionData.winrm_enabled || false,
          winrm_port: connectionData.winrm_port || 5985,
          winrm_use_https: connectionData.winrm_use_https || false,
          winrm_username: connectionData.winrm_username || '',
          winrm_password: connectionData.winrm_password || '',
          // SQL Server specific
          trust_server_certificate: connectionData.trust_server_certificate ?? true,
          // Oracle specific fields
          sid: connectionData.sid || '',
          service_name: connectionData.service_name || ''
        })

        if (newConn) {
          // Refresh the list
          await this.fetchConnections()
          return newConn
        }
        return null
      } catch (err) {
        this.error = err.message || 'Failed to create connection'
        console.error('createConnection error:', err)
        return null
      } finally {
        this.loading = false
      }
    },

    /**
     * Update an existing connection
     */
    async updateConnection(connectionData) {
      this.loading = true
      this.error = null

      try {
        const updatedConn = await UpdateConnection({
          id: connectionData.id,
          name: connectionData.name,
          host: connectionData.host,
          port: connectionData.port,
          database: connectionData.database || '',
          username: connectionData.username,
          password: connectionData.password || '',
          // SSH Configuration
          ssh_enabled: connectionData.ssh_enabled || false,
          ssh_port: connectionData.ssh_port || 22,
          ssh_username: connectionData.ssh_username || '',
          ssh_password: connectionData.ssh_password || '',
          // WinRM Configuration
          winrm_enabled: connectionData.winrm_enabled || false,
          winrm_port: connectionData.winrm_port || 5985,
          winrm_use_https: connectionData.winrm_use_https || false,
          winrm_username: connectionData.winrm_username || '',
          winrm_password: connectionData.winrm_password || '',
          // SQL Server specific
          trust_server_certificate: connectionData.trust_server_certificate ?? true,
          // Oracle specific fields
          sid: connectionData.sid || '',
          service_name: connectionData.service_name || '',
          connect_type: connectionData.connect_type || ''
        })

        if (updatedConn) {
          // Refresh the list
          await this.fetchConnections()
          return updatedConn
        }
        return null
      } catch (err) {
        this.error = err.message || 'Failed to update connection'
        console.error('updateConnection error:', err)
        return null
      } finally {
        this.loading = false
      }
    },

    /**
     * Delete a connection by ID
     */
    async deleteConnection(id) {
      this.loading = true
      this.error = null

      try {
        const success = await DeleteConnection(id)
        if (success) {
          // Remove from local list
          this.connections = this.connections.filter(c => c.id !== id)
          // Clear selection if deleted connection was selected
          if (this.selectedConnectionId === id) {
            this.selectedConnectionId = null
          }
          return true
        }
        return false
      } catch (err) {
        this.error = err.message || 'Failed to delete connection'
        console.error('deleteConnection error:', err)
        return false
      } finally {
        this.loading = false
      }
    },

    /**
     * Test a connection by ID
     */
    async testConnectionById(id, skipTunnel = false) {
      this.loading = true
      this.error = null
      this.testResult = null
      // Set per-connection testing state
      this.testingById[id] = true
      this.testResultById[id] = null

      try {
        const result = await TestConnection(id, skipTunnel)
        this.testResult = result
        this.testResultById[id] = result
        return result
      } catch (err) {
        this.error = err.message || 'Failed to test connection'
        const errorResult = {
          success: false,
          error: this.error
        }
        this.testResult = errorResult
        this.testResultById[id] = errorResult
        return errorResult
      } finally {
        this.loading = false
        this.testingById[id] = false
      }
    },

    /**
     * Test SSH tunnel connection
     * @param {string|object} idOrConfig - Connection ID or SSH config (for backward compatibility)
     * @param {object} sshConfigMaybe - SSH configuration (when first param is ID)
     */
    async testSSHConnection(idOrConfig, sshConfigMaybe) {
      // Support both signatures: (sshConfig) and (id, sshConfig)
      let id, sshConfig
      if (typeof idOrConfig === 'string') {
        id = idOrConfig
        sshConfig = sshConfigMaybe
      } else {
        sshConfig = idOrConfig
      }

      this.loading = true
      this.sshTestResult = null

      try {
        const result = await TestSSHConnection({
          host: sshConfig.host,
          port: sshConfig.port || 22,
          username: sshConfig.username,
          password: sshConfig.password
        })
        this.sshTestResult = result
        // Store result by connection ID if provided
        if (id) {
          this.sshTestResultById[id] = result
        }
        return result
      } catch (err) {
        const errorResult = {
          success: false,
          host: sshConfig.host,
          error: err.message || 'Failed to test SSH connection'
        }
        this.sshTestResult = errorResult
        if (id) {
          this.sshTestResultById[id] = errorResult
        }
        return errorResult
      } finally {
        this.loading = false
      }
    },
    /**
     * Test WinRM connection
     */
    async testWinRMConnection(winrmConfig) {
      this.loading = true
      this.winrmTestResult = null

      try {
        const result = await TestWinRMConnection({
          host: winrmConfig.host,
          port: winrmConfig.port || (winrmConfig.use_https ? 5986 : 5985),
          username: winrmConfig.username,
          password: winrmConfig.password,
          use_https: winrmConfig.use_https || false
        })
        this.winrmTestResult = result
        return result
      } catch (err) {
        this.winrmTestResult = {
          success: false,
          host: winrmConfig.host,
          error: err.message || 'Failed to test WinRM connection'
        }
        return this.winrmTestResult
      } finally {
        this.loading = false
      }
    },

    /**
     * Clear test result
     */
    clearTestResult() {
      this.testResult = null
      this.sshTestResult = null
      this.winrmTestResult = null
    },

    /**
     * Clear error
     */
    clearError() {
      this.error = null
    }
  }
})
