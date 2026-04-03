import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const autoBenchSource = fs.readFileSync(path.resolve(__dirname, '../src/components/tabs/AutoBenchTab.vue'), 'utf8')

test('T12.2: autobench renders active run section with status and metrics', () => {
  assert.match(autoBenchSource, /active-run-section/)
  assert.match(autoBenchSource, /run-strip/)
  assert.match(autoBenchSource, /status-badge/)
  assert.match(autoBenchSource, /status-metrics/)
  assert.match(autoBenchSource, /metric-label/)
  assert.match(autoBenchSource, /metric-value/)
  assert.match(autoBenchSource, /progress-bar-container/)
  assert.match(autoBenchSource, /progress-bar/)
  assert.match(autoBenchSource, /progress-label/)
  assert.match(autoBenchSource, /items-list/)
  assert.match(autoBenchSource, /item-row/)
  assert.match(autoBenchSource, /item-status/)
})

test('T12.2: autobench displays item-level status with connection, type, and phase', () => {
  assert.match(autoBenchSource, /items-header/)
  assert.match(autoBenchSource, />Connection</)
  assert.match(autoBenchSource, />Type</)
  assert.match(autoBenchSource, />Status</)
  assert.match(autoBenchSource, />Phase</)
  assert.match(autoBenchSource, /item-connection/)
  assert.match(autoBenchSource, /item-type/)
})

test('T12.2: autobench has progress bar with percentage display', () => {
  assert.match(autoBenchSource, /class="progress-bar-container"/)
  assert.match(autoBenchSource, /class="progress-bar"/)
  assert.match(autoBenchSource, /:style=.*width.*progress/)
  assert.match(autoBenchSource, /suiteSummary\.progress/)
})

test('T12.2: autobench has status badge styling for success/error/running states', () => {
  assert.match(autoBenchSource, /getStatusClass/)
  assert.match(autoBenchSource, /status-success/)
  assert.match(autoBenchSource, /status-error/)
  assert.match(autoBenchSource, /status-running/)
  assert.match(autoBenchSource, /status-warning/)
})

test('T12.3: autobench has handleRerunFailed function for re-running failed items', () => {
  assert.match(autoBenchSource, /handleRerunFailed/)
  assert.match(autoBenchSource, /AutoBenchBinding\.RerunFailed/)
})

test('T12.3: autobench does NOT render View Report button in item rows (P16)', () => {
  // View Report button was removed per P16 - no individual report column in items table
  assert.doesNotMatch(autoBenchSource, /View Report/)
  assert.doesNotMatch(autoBenchSource, /item-report/)
  assert.doesNotMatch(autoBenchSource, /link-button/)
})

test('T12.3: autobench renders Re-run Failed button when suite has failed items', () => {
  assert.match(autoBenchSource, /Re-run Failed/)
  assert.match(autoBenchSource, /hasFailedItems/)
  assert.match(autoBenchSource, /@click="handleRerunFailed"/)
})

test('T12.1: autobench uses real connections from store', () => {
  assert.match(autoBenchSource, /useConnectionStore/)
  assert.match(autoBenchSource, /realConnections/)
  assert.match(autoBenchSource, /connectionStore\.connections/)
  assert.match(autoBenchSource, /connectionStore\.fetchConnections/)
})

test('T12.1: autobench loads connections on mount', () => {
  assert.match(autoBenchSource, /onMounted/)
  assert.match(autoBenchSource, /await connectionStore\.fetchConnections/)
})

test('M10/T10.2: autobench has CreateSuite API binding', () => {
  assert.match(autoBenchSource, /AutoBenchBinding/)
  assert.match(autoBenchSource, /AutoBenchBinding\.CreateSuite/)
  assert.match(autoBenchSource, /handleCreateSuite/)
  assert.match(autoBenchSource, /canCreateSuite/)
})

test('M11/T11.3: autobench has StartSuite API binding', () => {
  assert.match(autoBenchSource, /AutoBenchBinding\.StartSuite/)
  assert.match(autoBenchSource, /handleStartSuite/)
  assert.match(autoBenchSource, /canStartSuite/)
})

test('M11/T11.3: autobench has GetSuiteStatus API binding for polling', () => {
  assert.match(autoBenchSource, /AutoBenchBinding\.GetSuiteStatus/)
  assert.match(autoBenchSource, /fetchSuiteStatus/)
  assert.match(autoBenchSource, /pollSuiteStatus/)
  assert.match(autoBenchSource, /isPolling/)
})

test('T12.2: autobench has computed suiteSummary with progress calculation', () => {
  assert.match(autoBenchSource, /const suiteSummary = computed/)
  assert.match(autoBenchSource, /total_items/)
  assert.match(autoBenchSource, /completed_items/)
  assert.match(autoBenchSource, /pending_items/)
  assert.match(autoBenchSource, /running_items/)
})

test('T12.2: autobench button states for create/start/running', () => {
  assert.match(autoBenchSource, /:disabled="!canCreateSuite"/)
  assert.match(autoBenchSource, /:disabled="!canStartSuite"/)
  assert.match(autoBenchSource, /isSuiteRunning/)
  assert.match(autoBenchSource, /resetSuite/)
  assert.match(autoBenchSource, /New Suite/)
})

test('T12.2: autobench has error handling for create and start operations', () => {
  assert.match(autoBenchSource, /createError/)
  assert.match(autoBenchSource, /startError/)
  assert.match(autoBenchSource, /error-banner/)
})

test('T12.2: autobench cleans up polling on unmount', () => {
  assert.match(autoBenchSource, /onUnmounted/)
  assert.match(autoBenchSource, /clearTimeout/)
})

test('T12.2: autobench displays connection type filter with database type options', () => {
  assert.match(autoBenchSource, /connectionFilterOptions/)
  assert.match(autoBenchSource, /activeConnectionFilter/)
  assert.match(autoBenchSource, /filteredConnections/)
})

test('T12.2: autobench uses wizard draft for connection and profile selection', () => {
  assert.match(autoBenchSource, /createAutoBenchWizardDraft/)
  assert.match(autoBenchSource, /toggleDraftConnectionSelection/)
  assert.match(autoBenchSource, /toggleDraftProfileSelection/)
  assert.match(autoBenchSource, /validateAutoBenchWizardDraft/)
  assert.match(autoBenchSource, /profileOptions/)
  assert.match(autoBenchSource, /selectedProfiles/)
  assert.match(autoBenchSource, /selectedConnectionIds/)
})

// UI Overhaul tests
test('UI Overhaul: autobench has two-column wizard layout', () => {
  assert.match(autoBenchSource, /wizard-columns/)
  assert.match(autoBenchSource, /grid-template-columns: 1fr 1fr/)
})

test('UI Overhaul: autobench uses compact connection rows instead of cards', () => {
  assert.match(autoBenchSource, /conn-row/)
  assert.match(autoBenchSource, /conn-row-name/)
  assert.match(autoBenchSource, /conn-row-type/)
  assert.match(autoBenchSource, /conn-row-host/)
})

test('UI Overhaul: autobench uses profile toggle pills instead of cards', () => {
  assert.match(autoBenchSource, /profile-toggle/)
  assert.match(autoBenchSource, /profile-toggles/)
})

test('UI Overhaul: autobench has active run section with left border strip', () => {
  assert.match(autoBenchSource, /active-run-section/)
  assert.match(autoBenchSource, /border-left: 4px solid var\(--primary\)/)
  assert.match(autoBenchSource, /run-strip/)
})

test('UI Overhaul: autobench has elapsed time tracking', () => {
  assert.match(autoBenchSource, /elapsedSeconds/)
  assert.match(autoBenchSource, /formatElapsed/)
})

test('UI Overhaul: autobench has currentItem computed for running item', () => {
  assert.match(autoBenchSource, /const currentItem = computed/)
})

test('UI Overhaul: autobench resolves connection IDs to names via connNameMap', () => {
  assert.match(autoBenchSource, /connNameMap/)
})
