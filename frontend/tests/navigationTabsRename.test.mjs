import { describe, it } from 'node:test'
import assert from 'node:assert'
import { navigationTabs } from '../src/constants/navigationTabs.mjs'

describe('Navigation Tabs Rename', () => {
  it('should have Benchmark label for tasks tab', () => {
    const tasksTab = navigationTabs.find(t => t.id === 'tasks')
    assert.ok(tasksTab, 'tasks tab should exist')
    assert.strictEqual(tasksTab.label, 'Benchmark')
  })

  it('should have Reports label for history tab', () => {
    const historyTab = navigationTabs.find(t => t.id === 'history')
    assert.ok(historyTab, 'history tab should exist')
    assert.strictEqual(historyTab.label, 'Reports')
  })

  it('should preserve all tab IDs unchanged', () => {
    const expectedIds = ['connections', 'templates', 'tasks', 'autobench', 'history']
    const actualIds = navigationTabs.map(t => t.id)
    assert.deepStrictEqual(actualIds, expectedIds)
  })
})
