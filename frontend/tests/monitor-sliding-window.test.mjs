import { describe, it } from 'node:test'
import assert from 'node:assert/strict'
import {
  TIME_WINDOWS,
  DEFAULT_TIME_WINDOW,
  cropSeriesToWindow,
  cropMetricsToWindow
} from '../src/components/tabs/tasksMonitorTimeWindow.mjs'

describe('Monitor sliding window', () => {
  describe('TIME_WINDOWS configuration', () => {
    it('should have exactly 3 windows (no All)', () => {
      assert.strictEqual(TIME_WINDOWS.length, 3)
    })

    it('should not contain an All entry', () => {
      const hasAll = TIME_WINDOWS.some(tw => tw.label === 'All' || tw.value === 0)
      assert.strictEqual(hasAll, false, 'No "All" window should exist')
    })

    it('should have 30s, 2m, and 10m entries', () => {
      const labels = TIME_WINDOWS.map(tw => tw.label)
      assert.deepStrictEqual(labels, ['30s', '2m', '10m'])
    })

    it('should have all positive values', () => {
      for (const tw of TIME_WINDOWS) {
        assert.ok(tw.value > 0, `${tw.label} should have positive value, got ${tw.value}`)
      }
    })
  })

  describe('cropSeriesToWindow', () => {
    it('should crop series to the specified window', () => {
      const series = Array.from({ length: 200 }, (_, i) => ({ value: i }))
      const cropped = cropSeriesToWindow(series, 30)
      assert.strictEqual(cropped.length, 30)
      assert.strictEqual(cropped[0].value, 170) // keeps the latest 30
      assert.strictEqual(cropped[29].value, 199)
    })

    it('should not crop if series fits within window', () => {
      const series = Array.from({ length: 10 }, (_, i) => ({ value: i }))
      const cropped = cropSeriesToWindow(series, 30)
      assert.strictEqual(cropped.length, 10)
    })

    it('should handle empty series', () => {
      assert.deepStrictEqual(cropSeriesToWindow([], 30), [])
    })

    it('should handle non-array input', () => {
      assert.strictEqual(cropSeriesToWindow(null, 30), null)
    })

    it('should slide forward as new data arrives', () => {
      const series = Array.from({ length: 100 }, (_, i) => ({ value: i }))
      const window1 = cropSeriesToWindow(series, 30)
      const series2 = [...series, { value: 100 }]
      const window2 = cropSeriesToWindow(series2, 30)
      assert.strictEqual(window1.length, 30)
      assert.strictEqual(window2.length, 30)
      assert.strictEqual(window2[0].value, 71) // shifted by 1
      assert.strictEqual(window2[29].value, 100) // includes new point
    })
  })

  describe('cropMetricsToWindow', () => {
    it('should crop all metric series', () => {
      const metrics = {
        tps: { current: 100, series: Array.from({ length: 200 }, (_, i) => ({ value: i })) },
        tpm: { current: 6000, series: Array.from({ length: 200 }, (_, i) => ({ value: i * 10 })) },
        cpu_user: { current: 50, series: Array.from({ length: 200 }, (_, i) => ({ value: i })) }
      }
      const cropped = cropMetricsToWindow(metrics, 30)
      assert.strictEqual(cropped.tps.series.length, 30)
      assert.strictEqual(cropped.tpm.series.length, 30)
      assert.strictEqual(cropped.cpu_user.series.length, 30)
      // non-series fields preserved
      assert.strictEqual(cropped.tps.current, 100)
      assert.strictEqual(cropped.tpm.current, 6000)
    })

    it('should handle metrics without series', () => {
      const metrics = { tps: { current: 50 } }
      const cropped = cropMetricsToWindow(metrics, 30)
      assert.deepStrictEqual(cropped.tps, { current: 50 })
    })
  })
})

describe('Monitor pause behavior', () => {
  it('pause flag should freeze displayed metrics', () => {
    const metrics = { tps: { current: 100, series: [{ value: 1 }, { value: 2 }] } }
    const windowed = cropMetricsToWindow(metrics, 30)
    // Simulating paused state: the displayed metrics should be the last-rendered snapshot
    let paused = false
    let lastRendered = null
    const getDisplayed = (fresh) => {
      if (paused) return lastRendered
      lastRendered = fresh
      return fresh
    }

    const d1 = getDisplayed(windowed)
    paused = true
    const newMetrics = { tps: { current: 200, series: [{ value: 1 }, { value: 2 }, { value: 3 }] } }
    const newWindowed = cropMetricsToWindow(newMetrics, 30)
    const d2 = getDisplayed(newWindowed)

    // Paused: should still return old data
    assert.strictEqual(d2.tps.current, 100)
    assert.strictEqual(d2.tps.series.length, 2)
  })

  it('resume after pause should show fresh data', () => {
    const metrics1 = { tps: { current: 100, series: [{ value: 1 }] } }
    let paused = false
    let lastRendered = null
    const getDisplayed = (fresh) => {
      if (paused) return lastRendered
      lastRendered = fresh
      return fresh
    }

    getDisplayed(cropMetricsToWindow(metrics1, 30))
    paused = true
    getDisplayed(cropMetricsToWindow({ tps: { current: 200, series: [{ value: 2 }] } }, 30))

    // Resume
    paused = false
    const metrics3 = { tps: { current: 300, series: [{ value: 3 }] } }
    const d3 = getDisplayed(cropMetricsToWindow(metrics3, 30))
    assert.strictEqual(d3.tps.current, 300)
  })

  it('rapid pause/resume toggles should not lose data', () => {
    let paused = false
    let lastRendered = null
    const getDisplayed = (fresh) => {
      if (paused) return lastRendered
      lastRendered = fresh
      return fresh
    }

    for (let i = 0; i < 100; i++) {
      paused = i % 2 === 0
      const m = { tps: { current: i, series: [{ value: i }] } }
      getDisplayed(cropMetricsToWindow(m, 30))
    }
    // After last iteration, paused = false (99 is odd), so lastRendered = metrics at i=99
    assert.strictEqual(lastRendered.tps.current, 99)
  })
})
