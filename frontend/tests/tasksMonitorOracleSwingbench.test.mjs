import test from 'node:test'
import assert from 'node:assert/strict'

import {
  getOracleSwingbenchMetricOverlayState
} from '../src/components/tabs/tasksMonitorOracleSwingbenchState.mjs'

const sampleSeries = (values) => values.map((value, index) => ({
  ts: index + 1,
  value
}))

function buildTask(overrides = {}) {
  return {
    status: 'running',
    current_phase: 'run',
    benchmark_tool: 'swingbench',
    connection_snapshot: {
      type: 'oracle'
    },
    log_tail: [],
    metrics: {
      tps: { current: 0, series: [] },
      tpm: { current: 0, series: [] }
    },
    ...overrides
  }
}

test('returns prepare overlay for Oracle Swingbench prepare phase', () => {
  const state = getOracleSwingbenchMetricOverlayState(buildTask({
    current_phase: 'prepare',
    status: 'preparing'
  }), 'TPS')

  assert.equal(state.kind, 'prepare')
  assert.match(state.title, /Prepare/i)
  assert.match(state.body, /Run/i)
  assert.match(state.body, /0/i)
})

test('does not show run waiting overlay while Oracle Swingbench is still preflighting', () => {
  const state = getOracleSwingbenchMetricOverlayState(buildTask({
    status: 'starting',
    current_phase: 'none',
    log_tail: [
      { content: 'Oracle Swingbench run preflight started' },
      { content: 'Checking SOE workload user and schema prerequisites' }
    ]
  }), 'TPS')

  assert.equal(state.kind, 'none')
})

test('returns run waiting overlay before first non-zero throughput sample appears', () => {
  const state = getOracleSwingbenchMetricOverlayState(buildTask({
    log_tail: [
      { content: 'Swingbench run phase started' },
      { content: 'Establishing workload sessions' }
    ]
  }), 'TPS')

  assert.equal(state.kind, 'run-waiting')
  assert.match(state.title, /Run/i)
  assert.match(state.body, /session/i)
})

test('returns no overlay for Oracle authentication failures because status area owns errors', () => {
  const state = getOracleSwingbenchMetricOverlayState(buildTask({
    log_tail: [
      { content: 'Could not establish/maintain connection' },
      { content: 'ORA-01017: invalid username/password; logon denied' }
    ]
  }), 'TPS')

  assert.equal(state.kind, 'none')
})

test('returns no overlay for missing SOE objects because status area owns errors', () => {
  const state = getOracleSwingbenchMetricOverlayState(buildTask({
    log_tail: [
      { content: 'ORA-00942: table or view does not exist' }
    ]
  }), 'TPM')

  assert.equal(state.kind, 'none')
})

test('returns no overlay after cleanup invalidated the run because status area owns errors', () => {
  const state = getOracleSwingbenchMetricOverlayState(buildTask({
    status: 'failed',
    error_message: 'Cleanup removed required SOE objects. Please run Prepare first.',
    phase_history: [
      { phase: 'cleanup', status: 'success' }
    ],
    log_tail: [
      { content: 'Cleanup removed required SOE objects. Please run Prepare first.' }
    ]
  }), 'TPS')

  assert.equal(state.kind, 'none')
})

test('returns no overlay when run is stuck at zero users and zero throughput', () => {
  const state = getOracleSwingbenchMetricOverlayState(buildTask({
    log_tail: [
      { content: '10:58:35 [0/10]      0        0       0        0     0     0     0     0     0' },
      { content: '10:58:36 [0/10]      0        0       0        0     0     0     0     0     0' },
      { content: '10:58:37 [0/10]      0        0       0        0     0     0     0     0     0' }
    ]
  }), 'TPS')

  assert.equal(state.kind, 'none')
})

test('returns no overlay when task is stopping because status area owns stop messaging', () => {
  const state = getOracleSwingbenchMetricOverlayState(buildTask({
    status: 'stopping',
    log_tail: [
      { content: 'Stop requested' }
    ]
  }), 'TPS')

  assert.equal(state.kind, 'none')
})

test('error-like Oracle run signals never return run-error or stopping overlays', () => {
  const states = [
    getOracleSwingbenchMetricOverlayState(buildTask({
      status: 'failed',
      current_phase: 'run',
      error_message: 'Cleanup removed required SOE objects. Please run Prepare first.'
    }), 'TPS'),
    getOracleSwingbenchMetricOverlayState(buildTask({
      status: 'failed',
      current_phase: 'run',
      error_message: 'ORA-01017: invalid username/password; logon denied'
    }), 'TPS'),
    getOracleSwingbenchMetricOverlayState(buildTask({
      status: 'stopping',
      current_phase: 'run',
      log_tail: [{ content: 'Stop requested' }]
    }), 'TPS')
  ]

  for (const state of states) {
    assert.equal(state.kind, 'none')
    assert.notEqual(state.kind, 'run-error')
    assert.notEqual(state.kind, 'cleanup-invalidated')
    assert.notEqual(state.kind, 'stopping')
  }
})

test('does not return waiting overlay after the first non-zero throughput sample has appeared', () => {
  const state = getOracleSwingbenchMetricOverlayState(buildTask({
    metrics: {
      tps: { current: 0, series: sampleSeries([0, 18.2, 0]) },
      tpm: { current: 0, series: sampleSeries([0, 1092, 0]) }
    },
    log_tail: [
      { content: 'Waiting for next sample window' }
    ]
  }), 'TPS')

  assert.equal(state.kind, 'none')
})

test('does not return overlay for non-Oracle or non-Swingbench tasks', () => {
  const mysqlState = getOracleSwingbenchMetricOverlayState(buildTask({
    connection_snapshot: { type: 'mysql' }
  }), 'TPS')
  const hammerState = getOracleSwingbenchMetricOverlayState(buildTask({
    benchmark_tool: 'hammerdb'
  }), 'TPS')

  assert.equal(mysqlState.kind, 'none')
  assert.equal(hammerState.kind, 'none')
})

test('falls back to a conservative waiting message when logs do not reveal a more specific state', () => {
  const state = getOracleSwingbenchMetricOverlayState(buildTask(), 'TPM')

  assert.equal(state.kind, 'run-waiting')
  assert.match(state.body, /waiting for first TPS\/TPM sample/i)
})

test('error-like logs suppress waiting overlay even when startup logs still exist', () => {
  const state = getOracleSwingbenchMetricOverlayState(buildTask({
    status: 'running',
    log_tail: [
      { content: 'Swingbench run phase started' },
      { content: 'Establishing workload sessions' },
      { content: 'ORA-01017: invalid username/password; logon denied' }
    ]
  }), 'TPS')

  assert.equal(state.kind, 'none')
})
