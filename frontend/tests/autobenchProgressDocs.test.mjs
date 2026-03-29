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

test('AutoBench progress json keeps the required fixed schema', () => {
  const progress = readJson(progressJsonPath)

  // Verify required fields exist (core fields)
  const requiredFields = [
    'blocked_tasks',
    'changed_files',
    'current_module',
    'current_task',
    'done_tasks',
    'known_risks',
    'next_task',
    'test_results'
  ]

  for (const field of requiredFields) {
    assert.ok(field in progress, `Should have required field: ${field}`)
  }

  // Verify types (current_module/task/next_task may be null when completed)
  if (progress.current_module != null) {
    assert.ok(typeof progress.current_module === 'string' && progress.current_module.length > 0)
  }
  if (progress.current_task != null) {
    assert.ok(typeof progress.current_task === 'string' && progress.current_task.length > 0)
  }
  if (progress.next_task != null) {
    assert.ok(typeof progress.next_task === 'string' && progress.next_task.length > 0)
  }
  assert.ok(Array.isArray(progress.done_tasks))
  assert.ok(Array.isArray(progress.blocked_tasks))
  assert.ok(Array.isArray(progress.changed_files))
  assert.ok(Array.isArray(progress.test_results))
  assert.ok(Array.isArray(progress.known_risks))
})

test('AutoBench progress json ensures done_tasks includes current_task and does not include next_task', () => {
  const progress = readJson(progressJsonPath)

  // When project is completed, current_task and next_task are null
  if (progress.current_task != null) {
    assert.ok(progress.done_tasks.includes(progress.current_task), 'done_tasks should include current_task')
  }

  // Special case: when completed, next_task is "COMPLETED" which is not a task ID
  if (progress.next_task != null && progress.next_task !== 'COMPLETED') {
    assert.ok(!progress.done_tasks.includes(progress.next_task), 'done_tasks should not include next_task')
  }
})

test('AutoBench progress json done_tasks entries follow T{module}.{task} pattern', () => {
  const progress = readJson(progressJsonPath)
  const taskPattern = /^T\d+\.\d+$/

  for (const task of progress.done_tasks) {
    assert.ok(taskPattern.test(task), `Task "${task}" should match pattern T{module}.{task}`)
  }
})

test('AutoBench progress markdown records the current round summary and task structure', () => {
  const progress = readJson(progressJsonPath)
  const source = fs.readFileSync(progressMdPath, 'utf8')

  // Verify markdown header structure (may be in Chinese or English)
  assert.match(source, /^# AutoBench/m)

  // When project is completed, current_module/task/next_task are null
  // so we skip the field matching
  if (progress.current_module != null) {
    const moduleMatch = source.match(/当前模块:\s*(\S+)/)
    const taskMatch = source.match(/当前任务:\s*(\S+)/)
    const nextMatch = source.match(/下一任务:\s*(\S+)/)

    assert.ok(moduleMatch, 'Should have current module field')
    assert.ok(taskMatch, 'Should have current task field')
    assert.ok(nextMatch, 'Should have next task field')

    assert.equal(moduleMatch[1], progress.current_module)
    assert.equal(taskMatch[1], progress.current_task)
    assert.equal(nextMatch[1], progress.next_task)
  }

  // Verify task references exist in the markdown
  // When completed, the summary may only reference a subset of tasks
  const taskRefs = progress.done_tasks.filter(t => source.includes(t))
  assert.ok(taskRefs.length > 0, 'Markdown should mention at least some task IDs')
})

test('AutoBench progress json completion_summary exists when project is completed', () => {
  const progress = readJson(progressJsonPath)

  // If next_task is COMPLETED, completion_summary should exist
  if (progress.next_task === 'COMPLETED') {
    assert.ok('completion_summary' in progress, 'Should have completion_summary when completed')
    assert.equal(progress.completion_summary.status, 'completed')
    assert.ok(Array.isArray(progress.completion_summary.modules_completed))
    assert.ok(typeof progress.completion_summary.total_tasks === 'number')
  }
})
