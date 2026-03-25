import test from 'node:test'
import assert from 'node:assert/strict'
import path from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const helperPath = path.resolve(__dirname, '../src/components/tabs/autobenchReportState.mjs')

async function loadHelperModule() {
  return import(pathToFileURL(helperPath).href)
}

test('autobench report state derives summary rows failures and export entries from suite report snapshot', async () => {
  const helper = await loadHelperModule()

  const reportState = helper.buildAutoBenchReportState({
    suiteId: 'suite-001',
    generatedAt: '2026-03-25T12:00:00Z',
    summary: {
      status: 'partial_success',
      totalItems: 4,
      completedItemCount: 4,
      successItemCount: 2,
      failedItemCount: 1,
      skippedItemCount: 1
    },
    failures: [
      {
        suiteItemId: 'suite-001:item-2',
        connectionId: 'placeholder-mysql',
        profileType: 'cpu_bound',
        errorSummary: 'timeout while collecting metrics'
      }
    ],
    recommendations: [
      'Review the failed cpu_bound run before exporting final artifacts.'
    ],
    artifactPaths: {
      html: '/tmp/autobench/suite-001/report.html',
      json: '/tmp/autobench/suite-001/report.json'
    }
  })

  assert.equal(reportState.statusLabel, 'partial_success')
  assert.equal(reportState.summaryCards.length, 5)
  assert.deepEqual(
    reportState.summaryCards.map((item) => `${item.label}:${item.value}`),
    [
      'Suite Status:partial_success',
      'Completed Items:4 / 4',
      'Successful Items:2',
      'Failed Items:1',
      'Skipped Items:1'
    ]
  )
  assert.deepEqual(reportState.failureRows, [
    {
      id: 'suite-001:item-2',
      connectionId: 'placeholder-mysql',
      profileType: 'cpu_bound',
      errorSummary: 'timeout while collecting metrics'
    }
  ])
  assert.deepEqual(reportState.exportEntries, [
    {
      id: 'html',
      label: 'HTML Report',
      path: '/tmp/autobench/suite-001/report.html',
      available: true
    },
    {
      id: 'json',
      label: 'JSON Result',
      path: '/tmp/autobench/suite-001/report.json',
      available: true
    }
  ])
  assert.equal(reportState.recommendations[0], 'Review the failed cpu_bound run before exporting final artifacts.')
  assert.equal(reportState.generatedAtLabel, '2026-03-25T12:00:00Z')
})

test('autobench report state falls back to an empty report summary when no report exists yet', async () => {
  const helper = await loadHelperModule()

  const reportState = helper.buildAutoBenchReportState(null)

  assert.equal(reportState.statusLabel, 'not_started')
  assert.equal(reportState.generatedAtLabel, 'Pending')
  assert.equal(reportState.summaryCards[0].value, 'not_started')
  assert.equal(reportState.failureRows.length, 0)
  assert.equal(reportState.recommendations.length, 1)
  assert.match(reportState.recommendations[0], /suite report will appear/i)
  assert.deepEqual(reportState.exportEntries, [
    {
      id: 'html',
      label: 'HTML Report',
      path: '',
      available: false
    },
    {
      id: 'json',
      label: 'JSON Result',
      path: '',
      available: false
    }
  ])
})

test('autobench report state handles undefined summary gracefully', async () => {
  const helper = await loadHelperModule()

  const reportState = helper.buildAutoBenchReportState({
    suiteId: 'suite-002',
    generatedAt: '2026-03-25T14:00:00Z'
  })

  assert.equal(reportState.statusLabel, 'not_started')
  assert.equal(reportState.generatedAtLabel, '2026-03-25T14:00:00Z')
  assert.equal(reportState.summaryCards.length, 5)
  assert.deepEqual(
    reportState.summaryCards.map((item) => `${item.label}:${item.value}`),
    [
      'Suite Status:not_started',
      'Completed Items:0 / 0',
      'Successful Items:0',
      'Failed Items:0',
      'Skipped Items:0'
    ]
  )
  assert.equal(reportState.failureRows.length, 0)
  assert.equal(reportState.exportEntries.length, 2)
  assert.equal(reportState.exportEntries[0].available, false)
})

test('autobench report state handles empty failures array', async () => {
  const helper = await loadHelperModule()

  const reportState = helper.buildAutoBenchReportState({
    suiteId: 'suite-003',
    generatedAt: '2026-03-25T15:00:00Z',
    summary: {
      status: 'success',
      totalItems: 2,
      completedItemCount: 2,
      successItemCount: 2,
      failedItemCount: 0,
      skippedItemCount: 0
    },
    failures: [],
    artifactPaths: {
      html: '/tmp/autobench/suite-003/report.html',
      json: '/tmp/autobench/suite-003/report.json'
    }
  })

  assert.equal(reportState.statusLabel, 'success')
  assert.equal(reportState.failureRows.length, 0)
  assert.equal(reportState.exportEntries[0].available, true)
  assert.equal(reportState.exportEntries[1].available, true)
})

test('autobench report state provides default recommendations when missing', async () => {
  const helper = await loadHelperModule()

  const reportState = helper.buildAutoBenchReportState({
    suiteId: 'suite-004',
    summary: { status: 'success' }
  })

  assert.equal(reportState.recommendations.length, 1)
  assert.match(reportState.recommendations[0], /no additional recommendations/i)
})

test('autobench report state handles failure items with missing fields', async () => {
  const helper = await loadHelperModule()

  const reportState = helper.buildAutoBenchReportState({
    suiteId: 'suite-005',
    summary: { status: 'partial_success' },
    failures: [
      { suiteItemId: 'item-1' },
      { connectionId: 'conn-x', profileType: 'io_bound' },
      {}
    ]
  })

  assert.equal(reportState.failureRows.length, 3)
  assert.equal(reportState.failureRows[0].id, 'item-1')
  assert.equal(reportState.failureRows[0].connectionId, 'unknown-connection')
  assert.equal(reportState.failureRows[1].connectionId, 'conn-x')
  assert.equal(reportState.failureRows[1].profileType, 'io_bound')
  assert.equal(reportState.failureRows[2].id, 'failure-3')
  assert.equal(reportState.failureRows[2].errorSummary, 'No error summary recorded.')
})

test('autobench report state treats non-numeric summary values as zero', async () => {
  const helper = await loadHelperModule()

  const reportState = helper.buildAutoBenchReportState({
    suiteId: 'suite-006',
    summary: {
      status: 'running',
      totalItems: 'five',
      completedItemCount: null,
      successItemCount: undefined,
      failedItemCount: {}
    }
  })

  assert.equal(reportState.statusLabel, 'running')
  assert.deepEqual(
    reportState.summaryCards.map((item) => item.value),
    ['running', '0 / 0', '0', '0', '0']
  )
})

test('autobench report state treats empty generatedAt as Pending', async () => {
  const helper = await loadHelperModule()

  const reportState = helper.buildAutoBenchReportState({
    suiteId: 'suite-007',
    generatedAt: '',
    summary: { status: '' }
  })

  assert.equal(reportState.generatedAtLabel, 'Pending')
  assert.equal(reportState.statusLabel, 'not_started')
})

test('autobench report state treats empty artifactPaths as unavailable', async () => {
  const helper = await loadHelperModule()

  const reportState = helper.buildAutoBenchReportState({
    suiteId: 'suite-008',
    artifactPaths: {}
  })

  assert.equal(reportState.exportEntries.length, 2)
  assert.equal(reportState.exportEntries[0].available, false)
  assert.equal(reportState.exportEntries[1].available, false)
  assert.equal(reportState.exportEntries[0].path, '')
  assert.equal(reportState.exportEntries[1].path, '')
})

test('autobench report state handles non-string artifact paths gracefully', async () => {
  const helper = await loadHelperModule()

  const reportState = helper.buildAutoBenchReportState({
    suiteId: 'suite-009',
    artifactPaths: {
      html: null,
      json: 12345
    }
  })

  assert.equal(reportState.exportEntries[0].path, '')
  assert.equal(reportState.exportEntries[0].available, false)
  assert.equal(reportState.exportEntries[1].path, '')
  assert.equal(reportState.exportEntries[1].available, false)
})
