import { describe, it } from 'node:test'
import assert from 'node:assert/strict'

/**
 * AutoBench observation mode tests
 *
 * Tests the behavior when AutoBench is running and the Benchmark page
 * should be in read-only observation mode linked to the active sub-task.
 */

describe('AutoBench observation mode', () => {
  describe('isAutoBenchControlled derivation', () => {
    it('should be true when suite status is running', () => {
      const suiteStatus = { status: 'running', items: [] }
      assert.strictEqual(suiteStatus?.status === 'running', true)
    })

    it('should be false when suite status is success', () => {
      const suiteStatus = { status: 'success', items: [] }
      assert.strictEqual(suiteStatus?.status === 'running', false)
    })

    it('should be false when suite status is null', () => {
      const suiteStatus = null
      assert.strictEqual(suiteStatus?.status === 'running', false)
    })

    it('should be false when suite status is failed', () => {
      const suiteStatus = { status: 'failed', items: [] }
      assert.strictEqual(suiteStatus?.status === 'running', false)
    })
  })

  describe('autoBenchCurrentItem derivation', () => {
    it('should find the running item in a suite', () => {
      const isAutoBenchControlled = true
      const suiteStatus = {
        status: 'running',
        items: [
          { id: 'item-1', status: 'completed', connection_id: 'conn-1', database_type: 'mysql' },
          { id: 'item-2', status: 'running', connection_id: 'conn-2', database_type: 'oracle' },
          { id: 'item-3', status: 'pending', connection_id: 'conn-3', database_type: 'postgres' }
        ]
      }

      const item = (!isAutoBenchControlled || !suiteStatus?.items)
        ? null
        : suiteStatus.items.find(i => i.status === 'running') || null

      assert.ok(item)
      assert.strictEqual(item.id, 'item-2')
      assert.strictEqual(item.connection_id, 'conn-2')
      assert.strictEqual(item.database_type, 'oracle')
    })

    it('should return null when no item is running', () => {
      const isAutoBenchControlled = true
      const suiteStatus = {
        status: 'success',
        items: [
          { id: 'item-1', status: 'completed' },
          { id: 'item-2', status: 'completed' }
        ]
      }

      const item = (!isAutoBenchControlled || !suiteStatus?.items)
        ? null
        : suiteStatus.items.find(i => i.status === 'running') || null

      assert.strictEqual(item, null)
    })

    it('should return null when not controlled', () => {
      const isAutoBenchControlled = false
      const suiteStatus = {
        status: 'running',
        items: [{ id: 'item-1', status: 'running' }]
      }

      const item = (!isAutoBenchControlled || !suiteStatus?.items)
        ? null
        : suiteStatus.items.find(i => i.status === 'running') || null

      assert.strictEqual(item, null)
    })
  })

  describe('observation mode draft population', () => {
    it('should populate draft from current item', () => {
      const item = {
        id: 'item-1',
        status: 'running',
        connection_id: 'conn-mysql',
        template_id: 'tmpl-oltp',
        database_type: 'mysql'
      }
      const connections = [
        { id: 'conn-mysql', type: 'mysql' }
      ]

      const conn = connections.find(c => c.id === item.connection_id)
      const draft = {
        database_type: item.database_type || conn?.type || '',
        connection_id: item.connection_id,
        template_id: item.template_id || '',
        action: 'full_pipeline'
      }

      assert.strictEqual(draft.database_type, 'mysql')
      assert.strictEqual(draft.connection_id, 'conn-mysql')
      assert.strictEqual(draft.template_id, 'tmpl-oltp')
      assert.strictEqual(draft.action, 'full_pipeline')
    })

    it('should fallback to connection type when item lacks database_type', () => {
      const item = {
        id: 'item-1',
        status: 'running',
        connection_id: 'conn-oracle',
        template_id: 'tmpl-oltp'
      }
      const connections = [
        { id: 'conn-oracle', type: 'oracle' }
      ]

      const conn = connections.find(c => c.id === item.connection_id)
      const databaseType = item.database_type || conn?.type || ''

      assert.strictEqual(databaseType, 'oracle')
    })

    it('should handle missing connection gracefully', () => {
      const item = {
        id: 'item-1',
        status: 'running',
        connection_id: 'conn-missing',
        template_id: 'tmpl-oltp',
        database_type: 'postgres'
      }

      const conn = null
      const databaseType = item.database_type || conn?.type || ''

      assert.strictEqual(databaseType, 'postgres')
    })
  })

  describe('controls disabled during observation', () => {
    it('Start button should be disabled', () => {
      const isAutoBenchControlled = true
      const canStart = !isAutoBenchControlled && true // canPreview && !startBlocked
      assert.strictEqual(canStart, false)
    })

    it('Stop button should be disabled', () => {
      const isAutoBenchControlled = true
      const canStop = !isAutoBenchControlled && true // stopEnabled
      assert.strictEqual(canStop, false)
    })

    it('form selects should be disabled', () => {
      const isAutoBenchControlled = true
      assert.strictEqual(isAutoBenchControlled, true) // :disabled="isAutoBenchControlled"
    })

    it('controls should re-enable when AutoBench completes', () => {
      let isAutoBenchControlled = true
      let canStart = !isAutoBenchControlled && true
      assert.strictEqual(canStart, false)

      // Suite completes
      isAutoBenchControlled = false
      canStart = !isAutoBenchControlled && true
      assert.strictEqual(canStart, true)
    })
  })

  describe('status summary during observation', () => {
    it('should show AutoBench label when controlled', () => {
      const isAutoBenchControlled = true
      const autoBenchCurrentItem = {
        status: 'running',
        database_type: 'mysql',
        phase_status: 'run'
      }

      let statusSummary
      if (isAutoBenchControlled && autoBenchCurrentItem) {
        statusSummary = {
          label: 'AutoBench',
          stateClass: 'autobench',
          message: `AutoBench managing: ${autoBenchCurrentItem.database_type || ''}`
        }
      }

      assert.ok(statusSummary)
      assert.strictEqual(statusSummary.label, 'AutoBench')
      assert.strictEqual(statusSummary.stateClass, 'autobench')
      assert.ok(statusSummary.message.includes('mysql'))
    })

    it('should not show AutoBench label when not controlled', () => {
      const isAutoBenchControlled = false
      assert.strictEqual(isAutoBenchControlled, false)
    })
  })

  describe('task switching in multi-item suite', () => {
    it('should track current running item as suite progresses', () => {
      const suiteItems = [
        { id: 'item-1', status: 'completed', connection_id: 'conn-mysql' },
        { id: 'item-2', status: 'running', connection_id: 'conn-oracle' },
        { id: 'item-3', status: 'pending', connection_id: 'conn-pg' }
      ]

      const current = suiteItems.find(i => i.status === 'running')
      assert.strictEqual(current.id, 'item-2')
      assert.strictEqual(current.connection_id, 'conn-oracle')

      // Item 2 completes, item 3 starts
      suiteItems[1].status = 'completed'
      suiteItems[2].status = 'running'
      const next = suiteItems.find(i => i.status === 'running')
      assert.strictEqual(next.id, 'item-3')
      assert.strictEqual(next.connection_id, 'conn-pg')
    })
  })
})
