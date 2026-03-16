import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const taskBindingPath = path.resolve(__dirname, '../src/lib/taskBinding.js')

test('task binding only exposes in-app log retrieval and no system log opener', () => {
  const source = fs.readFileSync(taskBindingPath, 'utf8')

  assert.match(source, /export function getTaskLogs/)
  assert.doesNotMatch(source, /export function openTaskLog/)
  assert.doesNotMatch(source, /openPathInSystemViewer/)
})
