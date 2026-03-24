<script setup>
/**
 * App.vue
 * Main application component with Tab navigation.
 * Navicat-style light desktop tool theme.
 */
import { ref, onMounted, onUnmounted, onErrorCaptured } from 'vue'
import { useMonitorStore } from './stores/monitor'
import { useAppStore } from './stores/app'
import { navigationTabs } from './constants/navigationTabs.mjs'

// Tab components
import ConnectionsTab from './components/tabs/ConnectionsTab.vue'
import TemplatesTab from './components/tabs/TemplatesTab.vue'
import TasksMonitorTab from './components/tabs/TasksMonitorTab.vue'
import AutoBenchTab from './components/tabs/AutoBenchTab.vue'
import HistoryTab from './components/tabs/HistoryTab.vue'

// Global error state
const globalError = ref(null)
const showError = ref(false)

// Monitor store
const monitorStore = useMonitorStore()
const appStore = useAppStore()

// Error boundary
onErrorCaptured((error, instance, info) => {
  console.error('Global error captured:', error)
  globalError.value = {
    message: error.message || 'An unexpected error occurred',
    component: info?.type?.name || 'Unknown',
    stack: error.stack
  }
  showError.value = true
  return false
})

// Dismiss global error
const dismissError = () => {
  showError.value = false
  globalError.value = null
}

// Lifecycle hooks
onMounted(async () => {
  console.log('DB-BenchMind Wails App mounted')
  try {
    await monitorStore.initSystemMonitoring()
  } catch (err) {
    console.error('Failed to initialize system monitoring:', err)
  }
})

onUnmounted(async () => {
  console.log('DB-BenchMind Wails App unmounted')
  try {
    await monitorStore.stopSystemMonitoringAction()
  } catch (err) {
    console.error('Failed to stop system monitoring:', err)
  }
})
</script>

<template>
  <div class="app-container">
    <!-- Global error modal -->
    <div v-if="showError" class="error-overlay" @click.self="dismissError">
      <div class="error-modal">
        <div class="error-header">
          <span class="error-icon">⚠️</span>
          <span class="error-title">Error</span>
        </div>
        <div class="error-body">
          <p class="error-message">{{ globalError?.message }}</p>
          <details v-if="globalError?.stack" class="error-details">
            <summary>Stack Trace</summary>
            <pre>{{ globalError.stack }}</pre>
          </details>
        </div>
        <div class="error-footer">
          <button class="btn btn-primary" @click="dismissError">Dismiss</button>
        </div>
      </div>
    </div>

    <!-- Tab Navigation -->
    <div class="tab-nav">
      <div class="tab-list">
        <button
          v-for="tab in navigationTabs"
          :key="tab.id"
          :class="['tab-item', { active: appStore.activeTab === tab.id }]"
          @click="appStore.setActiveTab(tab.id)"
        >
          <span class="tab-icon">{{ tab.icon }}</span>
          <span class="tab-label">{{ tab.label }}</span>
        </button>
      </div>
    </div>

    <!-- Tab Content -->
    <div class="tab-content">
      <ConnectionsTab v-if="appStore.activeTab === 'connections'" />
      <TemplatesTab v-else-if="appStore.activeTab === 'templates'" />
      <TasksMonitorTab v-else-if="appStore.activeTab === 'tasks'" />
      <AutoBenchTab v-else-if="appStore.activeTab === 'autobench'" />
      <HistoryTab v-else-if="appStore.activeTab === 'history'" />
    </div>
  </div>
</template>

<style>
/* ============================================================
   App Layout - Navicat-style Light Theme
   ============================================================ */

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.app-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  background-color: var(--bg-app);
  color: var(--text-primary);
}

/* ============================================================
   Tab Navigation - Desktop Tool Style
   ============================================================ */
.tab-nav {
  background-color: var(--bg-primary);
  border-bottom: 1px solid var(--border-color);
  padding: 0 16px;
  flex-shrink: 0;
}

.tab-list {
  display: flex;
  gap: 0;
  overflow-x: auto;
  height: var(--tab-height);
}

.tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 16px;
  height: 100%;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--text-secondary);
  font-size: var(--font-size-base);
  font-weight: 500;
  cursor: pointer;
  transition: none;
  white-space: nowrap;
}

.tab-item:hover {
  color: var(--text-primary);
  background-color: var(--bg-secondary);
}

.tab-item.active {
  color: var(--primary);
  border-bottom-color: var(--primary);
  background-color: transparent;
}

.tab-icon {
  font-size: 14px;
  line-height: 1;
}

.tab-label {
  font-weight: 500;
}

/* ============================================================
   Tab Content
   ============================================================ */
.tab-content {
  flex: 1;
  overflow: auto;
  padding: var(--spacing-lg);
  background-color: var(--bg-app);
}

/* ============================================================
   Error Modal - Light Theme
   ============================================================ */
.error-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.error-modal {
  background-color: var(--bg-primary);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-modal);
  max-width: 500px;
  width: 90%;
  border: 1px solid var(--danger-border);
}

.error-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  border-bottom: 1px solid var(--border-light);
  background-color: var(--danger-bg);
}

.error-icon {
  font-size: 20px;
}

.error-title {
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--danger);
}

.error-body {
  padding: 16px;
}

.error-message {
  font-size: var(--font-size-base);
  color: var(--text-primary);
  margin-bottom: 12px;
}

.error-details {
  font-size: var(--font-size-xs);
  color: var(--text-muted);
}

.error-details summary {
  cursor: pointer;
  margin-bottom: 8px;
  color: var(--text-secondary);
}

.error-details pre {
  background-color: var(--bg-secondary);
  padding: 12px;
  border-radius: var(--radius-md);
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 200px;
  overflow-y: auto;
  font-family: var(--font-family-mono);
  border: 1px solid var(--border-light);
}

.error-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--border-light);
  display: flex;
  justify-content: flex-end;
}
</style>
