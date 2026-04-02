const ACTIVE_STATUSES = new Set(['starting', 'preparing', 'running', 'cleaning', 'stopping'])

export function isOracleSwingbenchTask(task) {
  return task?.benchmark_tool === 'swingbench' && task?.connection_snapshot?.type === 'oracle'
}

// getMetricOverlayState returns an overlay descriptor for any task type.
// For Oracle Swingbench tasks, it delegates to the full Oracle-specific logic.
// For other adapters, it shows a simple prepare-phase overlay when applicable.
export function getMetricOverlayState(task, metricLabel = 'TPS') {
  // Oracle Swingbench gets the full specialized treatment
  if (isOracleSwingbenchTask(task)) {
    return getOracleSwingbenchMetricOverlayState(task, metricLabel)
  }

  // Generic prepare phase overlay for any adapter
  if (task?.current_phase === 'preparing' || task?.status === 'preparing') {
    return {
      kind: 'prepare',
      title: 'Prepare phase',
      body: 'TPS/TPM will appear after Run starts. Current 0 is expected here.'
    }
  }

  return { kind: 'none', title: '', body: '' }
}

export function getOracleSwingbenchMetricOverlayState(task, metricLabel = 'TPS') {
  if (!isOracleSwingbenchTask(task)) {
    return { kind: 'none', title: '', body: '' }
  }

  if (task?.current_phase === 'prepare') {
    return {
      kind: 'prepare',
      title: 'Prepare phase',
      body: 'TPS/TPM will appear after Run starts. Current 0 is expected here.'
    }
  }

  if (!isHealthyRunWaitingTask(task) || hasSeenFirstThroughputSample(task) || hasBlockingRunSignal(task)) {
    return { kind: 'none', title: '', body: '' }
  }

  return {
    kind: 'run-waiting',
    title: 'Run started',
    body: inferRunWaitingMessage(task, metricLabel)
  }
}

function isActiveTask(task) {
  return ACTIVE_STATUSES.has(task?.status)
}

function isHealthyRunWaitingTask(task) {
  return isActiveTask(task) && task?.status === 'running' && task?.current_phase === 'run'
}

function hasSeenFirstThroughputSample(task) {
  return ['tps', 'tpm'].some((metricKey) => {
    const series = Array.isArray(task?.metrics?.[metricKey]?.series) ? task.metrics[metricKey].series : []
    return series.some((point) => Number(point?.value || 0) > 0)
  })
}

function hasBlockingRunSignal(task) {
  if (task?.status === 'failed' || task?.status === 'stopping' || task?.status === 'stopped') {
    return true
  }

  const corpus = [
    task?.error_message,
    ...collectRecentLogContent(task)
  ]
    .filter(Boolean)
    .join('\n')

  if (!corpus.trim()) {
    return false
  }

  return /cleanup removed required soe objects/i.test(corpus) ||
    /ora-01017/i.test(corpus) ||
    /invalid username\/password/i.test(corpus) ||
    /could not establish\/maintain connection/i.test(corpus) ||
    /ora-12154/i.test(corpus) ||
    /ora-125\d+/i.test(corpus) ||
    /tns:/i.test(corpus) ||
    /logon denied/i.test(corpus) ||
    /ora-00942/i.test(corpus) ||
    /table or view does not exist/i.test(corpus) ||
    /schema.*missing/i.test(corpus) ||
    /\bsoe\b.*(missing|does not exist)/i.test(corpus) ||
    /object.*does not exist/i.test(corpus) ||
    /child process failed/i.test(corpus) ||
    /process exited/i.test(corpus) ||
    /phase finished: run \(failed\)/i.test(corpus) ||
    hasStalledZeroThroughputRun(task)
}

function collectRecentLogContent(task) {
  const lines = Array.isArray(task?.log_tail) ? task.log_tail : []
  return lines
    .slice(-16)
    .map((line) => String(line?.content || '').trim())
    .filter(Boolean)
}

function inferRunWaitingMessage(task, metricLabel) {
  const recentContent = collectRecentLogContent(task).reverse()

  for (const content of recentContent) {
    const lower = content.toLowerCase()
    if (/(warmup|warming up|ramp up)/.test(lower)) {
      return 'Warmup in progress, first TPS/TPM sample is still pending.'
    }
    if (/(session|logon|login|connect|connection pool)/.test(lower)) {
      return 'Establishing workload sessions before the first TPS/TPM sample.'
    }
    if (/(starting|launch|charbench|initializ|boot)/.test(lower)) {
      return 'Swingbench is starting up, waiting for the first TPS/TPM sample window.'
    }
    if (/(sample|snapshot|interval|waiting)/.test(lower)) {
      return 'Run started, waiting for first TPS/TPM sample.'
    }
  }

  const metricName = metricLabel === 'TPM' ? 'TPS/TPM' : 'TPS/TPM'
  return `Run started, waiting for first ${metricName} sample.`
}

function hasStalledZeroThroughputRun(task) {
  const zeroRows = collectRecentLogContent(task)
    .filter((content) => /\[\s*0\/\d+\s*\]/.test(content) && /\b0(\.0+)?\b/.test(content))

  return zeroRows.length >= 3
}
