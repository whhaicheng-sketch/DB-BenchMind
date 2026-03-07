/**
 * Connection Pinia Store
 * Manages database connection state for DB-BenchMind Wails frontend.
 */
import { defineStore } from 'pinia'
import {
  ListConnections,
  GetConnection,
  CreateConnection,
  UpdateConnection,
  DeleteConnection,
  TestConnection
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
    testResult: null
  }),

  getters: {
    // Get selected connection object
    selectedConnection: (state) => {
      if (!state.selectedConnectionId) return null
      return state.connections.find(c => c.id === state.selectedConnectionId) || null
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
          ssl_mode: connectionData.ssl_mode || ''
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
          ssl_mode: connectionData.ssl_mode || ''
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
    async testConnectionById(id) {
      this.loading = true
      this.error = null
      this.testResult = null

      try {
        const result = await TestConnection(id)
        this.testResult = result
        return result
      } catch (err) {
        this.error = err.message || 'Failed to test connection'
        this.testResult = {
          success: false,
          error: this.error
        }
        return this.testResult
      } finally {
        this.loading = false
      }
    },

    /**
     * Clear test result
     */
    clearTestResult() {
      this.testResult = null
    },

    /**
     * Clear error
     */
    clearError() {
      this.error = null
    }
  }
})
