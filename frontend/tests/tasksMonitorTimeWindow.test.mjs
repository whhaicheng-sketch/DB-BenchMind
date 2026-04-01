import test from 'node:test'
import assert from 'node:assert/strict'

import {
  TIME_WINDOWS,
  DEFAULT_TIME_WINDOW,
  cropSeriesToWindow,
  cropMetricsToWindow
} from '../src/components/tabs/tasksMonitorTimeWindow.mjs'

test('cropSeriesToWindow returns full series when window is 0', () => {
  const series = Array.from({ length: 100 }, (_, i) => ({ value: i }))
  const result = cropSeriesToWindow(series, 0)
  assert.equal(result.length, 100)
})

test('cropSeriesToWindow crops to the last N points', () => {
  const series = Array.from({ length: 100 }, (_, i) => ({ value: i }))
  const result = cropSeriesToWindow(series, 30)
  assert.equal(result.length, 30)
  assert.equal(result[0].value, 70)
  assert.equal(result[29].value, 99)
})

test('cropSeriesToWindow returns full series when shorter than window', () => {
  const series = [{ value: 1 }, { value: 2 }]
  const result = cropSeriesToWindow(series, 30)
  assert.equal(result.length, 2)
})

test('cropSeriesToWindow handles empty series', () => {
  assert.deepEqual(cropSeriesToWindow([], 30), [])
  assert.deepEqual(cropSeriesToWindow(null, 30), null)
})

test('cropMetricsToWindow crops all metric series', () => {
  const metrics = {
    tps: { current: 10, avg: 8, max: 12, series: Array.from({ length: 100 }, (_, i) => ({ value: i })) },
    tpm: { current: 600, avg: 480, max: 720, series: Array.from({ length: 100 }, (_, i) => ({ value: i * 60 })) }
  }
  const result = cropMetricsToWindow(metrics, 30)
  assert.equal(result.tps.series.length, 30)
  assert.equal(result.tpm.series.length, 30)
  assert.equal(result.tps.current, 10) // non-series fields preserved
})

test('TIME_WINDOWS has expected entries', () => {
  assert.equal(TIME_WINDOWS.length, 3)
  assert.equal(TIME_WINDOWS[0].label, '30s')
  assert.equal(TIME_WINDOWS[1].label, '2m')
  assert.equal(TIME_WINDOWS[2].label, '10m')
})

test('DEFAULT_TIME_WINDOW is 120 seconds (2 minutes)', () => {
  assert.equal(DEFAULT_TIME_WINDOW, 120)
})
