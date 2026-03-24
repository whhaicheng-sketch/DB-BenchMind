import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { navigationTabs } from '../src/constants/navigationTabs.mjs'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const appSource = fs.readFileSync(path.resolve(__dirname, '../src/App.vue'), 'utf8')

test('navigation only exposes supported GUI tabs', () => {
  const tabIds = navigationTabs.map((tab) => tab.id)

  assert.deepEqual(tabIds, ['connections', 'templates', 'tasks', 'autobench', 'history'])
  assert.equal(tabIds.includes('comparison'), false)
  assert.equal(tabIds.includes('reports'), false)
  assert.equal(tabIds.includes('settings'), false)
})

test('app mounts AutoBench as a dedicated page component instead of reusing performance analysis', () => {
  assert.match(appSource, /import AutoBenchTab from '\.\/components\/tabs\/AutoBenchTab\.vue'/)
  assert.match(appSource, /<TasksMonitorTab v-else-if="appStore\.activeTab === 'tasks'" \/>/)
  assert.match(appSource, /<AutoBenchTab v-else-if="appStore\.activeTab === 'autobench'" \/>/)
})
