<script setup>
/**
 * App.vue
 * Main application component with Tab navigation.
 * Follows Fyne framework layout with tabs.
 */
import { ref, onMounted, onUnmounted, onErrorCaptured } from 'vue'
import { useMonitorStore } from './stores/monitor'
import { useAppStore } from './stores/app'
import { navigationTabs } from './constants/navigationTabs.mjs'

// Tab components
import ConnectionsTab from './components/tabs/ConnectionsTab.vue'
import TemplatesTab from './components/tabs/TemplatesTab.vue'
import TasksMonitorTab from './components/tabs/TasksMonitorTab.vue'
import HistoryTab from './components/tabs/HistoryTab.vue'
import ImpactAnalysisTab from './components/tabs/ImpactAnalysisTab.vue'

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
    <div v-if="showError" class="global-error-overlay" @click.self="dismissError">
      <div class="global-error-modal">
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
          <button class="btn-dismiss" @click="dismissError">Dismiss</button>
        </div>
      </div>
    </div>

    <!-- Tab Navigation -->
    <div class="tab-navigation">
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
      <HistoryTab v-else-if="appStore.activeTab === 'history'" />
      <ImpactAnalysisTab v-else-if="appStore.activeTab === 'impact-analysis'" />
    </div>
  </div>
</template>

<style>
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
  background-color: #1b2636;
  color: #e2e8f0;
}

/* Tab Navigation */
.tab-navigation {
  background-color: #0f1724;
  border-bottom: 1px solid #2a3a4a;
  padding: 0 16px;
}

.tab-list {
  display: flex;
  gap: 4px;
  overflow-x: auto;
}

.tab-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: transparent;
  border: none;
  color: #a0aec0;
  font-size: 14px;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
  white-space: nowrap;
}

.tab-item:hover {
  background-color: rgba(255, 255, 255, 0.05);
  color: #e2e8f0;
}

.tab-item.active {
  color: #4299e1;
  border-bottom-color: #4299e1;
  background-color: rgba(66, 153, 225, 0.1);
}

.tab-icon {
  font-size: 16px;
}

.tab-label {
  font-weight: 500;
}

/* Tab Content */
.tab-content {
  flex: 1;
  overflow: auto;
  padding: 20px;
}

/* Error Modal */
.global-error-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.global-error-modal {
  background-color: #2a3a4a;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
  max-width: 500px;
  width: 90%;
  border: 1px solid #e53e3e;
}

.error-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  border-bottom: 1px solid #3a4a5a;
  background-color: rgba(229, 62, 62, 0.1);
}

.error-icon {
  font-size: 20px;
}

.error-title {
  font-size: 16px;
  font-weight: 600;
  color: #e53e3e;
}

.error-body {
  padding: 16px;
}

.error-message {
  font-size: 14px;
  color: #e2e8f0;
  margin-bottom: 12px;
}

.error-details {
  font-size: 11px;
  color: #718096;
}

.error-details summary {
  cursor: pointer;
  margin-bottom: 8px;
}

.error-details pre {
  background-color: #1e2a3a;
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 200px;
  overflow-y: auto;
}

.error-footer {
  padding: 12px 16px;
  border-top: 1px solid #3a4a5a;
  display: flex;
  justify-content: flex-end;
}

.btn-dismiss {
  padding: 8px 16px;
  background-color: #e53e3e;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
}

.btn-dismiss:hover {
  background-color: #c53030;
}
</style>
