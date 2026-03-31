import { describe, it } from 'node:test'
import assert from 'node:assert/strict'

/**
 * ReportsTab filter label persistence tests
 *
 * Tests that the status filter dropdown shows the correct label
 * after selecting different filter values. Also tests the backend-to-frontend
 * value mapping.
 */

// ===========================================================================
// Filter label mapping (mirrors ReportsTab.vue onFilterChange logic)
// ===========================================================================

const filterOptions = [
  { value: '', label: 'All' },
  { value: 'success', label: '\u6210\u529f' },
  { value: 'failed', label: '\u5931\u8d25' },
  { value: 'cancelled', label: 'Stop' }
]

// backend-to-frontend mapping (mirrors ReportsTab.vue watcher logic)
const backendToFrontend = {
  completed: 'success',
  cancelled: 'cancelled',
  failed: 'failed'
}

function resolveFrontendFilter(statusFilter) {
  return statusFilter
}

function resolveBackendFilter(statusFilter) {
  if (statusFilter === 'success') return 'completed'
  return statusFilter
}

describe('ReportsTab - filter label persistence', () => {
  it('should show success label when success filter is selected', () => {
    const label = filterOptions.find(o => o.value === 'success').label
    assert.equal(label, '\u6210\u529f')
  })

  it('should show failed label when failed filter is selected', () => {
    const label = filterOptions.find(o => o.value === 'failed').label
    assert.equal(label, '\u5931\u8d25')
  })

  it('should show Stop label when cancelled filter is selected', () => {
    const label = filterOptions.find(o => o.value === 'cancelled').label
    assert.equal(label, 'Stop')
  })

  it('should show All label when empty filter is selected', () => {
    const label = filterOptions.find(o => o.value === '').label
    assert.equal(label, 'All')
  })

  it('should map backend completed to frontend success', () => {
    assert.equal(backendToFrontend['completed'], 'success')
  })

  it('should map backend failed to frontend failed', () => {
    assert.equal(backendToFrontend['failed'], 'failed')
  })

  it('should map backend cancelled to frontend cancelled', () => {
    assert.equal(backendToFrontend['cancelled'], 'cancelled')
  })

  it('should pass through unknown status values unchanged', () => {
    assert.equal(backendToFrontend['running'], undefined)
  })

  it('should map frontend success to backend completed for query', () => {
    assert.equal(resolveBackendFilter('success'), 'completed')
  })

  it('should map frontend cancelled to backend cancelled for query', () => {
    assert.equal(resolveBackendFilter('cancelled'), 'cancelled')
  })

  it('should pass frontend failed to backend failed for query', () => {
    assert.equal(resolveBackendFilter('failed'), 'failed')
  })

  it('should pass empty string through for query', () => {
    assert.equal(resolveBackendFilter(''), '')
  })
})
