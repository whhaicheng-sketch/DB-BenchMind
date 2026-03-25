import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const helperPath = path.resolve(__dirname, '../src/components/tabs/autobenchMonitorState.mjs')

async function loadHelperModule() {
  if (!fs.existsSync(helperPath)) {
    assert.fail('autobenchMonitorState.mjs is required for T5.1 monitor rendering')
  }

  return import(pathToFileURL(helperPath).href)
}

test('autobench monitor state derives overall progress and current item from suite snapshot', async () => {
  const helper = await loadHelperModule()

  const monitor = helper.buildAutoBenchMonitorState({
    suiteId: 'suite-1',
    status: 'running',
    totalItems: 4,
    pendingItems: 1,
    runningItems: 1,
    completedItems: 2,
    items: [
      { id: 'item-1', connectionId: 'conn-a', profileType: 'test', status: 'success' },
      { id: 'item-2', connectionId: 'conn-a', profileType: 'cpu_bound', status: 'failed' },
      { id: 'item-3', connectionId: 'conn-b', profileType: 'test', status: 'running' },
      { id: 'item-4', connectionId: 'conn-b', profileType: 'cpu_bound', status: 'pending' }
    ]
  })

  assert.equal(monitor.statusLabel, 'running')
  assert.equal(monitor.progressPercent, 50)
  assert.equal(monitor.completedLabel, '2 / 4 completed')
  assert.equal(monitor.currentItem.id, 'item-3')
  assert.equal(monitor.currentItem.status, 'running')
  assert.equal(monitor.currentItem.connectionId, 'conn-b')
  assert.equal(monitor.itemRows.length, 4)
  assert.deepEqual(
    monitor.itemRows.map((item) => ({
      id: item.id,
      status: item.status,
      phaseLabel: item.phaseLabel,
      logLabel: item.logLabel,
      resultSummary: item.resultSummary
    })),
    [
      { id: 'item-1', status: 'success', phaseLabel: 'none', logLabel: 'Logs available later', resultSummary: '' },
      { id: 'item-2', status: 'failed', phaseLabel: 'none', logLabel: 'Logs available later', resultSummary: '' },
      { id: 'item-3', status: 'running', phaseLabel: 'none', logLabel: 'Logs available later', resultSummary: '' },
      { id: 'item-4', status: 'pending', phaseLabel: 'none', logLabel: 'Logs available later', resultSummary: '' }
    ]
  )
})

test('autobench monitor state falls back to an empty summary when no suite exists yet', async () => {
  const helper = await loadHelperModule()

  const monitor = helper.buildAutoBenchMonitorState(null)

  assert.equal(monitor.statusLabel, 'idle')
  assert.equal(monitor.progressPercent, 0)
  assert.equal(monitor.completedLabel, '0 / 0 completed')
  assert.equal(monitor.currentItem, null)
  assert.deepEqual(monitor.itemRows, [])
  assert.match(monitor.emptyMessage, /No AutoBench suite/i)
})
