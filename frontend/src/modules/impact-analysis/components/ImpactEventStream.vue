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
  background-color: #1a2332;
  border: 1px solid #2a3a4a;
  border-radius: 8px;
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
  border-bottom: 1px solid #2a3a4a;
}

.stream-title {
  font-size: 14px;
  font-weight: 600;
  color: #e2e8f0;
  margin: 0;
}

.event-count {
  font-size: 12px;
  color: #718096;
  background-color: rgba(255, 255, 255, 0.05);
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
  color: #718096;
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
  background-color: rgba(0, 0, 0, 0.2);
  border-radius: 6px;
  border-left: 3px solid transparent;
  transition: background-color 0.2s;
}

.event-item:hover {
  background-color: rgba(0, 0, 0, 0.3);
}

.event-item.info {
  border-left-color: #4299e1;
}

.event-item.warn {
  border-left-color: #ecc94b;
}

.event-item.error {
  border-left-color: #f56565;
}

.event-item.success {
  border-left-color: #48bb78;
}

.event-icon {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  background-color: rgba(255, 255, 255, 0.05);
  border-radius: 6px;
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
  color: #e2e8f0;
}

.event-time {
  font-size: 11px;
  color: #718096;
  font-family: 'SF Mono', Monaco, monospace;
}

.event-message {
  font-size: 12px;
  color: #a0aec0;
  line-height: 1.4;
  word-break: break-word;
}

.event-level-badge {
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 9px;
  font-weight: 600;
  text-transform: uppercase;
  flex-shrink: 0;
}

.event-level-badge.info {
  background-color: rgba(66, 153, 225, 0.2);
  color: #63b3ed;
}

.event-level-badge.warn {
  background-color: rgba(236, 201, 75, 0.2);
  color: #f6e05e;
}

.event-level-badge.error {
  background-color: rgba(245, 101, 101, 0.2);
  color: #fc8181;
}

.event-level-badge.success {
  background-color: rgba(72, 187, 120, 0.2);
  color: #68d391;
}

/* Scrollbar styling */
.stream-container::-webkit-scrollbar {
  width: 6px;
}

.stream-container::-webkit-scrollbar-track {
  background: transparent;
}

.stream-container::-webkit-scrollbar-thumb {
  background-color: #3a4a5a;
  border-radius: 3px;
}

.stream-container::-webkit-scrollbar-thumb:hover {
  background-color: #4a5a6a;
}
</style>
