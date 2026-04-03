<template>
  <div class="report-metric-chart">
    <!-- Chart Header -->
    <div class="chart-header">
      <span class="chart-label">{{ label }}</span>
      <span class="chart-value" :style="{ color: color }">{{ formattedValue }}</span>
      <span v-if="unit" class="chart-unit">{{ unit }}</span>
    </div>

    <!-- Has Data: show charts -->
    <template v-if="hasData">
      <!-- Stats -->
      <div v-if="showStats" class="chart-stats">
        <span class="stat">平均值: {{ formatNumber(stats.avg) }}</span>
        <span class="stat">最大值: {{ formatNumber(stats.max) }}</span>
        <span class="stat">最小值: {{ formatNumber(stats.min) }}</span>
        <span v-if="stats.isSinglePoint" class="stat stat-note">单点数据</span>
      </div>

      <!-- Bar Chart -->
      <div ref="chartRef" class="bar-chart"></div>

      <!-- Sparkline -->
      <div v-if="showSparkline && hasTimeSeries" class="sparkline-section">
        <span class="sparkline-label">时间序列趋势</span>
        <div ref="sparklineRef" class="sparkline-chart"></div>
      </div>
    </template>

    <!-- No Data State -->
    <div v-else class="no-data-state">
      <span class="no-data-text">暂无数据</span>
      <span class="no-data-hint">该指标在报告中未记录数值</span>
    </div>
  </div>
</template>

<script setup>
/**
 * ReportMetricChart.vue
 * 报告指标图表组件 - 用于显示历史报告中的静态指标数据
 * 不依赖实时监控 store，仅使用传入的静态数据
 */
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import * as echarts from 'echarts'

const props = defineProps({
  metricKey: { type: String, required: true },
  label: { type: String, default: '' },
  currentValue: { type: Number, default: null },
  dataPoints: { type: Array, default: () => [] },
  color: { type: String, default: '#e53e3e' },
  height: { type: String, default: 'auto' },
  width: { type: String, default: '100%' },
  showStats: { type: Boolean, default: true },
  showSparkline: { type: Boolean, default: true },
  unit: { type: String, default: '' }
})

const chartRef = ref(null)
const sparklineRef = ref(null)
const chart = ref(null)
const sparklineChart = ref(null)

// Computed properties
const hasData = computed(() => {
  return props.currentValue !== null && props.currentValue !== undefined
})

const hasTimeSeries = computed(() => {
  return props.dataPoints && props.dataPoints.length > 0
})

const formattedValue = computed(() => {
  if (props.currentValue === null || props.currentValue === undefined) return 'N/A'
  return props.currentValue.toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  })
})

const stats = computed(() => {
  if (!hasTimeSeries.value) {
    return {
      avg: props.currentValue || 0,
      max: props.currentValue || 0,
      min: props.currentValue || 0,
      isSinglePoint: props.currentValue !== null && props.currentValue !== undefined
    }
  }

  const values = props.dataPoints
    .map(p => p.value || p.tpm || p.tps || p.latency_avg || 0)
    .filter(v => v !== null && v !== undefined)

  if (values.length === 0) {
    return {
      avg: props.currentValue || 0,
      max: props.currentValue || 0,
      min: props.currentValue || 0,
      isSinglePoint: props.currentValue !== null && props.currentValue !== undefined
    }
  }

  return {
    avg: values.reduce((a, b) => a + b, 0) / values.length,
    max: Math.max(...values),
    min: Math.min(...values),
    isSinglePoint: false
  }
})

const sparklineData = computed(() => {
  if (!hasTimeSeries.value) return []
  return props.dataPoints
    .slice(-60) // 最多显示 60 个点
    .filter(p => {
      const v = p.value ?? p.tpm ?? p.tps ?? p.latency_avg ?? null
      return v !== null && v !== undefined
    })
    .sort((a, b) => (a.timestamp ?? 0) - (b.timestamp ?? 0))
    .map(p => p.value ?? p.tpm ?? p.tps ?? p.latency_avg ?? 0)
})

const formatNumber = (value) => {
  if (value === null || value === undefined) return 'N/A'
  if (value >= 1000000) return (value / 1000000).toFixed(1) + 'M'
  if (value >= 1000) return (value / 1000).toFixed(1) + 'K'
  return value.toFixed(2)
}

const initCharts = () => {
  // 初始化柱状图
  if (chartRef.value && hasData.value) {
    chart.value = echarts.init(chartRef.value, null, {
      renderer: 'canvas',
      useDirtyRect: true
    })
  }

  // 初始化趋势线图
  if (sparklineRef.value && hasTimeSeries.value && props.showSparkline) {
    sparklineChart.value = echarts.init(sparklineRef.value, null, {
      renderer: 'canvas',
      useDirtyRect: true
    })
  }
}

const updateCharts = () => {
  // 更新柱状图
  if (chart.value && hasData.value) {
    const maxVal = stats.value.max > 0 ? stats.value.max : props.currentValue || 100
    chart.value.setOption({
      animation: false,
      grid: { top: 5, right: 5, bottom: 5, left: 5, containLabel: false },
      xAxis: {
        type: 'value',
        show: false,
        max: Math.max(maxVal * 1.1, props.currentValue || 1)
      },
      yAxis: { type: 'category', show: false, data: [''] },
      series: [{
        type: 'bar',
        data: [props.currentValue || 0],
        barWidth: '60%',
        itemStyle: {
          color: props.color,
          borderRadius: [0, 4, 4, 0]
        }
      }],
      tooltip: { show: false }
    }, true)
  }

  // 更新趋势线
  if (sparklineChart.value && sparklineData.value.length > 0) {
    sparklineChart.value.setOption({
      animation: false,
      grid: { show: false, left: 0, right: 0, top: 0, bottom: 0 },
      xAxis: { type: 'category', show: false },
      yAxis: {
        type: 'value',
        show: false,
        splitLine: { show: false }
      },
      series: [{
        type: 'line',
        data: sparklineData.value,
        smooth: true,
        connectNulls: true,
        symbol: 'none',
        lineStyle: { color: props.color, width: 2 },
        areaStyle: { color: props.color, opacity: 0.2 }
      }]
    }, true)
  }
}

const handleResize = () => {
  if (chart.value) chart.value.resize()
  if (sparklineChart.value) sparklineChart.value.resize()
}

// 监听数据变化
watch(() => [props.currentValue, props.dataPoints, props.color], () => {
  nextTick(() => {
    updateCharts()
  })
}, { deep: true })

onMounted(() => {
  nextTick(() => {
    initCharts()
    updateCharts()
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

<style scoped>
.report-metric-chart {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  background-color: var(--bg-primary);
  border-radius: var(--radius-md);
}

/* Header */
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
  font-family: var(--font-family-mono);
}

.chart-unit {
  font-size: 12px;
  color: var(--text-muted);
  margin-left: 4px;
}

/* Stats */
.chart-stats {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 11px;
  color: var(--text-muted);
}

.stat {
  padding: 4px 8px;
  background-color: var(--bg-secondary);
  border-radius: var(--radius-xs);
}

.stat-note {
  font-style: italic;
  opacity: 0.7;
}

/* Bar Chart */
.bar-chart {
  height: 24px;
  background-color: var(--bg-secondary);
  border-radius: var(--radius-xs);
}

/* Sparkline */
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

/* No Data State */
.no-data-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 16px;
  background-color: var(--bg-secondary);
  border-radius: var(--radius-sm);
  gap: 4px;
}

.no-data-text {
  font-size: 12px;
  color: var(--text-muted);
  font-style: italic;
}

.no-data-hint {
  font-size: 11px;
  color: var(--text-muted);
  opacity: 0.7;
}
</style>
