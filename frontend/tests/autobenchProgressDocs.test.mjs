import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const progressJsonPath = path.resolve(__dirname, '../../docs/AutoBench/progress.json')
const progressMdPath = path.resolve(__dirname, '../../docs/AutoBench/progress.md')

function readJson(filePath) {
  return JSON.parse(fs.readFileSync(filePath, 'utf8'))
}

test('AutoBench progress json keeps the required fixed schema and advances one next task only', () => {
  const progress = readJson(progressJsonPath)

  assert.deepEqual(Object.keys(progress).sort(), [
    'blocked_tasks',
    'changed_files',
    'current_module',
    'current_task',
    'done_tasks',
    'known_risks',
    'next_task',
    'test_results'
  ])
  assert.equal(progress.current_module, 'M3')
  assert.equal(progress.current_task, 'T3.2')
  assert.deepEqual(progress.done_tasks, ['T0.1', 'T0.2', 'T1.1', 'T1.2', 'T1.3', 'T2.1', 'T2.2', 'T2.3', 'T3.1', 'T3.2'])
  assert.equal(progress.next_task, 'T3.3')
  assert.deepEqual(progress.blocked_tasks, [])
  assert.ok(Array.isArray(progress.changed_files))
  assert.ok(Array.isArray(progress.test_results))
  assert.ok(Array.isArray(progress.known_risks))
})

test('AutoBench progress markdown records the current round summary and next task', () => {
  const source = fs.readFileSync(progressMdPath, 'utf8')

  assert.match(source, /^# AutoBench Progress/m)
  assert.match(source, /当前模块:\s*M3/)
  assert.match(source, /当前任务:\s*T3\.2/)
  assert.match(source, /下一任务:\s*T3\.3/)
  assert.match(source, /已完成任务:\s*T0\.1,\s*T0\.2,\s*T1\.1,\s*T1\.2,\s*T1\.3,\s*T2\.1,\s*T2\.2,\s*T2\.3,\s*T3\.1,\s*T3\.2/)
})
