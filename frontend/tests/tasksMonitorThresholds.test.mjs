import test from 'node:test'
import assert from 'node:assert/strict'

import { THRESHOLDS, evaluateThreshold, thresholdIcon, getThresholdDescriptions } from '../src/components/tabs/tasksMonitorThresholds.mjs'

test('evaluateThreshold returns normal for values below warning', () => {
  assert.equal(evaluateThreshold('cpu_user', 30, 0), 'normal')
})

test('evaluateThreshold returns warning for CPU above 70%', () => {
  assert.equal(evaluateThreshold('cpu_user', 75, 0), 'warning')
})

test('evaluateThreshold returns critical for CPU above 85%', () => {
  assert.equal(evaluateThreshold('cpu_user', 90, 0), 'critical')
})

test('evaluateThreshold returns warning for TPS drop > 30% from avg', () => {
  assert.equal(evaluateThreshold('tps', 60, 100), 'warning')
})

test('evaluateThreshold returns critical for TPS drop > 60% from avg', () => {
  assert.equal(evaluateThreshold('tps', 30, 100), 'critical')
})

test('evaluateThreshold returns normal for TPS near avg', () => {
  assert.equal(evaluateThreshold('tps', 95, 100), 'normal')
})

test('evaluateThreshold returns normal for unknown metric key', () => {
  assert.equal(evaluateThreshold('unknown_metric', 999, 0), 'normal')
})

test('thresholdIcon returns correct icons', () => {
  assert.equal(thresholdIcon('normal'), '✓')
  assert.equal(thresholdIcon('warning'), '⚡')
  assert.equal(thresholdIcon('critical'), '🔴')
})

test('getThresholdDescriptions returns descriptions for all metrics', () => {
  const descriptions = getThresholdDescriptions()
  assert.ok(descriptions.length >= 9)
  const cpuDesc = descriptions.find(d => d.key === 'cpu_user')
  assert.ok(cpuDesc)
  assert.ok(cpuDesc.description.includes('CPU'))
})

test('disk latency threshold works correctly', () => {
  assert.equal(evaluateThreshold('disk_latency', 5, 0), 'normal')
  assert.equal(evaluateThreshold('disk_latency', 15, 0), 'warning')
  assert.equal(evaluateThreshold('disk_latency', 60, 0), 'critical')
})

test('disk spike threshold works correctly', () => {
  assert.equal(evaluateThreshold('disk_read', 100, 50), 'normal')   // 100% spike < 200%
  assert.equal(evaluateThreshold('disk_read', 150, 50), 'warning')  // 200% spike
  assert.equal(evaluateThreshold('disk_read', 250, 50), 'critical') // 400% spike
})
