import test from 'node:test'
import assert from 'node:assert/strict'

import { resolveTasksMonitorBinding } from '../src/components/tabs/tasksMonitorTaskState.mjs'

function buildTask(overrides = {}) {
  return {
    id: 'task-active',
    status: 'running',
    current_phase: 'run',
    started_at: '2026-03-15T12:00:00Z',
    template_snapshot: { name: 'Oracle Swingbench Test' },
    connection_snapshot: { name: 'Oracle Prod', type: 'oracle' },
    ...overrides
  }
}

test('stop remains enabled for the active task even when draft changes', () => {
  const tasks = [buildTask()]

  const before = resolveTasksMonitorBinding({
    tasks,
    draft: {
      connection_id: 'conn-a',
      template_id: 'tpl-a',
      action: 'run',
      overrides: { virtual_users: 8, duration: 60 }
    }
  })

  const after = resolveTasksMonitorBinding({
    tasks,
    draft: {
      connection_id: 'conn-b',
      template_id: 'tpl-b',
      action: 'cleanup',
      overrides: { virtual_users: 64, duration: 600 }
    }
  })

  assert.equal(before.stopEnabled, true)
  assert.equal(after.stopEnabled, true)
  assert.equal(before.stopTaskId, 'task-active')
  assert.equal(after.stopTaskId, 'task-active')
})

test('stop always targets the active task identity instead of the latest non-active task or draft', () => {
  const tasks = [
    buildTask({
      id: 'task-completed',
      status: 'success',
      current_phase: 'none',
      started_at: '2026-03-15T12:05:00Z'
    }),
    buildTask({
      id: 'task-active',
      status: 'running',
      started_at: '2026-03-15T12:00:00Z'
    })
  ]

  const result = resolveTasksMonitorBinding({
    tasks,
    draft: {
      connection_id: 'conn-different',
      template_id: 'tpl-different',
      action: 'prepare'
    }
  })

  assert.equal(result.activeTask?.id, 'task-active')
  assert.equal(result.stopTaskId, 'task-active')
  assert.equal(result.logViewerTask?.id, 'task-active')
})

test('stopping task remains the active binding target for stop and log viewer', () => {
  const result = resolveTasksMonitorBinding({
    tasks: [
      buildTask({
        id: 'task-stopping',
        status: 'stopping',
        current_phase: 'run'
      })
    ],
    draft: {
      connection_id: 'conn-next',
      template_id: 'tpl-next',
      action: 'prepare'
    }
  })

  assert.equal(result.activeTask?.id, 'task-stopping')
  assert.equal(result.stopTaskId, 'task-stopping')
  assert.equal(result.logViewerTask?.id, 'task-stopping')
})

test('stop enters stop-in-flight and disables re-entry for the same active task', () => {
  const result = resolveTasksMonitorBinding({
    tasks: [buildTask()],
    draft: {
      connection_id: 'conn-a',
      template_id: 'tpl-a',
      action: 'run'
    },
    pendingStopTaskId: 'task-active'
  })

  assert.equal(result.stopEnabled, false)
  assert.equal(result.stopInFlight, true)
  assert.equal(result.stopTaskId, 'task-active')
})

test('stop is disabled when there is no active task', () => {
  const result = resolveTasksMonitorBinding({
    tasks: [
      buildTask({
        id: 'task-success',
        status: 'success',
        current_phase: 'none'
      })
    ],
    draft: {
      connection_id: 'conn-next',
      template_id: 'tpl-next',
      action: 'run'
    }
  })

  assert.equal(result.stopEnabled, false)
  assert.equal(result.stopTaskId, '')
  assert.equal(result.activeTask, null)
})

test('terminal task clears stale pending stop state and keeps start path unblocked', () => {
  const result = resolveTasksMonitorBinding({
    tasks: [
      buildTask({
        id: 'task-stopped',
        status: 'stopped',
        current_phase: 'none'
      })
    ],
    pendingStopTaskId: 'task-stopped'
  })

  assert.equal(result.activeTask, null)
  assert.equal(result.stopEnabled, false)
  assert.equal(result.stopInFlight, false)
  assert.equal(result.stopTaskId, '')
  assert.equal(result.pendingStopTaskId, '')
  assert.equal(result.startBlocked, false)
})

test('failed preflight task does not keep start blocked or stop enabled', () => {
  const result = resolveTasksMonitorBinding({
    tasks: [
      buildTask({
        id: 'task-failed',
        status: 'failed',
        current_phase: 'none',
        error_message: 'Sysbench run failed: benchmark tables are not prepared. Run Prepare first.'
      })
    ],
    pendingStopTaskId: 'task-failed'
  })

  assert.equal(result.activeTask, null)
  assert.equal(result.stopEnabled, false)
  assert.equal(result.stopTaskId, '')
  assert.equal(result.pendingStopTaskId, '')
  assert.equal(result.startBlocked, false)
})

test('stopping task keeps start blocked and stop locked to prevent re-entry', () => {
  const result = resolveTasksMonitorBinding({
    tasks: [
      buildTask({
        id: 'task-stopping',
        status: 'stopping',
        current_phase: 'run'
      })
    ],
    pendingStopTaskId: 'task-stopping'
  })

  assert.equal(result.activeTask?.id, 'task-stopping')
  assert.equal(result.stopEnabled, false)
  assert.equal(result.stopInFlight, true)
  assert.equal(result.pendingStopTaskId, 'task-stopping')
  assert.equal(result.startBlocked, true)
})
