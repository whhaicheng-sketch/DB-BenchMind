import test from 'node:test'
import assert from 'node:assert/strict'
import path from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const helperPath = path.resolve(__dirname, '../src/components/tabs/autobenchMonitorState.mjs')

async function loadHelperModule() {
  return import(pathToFileURL(helperPath).href)
}

test('autobench monitor regression keeps aggregate progress current item and item rows in sync', async () => {
  const helper = await loadHelperModule()

  const monitor = helper.buildAutoBenchMonitorState({
    suiteId: 'suite-regression',
    status: 'partial_success',
    totalItems: 5,
    pendingItems: 0,
    runningItems: 0,
    completedItems: 5,
    items: [
      { id: 'item-1', connectionId: 'conn-a', profileType: 'test', status: 'success', phaseStatus: 'completed', resultSummary: 'smoke ok' },
      { id: 'item-2', connectionId: 'conn-a', profileType: 'cpu_bound', status: 'failed', phaseStatus: 'running', resultSummary: 'throughput below threshold' },
      { id: 'item-3', connectionId: 'conn-a', profileType: 'io_bound', status: 'skipped', phaseStatus: 'none' },
      { id: 'item-4', connectionId: 'conn-b', profileType: 'test', status: 'success', phaseStatus: 'completed', resultSummary: 'smoke ok' },
      { id: 'item-5', connectionId: 'conn-b', profileType: 'cpu_bound', status: 'success', phaseStatus: 'completed', resultSummary: 'stable throughput' }
    ]
  })

  assert.equal(monitor.statusLabel, 'partial_success')
  assert.equal(monitor.progressPercent, 100)
  assert.equal(monitor.completedLabel, '5 / 5 completed')
  assert.equal(monitor.currentItem, null)
  assert.equal(monitor.itemRows.length, 5)
  assert.equal(monitor.itemRows[1].status, 'failed')
  assert.equal(monitor.itemRows[1].phaseLabel, 'running')
  assert.equal(monitor.itemRows[1].resultSummary, 'throughput below threshold')
  assert.equal(monitor.itemRows[2].status, 'skipped')
})
