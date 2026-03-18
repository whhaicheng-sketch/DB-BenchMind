<template>
  <div class="impact-event-stream">
    <div class="stream-header">
      <h3 class="stream-title">Event Stream</h3>
      <span class="event-count">{{ events.length }} events</span>
    </div>

    <div class="stream-container" ref="streamContainer">
      <div v-if="events.length === 0" class="stream-empty">
        <span class="empty-icon">📋</span>
        <span>Waiting for events...</span>
      </div>

      <div v-else class="event-list">
        <div
          v-for="event in events"
          :key="event.eventId"
          class="event-item"
          :class="event.level"
        >
          <div class="event-icon">
            {{ getEventIcon(event.type) }}
          </div>
          <div class="event-content">
            <div class="event-header">
              <span class="event-type">{{ getEventLabel(event.type) }}</span>
              <span class="event-time">{{ formatTime(event.timestamp) }}</span>
            </div>
            <div class="event-message">{{ event.message }}</div>
          </div>
          <div class="event-level-badge" :class="event.level">
            {{ event.level }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue'
import { EventTypeConfig } from '../constants'
import { formatTimestamp } from '../types'

const props = defineProps({
  events: {
    type: Array,
    default: () => []
  }
})

const streamContainer = ref(null)

function getEventIcon(eventType) {
  return EventTypeConfig[eventType]?.icon || '📌'
}

function getEventLabel(eventType) {
  return EventTypeConfig[eventType]?.label || eventType
}

function formatTime(timestamp) {
  return formatTimestamp(timestamp)
}

// Auto-scroll to bottom when new events arrive
watch(() => props.events.length, async () => {
  await nextTick()
  if (streamContainer.value) {
    streamContainer.value.scrollTop = streamContainer.value.scrollHeight
  }
})
</script>

<style scoped>
.impact-event-stream {
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 300px;
}

.stream-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-light);
}

.stream-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.event-count {
  font-size: 12px;
  color: var(--text-muted);
  background-color: var(--bg-secondary);
  padding: 2px 8px;
  border-radius: 10px;
}

.stream-container {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

.stream-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-muted);
  gap: 12px;
}

.empty-icon {
  font-size: 32px;
  opacity: 0.5;
}

.event-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.event-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 10px 12px;
  background-color: var(--bg-secondary);
  border-radius: var(--radius-sm);
  border-left: 3px solid transparent;
  transition: background-color var(--transition-fast);
}

.event-item:hover {
  background-color: var(--bg-hover);
}

.event-item.info {
  border-left-color: var(--primary);
}

.event-item.warn {
  border-left-color: var(--warning);
}

.event-item.error {
  border-left-color: var(--danger);
}

.event-item.success {
  border-left-color: var(--success);
}

.event-icon {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  background-color: var(--bg-hover);
  border-radius: var(--radius-sm);
  flex-shrink: 0;
}

.event-content {
  flex: 1;
  min-width: 0;
}

.event-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.event-type {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.event-time {
  font-size: 11px;
  color: var(--text-muted);
  font-family: var(--font-family-mono);
}

.event-message {
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.4;
  word-break: break-word;
}

.event-level-badge {
  padding: 2px 6px;
  border-radius: var(--radius-xs);
  font-size: 9px;
  font-weight: 600;
  text-transform: uppercase;
  flex-shrink: 0;
}

.event-level-badge.info {
  background-color: var(--primary-light);
  color: var(--primary);
}

.event-level-badge.warn {
  background-color: var(--warning-bg);
  color: var(--warning);
}

.event-level-badge.error {
  background-color: var(--danger-bg);
  color: var(--danger);
}

.event-level-badge.success {
  background-color: var(--success-bg);
  color: var(--success);
}

/* Scrollbar styling */
.stream-container::-webkit-scrollbar {
  width: 6px;
}

.stream-container::-webkit-scrollbar-track {
  background: transparent;
}

.stream-container::-webkit-scrollbar-thumb {
  background-color: var(--border-color);
  border-radius: 3px;
}

.stream-container::-webkit-scrollbar-thumb:hover {
  background-color: var(--border-dark);
}
</style>
