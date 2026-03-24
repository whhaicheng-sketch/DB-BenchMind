import test from 'node:test'
import assert from 'node:assert/strict'

import {
  getPreferredTemplateId,
  isTestTemplate
} from '../src/components/tabs/tasksMonitorTemplateSelection.mjs'

function buildTemplate(overrides = {}) {
  return {
    id: 'tpl-default',
    name: 'Default Template',
    profile_type: '',
    tags: [],
    ...overrides
  }
}

test('prefers the database test template when profile_type marks it as test', () => {
  const templates = [
    buildTemplate({ id: 'mysql-cpu', name: 'MySQL CPU Bound', profile_type: 'cpu_bound' }),
    buildTemplate({ id: 'mysql-test', name: 'MySQL Test', profile_type: 'test' }),
    buildTemplate({ id: 'mysql-io', name: 'MySQL IO Bound', profile_type: 'io_bound' })
  ]

  assert.equal(getPreferredTemplateId(templates, { fallbackTemplateId: 'mysql-cpu' }), 'mysql-test')
})

test('falls back to the existing selection when no test template exists', () => {
  const templates = [
    buildTemplate({ id: 'oracle-cpu', name: 'Oracle CPU Bound', profile_type: 'cpu_bound' }),
    buildTemplate({ id: 'oracle-io', name: 'Oracle IO Bound', profile_type: 'io_bound' })
  ]

  assert.equal(getPreferredTemplateId(templates, { fallbackTemplateId: 'oracle-io' }), 'oracle-io')
  assert.equal(getPreferredTemplateId(templates, { fallbackTemplateId: 'missing-template' }), '')
})

test('uses tags when profile_type is unavailable', () => {
  assert.equal(isTestTemplate(buildTemplate({
    id: 'sqlserver-smoke',
    name: 'SQL Server Smoke',
    profile_type: '',
    tags: ['builtin', 'test', 'minimal']
  })), true)
})

test('falls back to stable name or id matching when structured test metadata is absent', () => {
  assert.equal(isTestTemplate(buildTemplate({
    id: 'postgresql_test_template',
    name: 'PostgreSQL Baseline'
  })), true)

  assert.equal(isTestTemplate(buildTemplate({
    id: 'postgresql-baseline',
    name: 'PostgreSQL Test : sysbench'
  })), true)
})
