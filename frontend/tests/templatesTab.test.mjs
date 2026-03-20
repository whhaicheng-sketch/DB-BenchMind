import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { templateMocks } from '../src/mock/templates.js'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const editorDialogSource = fs.readFileSync(path.resolve(__dirname, '../src/components/template/TemplateEditorDialog.vue'), 'utf8')
const basicSectionSource = fs.readFileSync(path.resolve(__dirname, '../src/components/template/TemplateBasicSection.vue'), 'utf8')
const previewSectionSource = fs.readFileSync(path.resolve(__dirname, '../src/components/template/TemplatePreviewSection.vue'), 'utf8')
const runtimeSectionSource = fs.readFileSync(path.resolve(__dirname, '../src/components/template/TemplateRuntimeSection.vue'), 'utf8')

test('template editor dialog uses the shared light theme tokens instead of hard-coded dark panels', () => {
  assert.match(editorDialogSource, /\.editor-dialog\s*\{[\s\S]*background:\s*var\(--bg-primary\)/)
  assert.match(editorDialogSource, /\.dialog-header\s*\{[\s\S]*background:\s*var\(--bg-primary\)/)
  assert.match(editorDialogSource, /\.dialog-body\s*\{[\s\S]*background:\s*var\(--bg-secondary\)/)
  assert.match(editorDialogSource, /\.dialog-footer\s*\{[\s\S]*background:\s*var\(--bg-primary\)/)
  assert.match(editorDialogSource, /\.mode-switch\s*\{[\s\S]*background:\s*var\(--bg-tertiary\)/)
  assert.match(editorDialogSource, /\.mode-btn\.active\s*\{[\s\S]*background:\s*var\(--bg-primary\)/)
  assert.doesNotMatch(editorDialogSource, /#0f172a|#111827|#0b1220|#1e293b|#334155|#f8fafc|#94a3b8|rgba\(15,\s*23,\s*42|rgba\(10,\s*15,\s*27/)
})

test('template basic and runtime sections keep inputs on the light dialog surfaces', () => {
  assert.match(basicSectionSource, /\.section-card\s*\{[\s\S]*background:\s*var\(--bg-primary\)/)
  assert.match(basicSectionSource, /\.field-input\s*\{[\s\S]*background:\s*var\(--bg-primary\)/)
  assert.match(runtimeSectionSource, /\.section-card\s*\{[\s\S]*background:\s*var\(--bg-primary\)/)
  assert.match(runtimeSectionSource, /\.phase-pill\s*\{[\s\S]*background:\s*var\(--bg-secondary\)/)
  assert.match(runtimeSectionSource, /\.field-input\s*\{[\s\S]*background:\s*var\(--bg-primary\)/)
})

test('template preview section stays on the same light surface as the rest of the dialog', () => {
  assert.match(previewSectionSource, /\.section-card\s*\{[\s\S]*background:\s*var\(--bg-primary\)/)
  assert.match(previewSectionSource, /\.summary-item\s*\{[\s\S]*background:\s*var\(--bg-secondary\)/)
  assert.match(previewSectionSource, /\.preview-notes\s*\{[\s\S]*background:\s*var\(--bg-secondary\)/)
  assert.doesNotMatch(previewSectionSource, /rgba\(30,\s*41,\s*59|rgba\(15,\s*23,\s*42|#1f2937|#1e293b|#f8fafc|#cbd5e1/)
})

test('template fallback data keeps only the single default test template', () => {
  assert.equal(templateMocks.length, 1)
  assert.equal(templateMocks[0].id, 'tpl_test_mysql_sysbench')
  assert.equal(templateMocks[0].scope, 'test')
})
