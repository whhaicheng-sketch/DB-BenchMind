<script setup>
/**
 * LogPanel.vue
 * Real-time log display panel for benchmark output.
 * Supports auto-scrolling and line limiting.
 */
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { useBenchmarkStore } from '../../stores/benchmark'

// Props
const props = defineProps({
  maxHeight: {
    type: String,
    default: '300px'
  },
  autoScroll: {
    type: Boolean,
    default: true
  },
  maxLines: {
    type: Number,
    default: 200
  }
})

// Store
const benchmarkStore = useBenchmarkStore()

// Refs
const logContainer = ref(null)
const isAutoScrollEnabled = ref(props.autoScroll)

// Computed
const logLines = computed(() => {
  const lines = benchmarkStore.logLines
  // Limit lines to maxLines
  if (lines.length > props.maxLines) {
    return lines.slice(-props.maxLines)
  }
  return lines
})

const hasLogs = computed(() => logLines.value.length > 0)

// Methods
const scrollToBottom = () => {
  nextTick(() => {
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight
    }
  })
}

const toggleAutoScroll = () => {
  isAutoScrollEnabled.value = !isAutoScrollEnabled.value
  if (isAutoScrollEnabled.value) {
    scrollToBottom()
  }
}

const clearLogs = () => {
  benchmarkStore.clearLogs()
}

const getLineClass = (line) => {
  const lowerLine = line.toLowerCase()
  if (lowerLine.includes('error') || lowerLine.includes('fail') || lowerLine.includes('✗')) {
    return 'error'
  }
  if (lowerLine.includes('warning') || lowerLine.includes('warn')) {
    return 'warning'
  }
  if (lowerLine.includes('success') || lowerLine.includes('completed') || lowerLine.includes('✓')) {
    return 'success'
  }
  if (lowerLine.includes('info:') || lowerLine.includes('starting') || lowerLine.includes('loaded')) {
    return 'info'
  }
  return ''
}

// Watch for new logs and auto-scroll
watch(() => benchmarkStore.logLines.length, (newLength, oldLength) => {
  if (isAutoScrollEnabled.value && newLength > (oldLength || 0)) {
    scrollToBottom()
  }
}, { flush: 'sync' })

// Scroll to bottom on mount if logs exist
onMounted(() => {
  if (hasLogs.value && isAutoScrollEnabled.value) {
    scrollToBottom()
  }
})
</script>

<template>
  <div class="log-panel">
    <div class="log-header">
      <span class="log-title">Log Output</span>
      <div class="log-controls">
        <button
          class="control-btn"
          :class="{ active: isAutoScrollEnabled }"
          @click="toggleAutoScroll"
          title="Toggle auto-scroll"
        >
          <span class="icon-scroll">&#8595;</span>
        </button>
        <button
          class="control-btn"
          @click="clearLogs"
          title="Clear logs"
        >
          <span class="icon-clear">&#10005;</span>
        </button>
      </div>
    </div>
    <div
      ref="logContainer"
      class="log-container"
      :style="{ maxHeight: maxHeight }"
    >
      <div v-if="hasLogs" class="log-content">
        <div
          v-for="(line, index) in logLines"
          :key="index"
          class="log-line"
          :class="getLineClass(line)"
        >
          {{ line }}
        </div>
      </div>
      <div v-else class="log-empty">
        Waiting for benchmark data...
      </div>
    </div>
  </div>
</template>

<style scoped>
.log-panel {
  display: flex;
  flex-direction: column;
  background-color: #fafbfc;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 12px;
  background-color: var(--bg-secondary);
  border-bottom: 1px solid var(--border-light);
}

.log-title {
  font-size: var(--font-size-xs);
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.log-controls {
  display: flex;
  gap: 4px;
}

.control-btn {
  width: 24px;
  height: 24px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-xs);
  background-color: var(--bg-primary);
  color: var(--text-muted);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
  font-size: var(--font-size-sm);
}

.control-btn:hover {
  background-color: var(--bg-hover);
  color: var(--text-primary);
  border-color: var(--border-dark);
}

.control-btn.active {
  background-color: var(--primary-light);
  color: var(--primary);
  border-color: var(--primary);
}

.icon-scroll,
.icon-clear {
  font-size: var(--font-size-sm);
  line-height: 1;
}

.log-container {
  flex: 1;
  overflow-y: auto;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  line-height: 1.5;
  background-color: #282c34;
}

.log-content {
  padding: 8px 12px;
}

.log-line {
  white-space: pre-wrap;
  word-break: break-all;
  color: #abb2bf;
  padding: 1px 0;
}

.log-line.error {
  color: #e06c75;
}

.log-line.warning {
  color: #e5c07b;
}

.log-line.success {
  color: #98c379;
}

.log-line.info {
  color: #61afef;
}

.log-empty {
  padding: 20px;
  text-align: center;
  color: #5c6370;
  font-style: italic;
}

/* Scrollbar styling */
.log-container::-webkit-scrollbar {
  width: 8px;
}

.log-container::-webkit-scrollbar-track {
  background: #21252b;
}

.log-container::-webkit-scrollbar-thumb {
  background: #4b5263;
  border-radius: 4px;
}

.log-container::-webkit-scrollbar-thumb:hover {
  background: #5c6370;
}
</style>
