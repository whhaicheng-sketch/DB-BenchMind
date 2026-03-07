<script setup>
/**
 * App.vue
 * Main application component.
 * Initializes system monitoring on mount.
 * Includes global error boundary for error handling.
 */
import { ref, onMounted, onUnmounted, onErrorCaptured } from 'vue'
import Sidebar from './components/layout/Sidebar.vue'
import MainContent from './components/layout/MainContent.vue'
import { useMonitorStore } from './stores/monitor'

// Application state
const isRunning = ref(false)
const currentStatus = ref('Idle')

// Global error state (T29.3)
const globalError = ref(null)
const showError = ref(false)

// Monitor store
const monitorStore = useMonitorStore()

// T29.3: Global error boundary
onErrorCaptured((error, instance, info) => {
  console.error('Global error captured:', error)
  console.error('Component:', info)

  // Format error message
  globalError.value = {
    message: error.message || 'An unexpected error occurred',
    component: info?.type?.name || 'Unknown',
    stack: error.stack
  }
  showError.value = true

  // Return false to prevent the error from propagating further
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
  
  // Initialize system monitoring on app start
  try {
    await monitorStore.initSystemMonitoring()
    console.log('System monitoring initialized')
  } catch (err) {
    console.error('Failed to initialize system monitoring:', err)
  }
})

onUnmounted(async () => {
  console.log('DB-BenchMind Wails App unmounted')
  
  // Stop system monitoring on unmount
  try {
    await monitorStore.stopSystemMonitoringAction()
    console.log('System monitoring stopped')
  } catch (err) {
    console.error('Failed to stop system monitoring:', err)
  }
})
</script>

<template>
  <div class="app-container">
    <!-- Global error modal (T29.3) -->
    <div v-if="showError" class="global-error-overlay">
      <div class="global-error-modal">
        <div class="error-header">
          <span class="error-icon">⚠️</span>
          <span class="error-title">Error</span>
        </div>
        <div class="error-body">
          <p class="error-message">{{ globalError?.message }}</p>
          <p v-if="globalError?.component" class="error-component">
            Component: {{ globalError.component }}
          </p>
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

    <!-- Left sidebar: Control panel (300px fixed) -->
    <Sidebar
      class="sidebar"
      :is-running="isRunning"
      :current-status="currentStatus"
    />

    <!-- Right main content: Charts area (flex-grow) -->
    <MainContent
      class="main-content"
      :is-running="isRunning"
    />
  </div>
</template>

<style>
.app-container {
  display: flex;
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.sidebar {
  width: 300px;
  min-width: 300px;
  height: 100%;
  overflow: hidden;
}

.main-content {
  flex: 1;
  height: 100%;
  overflow: hidden;
}

/* Status indicator styles */
.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background-color: #2a3a4a;
  border-radius: 6px;
  margin-bottom: 16px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #718096;
}

.status-indicator.active .status-dot {
  background-color: #48bb78;
  animation: pulse 2s ease-in-out infinite;
}

.status-text {
  font-size: 12px;
  color: #a0aec0;
}

.status-indicator.active .status-text {
  color: #48bb78;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* T29.3: Global error modal styles */
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

.error-component {
  font-size: 12px;
  color: #a0aec0;
  margin-bottom: 8px;
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
