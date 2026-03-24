import test from 'node:test'
import assert from 'node:assert/strict'

import {
  createEmptyRetainedBusinessMetricsState,
  updateRetainedBusinessMetricsState,
  resolveDisplayedTaskMetrics
} from '../src/components/tabs/tasksMonitorPerformanceMetrics.mjs'

function sampleSeries(values) {
  return values.map((value, index) => ({
    ts: index + 1,
    value
  }))
}

function buildTask(overrides = {}) {
  return {
    id: 'task-1',
    status: 'running',
    metrics: {
      tps: {
        current: 21.5,
        avg: 18.1,
        max: 24.3,
        series: sampleSeries([12.1, 18.4, 21.5])
      },
      tpm: {
        current: 1290,
        avg: 1086,
        max: 1458,
        series: sampleSeries([726, 1104, 1290])
      },
      cpu_user: {
        current: 33,
        series: sampleSeries([24, 28, 33])
      }
    },
    ...overrides
  }
}

test('running to success with empty TPS and TPM payload preserves the last business curves and stats', () => {
  let retained = createEmptyRetainedBusinessMetricsState()
  retained = updateRetainedBusinessMetricsState(retained, buildTask())

  const finishedTask = buildTask({
    status: 'success',
    metrics: {
      tps: { current: 0, avg: 0, max: 0, series: [] },
      tpm: { current: 0, avg: 0, max: 0, series: [] },
      cpu_user: { current: 0, series: [] }
    }
  })

  retained = updateRetainedBusinessMetricsState(retained, finishedTask)
  const displayedMetrics = resolveDisplayedTaskMetrics(finishedTask, retained)

  assert.equal(displayedMetrics.tps.current, 21.5)
  assert.equal(displayedMetrics.tps.avg, 18.1)
  assert.equal(displayedMetrics.tps.max, 24.3)
  assert.deepEqual(displayedMetrics.tps.series, sampleSeries([12.1, 18.4, 21.5]))

  assert.equal(displayedMetrics.tpm.current, 1290)
  assert.equal(displayedMetrics.tpm.avg, 1086)
  assert.equal(displayedMetrics.tpm.max, 1458)
  assert.deepEqual(displayedMetrics.tpm.series, sampleSeries([726, 1104, 1290]))
})

test('terminal task keeps retained TPS and TPM across repeated empty rerenders for the same task', () => {
  let retained = createEmptyRetainedBusinessMetricsState()
  retained = updateRetainedBusinessMetricsState(retained, buildTask())

  const finishedTask = buildTask({
    status: 'completed',
    metrics: {
      tps: { current: 0, avg: 0, max: 0, series: [] },
      tpm: { current: 0, avg: 0, max: 0, series: [] }
    }
  })

  retained = updateRetainedBusinessMetricsState(retained, finishedTask)
  retained = updateRetainedBusinessMetricsState(retained, finishedTask)

  const displayedMetrics = resolveDisplayedTaskMetrics(finishedTask, retained)
  assert.equal(displayedMetrics.tps.current, 21.5)
  assert.equal(displayedMetrics.tpm.current, 1290)
  assert.equal(displayedMetrics.tps.series.length, 3)
  assert.equal(displayedMetrics.tpm.series.length, 3)
})

test('starting a new active task clears retained TPS and TPM from the previous run', () => {
  let retained = createEmptyRetainedBusinessMetricsState()
  retained = updateRetainedBusinessMetricsState(retained, buildTask())

  const newRunningTask = buildTask({
    id: 'task-2',
    status: 'starting',
    metrics: {
      tps: { current: 0, avg: 0, max: 0, series: [] },
      tpm: { current: 0, avg: 0, max: 0, series: [] }
    }
  })

  retained = updateRetainedBusinessMetricsState(retained, newRunningTask)
  const displayedMetrics = resolveDisplayedTaskMetrics(newRunningTask, retained)

  assert.equal(retained.taskId, 'task-2')
  assert.equal(retained.metrics, null)
  assert.deepEqual(displayedMetrics.tps.series, [])
  assert.deepEqual(displayedMetrics.tpm.series, [])
  assert.equal(displayedMetrics.tps.current, 0)
  assert.equal(displayedMetrics.tpm.current, 0)
})

test('system metrics stay sourced from the current task while only TPS and TPM are retained', () => {
  let retained = createEmptyRetainedBusinessMetricsState()
  retained = updateRetainedBusinessMetricsState(retained, buildTask())

  const finishedTask = buildTask({
    status: 'finished',
    metrics: {
      tps: { current: 0, avg: 0, max: 0, series: [] },
      tpm: { current: 0, avg: 0, max: 0, series: [] },
      cpu_user: { current: 7, series: sampleSeries([7]) }
    }
  })

  retained = updateRetainedBusinessMetricsState(retained, finishedTask)
  const displayedMetrics = resolveDisplayedTaskMetrics(finishedTask, retained)

  assert.equal(displayedMetrics.tps.current, 21.5)
  assert.equal(displayedMetrics.tpm.current, 1290)
  assert.equal(displayedMetrics.cpu_user.current, 7)
  assert.deepEqual(displayedMetrics.cpu_user.series, sampleSeries([7]))
})
