import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const helperPath = path.resolve(__dirname, '../src/components/tabs/autobenchWizardDraft.mjs')

async function loadHelperModule() {
  if (!fs.existsSync(helperPath)) {
    assert.fail('autobenchWizardDraft.mjs is required for T4.1 suite plan generation')
  }

  return import(pathToFileURL(helperPath).href)
}

test('autobench suite items are generated deterministically from selected connections and profiles', async () => {
  const helper = await loadHelperModule()

  const items = helper.buildLocalSuiteItems({
    ...helper.createAutoBenchWizardDraft(),
    selectedConnectionIds: ['placeholder-mysql', 'placeholder-oracle'],
    selectedProfiles: ['io_bound', 'test']
  }, helper.placeholderConnections)

  assert.equal(items.length, 4)
  assert.deepEqual(
    items.map((item) => ({
      id: item.id,
      order: item.order,
      suiteId: item.suiteId,
      connectionId: item.connectionId,
      profileId: item.profileId,
      status: item.status,
      phase: item.phase,
      linkedTaskId: item.linkedTaskId,
      resultSummary: item.resultSummary
    })),
    [
      {
        id: 'autobench-local-suite:placeholder-oracle:test',
        order: 1,
        suiteId: 'autobench-local-suite',
        connectionId: 'placeholder-oracle',
        profileId: 'test',
        status: 'pending',
        phase: 'pending',
        linkedTaskId: '',
        resultSummary: ''
      },
      {
        id: 'autobench-local-suite:placeholder-oracle:io_bound',
        order: 2,
        suiteId: 'autobench-local-suite',
        connectionId: 'placeholder-oracle',
        profileId: 'io_bound',
        status: 'pending',
        phase: 'pending',
        linkedTaskId: '',
        resultSummary: ''
      },
      {
        id: 'autobench-local-suite:placeholder-mysql:test',
        order: 3,
        suiteId: 'autobench-local-suite',
        connectionId: 'placeholder-mysql',
        profileId: 'test',
        status: 'pending',
        phase: 'pending',
        linkedTaskId: '',
        resultSummary: ''
      },
      {
        id: 'autobench-local-suite:placeholder-mysql:io_bound',
        order: 4,
        suiteId: 'autobench-local-suite',
        connectionId: 'placeholder-mysql',
        profileId: 'io_bound',
        status: 'pending',
        phase: 'pending',
        linkedTaskId: '',
        resultSummary: ''
      }
    ]
  )
})

test('autobench suite plan assembles suite-level metadata without runtime side effects', async () => {
  const helper = await loadHelperModule()

  const plan = helper.buildLocalSuitePlan({
    ...helper.createAutoBenchWizardDraft(),
    selectedConnectionIds: ['placeholder-sqlserver', 'placeholder-oracle'],
    selectedProfiles: ['cpu_bound', 'test']
  }, helper.placeholderConnections)

  assert.deepEqual(
    {
      suiteId: plan.suiteId,
      suiteName: plan.suiteName,
      status: plan.status,
      executionMode: plan.executionMode,
      failurePolicy: plan.failurePolicy,
      cleanupEnabled: plan.cleanupEnabled,
      selectedConnectionIds: plan.selectedConnectionIds,
      selectedProfiles: plan.selectedProfiles,
      totalItems: plan.totalItems
    },
    {
      suiteId: 'autobench-local-suite',
      suiteName: 'AutoBench Local Plan',
      status: 'ready',
      executionMode: 'serial',
      failurePolicy: 'continue_by_connection',
      cleanupEnabled: true,
      selectedConnectionIds: ['placeholder-oracle', 'placeholder-sqlserver'],
      selectedProfiles: ['test', 'cpu_bound'],
      totalItems: 4
    }
  )
  assert.equal(plan.items.length, 4)
  assert.equal(plan.items[0].connectionLabel, 'Oracle Placeholder')
  assert.equal(plan.items[3].databaseType, 'sqlserver')
})

test('autobench suite plan remains draft and empty when selections are incomplete', async () => {
  const helper = await loadHelperModule()

  const noConnections = helper.buildLocalSuitePlan({
    ...helper.createAutoBenchWizardDraft(),
    selectedConnectionIds: [],
    selectedProfiles: ['test']
  }, helper.placeholderConnections)

  assert.equal(noConnections.status, 'draft')
  assert.equal(noConnections.totalItems, 0)
  assert.deepEqual(noConnections.items, [])

  const noProfiles = helper.buildLocalSuitePlan({
    ...helper.createAutoBenchWizardDraft(),
    selectedConnectionIds: ['placeholder-oracle'],
    selectedProfiles: []
  }, helper.placeholderConnections)

  assert.equal(noProfiles.status, 'draft')
  assert.equal(noProfiles.totalItems, 0)
  assert.deepEqual(noProfiles.items, [])
})
