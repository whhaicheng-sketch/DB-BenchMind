import { describe, it } from 'node:test'
import assert from 'node:assert/strict'

/**
 * Benchmark observation mode tests
 *
 * Tests that when AutoBench is running, the form fields should be disabled
 * and the draft is populated with AutoBench item data.
 * The tests verify canStart, canStop, and observation mode logic.
 */

// ===========================================================================
// canStart / canStop (observation mode mirrors TasksMonitorTab.vue logic)
// ===========================================================================

function computeCanStart(observationMode, hasConnection, hasTemplate) {
  if (!observationMode) return false
  if (!hasConnection) return false
  if (!hasTemplate) return false
  return true
}

function computeCanStop(observationMode) {
  if (!observationMode) return false
  return true
}

// ===========================================================================
// Populate draft from running item
// ===========================================================================

function populateDraft(runningItem) {
  if (!runningItem) return null
  return {
    id: runningItem.id,
    connectionId: runningItem.connection_id,
    templateId: runningItem.template_id
  }
}

function computeObservationModeState(runningItem) {
  if (!runningItem) return { observationMode: false }
  return { observationMode: true }
}

// ===========================================================================
// Tests
// ===========================================================================

describe('computeCanStart', () => {
  const testCases = [
    {
      name: 'returns false when observationMode is false',
      observationMode: false,
      hasConnection: true,
      hasTemplate: true,
      expected: false
    },
    {
      name: 'returns false when no connection selected',
      observationMode: true,
      hasConnection: false,
      hasTemplate: true,
      expected: false
    },
    {
      name: 'returns false when no template selected',
      observationMode: true,
      hasConnection: true,
      hasTemplate: false,
      expected: false
    },
    {
      name: 'returns true when observationMode active and connection and template present',
      observationMode: true,
      hasConnection: true,
      hasTemplate: true,
      expected: true
    }
  ]

  for (const tt of testCases) {
    it(tt.name, () => {
      const result = computeCanStart(tt.observationMode, tt.hasConnection, tt.hasTemplate)
      assert.strictEqual(result, tt.expected)
    })
  }
})

describe('computeCanStop', () => {
  it('returns false when observationMode is false', () => {
    assert.strictEqual(computeCanStop(false), false)
  })

  it('returns true when observationMode is true', () => {
    assert.strictEqual(computeCanStop(true), true)
  })
})

describe('populateDraft', () => {
  it('returns null when runningItem is null', () => {
    assert.strictEqual(populateDraft(null), null)
  })

  it('returns null when runningItem is undefined', () => {
    assert.strictEqual(populateDraft(undefined), null)
  })

  it('populates draft with running item data', () => {
    const runningItem = {
      id: 'item-1',
      connection_id: 'conn-a',
      template_id: 'mysql_test'
    }
    const draft = populateDraft(runningItem)
    assert.deepStrictEqual(draft, {
      id: 'item-1',
      connectionId: 'conn-a',
      templateId: 'mysql_test'
    })
  })

  it('populates draft for different item', () => {
    const runningItem = {
      id: 'item-2',
      connection_id: 'conn-b',
      template_id: 'pg_oltp'
    }
    const draft = populateDraft(runningItem)
    assert.deepStrictEqual(draft, {
      id: 'item-2',
      connectionId: 'conn-b',
      templateId: 'pg_oltp'
    })
  })
})

describe('computeObservationModeState', () => {
  it('returns observationMode false when no running item', () => {
    const result = computeObservationModeState(null)
    assert.strictEqual(result.observationMode, false)
  })

  it('returns observationMode false when running item is undefined', () => {
    const result = computeObservationModeState(undefined)
    assert.strictEqual(result.observationMode, false)
  })

  it('returns observationMode true when running item is present', () => {
    const runningItem = { id: 'item-1', connection_id: 'conn-a', template_id: 'mysql_test' }
    const result = computeObservationModeState(runningItem)
    assert.strictEqual(result.observationMode, true)
  })
})

describe('observation mode integration', () => {
  it('canStart is false without running item because observationMode is false', () => {
    const runningItem = null
    const state = computeObservationModeState(runningItem)
    const canStart = computeCanStart(state.observationMode, true, true)
    assert.strictEqual(canStart, false)
  })

  it('canStart is true when running item present and connection/template selected', () => {
    const runningItem = { id: 'item-1', connection_id: 'conn-a', template_id: 'mysql_test' }
    const state = computeObservationModeState(runningItem)
    const canStart = computeCanStart(state.observationMode, true, true)
    assert.strictEqual(canStart, true)
  })

  it('canStop is false without running item because observationMode is false', () => {
    const runningItem = null
    const state = computeObservationModeState(runningItem)
    const canStop = computeCanStop(state.observationMode)
    assert.strictEqual(canStop, false)
  })

  it('canStop is true when running item present', () => {
    const runningItem = { id: 'item-1', connection_id: 'conn-a', template_id: 'mysql_test' }
    const state = computeObservationModeState(runningItem)
    const canStop = computeCanStop(state.observationMode)
    assert.strictEqual(canStop, true)
  })

  it('draft is populated with running item data during observation', () => {
    const runningItem = { id: 'item-1', connection_id: 'conn-a', template_id: 'mysql_test' }
    const draft = populateDraft(runningItem)
    assert.strictEqual(draft.id, 'item-1')
    assert.strictEqual(draft.connectionId, 'conn-a')
    assert.strictEqual(draft.templateId, 'mysql_test')
  })

  it('canStart is false during observation when connection is missing', () => {
    const runningItem = { id: 'item-1', connection_id: 'conn-a', template_id: 'mysql_test' }
    const state = computeObservationModeState(runningItem)
    const canStart = computeCanStart(state.observationMode, false, true)
    assert.strictEqual(canStart, false)
  })

  it('canStart is false during observation when template is missing', () => {
    const runningItem = { id: 'item-1', connection_id: 'conn-a', template_id: 'mysql_test' }
    const state = computeObservationModeState(runningItem)
    const canStart = computeCanStart(state.observationMode, true, false)
    assert.strictEqual(canStart, false)
  })
})
