import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const helperPath = path.resolve(__dirname, '../src/components/tabs/autobenchWizardDraft.mjs')

async function loadHelperModule() {
  if (!fs.existsSync(helperPath)) {
    assert.fail('autobenchWizardDraft.mjs is required for T2.3 local draft behavior')
  }

  return import(pathToFileURL(helperPath).href)
}

test('autobench wizard draft defaults match the first-phase orchestration policy', async () => {
  const helper = await loadHelperModule()
  const draft = helper.createAutoBenchWizardDraft()

  assert.deepEqual(draft.selectedConnectionIds, [])
  assert.deepEqual(draft.selectedProfiles, ['test', 'cpu_bound', 'io_bound'])
  assert.equal(draft.executionMode, 'serial')
  assert.equal(draft.failurePolicy, 'continue_by_connection')
  assert.equal(draft.cleanupEnabled, true)
})

test('autobench wizard draft toggles local connection and profile selection without backend data', async () => {
  const helper = await loadHelperModule()
  let draft = helper.createAutoBenchWizardDraft()

  draft = helper.toggleDraftConnectionSelection(draft, 'placeholder-mysql')
  assert.deepEqual(draft.selectedConnectionIds, ['placeholder-mysql'])

  draft = helper.toggleDraftConnectionSelection(draft, 'placeholder-mysql')
  assert.deepEqual(draft.selectedConnectionIds, [])

  draft = helper.toggleDraftProfileSelection(draft, 'cpu_bound')
  assert.deepEqual(draft.selectedProfiles, ['test', 'io_bound'])

  draft = helper.toggleDraftProfileSelection(draft, 'cpu_bound')
  assert.deepEqual(draft.selectedProfiles, ['test', 'cpu_bound', 'io_bound'])
})

test('autobench wizard validation keeps create action disabled until at least one connection and profile are selected', async () => {
  const helper = await loadHelperModule()
  const emptyDraft = {
    ...helper.createAutoBenchWizardDraft(),
    selectedProfiles: []
  }

  assert.deepEqual(helper.validateAutoBenchWizardDraft(emptyDraft), {
    canCreateSuite: false,
    connectionError: 'Select at least one placeholder connection to continue.',
    profileError: 'Select at least one profile to continue.'
  })

  const validDraft = {
    ...helper.createAutoBenchWizardDraft(),
    selectedConnectionIds: ['placeholder-oracle']
  }

  assert.deepEqual(helper.validateAutoBenchWizardDraft(validDraft), {
    canCreateSuite: true,
    connectionError: '',
    profileError: ''
  })
})

test('autobench connection filter options and local filtering stay fully frontend-side', async () => {
  const helper = await loadHelperModule()

  assert.deepEqual(helper.connectionFilterOptions.map((item) => item.id), [
    'all',
    'oracle',
    'mysql',
    'postgresql',
    'sqlserver'
  ])

  const oracleConnections = helper.filterPlaceholderConnections(helper.placeholderConnections, 'oracle')
  assert.equal(oracleConnections.length, 1)
  assert.equal(oracleConnections[0].databaseType, 'oracle')

  const sqlserverConnections = helper.filterPlaceholderConnections(helper.placeholderConnections, 'sqlserver')
  assert.equal(sqlserverConnections.length, 1)
  assert.equal(sqlserverConnections[0].databaseType, 'sqlserver')

  const allConnections = helper.filterPlaceholderConnections(helper.placeholderConnections, 'all')
  assert.equal(allConnections.length, helper.placeholderConnections.length)
})

test('autobench profile options keep canonical order and expose local scope metadata', async () => {
  const helper = await loadHelperModule()

  assert.deepEqual(helper.profileOptions.map((item) => item.id), ['test', 'cpu_bound', 'io_bound'])
  assert.deepEqual(helper.profileOptions.map((item) => item.scope), ['smoke', 'throughput', 'storage'])
  assert.match(helper.profileOptions[0].description, /fast|smoke/i)
  assert.match(helper.profileOptions[1].description, /cpu/i)
  assert.match(helper.profileOptions[2].description, /io|storage/i)
})

test('autobench selected profile summary stays ordered and frontend-only', async () => {
  const helper = await loadHelperModule()

  assert.equal(
    helper.describeSelectedProfiles(['io_bound', 'test']),
    'test -> io_bound'
  )

  assert.equal(
    helper.describeSelectedProfiles([]),
    'none'
  )
})

test('autobench local plan preview expands selected connections and profiles in stable serial order', async () => {
  const helper = await loadHelperModule()

  const preview = helper.buildLocalPlanPreview({
    selectedConnectionIds: ['placeholder-mysql', 'placeholder-oracle'],
    selectedProfiles: ['cpu_bound', 'test'],
    executionMode: 'serial',
    failurePolicy: 'continue_by_connection',
    cleanupEnabled: true
  }, helper.placeholderConnections)

  assert.equal(preview.totalItems, 4)
  assert.equal(preview.summary.executionMode, 'serial')
  assert.equal(preview.summary.failurePolicy, 'continue_by_connection')
  assert.equal(preview.summary.cleanupEnabled, true)
  assert.deepEqual(
    preview.items.map((item) => `${item.connectionId}:${item.profileId}`),
    [
      'placeholder-oracle:test',
      'placeholder-oracle:cpu_bound',
      'placeholder-mysql:test',
      'placeholder-mysql:cpu_bound'
    ]
  )
})

test('autobench local plan preview remains empty when local draft has no selected connections or profiles', async () => {
  const helper = await loadHelperModule()

  const previewWithoutConnections = helper.buildLocalPlanPreview({
    ...helper.createAutoBenchWizardDraft(),
    selectedProfiles: ['test']
  }, helper.placeholderConnections)

  assert.equal(previewWithoutConnections.totalItems, 0)
  assert.deepEqual(previewWithoutConnections.items, [])

  const previewWithoutProfiles = helper.buildLocalPlanPreview({
    ...helper.createAutoBenchWizardDraft(),
    selectedConnectionIds: ['placeholder-oracle'],
    selectedProfiles: []
  }, helper.placeholderConnections)

  assert.equal(previewWithoutProfiles.totalItems, 0)
  assert.deepEqual(previewWithoutProfiles.items, [])
})
