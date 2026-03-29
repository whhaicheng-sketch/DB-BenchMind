import test from 'node:test'
import assert from 'node:assert/strict'

import { detectAnomalies, mapHoverIndex, getValuesAtHover, computeSystemAvgs } from '../src/components/tabs/tasksMonitorCorrelation.mjs'

test('detectAnomalies returns empty for normal values', () => {
  const series = {
    cpu_user: Array.from({ length: 10 }, () => ({ value: 50 })),
    cpu_sys: Array.from({ length: 10 }, () => ({ value: 20 }))
  }
  const avgs = { cpu_user: 50, cpu_sys: 20 }
  const anomalies = detectAnomalies(series, avgs)
  assert.equal(anomalies.length, 0)
})

test('detectAnomalies detects critical CPU spike', () => {
  const series = {
    cpu_user: [{ value: 50 }, { value: 50 }, { value: 90 }, { value: 50 }]
  }
  const avgs = { cpu_user: 50 }
  const anomalies = detectAnomalies(series, avgs)
  assert.equal(anomalies.length, 1)
  assert.equal(anomalies[0].index, 2)
  assert.equal(anomalies[0].worstLevel, 'critical')
})

test('detectAnomalies handles empty series', () => {
  assert.deepEqual(detectAnomalies({}, {}), [])
  assert.deepEqual(detectAnomalies(null, null), [])
})

test('mapHoverIndex maps correctly between same-length series', () => {
  assert.equal(mapHoverIndex(5, 10, 10), 5)
})

test('mapHoverIndex maps correctly between different-length series', () => {
  assert.equal(mapHoverIndex(5, 10, 5), 2)
})

test('mapHoverIndex returns -1 for empty target', () => {
  assert.equal(mapHoverIndex(5, 10, 0), -1)
})

test('getValuesAtHover returns values at correct index', () => {
  const series = {
    tps: [{ value: 10 }, { value: 20 }, { value: 30 }],
    cpu_user: [{ value: 40 }, { value: 50 }, { value: 60 }]
  }
  const values = getValuesAtHover(series, 1)
  assert.equal(values.tps, 20)
  assert.equal(values.cpu_user, 50)
})

test('getValuesAtHover returns null for invalid index', () => {
  assert.equal(getValuesAtHover({}, -1), null)
  assert.equal(getValuesAtHover(null, 0), null)
})

test('computeSystemAvgs calculates average from series', () => {
  const metrics = {
    cpu_user: [{ value: 40 }, { value: 60 }],
    cpu_sys: [{ value: 10 }, { value: 30 }]
  }
  const avgs = computeSystemAvgs(metrics)
  assert.equal(avgs.cpu_user, 50)
  assert.equal(avgs.cpu_sys, 20)
})

test('computeSystemAvgs handles empty metrics', () => {
  assert.deepEqual(computeSystemAvgs(null), {})
  assert.deepEqual(computeSystemAvgs({}), {})
})
