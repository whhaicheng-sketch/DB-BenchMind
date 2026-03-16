import test from 'node:test'
import assert from 'node:assert/strict'

import { navigationTabs } from '../src/constants/navigationTabs.mjs'

test('navigation only exposes supported GUI tabs', () => {
  const tabIds = navigationTabs.map((tab) => tab.id)

  assert.deepEqual(tabIds, ['connections', 'templates', 'tasks', 'history'])
  assert.equal(tabIds.includes('comparison'), false)
  assert.equal(tabIds.includes('reports'), false)
  assert.equal(tabIds.includes('settings'), false)
})
