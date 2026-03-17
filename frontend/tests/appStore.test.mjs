import test from 'node:test'
import assert from 'node:assert/strict'
import { clearPendingTaskTemplateState, queueTemplateForTaskState } from '../src/stores/appState.mjs'

test('queueTemplateForTask moves the app to tasks and stores the pending template payload', () => {
  const initialState = {
    activeTab: 'connections',
    pendingTaskTemplate: null
  }
  const payload = { templateId: 'tpl-sysbench-smoke', source: 'template-list' }
  const nextState = queueTemplateForTaskState(initialState, payload)

  assert.equal(nextState.activeTab, 'tasks')
  assert.deepEqual(nextState.pendingTaskTemplate, payload)
})

test('clearPendingTaskTemplate only clears the payload and keeps the active tab unchanged', () => {
  const nextState = clearPendingTaskTemplateState({
    activeTab: 'tasks',
    pendingTaskTemplate: { templateId: 'tpl-sysbench-smoke' }
  })

  assert.equal(nextState.activeTab, 'tasks')
  assert.equal(nextState.pendingTaskTemplate, null)
})
