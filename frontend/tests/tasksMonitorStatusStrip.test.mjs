import test from 'node:test'
import assert from 'node:assert/strict'

import { buildStatusStripModel } from '../src/components/tabs/tasksMonitorStatusStrip.mjs'

test('idle status strip always returns timings and metaItems for template consumers', () => {
  const result = buildStatusStripModel(null, { prepare: '00:11', run: '00:22', total: '00:33' }, '01:00')

  assert.equal(result.status, 'Idle')
  assert.ok(Array.isArray(result.timings))
  assert.equal(result.timings.length, 3)
  assert.ok(Array.isArray(result.metaItems))
  assert.equal(result.metaItems.length, 0)
})

test('active status strip preserves structured meta chips', () => {
  const result = buildStatusStripModel({
    status: 'running',
    current_phase: 'run',
    benchmark_tool: 'swingbench',
    action: 'run',
    template_snapshot: { name: 'Swingbench SOE' },
    connection_snapshot: { type: 'oracle' }
  }, { prepare: '00:12', run: '01:20', total: '01:32' }, '05:00', {
    statusLabel: () => 'Running',
    phaseLabel: () => 'run',
    isActive: () => true
  })

  assert.equal(result.stateClass, 'active')
  assert.deepEqual(result.metaItems.map((item) => item.label), ['Tool', 'Database', 'Template', 'Action', 'Run target'])
})

test('preflight failure with phase none keeps run timing at zero in the strip', () => {
  const result = buildStatusStripModel({
    status: 'failed',
    current_phase: 'none',
    benchmark_tool: 'swingbench',
    action: 'run',
    template_snapshot: { name: 'Oracle Swingbench Test' },
    connection_snapshot: { type: 'oracle' },
    phase_history: []
  }, { prepare: '00:00', run: '00:00', total: '00:07' }, '00:47', {
    statusLabel: () => 'Failed',
    phaseLabel: () => 'none',
    isActive: () => false
  })

  assert.equal(result.run, '00:00')
  assert.match(result.phase, /failed/i)
})

test('sysbench preflight failure keeps failed phase semantics instead of run waiting or finished', () => {
  const result = buildStatusStripModel({
    status: 'failed',
    current_phase: 'none',
    benchmark_tool: 'sysbench',
    action: 'run',
    template_snapshot: { name: 'MySQL Sysbench Test' },
    connection_snapshot: { type: 'mysql' },
    phase_history: []
  }, { prepare: '00:00', run: '00:00', total: '00:03' }, '01:00', {
    statusLabel: () => 'Failed',
    phaseLabel: () => 'none',
    isActive: () => false
  })

  assert.match(result.phase, /failed/i)
  assert.notEqual(result.phase, 'finished')
  assert.equal(result.run, '00:00')
  assert.equal(result.stateClass, 'failed')
})

test('preflight failure surfaces explicit failure reason instead of generic finished phase', () => {
  const result = buildStatusStripModel({
    status: 'failed',
    current_phase: 'none',
    benchmark_tool: 'sysbench',
    action: 'run',
    error_message: 'Sysbench run failed: benchmark tables are not prepared. Please run Prepare first.',
    template_snapshot: { name: 'MySQL Sysbench Test' },
    connection_snapshot: { type: 'mysql' },
    phase_history: []
  }, { prepare: '00:00', run: '00:00', total: '00:03' }, '01:00', {
    statusLabel: () => 'Failed',
    phaseLabel: () => 'finished',
    isActive: () => false
  })

  assert.notEqual(result.phase, 'finished')
  assert.match(result.phase, /preflight failed|run failed|failed/i)
  assert.equal(result.detail, 'See logs for details')
})

test('oracle sysbench and hammerdb direct-run failures share the same reason-first strip behavior', () => {
  const tasks = [
    {
      benchmark_tool: 'swingbench',
      connection_snapshot: { type: 'oracle' },
      error_message: 'Oracle Swingbench run failed: Cleanup removed required SOE objects. Please run Prepare first.'
    },
    {
      benchmark_tool: 'sysbench',
      connection_snapshot: { type: 'postgresql' },
      error_message: 'Sysbench run failed: benchmark tables are not prepared. Please run Prepare first.'
    },
    {
      benchmark_tool: 'hammerdb',
      connection_snapshot: { type: 'sqlserver' },
      error_message: 'HammerDB run failed: benchmark objects are missing. Please run Prepare first.'
    }
  ]

  for (const task of tasks) {
    const result = buildStatusStripModel({
      status: 'failed',
      current_phase: 'none',
      action: 'run',
      template_snapshot: { name: 'Direct run' },
      phase_history: [],
      ...task
    }, { prepare: '00:00', run: '00:00', total: '00:04' }, '01:00', {
      statusLabel: () => 'Failed',
      phaseLabel: () => 'finished',
      isActive: () => false
    })

    assert.equal(result.status, 'Failed')
    assert.equal(result.detail, 'See logs for details')
    assert.match(result.phase, /run failed/i)
    assert.notEqual(result.phase, 'finished')
  }
})

test('generic failed fallback is only used when no failure reason exists', () => {
  const result = buildStatusStripModel({
    status: 'failed',
    current_phase: 'none',
    benchmark_tool: 'sysbench',
    action: 'run',
    template_snapshot: { name: 'MySQL Sysbench Test' },
    connection_snapshot: { type: 'mysql' },
    phase_history: []
  }, { prepare: '00:00', run: '00:00', total: '00:03' }, '01:00', {
    statusLabel: () => 'Failed',
    phaseLabel: () => 'finished',
    isActive: () => false
  })

  assert.match(result.phase, /failed/i)
  assert.equal(result.detail, '')
})

test('failed prepare strip keeps only short summary text and does not expose raw error details', () => {
  const longError = 'step 4 (Create Schema) failed: exit status 255: ORA-01920 user name conflicts with another user or role and a very long diagnostic body that should only exist in logs'

  const result = buildStatusStripModel({
    status: 'failed',
    current_phase: 'prepare',
    action: 'prepare',
    benchmark_tool: 'swingbench',
    error_message: longError,
    template_snapshot: { name: 'Oracle Swingbench Test' },
    connection_snapshot: { type: 'oracle' },
    phase_history: [
      { phase: 'prepare', status: 'failed', message: longError }
    ]
  }, { prepare: '00:31', run: '00:00', total: '00:31' }, '01:00', {
    statusLabel: () => 'Failed',
    phaseLabel: () => 'prepare',
    isActive: () => false
  })

  assert.equal(result.status, 'Failed')
  assert.equal(result.phase, 'Prepare failed')
  assert.equal(result.detail, 'See logs for details')
  assert.doesNotMatch(result.phase, /step 4|exit status|ORA-01920/i)
  assert.doesNotMatch(result.detail, /step 4|exit status|ORA-01920/i)
})

test('failed run strip keeps only short summary text and routes details to logs', () => {
  const longError = 'run failed: exit status 255 with ORA-01017 invalid username/password and extra stderr payload that should not appear in the status card'

  const result = buildStatusStripModel({
    status: 'failed',
    current_phase: 'run',
    action: 'run',
    benchmark_tool: 'swingbench',
    error_message: longError,
    template_snapshot: { name: 'Oracle Swingbench Test' },
    connection_snapshot: { type: 'oracle' },
    phase_history: [
      { phase: 'run', status: 'failed', message: longError }
    ]
  }, { prepare: '00:30', run: '00:04', total: '00:34' }, '01:00', {
    statusLabel: () => 'Failed',
    phaseLabel: () => 'run',
    isActive: () => false
  })

  assert.equal(result.phase, 'Run failed')
  assert.equal(result.detail, 'See logs for details')
  assert.doesNotMatch(result.detail, /exit status|ORA-01017/i)
})
