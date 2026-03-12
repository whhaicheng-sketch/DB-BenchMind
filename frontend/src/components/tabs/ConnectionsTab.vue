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
        <button class="btn btn-primary btn-add" @click="showAddModal = true">
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
            @click="selectConnection(conn)"
          >
            <!-- Left: Name + Info + SSH/WinRM 标志(内联显示) -->
            <div class="conn-main">
              <div class="conn-name">{{ conn.name }}</div>
              <div class="conn-host">
                {{ conn.host }}:{{ conn.port }}
                <span v-if="conn.ssh_enabled" class="tag tag-ssh tag-inline">SSH</span>
                <span v-if="conn.winrm_enabled" class="tag tag-winrm tag-inline">WinRM</span>
              </div>
            </div>

            <!-- Tags(只显示非默认数据库名) -->
            <div class="conn-tags">
              <span v-if="shouldShowDatabaseTag(conn)" class="tag tag-db">{{ conn.database }}</span>
            </div>

            <!-- Right: Actions -->
            <div class="conn-actions">
              <!-- Test Button with Status -->
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
                <span v-if="connectionStore.isTestingById(conn.id)">⟳ Testing...</span>
                <span v-else>Test</span>
              </button>
              <button class="btn-action btn-edit" @click.stop="editConnection(conn)">Edit</button>
              <button class="btn-action btn-more" @click.stop="toggleMoreMenu(conn)">
                ⋮
              </button>
              <!-- More Menu -->
              <div v-if="moreMenuId === conn.id" class="more-menu">
                <button class="menu-item danger" @click.stop="deleteConnection(conn)">Delete</button>
              </div>
              <!-- Test Result Message -->
              <div v-if="connectionStore.getTestResultById(conn.id) || (conn.ssh_enabled && connectionStore.getSSHTestResultById(conn.id)) || (conn.winrm_enabled && connectionStore.getWinRMTestResultById(conn.id))"
                   class="conn-test-result">
                <!-- Database 测试结果 -->
                <span :class="connectionStore.getTestResultById(conn.id)?.success ? 'conn-test-result--success' : 'conn-test-result--error'">
                  DB: {{ connectionStore.getTestResultById(conn.id)?.success ? 'OK' : 'FAIL' }}
                </span>
                <!-- SSH 测试结果 (如果有SSH配置) -->
                <span v-if="conn.ssh_enabled && connectionStore.getSSHTestResultById(conn.id)" style="margin-left: 10px;"
                      :class="connectionStore.getSSHTestResultById(conn.id).success ? 'conn-test-result--success' : 'conn-test-result--error'">
                  SSH: {{ connectionStore.getSSHTestResultById(conn.id).success ? 'OK' : 'FAIL' }}
                </span>
                <!-- WinRM 测试结果 (如果有WinRM配置) -->
                <span v-if="conn.winrm_enabled && connectionStore.getWinRMTestResultById(conn.id)" style="margin-left: 10px;"
                      :class="connectionStore.getWinRMTestResultById(conn.id).success ? 'conn-test-result--success' : 'conn-test-result--error'">
                  WinRM: {{ connectionStore.getWinRMTestResultById(conn.id).success ? 'OK' : 'FAIL' }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Empty State -->
        <div v-else class="empty-state">
          <span class="empty-text">No {{ group.label }} connections</span>
          <button class="btn-link" @click="addConnectionOfType(group.type)">
            Add {{ group.label }} Connection
          </button>
        </div>
      </template>

      <!-- No connections at all -->
      <div v-if="totalConnections === 0" class="no-connections">
        <div class="no-connections-icon">🔌</div>
        <p>No database connections configured</p>
        <button class="btn btn-primary" @click="showAddModal = true">
          + Add Your First Connection
        </button>
      </div>
    </div>

    <!-- Add/Edit Modal -->
    <div v-if="showAddModal || showEditModal" class="modal-overlay" @click.self="closeModal">
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
import { ref, computed, onMounted } from 'vue'
import { useConnectionStore } from '../../stores/connection'
import ConnectionForm from '../connection/ConnectionForm.vue'

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
    { label: '🐬 MySQL', value: 'mysql', count: groups.find(g => g.type === 'mysql')?.connections.length || 0 },
    { label: '🐘 PostgreSQL', value: 'postgresql', count: groups.find(g => g.type === 'postgresql')?.connections.length || 0 },
    { label: '🔴 Oracle', value: 'oracle', count: groups.find(g => g.type === 'oracle')?.connections.length || 0 },
    { label: '🔷 SQL Server', value: 'sqlserver', count: groups.find(g => g.type === 'sqlserver')?.connections.length || 0 }
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
onMounted(async () => {
  await connectionStore.fetchConnections()
  document.addEventListener('click', () => {
    moreMenuId.value = null
  })
})

const selectConnection = (conn) => {
  selectedId.value = conn.id
  selectedConnection.value = conn
}

const testConnection = async (conn) => {
  selectedId.value = conn.id
  selectedConnection.value = conn

  // 运行数据库连接测试
  await connectionStore.testConnectionById(conn.id)

  // 如果启用了 SSH，同时运行 SSH 测试
  if (conn.ssh_enabled) {
    await connectionStore.testSSHConnection(conn.id, {
      host: conn.host,
      port: conn.ssh_port || 22,
      username: conn.ssh_username,
      password: conn.ssh_password || ''
    })
  }

  // 如果启用了 WinRM，同时运行 WinRM 测试
  if (conn.winrm_enabled) {
    await connectionStore.testWinRMConnection(conn.id, {
      host: conn.host,
      port: conn.winrm_port || 5985,
      username: conn.winrm_username,
      password: conn.winrm_password || '',
      use_https: conn.winrm_use_https || false
    })
  }
}

const editConnection = (conn) => {
  selectedConnection.value = conn
  showEditModal.value = true
}

const deleteConnection = async (conn) => {
  if (confirm(`Delete connection "${conn.name}"?`)) {
    await connectionStore.deleteConnection(conn.id)
  }
  moreMenuId.value = null
}

// 判断是否应该显示数据库名标签
// 只显示非默认数据库名（PostgreSQL 默认 "postgres" 不显示)
const shouldShowDatabaseTag = (conn) => {
  if (!conn.database) return false
  // PostgreSQL 默认数据库名是"postgres"， 不显示
  if (conn.type === 'postgresql' && conn.database === 'postgres') return false
  // 其他数据库类型： 如果数据库名是默认值， 也不显示
  // MySQL 默认为空
  if (conn.type === 'mysql' && !conn.database) return false
  // Oracle 默认 SID/Service Name 可能是 "orcl" 或 "XE"
  if (conn.type === 'oracle' && (conn.database === 'orcl' || conn.database === 'XE')) return false
  // SQL Server 默认为空
  if (conn.type === 'sqlserver' && !conn.database) return false
  // 其他情况显示
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
  padding-bottom: 20px;
  border-bottom: 1px solid #2d3748;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 16px;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #f7fafc;
  margin: 0;
}

.page-subtitle {
  font-size: 14px;
  color: #718096;
  margin: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  gap: 4px;
  background: #1a202c;
  padding: 4px;
  border-radius: 8px;
}

.filter-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: #a0aec0;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
}

.filter-btn:hover {
  background: #2d3748;
  color: #e2e8f0;
}

.filter-btn.active {
  background: #2d3748;
  color: #4299e1;
}

.filter-count {
  background: #4a5568;
  padding: 2px 6px;
  border-radius: 10px;
  font-size: 11px;
}

.filter-btn.active .filter-count {
  background: #4299e1;
  color: white;
}

/* Main Button */
.btn-add {
  font-weight: 600;
  padding: 10px 20px;
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
  gap: 8px;
  padding: 12px 0 8px;
  color: #a0aec0;
  font-size: 13px;
  font-weight: 500;
  border-bottom: 1px solid #2d3748;
  margin-bottom: 8px;
}

.group-icon {
  font-size: 14px;
}

.group-name {
  color: #e2e8f0;
}

.group-count {
  color: #718096;
  font-weight: 400;
}

/* Connection Items */
.connection-items {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-bottom: 16px;
}

.connection-row {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 16px;
  background: #1a202c;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.15s;
}

.connection-row:hover {
  background: #242d3c;
}

.conn-main {
  flex: 1;
  min-width: 0;
}

.conn-name {
  font-weight: 500;
  color: #e2e8f0;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.conn-host {
  font-size: 12px;
  color: #718096;
  font-family: monospace;
}

/* Tags */
.conn-tags {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.tag {
  font-size: 10px;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.tag-ssh {
  background: rgba(66, 153, 225, 0.2);
  color: #63b3ed;
}

.tag-winrm {
  background: rgba(237, 137, 54, 0.2);
  color: #ed8936;
}

.tag-ssl {
  background: rgba(72, 187, 120, 0.2);
  color: #68d391;
}

.tag-db {
  background: rgba(160, 174, 192, 0.2);
  color: #a0aec0;
}

/* Actions */
.conn-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  position: relative;
}

.btn-action {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
  background: #2d3748;
  color: #a0aec0;
}

.btn-action:hover {
  background: #4a5568;
  color: #e2e8f0;
}

.btn-test {
  color: #68d391;
}

.btn-test:hover {
  background: rgba(72, 187, 120, 0.2);
}

.btn-edit {
  color: #63b3ed;
}

.btn-edit:hover {
  background: rgba(66, 153, 225, 0.2);
}

.btn-more {
  padding: 6px 10px;
  font-size: 16px;
  font-weight: bold;
}

/* Test button states */
.btn-testing {
  color: #63b3ed;
  animation: pulse 1s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.btn-test-success {
  color: #68d391;
}

.btn-test-error {
  color: #fc8181;
}

/* Test result display */
.conn-test-result {
  font-size: 11px;
  padding: 4px 8px;
  border-radius: 4px;
  margin-left: 8px;
  max-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.conn-test-result--success {
  background: rgba(72, 187, 120, 0.2);
  color: #68d391;
}

.conn-test-result--error {
  background: rgba(229, 62, 62, 0.2);
  color: #fc8181;
}

/* More Menu */
.more-menu {
  position: absolute;
  top: 100%;
  right: 0;
  background: #1a202c;
  border: 1px solid #2d3748;
  border-radius: 6px;
  padding: 4px;
  z-index: 100;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.menu-item {
  display: block;
  width: 100%;
  padding: 8px 16px;
  border: none;
  background: transparent;
  color: #a0aec0;
  font-size: 13px;
  text-align: left;
  cursor: pointer;
  border-radius: 4px;
}

.menu-item:hover {
  background: #2d3748;
}

.menu-item.danger {
  color: #fc8181;
}

.menu-item.danger:hover {
  background: rgba(229, 62, 62, 0.2);
}

/* Empty State */
.empty-state {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  background: #1a202c;
  border-radius: 6px;
  margin-bottom: 8px;
}

.empty-text {
  font-size: 13px;
  color: #718096;
  flex: 1;
}

.btn-link {
  background: none;
  border: none;
  color: #4299e1;
  font-size: 12px;
  cursor: pointer;
  padding: 4px 8px;
}

.btn-link:hover {
  text-decoration: underline;
}

/* No Connections */
.no-connections {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #718096;
}

.no-connections-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.no-connections p {
  margin-bottom: 20px;
}

/* Buttons */
.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
}

.btn-primary {
  background-color: #4299e1;
  color: white;
}

.btn-primary:hover {
  background-color: #3182ce;
}

/* Modal */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background-color: #1a202c;
  border-radius: 8px;
  max-width: 820px;
  width: 90%;
  max-height: 90vh;
  overflow: visible;
  display: flex;
  flex-direction: column;
  border: 1px solid #2d3748;
}
</style>
