<script setup>
/**
 * MetricChart.vue - Performance Optimized
 * T27: Optimized for minimal redraw and memory usage
 */
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import * as echarts from 'echarts'

// T27.1: Throttle helper to limit redraw frequency
const useThrottle = (fn, delay) => {
  let lastCall = 0
  let pendingArgs = null
  let timeoutId = null

  return (...args) => {
    const now = Date.now()
    pendingArgs = args

    if (now - lastCall >= delay) {
      lastCall = now
      fn(...args)
    } else if (!timeoutId) {
      timeoutId = setTimeout(() => {
        lastCall = Date.now()
        timeoutId = null
        if (pendingArgs) {
          fn(...pendingArgs)
        }
      }, delay - (now - lastCall))
    }
  }
}

const props = defineProps({
  metricKey: { type: String, required: true },
  label: { type: String, default: '' },
  currentValue: { type: Number, default: 0 },
  unit: { type: String, default: '' },
  dataPoints: { type: Array, default: () => [] },
  color: { type: String, default: '#e53e3e' },
  height: { type: String, default: '120px' },
  width: { type: String, default: '100%' },
  showStats: { type: Boolean, default: true },
  showSparkline: { type: Boolean, default: true }
})

const chartRef = ref(null)
const sparklineRef = ref(null)
const chart = ref(null)
const sparklineChart = ref(null)

// T27.4: Cache last values to avoid unnecessary updates
let lastCurrentValue = null
let lastDataLength = 0
let lastColor = null

const formattedValue = computed(() => {
  if (props.currentValue === null || props.currentValue === undefined) return 'N/A'
  return props.currentValue.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
})

// T27.3: Optimize stats calculation with memoization
const stats = computed(() => {
  if (!props.dataPoints || props.dataPoints.length === 0) return null
  // Only process last 60 points for performance
  const points = props.dataPoints.slice(-60)
  const values = points.map(p => p.value || p.tpm || p.tps || 0)
  const avg = values.reduce((a, b) => a + b, 0) / values.length
  return { avg, max: Math.max(...values), min: Math.min(...values), count: values.length }
})

// T27.3: Limit sparkline data to last 60 points
const sparklineData = computed(() => {
  if (!props.dataPoints || props.dataPoints.length === 0) return []
  // Only take last 60 points for sparkline
  return props.dataPoints.slice(-60).map(p => p.value || p.tpm || p.tps || 0)
})

const formatNumber = (value) => {
  if (value === null || value === undefined) return 'N/A'
  if (value >= 1000000) return (value / 1000000).toFixed(1) + 'M'
  if (value >= 1000) return (value / 1000).toFixed(1) + 'K'
  return value.toFixed(1)
}

const initCharts = () => {
  if (chartRef.value) {
    chart.value = echarts.init(chartRef.value, null, {
      renderer: 'canvas', // Canvas is more performant than SVG
      useDirtyRect: true  // Enable dirty rectangle rendering
    })
  }
  if (sparklineRef.value && props.showSparkline) {
    sparklineChart.value = echarts.init(sparklineRef.value, null, {
      renderer: 'canvas',
      useDirtyRect: true
    })
  }
}

// T27.2: Throttled update function (max 10fps for charts)
const updateChartsThrottled = useThrottle(() => {
  updateChartsInternal()
}, 100)

const updateChartsInternal = () => {
  // T27.4: Skip if values haven't changed
  const dataLength = props.dataPoints?.length || 0
  const valueChanged = props.currentValue !== lastCurrentValue
  const lengthChanged = dataLength !== lastDataLength
  const colorChanged = props.color !== lastColor

  if (!valueChanged && !lengthChanged && !colorChanged) {
    return
  }

  lastCurrentValue = props.currentValue
  lastDataLength = dataLength
  lastColor = props.color

  // Bar chart - only update data series
  if (chart.value) {
    const maxVal = stats.value ? stats.value.max : 100
    chart.value.setOption({
      animation: false,
      grid: { top: 5, right: 5, bottom: 5, left: 5, containLabel: false },
      xAxis: { type: 'value', show: false, max: Math.max(maxVal, props.currentValue || 1) },
      yAxis: { type: 'category', show: false, data: [''] },
      series: [{
        type: 'bar',
        data: [props.currentValue || 0],
        barWidth: '60%',
        itemStyle: { color: props.color, borderRadius: [0, 4, 4, 0] }
      }],
      tooltip: { show: false }
    }, false, true) // T27.2: notMerge=false, lazyUpdate=true
  }

  // Sparkline - only update if data length changed
  if (sparklineChart.value && sparklineData.value.length > 0) {
    sparklineChart.value.setOption({
      animation: false,
      grid: { show: false, left: 0, right: 0, top: 0, bottom: 0 },
      xAxis: { type: 'category', show: false },
      yAxis: { type: 'value', show: false, splitLine: { show: false } },
      series: [{
        type: 'line',
        data: sparklineData.value,
        smooth: 0.3,
        symbol: 'none',
        lineStyle: { color: props.color, width: 2 },
        areaStyle: { color: props.color, opacity: 0.2 }
      }]
    }, false, true) // T27.2: notMerge=false, lazyUpdate=true
  }
}

// T27.2: Use throttled updates
watch(() => props.currentValue, updateChartsThrottled)
// T27.3: Don't use deep watch, only watch length
watch(() => props.dataPoints?.length, updateChartsThrottled)
watch(() => props.color, updateChartsThrottled)

// Handle window resize
const handleResize = useThrottle(() => {
  if (chart.value) chart.value.resize()
  if (sparklineChart.value) sparklineChart.value.resize()
}, 200)

onMounted(() => {
  nextTick(() => {
    initCharts()
    updateChartsInternal()
    window.addEventListener('resize', handleResize)
  })
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  if (chart.value) {
    chart.value.dispose()
    chart.value = null
  }
  if (sparklineChart.value) {
    sparklineChart.value.dispose()
    sparklineChart.value = null
  }
})
</script>

<template>
  <div class="metric-chart" :style="{ height: height, width: width }">
    <div class="chart-header">
      <span class="chart-label">{{ label || metricKey }}</span>
      <span class="chart-value" :style="{ color: color }">{{ formattedValue }}</span>
      <span v-if="unit" class="chart-unit">{{ unit }}</span>
    </div>
    <div v-if="showStats && stats" class="chart-stats">
      <span class="stat">Avg: {{ formatNumber(stats.avg) }}</span>
      <span class="stat">Max: {{ formatNumber(stats.max) }}</span>
    </div>
    <div ref="chartRef" class="bar-chart"></div>
    <div v-if="showSparkline" class="sparkline-section">
      <span class="sparkline-label">Trend (60s)</span>
      <div ref="sparklineRef" class="sparkline-chart"></div>
    </div>
  </div>
</template>

<style scoped>
.metric-chart {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
}
.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.chart-label {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}
.chart-value {
  font-size: 20px;
  font-weight: 600;
}
.chart-unit {
  font-size: 12px;
  color: var(--text-muted);
  margin-left: 4px;
}
.chart-stats {
  display: flex;
  gap: 12px;
  font-size: 11px;
  color: var(--text-muted);
}
.stat {
  padding: 4px 8px;
  background-color: var(--bg-secondary);
  border-radius: var(--radius-xs);
}
.bar-chart {
  height: 24px;
  background-color: var(--bg-secondary);
  border-radius: var(--radius-xs);
}
.sparkline-section {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 8px;
}
.sparkline-label {
  font-size: 10px;
  color: var(--text-muted);
  text-transform: uppercase;
}
.sparkline-chart {
  height: 40px;
  background-color: var(--bg-secondary);
  border-radius: var(--radius-xs);
}
</style>
