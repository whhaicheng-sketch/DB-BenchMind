<template>
  <div class="impact-trend-chart">
    <div class="chart-header">
      <h3 class="chart-title">TPS & Error Trend</h3>
      <div class="chart-legend">
        <div class="legend-item success">
          <span class="legend-color"></span>
          <span>Success TPS</span>
        </div>
        <div class="legend-item error">
          <span class="legend-color"></span>
          <span>Error Count</span>
        </div>
      </div>
    </div>
    <div class="chart-container" ref="chartContainer">
      <div v-if="!trendData || trendData.length === 0" class="chart-empty">
        <span class="empty-icon">📈</span>
        <span>No trend data yet</span>
      </div>
      <svg v-else class="trend-chart" :viewBox="`0 0 ${chartWidth} ${chartHeight}`">
        <!-- Grid lines -->
        <g class="grid-lines">
          <line
            v-for="i in 5"
            :key="'h' + i"
            :x1="padding.left"
            :y1="padding.top + (chartHeight - padding.top - padding.bottom) * (i - 1) / 4"
            :x2="chartWidth - padding.right"
            :y2="padding.top + (chartHeight - padding.top - padding.bottom) * (i - 1) / 4"
            class="grid-line"
          />
        </g>

        <!-- Y-axis labels -->
        <g class="y-axis">
          <text
            v-for="(label, i) in yLabels"
            :key="'y' + i"
            :x="padding.left - 10"
            :y="padding.top + (chartHeight - padding.top - padding.bottom) * i / 4"
            class="axis-label"
            text-anchor="end"
          >
            {{ label }}
          </text>
        </g>

        <!-- Success TPS Line -->
        <path
          v-if="successPath"
          :d="successPath"
          class="line success-line"
          fill="none"
        />
        <path
          v-if="successArea"
          :d="successArea"
          class="area success-area"
        />

        <!-- Error Line -->
        <path
          v-if="errorPath"
          :d="errorPath"
          class="line error-line"
          fill="none"
        />

        <!-- Event Markers -->
        <g class="event-markers">
          <g
            v-for="(event, index) in eventMarkers"
            :key="'event-' + index"
            class="event-marker"
          >
            <line
              :x1="event.x"
              :y1="padding.top"
              :x2="event.x"
              :y2="chartHeight - padding.bottom"
              class="marker-line"
            />
            <circle
              :cx="event.x"
              :cy="padding.top"
              r="4"
              class="marker-dot"
              :class="event.level"
            />
          </g>
        </g>

        <!-- X-axis time labels -->
        <g class="x-axis">
          <text
            v-for="(label, i) in xLabels"
            :key="'x' + i"
            :x="padding.left + (chartWidth - padding.left - padding.right) * i / (xLabels.length - 1)"
            :y="chartHeight - 5"
            class="axis-label"
            text-anchor="middle"
          >
            {{ label }}
          </text>
        </g>
      </svg>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { EventType, EventLevel, ChartConfig } from '../constants'
import { formatTimestamp } from '../types'

const props = defineProps({
  trendData: {
    type: Array,
    default: () => []
  },
  events: {
    type: Array,
    default: () => []
  }
})

const chartContainer = ref(null)
const chartWidth = ref(800)
const chartHeight = ref(300)
const padding = { top: 30, right: 30, bottom: 30, left: 50 }

// Compute max TPS for scaling
const maxTps = computed(() => {
  if (!props.trendData || props.trendData.length === 0) return 100
  const max = Math.max(...props.trendData.map(d => d.successTps || 0))
  return Math.max(max, 100)
})

// Compute max errors for scaling
const maxErrors = computed(() => {
  if (!props.trendData || props.trendData.length === 0) return 10
  const max = Math.max(...props.trendData.map(d => d.errorCount || 0))
  return Math.max(max, 10)
})

// Y-axis labels
const yLabels = computed(() => {
  const labels = []
  for (let i = 0; i <= 4; i++) {
    labels.push(Math.round(maxTps.value * (4 - i) / 4))
  }
  return labels
})

// X-axis labels (time)
const xLabels = computed(() => {
  if (!props.trendData || props.trendData.length === 0) return []

  const count = 5
  const labels = []
  const step = Math.floor(props.trendData.length / (count - 1))

  for (let i = 0; i < count; i++) {
    const index = Math.min(i * step, props.trendData.length - 1)
    labels.push(formatTimestamp(props.trendData[index]?.timestamp))
  }

  return labels
})

// Scale functions
function scaleX(index) {
  const chartAreaWidth = chartWidth.value - padding.left - padding.right
  return padding.left + (index / Math.max(props.trendData.length - 1, 1)) * chartAreaWidth
}

function scaleYTps(value) {
  const chartAreaHeight = chartHeight.value - padding.top - padding.bottom
  return padding.top + chartAreaHeight - (value / maxTps.value) * chartAreaHeight
}

function scaleYError(value) {
  const chartAreaHeight = chartHeight.value - padding.top - padding.bottom
  return padding.top + chartAreaHeight - (value / maxErrors.value) * chartAreaHeight
}

// Success TPS path
const successPath = computed(() => {
  if (!props.trendData || props.trendData.length === 0) return null

  const points = props.trendData
    .map((d, i) => `${scaleX(i)},${scaleYTps(d.successTps || 0)}`)
    .join(' L ')

  return `M ${points}`
})

// Success area (filled)
const successArea = computed(() => {
  if (!props.trendData || props.trendData.length === 0) return null

  const points = props.trendData
    .map((d, i) => `${scaleX(i)},${scaleYTps(d.successTps || 0)}`)
    .join(' L ')

  const baseY = chartHeight.value - padding.bottom
  const startX = padding.left
  const endX = scaleX(props.trendData.length - 1)

  return `M ${startX},${baseY} L ${points} L ${endX},${baseY} Z`
})

// Error path
const errorPath = computed(() => {
  if (!props.trendData || props.trendData.length === 0) return null

  const points = props.trendData
    .map((d, i) => `${scaleX(i)},${scaleYError(d.errorCount || 0)}`)
    .join(' L ')

  return `M ${points}`
})

// Event markers on chart
const eventMarkers = computed(() => {
  if (!props.trendData || props.trendData.length === 0) return []

  const markers = []

  props.trendData.forEach((point, index) => {
    if (point.eventMarkers && point.eventMarkers.length > 0) {
      point.eventMarkers.forEach(eventType => {
        // Find corresponding event for level
        const event = props.events.find(e => e.type === eventType && Math.abs(e.timestamp - point.timestamp) < 2000)
        markers.push({
          x: scaleX(index),
          type: eventType,
          level: event?.level || 'info'
        })
      })
    }
  })

  return markers
})

// Resize handler
function handleResize() {
  if (chartContainer.value) {
    const rect = chartContainer.value.getBoundingClientRect()
    chartWidth.value = rect.width
    chartHeight.value = 300
  }
}

onMounted(() => {
  handleResize()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped>
.impact-trend-chart {
  background-color: #1a2332;
  border: 1px solid #2a3a4a;
  border-radius: 8px;
  padding: 16px 20px;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.chart-title {
  font-size: 14px;
  font-weight: 600;
  color: #e2e8f0;
  margin: 0;
}

.chart-legend {
  display: flex;
  gap: 20px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #a0aec0;
}

.legend-color {
  width: 12px;
  height: 3px;
  border-radius: 2px;
}

.legend-item.success .legend-color {
  background-color: #48bb78;
}

.legend-item.error .legend-color {
  background-color: #f56565;
}

.chart-container {
  width: 100%;
  height: 300px;
  position: relative;
}

.chart-empty {
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

.trend-chart {
  width: 100%;
  height: 100%;
  overflow: visible;
}

.grid-line {
  stroke: #2a3a4a;
  stroke-width: 1;
  stroke-dasharray: 4, 4;
}

.axis-label {
  fill: #718096;
  font-size: 10px;
}

.line {
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.success-line {
  stroke: #48bb78;
}

.success-area {
  fill: rgba(72, 187, 120, 0.1);
}

.error-line {
  stroke: #f56565;
}

.marker-line {
  stroke: #ecc94b;
  stroke-width: 1;
  stroke-dasharray: 3, 3;
  opacity: 0.6;
}

.marker-dot {
  stroke-width: 0;
}

.marker-dot.info {
  fill: #4299e1;
}

.marker-dot.warn {
  fill: #ecc94b;
}

.marker-dot.error {
  fill: #f56565;
}

.marker-dot.success {
  fill: #48bb78;
}
</style>
