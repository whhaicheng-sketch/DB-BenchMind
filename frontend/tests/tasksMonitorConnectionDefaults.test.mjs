import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

import { getPreferredConnectionId } from '../src/components/tabs/tasksMonitorConnectionDefaults.mjs'

const tasksMonitorTabSource = fs.readFileSync(
  path.resolve('frontend/src/components/tabs/TasksMonitorTab.vue'),
  'utf8'
)

test('defaults to the first available connection when current selection is empty', () => {
  const preferredConnectionId = getPreferredConnectionId([
    { id: 'mysql-a', name: 'MySQL A' },
    { id: 'mysql-b', name: 'MySQL B' }
  ], '')

  assert.equal(preferredConnectionId, 'mysql-a')
})

test('returns empty when there are no candidate connections', () => {
  assert.equal(getPreferredConnectionId([], ''), '')
})

test('falls back to the first candidate when current selection is invalid', () => {
  const preferredConnectionId = getPreferredConnectionId([
    { id: 'oracle-a', name: 'Oracle A' },
    { id: 'oracle-b', name: 'Oracle B' }
  ], 'mysql-a')

  assert.equal(preferredConnectionId, 'oracle-a')
})

test('keeps the current connection when it remains valid', () => {
  const preferredConnectionId = getPreferredConnectionId([
    { id: 'pg-a', name: 'Postgres A' },
    { id: 'pg-b', name: 'Postgres B' }
  ], 'pg-b')

  assert.equal(preferredConnectionId, 'pg-b')
})

test('connection placeholder uses database type when no candidate connection exists', () => {
  assert.match(tasksMonitorTabSource, /Select one/)
  assert.doesNotMatch(tasksMonitorTabSource, /Select connection/)
})
