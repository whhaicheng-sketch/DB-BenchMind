<template>
  <div class="connections-tab">
    <!-- Page Header -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">Connections</h1>
        <p class="page-subtitle">Manage your database connections</p>
      </div>
      <div class="header-right">
        <div class="filter-group">
          <button
            v-for="filter in filters"
            :key="filter.value"
            :class="['filter-btn', { active: activeFilter === filter.value }]"
            @click="activeFilter = filter.value"
          >
            {{ filter.label }}
            <span v-if="filter.count > 0" class="filter-count">{{ filter.count }}</span>
          </button>
        </div>
        <button class="btn btn-sm" @click="handleTestAll">Test All</button>
        <button class="btn btn-primary" @click="showAddModal = true">
          + Add Connection
        </button>
      </div>
    </div>

    <!-- Connection List -->
    <div class="connection-list-container">
      <!-- Grouped by type -->
      <template v-for="group in filteredGroups" :key="group.type">
        <!-- Group Header -->
        <div class="group-header">
          <span class="group-icon">{{ group.icon }}</span>
          <span class="group-name">{{ group.label }}</span>
          <span class="group-count">({{ group.connections.length }})</span>
        </div>

        <!-- Connections -->
        <div v-if="group.connections.length > 0" class="connection-items">
          <div
            v-for="conn in group.connections"
            :key="conn.id"
            class="connection-row"
            :class="{ selected: selectedId === conn.id }"
            @click="selectConnection(conn)"
          >
            <!-- Left: Name + Info + SSH/WinRM flags -->
            <div class="conn-main">
              <div class="conn-name">{{ conn.name }}</div>
              <div class="conn-host">
                {{ conn.host }}:{{ conn.port }}
                <span v-if="isRemoteTypeSSH(conn)" class="tag tag-ssh">SSH</span>
                <span v-if="hasConfiguredAiAssistants(conn)" class="tag tag-ai" :title="getAiBadgeTooltip(conn)">AI</span>
                <span v-if="isRemoteTypeWinRM(conn)" class="tag tag-winrm">WinRM</span>
              </div>
            </div>

            <!-- Tags -->
            <div class="conn-tags">
              <span v-if="shouldShowDatabaseTag(conn)" class="tag tag-db">{{ conn.database }}</span>
            </div>

            <!-- Right: Actions -->
            <div class="conn-actions">
              <!-- Test Button -->
              <button
                class="btn-action btn-test"
                :class="{
                  'btn-testing': connectionStore.isTestingById(conn.id),
                  'btn-test-success': connectionStore.getTestResultById(conn.id)?.success,
                  'btn-test-error': connectionStore.getTestResultById(conn.id) && !connectionStore.getTestResultById(conn.id).success
                }"
                @click.stop="testConnection(conn)"
                :disabled="connectionStore.isTestingById(conn.id)"
              >
                <span v-if="connectionStore.isTestingById(conn.id)" class="spinner-sm"></span>
                <span v-else>Test</span>
              </button>
              <button class="btn-action" @click.stop="editConnection(conn)">Edit</button>
              <button class="btn-action btn-more" @click.stop="toggleMoreMenu(conn)">⋮</button>
              <!-- More Menu -->
              <div v-if="moreMenuId === conn.id" class="more-menu">
                <button @click.stop="cloneConnection(conn)" class="more-menu-item">Clone</button>
                <button class="menu-item menu-item-danger" @click.stop="deleteConnection(conn)">Delete</button>
              </div>
              <!-- Test Result -->
              <div class="conn-test-result">
                <span v-if="connectionStore.getTestResultById(conn.id)"
                      :class="getDbTestStatusClass(conn)">
                  DB: {{ getDbTestStatusText(conn) }}
                </span>
                <span v-if="isRemoteTypeSSH(conn) && connectionStore.getSSHTestResultById(conn.id)"
                      class="result-spacing"
                      :class="getSshTestStatusClass(conn)">
                  SSH: {{ getSshTestStatusText(conn) }}
                </span>
                <span v-if="isRemoteTypeWinRM(conn) && connectionStore.getWinRMTestResultById(conn.id)"
                      class="result-spacing"
                      :class="getWinrmTestStatusClass(conn)">
                  WinRM: {{ getWinrmTestStatusText(conn) }}
                </span>
                <span v-if="hasConfiguredAiAssistants(conn) && connectionStore.getAITestResultById(conn.id)"
                      class="result-spacing"
                      :class="getAiTestStatusClass(conn)">
                  AI: {{ getAiTestStatusText(conn) }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Empty State -->
        <div v-else class="empty-state-inline">
          <span class="empty-text">No {{ group.label }} connections</span>
          <button class="btn-link" @click="addConnectionOfType(group.type)">
            Add {{ group.label }} Connection
          </button>
        </div>
      </template>

      <!-- No connections at all -->
      <div v-if="totalConnections === 0" class="empty-state">
        <div class="empty-state-icon">🔌</div>
        <p class="empty-state-title">No database connections</p>
        <p class="empty-state-description">Create your first database connection to get started.</p>
        <button class="btn btn-primary" @click="showAddModal = true">
          + Add Connection
        </button>
      </div>
    </div>

    <!-- Add/Edit Modal -->
    <div v-if="showAddModal || showEditModal" class="modal-overlay">
      <div class="modal-content">
        <ConnectionForm
          :mode="showEditModal ? 'edit' : 'create'"
          :connection-id="showEditModal ? selectedConnection?.id : null"
          :default-type="pendingType"
          @saved="handleSaved"
          @cancelled="closeModal"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, onActivated } from 'vue'
import { useConnectionStore } from '../../stores/connection'
import ConnectionForm from '../connection/ConnectionForm.vue'
import { getAiBadgeTooltip, hasConfiguredAiAssistants } from './connectionCardBadges'
import { getRemoteType, isRemoteTypeSSH, isRemoteTypeWinRM } from '../connection/connectionFormRemoteState.mjs'

const connectionStore = useConnectionStore()

const showAddModal = ref(false)
const showEditModal = ref(false)
const selectedId = ref('')
const selectedConnection = ref(null)
const activeFilter = ref('all')
const moreMenuId = ref(null)
const pendingType = ref('')

// Filter options
const filters = computed(() => {
  const groups = connectionGroups.value
  return [
    { label: 'All', value: 'all', count: connectionStore.connections.length },
    { label: 'MySQL', value: 'mysql', count: groups.find(g => g.type === 'mysql')?.connections.length || 0 },
    { label: 'PostgreSQL', value: 'postgresql', count: groups.find(g => g.type === 'postgresql')?.connections.length || 0 },
    { label: 'Oracle', value: 'oracle', count: groups.find(g => g.type === 'oracle')?.connections.length || 0 },
    { label: 'SQL Server', value: 'sqlserver', count: groups.find(g => g.type === 'sqlserver')?.connections.length || 0 }
  ]
})

const filteredGroups = computed(() => {
  const groups = connectionGroups.value
  if (activeFilter.value === 'all') {
    return groups.filter(g => g.connections.length > 0 || hasNearbyGroups(g.type))
  }
  return groups.filter(g => g.type === activeFilter.value)
})

const hasNearbyGroups = (type) => {
  const groups = connectionGroups.value
  const idx = groups.findIndex(g => g.type === type)
  if (idx > 0 && groups[idx - 1].connections.length > 0) return true
  if (idx < groups.length - 1 && groups[idx + 1].connections.length > 0) return true
  return false
}

const totalConnections = computed(() => connectionStore.connections.length)

const connectionGroups = computed(() => {
  const groups = [
    { type: 'mysql', label: 'MySQL', icon: '🐬', connections: [] },
    { type: 'postgresql', label: 'PostgreSQL', icon: '🐘', connections: [] },
    { type: 'oracle', label: 'Oracle', icon: '🔴', connections: [] },
    { type: 'sqlserver', label: 'SQL Server', icon: '🔷', connections: [] }
  ]

  connectionStore.connections.forEach(conn => {
    const group = groups.find(g => g.type === conn.type)
    if (group) {
      group.connections.push(conn)
    }
  })

  return groups
})

// Close more menu when clicking outside
const handleDocumentClick = () => {
  moreMenuId.value = null
}

onMounted(async () => {
  await connectionStore.fetchConnections()
  document.addEventListener('click', handleDocumentClick)
})

// Refresh connections when switching back to this tab
onActivated(async () => {
  await connectionStore.fetchConnections()
})

onUnmounted(() => {
  document.removeEventListener('click', handleDocumentClick)
})

const selectConnection = (conn) => {
  selectedId.value = conn.id
  selectedConnection.value = conn
}

const testConnection = async (conn) => {
  selectedId.value = conn.id
  selectedConnection.value = conn
  await connectionStore.testConnectionById(conn.id)
}

const editConnection = (conn) => {
  selectedConnection.value = conn
  showEditModal.value = true
}

const cloneConnection = async (conn) => {
  moreMenuId.value = null
  const result = await connectionStore.cloneConnection(conn.id)
  if (result) {
    selectedConnection.value = result
    showEditModal.value = true
  }
}

const deleteConnection = async (conn) => {
  if (confirm(`Delete connection "${conn.name}"?`)) {
    await connectionStore.deleteConnection(conn.id)
  }
  moreMenuId.value = null
}

const shouldShowDatabaseTag = (conn) => {
  if (!conn.database) return false
  if (conn.type === 'postgresql' && conn.database === 'postgres') return false
  if (conn.type === 'mysql' && !conn.database) return false
  if (conn.type === 'oracle' && (conn.database === 'orcl' || conn.database === 'XE')) return false
  if (conn.type === 'sqlserver' && !conn.database) return false
  return true
}

const toggleMoreMenu = (conn) => {
  moreMenuId.value = moreMenuId.value === conn.id ? null : conn.id
}

const addConnectionOfType = (type) => {
  pendingType.value = type
  showAddModal.value = true
}

const closeModal = () => {
  showAddModal.value = false
  showEditModal.value = false
  pendingType.value = ''
}

const handleSaved = () => {
  closeModal()
  connectionStore.fetchConnections()
}

const handleTestAll = async () => {
  const ids = connectionStore.connections.map(c => c.id)
  if (ids.length === 0) return
  await connectionStore.batchTestConnections(ids)
}

// Helper functions for test status display
const getDbTestStatusText = (conn) => {
  const result = connectionStore.getDBTestResultById(conn.id)
  if (!result) return ''
  return result.success ? 'OK' : 'FAIL'
}

const getDbTestStatusClass = (conn) => {
  const result = connectionStore.getDBTestResultById(conn.id)
  if (!result) return ''
  return result.success ? 'result-success' : 'result-error'
}

const getSshTestStatusText = (conn) => {
  const result = connectionStore.getSSHTestResultById(conn.id)
  if (!result) return ''
  return result.success ? 'OK' : 'FAIL'
}

const getSshTestStatusClass = (conn) => {
  const result = connectionStore.getSSHTestResultById(conn.id)
  if (!result) return ''
  return result.success ? 'result-success' : 'result-error'
}

const getWinrmTestStatusText = (conn) => {
  const result = connectionStore.getWinRMTestResultById(conn.id)
  if (!result) return ''
  return result.success ? 'OK' : 'FAIL'
}

const getWinrmTestStatusClass = (conn) => {
  const result = connectionStore.getWinRMTestResultById(conn.id)
  if (!result) return ''
  return result.success ? 'result-success' : 'result-error'
}

const getAiTestStatusText = (conn) => {
  const result = connectionStore.getAITestResultById(conn.id)
  if (!result) return ''
  return result.success ? 'OK' : 'FAIL'
}

const getAiTestStatusClass = (conn) => {
  const result = connectionStore.getAITestResultById(conn.id)
  if (!result) return ''
  return result.success ? 'result-success' : 'result-error'
}

const getConnectionRemoteType = (conn) => getRemoteType(conn)
</script>

<style scoped>
.connections-tab {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* Page Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding-bottom: var(--spacing-lg);
  border-bottom: 1px solid var(--border-light);
  margin-bottom: var(--spacing-lg);
  flex-wrap: wrap;
  gap: var(--spacing-md);
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  font-size: var(--font-size-title);
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.page-subtitle {
  font-size: var(--font-size-md);
  color: var(--text-muted);
  margin: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  flex-wrap: wrap;
}

/* Filter Group */
.filter-group {
  display: flex;
  gap: 2px;
  background-color: var(--bg-secondary);
  padding: 3px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
}

.filter-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 5px 10px;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
  cursor: pointer;
  transition: none;
}

.filter-btn:hover {
  background-color: var(--bg-tertiary);
  color: var(--text-primary);
}

.filter-btn.active {
  background-color: var(--bg-primary);
  color: var(--primary);
  box-shadow: var(--shadow-sm);
}

.filter-count {
  background-color: var(--bg-tertiary);
  padding: 1px 5px;
  border-radius: 10px;
  font-size: var(--font-size-xs);
  color: var(--text-muted);
}

.filter-btn.active .filter-count {
  background-color: var(--primary-light);
  color: var(--primary);
}

/* Connection List Container */
.connection-list-container {
  flex: 1;
  overflow-y: auto;
}

/* Group Header */
.group-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) 0;
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
  font-weight: 500;
  border-bottom: 1px solid var(--border-light);
  margin-bottom: var(--spacing-sm);
}

.group-icon {
  font-size: var(--font-size-base);
}

.group-name {
  color: var(--text-primary);
}

.group-count {
  color: var(--text-muted);
  font-weight: 400;
}

/* Connection Items */
.connection-items {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-bottom: var(--spacing-lg);
}

.connection-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: 10px var(--spacing-md);
  background-color: var(--bg-primary);
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: background-color var(--transition-fast);
}

.connection-row:hover {
  background-color: var(--bg-hover);
}

.connection-row.selected {
  background-color: var(--bg-selected);
  border-color: var(--primary);
}

.conn-main {
  flex: 1;
  min-width: 0;
}

.conn-name {
  font-weight: 500;
  color: var(--text-primary);
  font-size: var(--font-size-base);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.conn-host {
  font-size: var(--font-size-sm);
  color: var(--text-muted);
  font-family: var(--font-family-mono);
}

/* Tags */
.conn-tags {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.tag {
  font-size: var(--font-size-xs);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-weight: 500;
}

.tag-ssh {
  background-color: var(--primary-light);
  color: var(--primary);
}

.tag-ai {
  background-color: var(--success-bg);
  color: var(--success);
}

.tag-winrm {
  background-color: var(--warning-bg);
  color: var(--warning);
}

.tag-db {
  background-color: var(--bg-secondary);
  color: var(--text-secondary);
  border: 1px solid var(--border-light);
}

/* Actions */
.conn-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  position: relative;
}

.btn-action {
  padding: 4px 10px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
  background-color: var(--bg-primary);
  color: var(--text-secondary);
}

.btn-action:hover {
  background-color: var(--bg-secondary);
  border-color: var(--border-dark);
  color: var(--text-primary);
}

.btn-action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-test {
  color: var(--success);
}

.btn-test:hover:not(:disabled) {
  background-color: var(--success-bg);
  border-color: var(--success);
  color: var(--success);
}

.btn-more {
  padding: 4px 8px;
  font-weight: bold;
}

/* Test button states */
.btn-testing {
  color: var(--primary);
  display: flex;
  align-items: center;
  gap: 4px;
}

.spinner-sm {
  width: 12px;
  height: 12px;
  border: 2px solid var(--border-color);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.btn-test-success {
  color: var(--success);
  border-color: var(--success);
}

.btn-test-error {
  color: var(--danger);
  border-color: var(--danger);
}

/* Test result display */
.conn-test-result {
  font-size: var(--font-size-xs);
  padding: 3px 8px;
  border-radius: var(--radius-sm);
  margin-left: var(--spacing-sm);
  max-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.result-success {
  background-color: var(--success-bg);
  color: var(--success);
}

.result-error {
  background-color: var(--danger-bg);
  color: var(--danger);
}

.result-spacing {
  margin-left: var(--spacing-sm);
}

/* More Menu */
.more-menu {
  position: absolute;
  top: 100%;
  right: 0;
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 4px;
  z-index: 100;
  box-shadow: var(--shadow-md);
  min-width: 100px;
}

.menu-item {
  display: block;
  width: 100%;
  padding: 6px 12px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
  text-align: left;
  cursor: pointer;
  border-radius: var(--radius-sm);
}

.menu-item:hover {
  background-color: var(--bg-secondary);
  color: var(--text-primary);
}

.menu-item-danger {
  color: var(--danger);
}

.menu-item-danger:hover {
  background-color: var(--danger-bg);
  color: var(--danger);
}

/* Empty States */
.empty-state-inline {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: 8px var(--spacing-md);
  background-color: var(--bg-secondary);
  border-radius: var(--radius-md);
  margin-bottom: var(--spacing-sm);
}

.empty-text {
  font-size: var(--font-size-sm);
  color: var(--text-muted);
  flex: 1;
}

.btn-link {
  background: none;
  border: none;
  color: var(--primary);
  font-size: var(--font-size-sm);
  cursor: pointer;
  padding: 4px 8px;
}

.btn-link:hover {
  text-decoration: underline;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: var(--text-muted);
  text-align: center;
}

.empty-state-icon {
  font-size: 48px;
  margin-bottom: var(--spacing-lg);
  opacity: 0.5;
}

.empty-state-title {
  font-size: var(--font-size-lg);
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: var(--spacing-sm);
}

.empty-state-description {
  font-size: var(--font-size-md);
  color: var(--text-muted);
  margin-bottom: var(--spacing-lg);
  max-width: 400px;
}

/* Modal */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background-color: var(--bg-primary);
  border-radius: var(--radius-lg);
  max-width: 800px;
  width: 90%;
  max-height: 90vh;
  overflow: visible;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--border-color);
  box-shadow: var(--shadow-modal);
}

.btn-sm {
  padding: 6px 10px;
  font-size: 12px;
}

.conn-checkbox {
  width: 16px;
  height: 16px;
  accent-color: var(--primary);
  flex-shrink: 0;
}

.btn-primary {
  background: var(--primary);
  color: white;
  border-color: var(--primary);
}

.more-menu-item {
  display: block;
  width: 100%;
  padding: 6px 12px;
  background: none;
  border: none;
  text-align: left;
  color: var(--text-primary);
  cursor: pointer;
  font-size: 13px;
}

.more-menu-item:hover {
  background: var(--bg-hover);
}
</style>
