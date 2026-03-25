import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const autoBenchSource = fs.readFileSync(path.resolve(__dirname, '../src/components/tabs/AutoBenchTab.vue'), 'utf8')

test('autobench page renders dedicated wizard monitor and report sections', () => {
  assert.match(autoBenchSource, /<h1 class="page-title">AutoBench<\/h1>/)
  assert.match(autoBenchSource, /<section class="autobench-section autobench-wizard"/)
  assert.match(autoBenchSource, /<section class="autobench-section autobench-monitor"/)
  assert.match(autoBenchSource, /<section class="autobench-section autobench-report"/)
  assert.match(autoBenchSource, />Wizard<\/h2>/)
  assert.match(autoBenchSource, />Monitor<\/h2>/)
  assert.match(autoBenchSource, />Report<\/h2>/)
})

test('autobench wizard renders local-only connections profiles and execution policy groups', () => {
  assert.match(autoBenchSource, />Selected Connections<\/h3>/)
  assert.match(autoBenchSource, />Database Type Filter<\/h4>/)
  assert.match(autoBenchSource, />Profiles<\/h3>/)
  assert.match(autoBenchSource, />Profile Scope<\/h4>/)
  assert.match(autoBenchSource, />Plan Preview<\/h3>/)
  assert.match(autoBenchSource, />Execution Policy<\/h3>/)
  assert.match(autoBenchSource, />Suite Progress<\/h3>/)
  assert.match(autoBenchSource, />Current Item<\/h3>/)
  assert.match(autoBenchSource, />Item Status<\/h3>/)
  assert.match(autoBenchSource, /placeholderConnections/)
  assert.match(autoBenchSource, /connectionFilterOptions/)
  assert.match(autoBenchSource, /activeConnectionFilter/)
  assert.match(autoBenchSource, /filteredConnections/)
  assert.match(autoBenchSource, /profileOptions/)
  assert.match(autoBenchSource, /selectedProfileSummary/)
  assert.match(autoBenchSource, /planPreview/)
  assert.match(autoBenchSource, /buildLocalPlanPreview/)
  assert.match(autoBenchSource, /buildAutoBenchMonitorState/)
  assert.match(autoBenchSource, /monitorState/)
  assert.match(autoBenchSource, /monitorState\.itemRows/)
  assert.match(autoBenchSource, /policySummaryItems/)
  assert.doesNotMatch(autoBenchSource, /connectionStore|templateStore|monitorStore/)
})

test('autobench copy stays independent from performance analysis and real task monitor controls', () => {
  assert.doesNotMatch(autoBenchSource, />Tasks & Monitor<\/h1>/)
  assert.doesNotMatch(autoBenchSource, /Performance Analysis/i)
  assert.doesNotMatch(autoBenchSource, />Start<\/button>/)
  assert.doesNotMatch(autoBenchSource, />Stop<\/button>/)
})

test('autobench actions remain safe placeholders with no executable runtime wiring', () => {
  assert.match(autoBenchSource, /<button class="placeholder-action" type="button" disabled>Create Suite \(later task\)<\/button>/)
  assert.doesNotMatch(autoBenchSource, /CreateSuite|GetSuiteStatus|window\.go|Start AutoBench|BenchmarkTask/)
})

test('autobench skeleton uses local draft state so the page can render without backend data', () => {
  assert.match(autoBenchSource, /const draft = ref\(createAutoBenchWizardDraft\(\)\)/)
  assert.match(autoBenchSource, /toggleConnectionSelection/)
  assert.match(autoBenchSource, /toggleProfileSelection/)
  assert.match(autoBenchSource, /const monitorPlaceholder = /)
  assert.match(autoBenchSource, /const reportPlaceholder = /)
  assert.match(autoBenchSource, /const wizardValidation = /)
  assert.doesNotMatch(autoBenchSource, /onMounted\(|watch\(|fetch\(|async setup|window\.runtime/)
})
