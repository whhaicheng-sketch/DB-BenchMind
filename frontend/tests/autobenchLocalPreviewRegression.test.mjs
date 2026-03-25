import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const helperPath = path.resolve(__dirname, '../src/components/tabs/autobenchWizardDraft.mjs')
const tabPath = path.resolve(__dirname, '../src/components/tabs/AutoBenchTab.vue')

async function loadHelperModule() {
  return import(pathToFileURL(helperPath).href)
}

function buildPreview(helper, selectedConnectionIds, selectedProfiles) {
  return helper.buildLocalPlanPreview({
    ...helper.createAutoBenchWizardDraft(),
    selectedConnectionIds,
    selectedProfiles
  }, helper.placeholderConnections)
}

test('local preview stays empty until both connection and profile selections are present', async () => {
  const helper = await loadHelperModule()

  const withoutConnections = buildPreview(helper, [], ['test'])
  assert.equal(withoutConnections.totalItems, 0)
  assert.deepEqual(withoutConnections.items, [])

  const withoutProfiles = buildPreview(helper, ['placeholder-oracle'], [])
  assert.equal(withoutProfiles.totalItems, 0)
  assert.deepEqual(withoutProfiles.items, [])

  const withBoth = buildPreview(helper, ['placeholder-oracle'], ['test', 'cpu_bound'])
  assert.equal(withBoth.totalItems, 2)
  assert.deepEqual(
    withBoth.items.map((item) => item.id),
    ['placeholder-oracle:test', 'placeholder-oracle:cpu_bound']
  )
})

test('local preview count and ordering remain stable across connection and profile toggles', async () => {
  const helper = await loadHelperModule()
  let draft = helper.createAutoBenchWizardDraft()

  draft = helper.toggleDraftConnectionSelection(draft, 'placeholder-mysql')
  draft = helper.toggleDraftConnectionSelection(draft, 'placeholder-oracle')

  let preview = helper.buildLocalPlanPreview(draft, helper.placeholderConnections)
  assert.equal(preview.totalItems, 6)
  assert.deepEqual(
    preview.items.map((item) => `${item.connectionId}:${item.profileId}`),
    [
      'placeholder-oracle:test',
      'placeholder-oracle:cpu_bound',
      'placeholder-oracle:io_bound',
      'placeholder-mysql:test',
      'placeholder-mysql:cpu_bound',
      'placeholder-mysql:io_bound'
    ]
  )
  assert.deepEqual(preview.items.map((item) => item.order), [1, 2, 3, 4, 5, 6])

  draft = helper.toggleDraftConnectionSelection(draft, 'placeholder-oracle')
  preview = helper.buildLocalPlanPreview(draft, helper.placeholderConnections)
  assert.equal(preview.totalItems, 3)
  assert.deepEqual(
    preview.items.map((item) => `${item.connectionId}:${item.profileId}`),
    [
      'placeholder-mysql:test',
      'placeholder-mysql:cpu_bound',
      'placeholder-mysql:io_bound'
    ]
  )

  draft = helper.toggleDraftProfileSelection(draft, 'cpu_bound')
  preview = helper.buildLocalPlanPreview(draft, helper.placeholderConnections)
  assert.equal(preview.totalItems, 2)
  assert.deepEqual(
    preview.items.map((item) => `${item.connectionId}:${item.profileId}`),
    ['placeholder-mysql:test', 'placeholder-mysql:io_bound']
  )

  draft = helper.toggleDraftConnectionSelection(draft, 'placeholder-oracle')
  draft = helper.toggleDraftProfileSelection(draft, 'cpu_bound')
  preview = helper.buildLocalPlanPreview(draft, helper.placeholderConnections)
  assert.equal(preview.totalItems, 6)
  assert.deepEqual(
    preview.items.map((item) => `${item.connectionId}:${item.profileId}`),
    [
      'placeholder-oracle:test',
      'placeholder-oracle:cpu_bound',
      'placeholder-oracle:io_bound',
      'placeholder-mysql:test',
      'placeholder-mysql:cpu_bound',
      'placeholder-mysql:io_bound'
    ]
  )
})

test('connection filtering stays local-only and does not change selected preview semantics', async () => {
  const helper = await loadHelperModule()

  const mysqlOnly = helper.filterPlaceholderConnections(helper.placeholderConnections, 'mysql')
  assert.deepEqual(mysqlOnly.map((item) => item.id), ['placeholder-mysql'])

  const preview = buildPreview(helper, ['placeholder-mysql', 'placeholder-sqlserver'], ['io_bound', 'test'])
  assert.equal(preview.totalItems, 4)
  assert.deepEqual(
    preview.items.map((item) => `${item.connectionId}:${item.profileId}`),
    [
      'placeholder-mysql:test',
      'placeholder-mysql:io_bound',
      'placeholder-sqlserver:test',
      'placeholder-sqlserver:io_bound'
    ]
  )
})

test('AutoBench tab keeps preview copy local-only and leaves create action disabled', async () => {
  const source = fs.readFileSync(tabPath, 'utf8')

  assert.match(source, /{{ planPreview\.totalItems }} local items/)
  assert.match(source, /Select at least one connection and one profile to build the local preview\./)
  assert.match(source, /Preview only\. This is a local orchestration sketch and does not create or run a suite\./)
  assert.match(source, /<button class="placeholder-action" type="button" disabled>Create Suite \(later task\)<\/button>/)
  assert.doesNotMatch(source, /CreateSuite|AutoBenchSuiteUseCase|window\.go/)
})
