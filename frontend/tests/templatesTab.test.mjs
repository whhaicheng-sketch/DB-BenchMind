import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const headerSource = fs.readFileSync(path.resolve(__dirname, '../src/components/template/TemplateHeader.vue'), 'utf8')
const editorDialogSource = fs.readFileSync(path.resolve(__dirname, '../src/components/template/TemplateEditorDialog.vue'), 'utf8')
const basicSectionSource = fs.readFileSync(path.resolve(__dirname, '../src/components/template/TemplateBasicSection.vue'), 'utf8')
const mockTemplatesSource = fs.readFileSync(path.resolve(__dirname, '../src/mock/templates.js'), 'utf8')
const previewSectionSource = fs.readFileSync(path.resolve(__dirname, '../src/components/template/TemplatePreviewSection.vue'), 'utf8')
const runtimeSectionSource = fs.readFileSync(path.resolve(__dirname, '../src/components/template/TemplateRuntimeSection.vue'), 'utf8')
const filterBarSource = fs.readFileSync(path.resolve(__dirname, '../src/components/template/TemplateFilterBar.vue'), 'utf8')
const listItemSource = fs.readFileSync(path.resolve(__dirname, '../src/components/template/TemplateListItem.vue'), 'utf8')
const listSource = fs.readFileSync(path.resolve(__dirname, '../src/components/template/TemplateList.vue'), 'utf8')
const storeSource = fs.readFileSync(path.resolve(__dirname, '../src/stores/template.js'), 'utf8')
const modelSource = fs.readFileSync(path.resolve(__dirname, '../src/models/template.js'), 'utf8')
const bindingModelsSource = fs.readFileSync(path.resolve(__dirname, '../wailsjs/go/models.ts.body'), 'utf8')
const tasksMonitorSource = fs.readFileSync(path.resolve(__dirname, '../src/components/tabs/TasksMonitorTab.vue'), 'utf8')

function loadTemplateModelModule() {
  const executableSource = modelSource.replace(/\bexport\s+/g, '')
  return new Function(`${executableSource}; return { createDefaultTemplate, normalizeTemplateRecord, createPhaseState, PHASE_KEYS };`)()
}

test('template editor dialog uses the shared light theme tokens instead of hard-coded dark panels', () => {
  assert.match(editorDialogSource, /\.editor-dialog\s*\{[\s\S]*background:\s*var\(--bg-primary\)/)
  assert.match(editorDialogSource, /\.dialog-header\s*\{[\s\S]*background:\s*var\(--bg-primary\)/)
  assert.match(editorDialogSource, /\.dialog-body\s*\{[\s\S]*background:\s*var\(--bg-secondary\)/)
  assert.match(editorDialogSource, /\.dialog-footer\s*\{[\s\S]*background:\s*var\(--bg-primary\)/)
  assert.doesNotMatch(editorDialogSource, /#0f172a|#111827|#0b1220|#1e293b|#334155|#f8fafc|#94a3b8/)
})

test('template basic and runtime sections keep inputs on the light dialog surfaces', () => {
  assert.match(basicSectionSource, /\.section-card\s*\{[\s\S]*background:\s*var\(--bg-primary\)/)
  assert.match(basicSectionSource, /\.field-input\s*\{[\s\S]*background:\s*var\(--bg-primary\)/)
  assert.match(runtimeSectionSource, /\.section-card\s*\{[\s\S]*background:\s*var\(--bg-primary\)/)
  assert.match(runtimeSectionSource, /\.field-input\s*\{[\s\S]*background:\s*var\(--bg-primary\)/)
})

test('template preview section stays on the same light surface as the rest of the dialog', () => {
  assert.match(previewSectionSource, /\.section-card\s*\{[\s\S]*background:\s*var\(--bg-primary\)/)
  assert.match(previewSectionSource, /\.summary-item\s*\{[\s\S]*background:\s*var\(--bg-secondary\)/)
  assert.match(previewSectionSource, /\.preview-notes\s*\{[\s\S]*background:\s*var\(--bg-secondary\)/)
  assert.doesNotMatch(previewSectionSource, /rgba\(30,\s*41,\s*59|rgba\(15,\s*23,\s*42|#1f2937|#1e293b|#f8fafc|#cbd5e1/)
})

test('templates page and dialog remove verbose helper copy but keep core sections', () => {
  assert.match(headerSource, />Templates</)
  assert.match(basicSectionSource, />Basic</)
  assert.match(runtimeSectionSource, />Runtime</)
  assert.match(previewSectionSource, />Preview</)
  assert.doesNotMatch(headerSource, /Manage connection-agnostic benchmark scenario templates for future task binding/)
  assert.doesNotMatch(basicSectionSource, /Core identity, scope and applicability for the selected template/)
  assert.doesNotMatch(runtimeSectionSource, /Phase switches, runtime controls and tool-specific benchmark parameters/)
  assert.doesNotMatch(previewSectionSource, /A compact summary of the scenario that will later be bound to a connection in Tasks & Monitor/)
  assert.doesNotMatch(runtimeSectionSource, /High-frequency fields only|Expanded field set placeholder for later phases|Add operator guidance or future backend mapping notes/)
})

test('templates compact layout keeps dialog, filter bar, list, and cards denser than the old roomy version', () => {
  assert.doesNotMatch(editorDialogSource, /padding:\s*24px;/)
  assert.doesNotMatch(editorDialogSource, /width:\s*min\(80vw,\s*1360px\);/)
  assert.doesNotMatch(editorDialogSource, /height:\s*min\(86vh,\s*980px\);/)
  assert.doesNotMatch(filterBarSource, /padding:\s*var\(--spacing-md\);/)
  assert.doesNotMatch(listSource, /padding:\s*var\(--spacing-md\)\s*var\(--spacing-lg\);/)
  assert.doesNotMatch(listItemSource, /padding:\s*var\(--spacing-md\);/)
  assert.doesNotMatch(basicSectionSource, /min-height:\s*72px;/)
  assert.doesNotMatch(previewSectionSource, /padding:\s*12px;/)
})

test('templates empty states and task handoff messaging keep only short action-oriented copy', () => {
  assert.doesNotMatch(listSource, /Create your first benchmark scenario template to start building reusable workload definitions\./)
  assert.doesNotMatch(listSource, /Adjust database or tool filters to broaden the result set\./)
  assert.doesNotMatch(editorDialogSource, /Task shell opened in Tasks & Monitor for template/)
})

test('templates UI no longer renders removed metadata fields or filters', () => {
  assert.doesNotMatch(basicSectionSource, /Template Name[\s\S]*Tags|>Tags<|>Scope<|>Status<|Last Updated/)
  assert.doesNotMatch(editorDialogSource, /scopeLabels|statusLabels|Readonly Shared templates are read-only/)
  assert.doesNotMatch(previewSectionSource, />Scope</)
  assert.doesNotMatch(listItemSource, /status-pill|updated-label|tag scope|template\.tags|template\.status|updatedAt/)
  assert.doesNotMatch(filterBarSource, /Scope \/ Tag|All Scopes|All Tags|filters\.scope|filters\.tag|scopeOptions|tagOptions/)
  assert.doesNotMatch(listSource, /scope or tag filters/i)
})

test('template fallback data ships the 12 built-in database templates with profile metadata', () => {
  assert.equal((mockTemplatesSource.match(/createDefaultTemplate\(/g) || []).length, 12)
  for (const id of [
    'oracle_cpu_bound',
    'oracle_io_bound',
    'oracle_test',
    'mysql_cpu_bound',
    'mysql_io_bound',
    'mysql_test',
    'sqlserver_cpu_bound',
    'sqlserver_io_bound',
    'sqlserver_test',
    'postgresql_cpu_bound',
    'postgresql_io_bound',
    'postgresql_test'
  ]) {
    assert.match(mockTemplatesSource, new RegExp(`id:\\s*'${id}'`))
  }
  assert.match(mockTemplatesSource, /profile_type:\s*'cpu_bound'/)
  assert.match(mockTemplatesSource, /source_alignment:\s*'engineered_split_from_baseline'/)
  assert.match(mockTemplatesSource, /test_position:\s*'smoke'/)
  assert.doesNotMatch(mockTemplatesSource, /\bscope:\s*|\\bstatus:\s*|updatedAt/)
})

test('template model and store only keep compact filters and four lifecycle phases', () => {
  assert.match(modelSource, /export const PHASE_KEYS = \[\s*'prepare',\s*'warmup',\s*'run',\s*'cleanup'\s*\]/)
  assert.match(modelSource, /export function createPhaseState\(overrides = \{\}\)\s*\{\s*return \{\s*prepare: createPhaseConfig\(overrides\.prepare, \{ enabled: false, required: false \}\),\s*warmup: createPhaseConfig\(overrides\.warmup, \{ enabled: false, required: false \}\),\s*run: createPhaseConfig\(overrides\.run, \{ enabled: true, required: true \}\),\s*cleanup: createPhaseConfig\(overrides\.cleanup, \{ enabled: false, required: false \}\)\s*\}\s*\}/s)
  assert.doesNotMatch(modelSource, /scope:\s*partial\.scope|status:\s*partial\.status|updatedAt:\s*partial\.updatedAt/)
  assert.doesNotMatch(modelSource, /build:\s*\{|generate:\s*\{|verify:\s*\{|delete:\s*\{/)
  assert.doesNotMatch(storeSource, /scopeLabels|statusLabels|allTags|filters\.scope|filters\.tag|canEditScope|canDeleteScope|template\.scope|template\.tags|template\.status|updatedAt/)
})

test('template model keeps canonical four phases only as compatibility data', () => {
  assert.doesNotMatch(runtimeSectionSource, /'build'|'generate'|'verify'|'delete'/)
  assert.doesNotMatch(storeSource, /createTaskFromTemplate\(|saveAsTemplate\(/)
})

test('template store only accepts the compact filter whitelist', () => {
  assert.match(storeSource, /const TEMPLATE_FILTER_KEYS = \['search', 'dbFamily', 'tool'\]/)
  assert.match(storeSource, /setFilter\(key, value\)\s*\{\s*if \(!TEMPLATE_FILTER_KEYS\.includes\(key\)\) return\s*this\.filters\[key\] = value/s)
})

test('template editor removes execution phase toggles and keeps runtime parameters only', () => {
  assert.doesNotMatch(runtimeSectionSource, /phase-grid|phase-pill|handlePhaseToggle|visiblePhaseKeys|isPhaseAllowed|isPhaseRequired|updateDraftPhase/)
  assert.doesNotMatch(runtimeSectionSource, /errors\.phaseCombination|errors\.phaseRun/)
  assert.match(runtimeSectionSource, />Concurrency Mode</)
  assert.match(runtimeSectionSource, />Duration \(s\)</)
  assert.match(runtimeSectionSource, />Warm-up \(s\)</)
  assert.match(runtimeSectionSource, />Runtime Notes</)
})

test('template editor collapses to a single form without mode switch and extra footer actions', () => {
  assert.doesNotMatch(editorDialogSource, /Standard|Advanced|Expert|mode-switch|mode-btn|editorMode|editorModes/)
  assert.doesNotMatch(editorDialogSource, /Save As|Create Task from Template|Cancel|Edit Template|Delete/)
  assert.doesNotMatch(editorDialogSource, /class="close-btn"/)
  assert.match(editorDialogSource, /<button\s+class="btn btn-primary"[\s\S]*>\s*Save\s*<\/button>/)
  assert.match(editorDialogSource, /<button class="btn btn-secondary" @click="templateStore\.closeEditor\(\)">Close<\/button>/)
})

test('template preview and editor copy no longer surface removed phase guidance and built-in copy path is explicit', () => {
  assert.doesNotMatch(previewSectionSource, /Lifecycle|phase-grid|phase-pill|Prepare -> Run -> Cleanup/)
  assert.match(editorDialogSource, /Built-in template/)
  assert.match(editorDialogSource, /Copy in the template list/)
})

test('template detail surfaces built-in profile metadata and engineered split notes', () => {
  assert.match(previewSectionSource, /Profile/)
  assert.match(previewSectionSource, /Source Alignment/)
  assert.match(previewSectionSource, /Metrics/)
  assert.match(previewSectionSource, /Prepare/)
  assert.match(previewSectionSource, /Run/)
  assert.match(previewSectionSource, /Cleanup/)
  assert.match(previewSectionSource, /Smoke/)
  assert.match(previewSectionSource, /Engineered/)
})

test('template store opens built-in templates in readonly view, opens custom templates in editing mode, and close only discards local draft state', () => {
  assert.match(storeSource, /openTemplate\(id\)\s*\{[\s\S]*this\.editorState = selected && !selected\.is_builtin \? 'editing' : 'view'/)
  assert.match(storeSource, /openTemplate\(id\)\s*\{[\s\S]*this\.editingTemplateDraft = selected && !selected\.is_builtin \? cloneTemplate\(selected\) : null/)
  const closeEditorBlock = storeSource.match(/closeEditor\(\)\s*\{[\s\S]*?\n    \},\n\n    createTemplate\(\)/)
  assert.ok(closeEditorBlock, 'closeEditor block not found')
  assert.match(closeEditorBlock[0], /this\.isEditorOpen = false[\s\S]*this\.editingTemplateDraft = null[\s\S]*this\.editorState = 'view'[\s\S]*this\.isDirty = false[\s\S]*this\.validationErrors = \{\}/)
  assert.doesNotMatch(closeEditorBlock[0], /saveTemplate\(/)
})

test('template list actions remove the built-in delete placeholder instead of rendering a disabled empty slot', () => {
  assert.match(listItemSource, /const primaryActionLabel = computed\(\(\) => \(props\.template\.is_builtin \? 'View' : 'Edit'\)\)/)
  assert.match(listItemSource, /title="Open template" @click="\$emit\('open'\)">{{ primaryActionLabel }}<\/button>/)
  assert.match(listItemSource, /<button class="btn-action" title="Duplicate" @click="\$emit\('duplicate'\)">Copy<\/button>/)
  assert.match(listItemSource, /v-if="!template\.is_builtin"/)
  assert.doesNotMatch(listItemSource, /:disabled="template\.is_builtin"/)
})

test('template store drops editor mode, phase toggle, save-as, and task handoff helpers', () => {
  assert.doesNotMatch(storeSource, /editorMode|cancelEditing\(|updateDraftPhase\(|saveAsTemplate\(|createTaskFromTemplate\(|placeholderAction\(/)
  assert.doesNotMatch(storeSource, /phaseRun|phaseCombination/)
})

test('performance analysis action options are fixed and no longer derived from template phase toggles', () => {
  assert.match(tasksMonitorSource, /const actionOptions = computed\(\(\) => \[\s*\{ value: 'prepare', label: 'Prepare', disabled: false \},\s*\{ value: 'run', label: 'Run', disabled: false \},\s*\{ value: 'cleanup', label: 'Cleanup', disabled: false \},\s*\{ value: 'full_pipeline', label: 'Run Full Flow \(Prepare -> Run -> Cleanup\)', disabled: false \}\s*\]\)/s)
  assert.doesNotMatch(tasksMonitorSource, /selectedTemplate\.value\?\.phases|allowFullPipeline|phases\.prepare|phases\.run|phases\.cleanup/)
})

test('template phase normalization backfills missing canonical phases before runtime reads', () => {
  const { normalizeTemplateRecord, PHASE_KEYS } = loadTemplateModelModule()
  const normalized = normalizeTemplateRecord({
    id: 'tpl_legacy_partial',
    name: 'Legacy Partial Template',
    tool: 'sysbench',
    dbFamily: 'mysql',
    workloadFamily: 'oltp-read-write',
    phases: {
      prepare: { enabled: true, required: false, params: { seed: true } },
      cleanup: { enabled: true, required: false, params: {} }
    }
  })

  assert.deepEqual(Object.keys(normalized.phases), PHASE_KEYS)
  assert.deepEqual(normalized.phases.prepare, { enabled: true, required: false, params: { seed: true } })
  assert.deepEqual(normalized.phases.warmup, { enabled: true, required: false, params: {} })
  assert.deepEqual(normalized.phases.run, { enabled: true, required: true, params: {} })
  assert.deepEqual(normalized.phases.cleanup, { enabled: true, required: false, params: {} })
})

test('default template creation tolerates undefined run phase input and restores a safe run config', () => {
  const { createDefaultTemplate } = loadTemplateModelModule()
  const template = createDefaultTemplate({
    tool: 'sysbench',
    dbFamily: 'mysql',
    workloadFamily: 'oltp-read-write',
    phases: {
      prepare: { enabled: true, required: false, params: {} },
      run: undefined
    }
  })

  assert.deepEqual(template.phases.run, { enabled: true, required: true, params: {} })
  assert.deepEqual(template.phases.warmup, { enabled: true, required: false, params: {} })
})

test('wails template models drop removed metadata fields from the frontend binding surface', () => {
  const templateDtoBlock = bindingModelsSource.match(/export class TemplateDTO \{[\s\S]*?export class TemplateListResult/)
  const phaseSetBlock = bindingModelsSource.match(/export class PhaseSet \{[\s\S]*?export class Runtime/)

  assert.ok(templateDtoBlock, 'TemplateDTO block not found')
  assert.ok(phaseSetBlock, 'PhaseSet block not found')
  assert.match(templateDtoBlock[0], /profile_type:\s*string;/)
  assert.match(templateDtoBlock[0], /source_alignment:\s*string;/)
  assert.match(templateDtoBlock[0], /prepare_config:\s*Record<string, any>;/)
  assert.match(templateDtoBlock[0], /run_config:\s*Record<string, any>;/)
  assert.match(templateDtoBlock[0], /cleanup_config:\s*Record<string, any>;/)
  assert.match(templateDtoBlock[0], /metrics:\s*string\[];/)
  assert.match(templateDtoBlock[0], /tags:\s*string\[];/)
  assert.doesNotMatch(templateDtoBlock[0], /scope:\s*string;|status:\s*string;|updatedAt:\s*string;/)
  assert.doesNotMatch(phaseSetBlock[0], /build:\s*PhaseConfig;|generate:\s*PhaseConfig;|verify:\s*PhaseConfig;|delete:\s*PhaseConfig;/)
})
