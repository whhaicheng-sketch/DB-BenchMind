export function buildStatusStripModel(currentTask, phaseTiming, requestedRunDurationLabel, helpers = {}) {
  const statusLabel = helpers.statusLabel || ((value) => value)
  const phaseLabel = helpers.phaseLabel || ((value) => value)
  const isActive = helpers.isActive || (() => false)
  const lastRunPhase = Array.isArray(currentTask?.phase_history)
    ? [...currentTask.phase_history].reverse().find((entry) => entry?.phase === 'run')
    : null
  const failureReason = resolveFailureReason(currentTask, lastRunPhase)
  const failureSummary = resolveFailureSummary(currentTask, failureReason)

  if (!currentTask) {
    return {
      status: 'Idle',
      phase: 'No active task',
      detail: '',
      prepare: '00:00',
      run: '00:00',
      total: '00:00',
      timings: [
        { label: 'Prepare', value: '00:00' },
        { label: 'Run', value: '00:00' },
        { label: 'Total', value: '00:00' }
      ],
      metaItems: [],
      badgeClass: 'idle',
      stateClass: 'idle'
    }
  }

  return {
    status: statusLabel(currentTask.status),
    phase: failureSummary.phase || (currentTask.status === 'failed' && lastRunPhase?.status === 'failed'
      ? 'Run failed'
      : phaseLabel(currentTask.current_phase, currentTask.status)),
    detail: failureSummary.detail,
    prepare: phaseTiming.prepare,
    run: phaseTiming.run,
    total: phaseTiming.total,
    timings: [
      { label: 'Prepare', value: phaseTiming.prepare },
      { label: 'Run', value: phaseTiming.run },
      { label: 'Total', value: phaseTiming.total }
    ],
    metaItems: [
      { label: 'Tool', value: currentTask.benchmark_tool || 'N/A' },
      { label: 'Database', value: currentTask.connection_snapshot?.type || 'N/A' },
      { label: 'Template', value: currentTask.template_snapshot?.name || 'N/A' },
      { label: 'Action', value: currentTask.action || 'N/A' },
      { label: 'Run target', value: requestedRunDurationLabel }
    ].filter((item) => item.value && item.value !== 'N/A'),
    badgeClass: currentTask.status || 'idle',
    stateClass: isActive(currentTask.status) ? 'active' : currentTask.status || 'idle'
  }
}

function resolveFailureReason(currentTask, lastRunPhase) {
  const candidates = [
    currentTask?.error_message,
    lastRunPhase?.message
  ]

  for (const candidate of candidates) {
    const value = String(candidate || '').trim()
    if (value) {
      return value
    }
  }

  return ''
}

function resolveFailureSummary(currentTask, failureReason) {
  if (currentTask?.status !== 'failed') {
    return { phase: '', detail: '' }
  }

  const phase = resolveFailedPhaseLabel(currentTask)
  if (!failureReason) {
    return { phase, detail: '' }
  }
  return { phase, detail: 'See logs for details' }
}

function resolveFailedPhaseLabel(currentTask) {
  const failedPhase = Array.isArray(currentTask?.phase_history)
    ? [...currentTask.phase_history].reverse().find((entry) => entry?.status === 'failed')?.phase
    : ''

  if (failedPhase === 'prepare' || currentTask?.action === 'prepare' || currentTask?.current_phase === 'prepare') {
    return 'Prepare failed'
  }
  if (failedPhase === 'run' || currentTask?.action === 'run') {
    return 'Run failed'
  }
  return 'Failed'
}
