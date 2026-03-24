const ACTIVE_STATUSES = new Set(['starting', 'preparing', 'running', 'cleaning', 'stopping'])

export function createEmptyRetainedBusinessMetricsState() {
  return {
    taskId: null,
    metrics: null
  }
}

export function updateRetainedBusinessMetricsState(state = createEmptyRetainedBusinessMetricsState(), task = null) {
  if (!task?.id) {
    return createEmptyRetainedBusinessMetricsState()
  }

  const taskId = task.id
  const retainableMetrics = extractRetainableBusinessMetrics(task.metrics)
  const taskChanged = state.taskId !== taskId

  if (taskChanged) {
    if (ACTIVE_STATUSES.has(task.status)) {
      return {
        taskId,
        metrics: retainableMetrics
      }
    }

    return {
      taskId,
      metrics: retainableMetrics
    }
  }

  if (retainableMetrics) {
    return {
      taskId,
      metrics: retainableMetrics
    }
  }

  return state
}

export function resolveDisplayedTaskMetrics(task = null, retainedState = createEmptyRetainedBusinessMetricsState()) {
  const metrics = task?.metrics || {}
  const retainedMetrics = retainedState.taskId === task?.id ? retainedState.metrics : null

  if (!retainedMetrics) {
    return metrics
  }

  return {
    ...metrics,
    tps: hasDisplayableMetric(metrics.tps) ? metrics.tps : retainedMetrics.tps,
    tpm: hasDisplayableMetric(metrics.tpm) ? metrics.tpm : retainedMetrics.tpm
  }
}

function extractRetainableBusinessMetrics(metrics = {}) {
  if (!hasDisplayableMetric(metrics.tps) && !hasDisplayableMetric(metrics.tpm)) {
    return null
  }

  return {
    tps: cloneMetric(metrics.tps),
    tpm: cloneMetric(metrics.tpm)
  }
}

function hasDisplayableMetric(metric = {}) {
  const current = Number(metric?.current || 0)
  const avg = Number(metric?.avg || 0)
  const max = Number(metric?.max || 0)
  const series = Array.isArray(metric?.series) ? metric.series : []
  return current > 0 || avg > 0 || max > 0 || series.length > 0
}

function cloneMetric(metric = {}) {
  return {
    ...metric,
    series: Array.isArray(metric?.series)
      ? metric.series.map((point) => ({ ...point }))
      : []
  }
}
