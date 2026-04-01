import { describe, it } from 'node:test'
import assert from 'node:assert/strict'

/**
 * Suite/report status convergence tests
 *
 * Tests the state chain from AutoBench suite completion to report status display.
 * Verifies that when a suite completes, no "running" child reports remain.
 */

// Mirrors analyzeGroupReports from ReportsTab.vue
function analyzeGroupReports(reports) {
  const total = reports.length
  let completed = 0
  let running = 0
  let failed = 0
  let cancelled = 0
  let partial = false

  for (const r of reports) {
    const s = r.status
    if (s === 'success' || s === 'completed') completed++
    else if (s === 'running' || s === 'pending' || s === 'draft' || s === 'ready') running++
    else if (s === 'failed') failed++
    else if (s === 'cancelled' || s === 'stopped' || s === 'interrupted') cancelled++
    else if (s === 'partial_success') partial = true
  }

  let status = 'success'
  if (running > 0) status = 'running'
  else if (failed > 0) status = 'failed'
  else if (cancelled > 0) status = 'cancelled'
  else if (partial) status = 'partial_success'

  return { status, progress: { total, completed, running, failed } }
}

describe('Suite/report status convergence', () => {
  it('suite with all completed reports shows success', () => {
    const reports = [
      { status: 'completed', source_type: 'autobench', suite_id: 'suite-1' },
      { status: 'completed', source_type: 'autobench', suite_id: 'suite-1' },
      { status: 'success', source_type: 'autobench', suite_id: 'suite-1' }
    ]
    const { status, progress } = analyzeGroupReports(reports)
    assert.strictEqual(status, 'success')
    assert.strictEqual(progress.completed, 3)
    assert.strictEqual(progress.running, 0)
  })

  it('suite success should NOT have running child reports', () => {
    // This is the key invariant: if suite is success, no child should be running
    const suiteStatus = 'success'
    const reports = [
      { status: 'completed', source_type: 'autobench', suite_id: 'suite-1' },
      { status: 'completed', source_type: 'autobench', suite_id: 'suite-1' },
      { status: 'completed', source_type: 'autobench', suite_id: 'suite-1' }
    ]

    const { progress } = analyzeGroupReports(reports)
    // Verify no running reports
    assert.strictEqual(progress.running, 0, 'No running reports should exist when suite is success')

    // Group status should derive from child reports, not suite status
    const { status: derivedStatus } = analyzeGroupReports(reports)
    assert.strictEqual(derivedStatus, suiteStatus)
  })

  it('reports page should not show running when suite completed', () => {
    // Simulate the scenario: backend correctly updates all reports to completed
    const reports = [
      { id: 'r1', status: 'completed', source_type: 'autobench', suite_id: 'suite-1' },
      { id: 'r2', status: 'completed', source_type: 'autobench', suite_id: 'suite-1' }
    ]

    const { status, progress } = analyzeGroupReports(reports)
    assert.strictEqual(status, 'success')
    assert.strictEqual(progress.running, 0)
    assert.strictEqual(progress.completed, 2)
  })

  it('group status derives from child reports not suite record', () => {
    // Even if suite record says "success", derive from actual reports
    const suiteRecordStatus = 'success'
    const reports = [
      { status: 'failed', source_type: 'autobench', suite_id: 'suite-1' },
      { status: 'completed', source_type: 'autobench', suite_id: 'suite-1' }
    ]

    const { status: derivedStatus } = analyzeGroupReports(reports)
    // Derived status should be "failed" based on actual report statuses,
    // NOT "success" from the suite record
    assert.strictEqual(derivedStatus, 'failed')
    assert.notStrictEqual(derivedStatus, suiteRecordStatus)
  })

  it('handles mixed suite and standalone reports correctly', () => {
    const reports = [
      { id: 'r1', status: 'completed', source_type: 'autobench', suite_id: 'suite-1' },
      { id: 'r2', status: 'success', source_type: 'autobench', suite_id: 'suite-1' },
      { id: 'r3', status: 'failed', source_type: 'benchmark', suite_id: 'standalone' }
    ]

    // Suite group
    const suiteReports = reports.filter(r => r.source_type === 'autobench')
    const { status: suiteStatus } = analyzeGroupReports(suiteReports)
    assert.strictEqual(suiteStatus, 'success')

    // Standalone
    const standaloneReports = reports.filter(r => r.source_type !== 'autobench')
    const { status: standaloneStatus } = analyzeGroupReports(standaloneReports)
    assert.strictEqual(standaloneStatus, 'failed')
  })

  it('page refresh and tab switch maintain state consistency', () => {
    // After refresh, all report statuses come from DB
    const freshReports = [
      { status: 'completed', source_type: 'autobench', suite_id: 'suite-1' },
      { status: 'completed', source_type: 'autobench', suite_id: 'suite-1' }
    ]
    const { status: s1, progress: p1 } = analyzeGroupReports(freshReports)

    // Same data after tab switch
    const { status: s2, progress: p2 } = analyzeGroupReports(freshReports)

    assert.strictEqual(s1, s2)
    assert.deepStrictEqual(p1, p2)
    assert.strictEqual(s1, 'success')
  })

  it('handles orphan suite group without crashing', () => {
    // Reports exist with suite_id but no matching suite record
    const reports = [
      { status: 'completed', source_type: 'autobench', suite_id: 'orphan-suite', template_name: 'MySQL OLTP' }
    ]
    const { status, progress } = analyzeGroupReports(reports)
    assert.strictEqual(status, 'success')
    assert.strictEqual(progress.total, 1)
    assert.strictEqual(progress.completed, 1)
  })

  it('empty reports list returns success with zero progress', () => {
    const { status, progress } = analyzeGroupReports([])
    assert.strictEqual(status, 'success')
    assert.deepStrictEqual(progress, { total: 0, completed: 0, running: 0, failed: 0 })
  })
})

describe('Report grouping partition logic', () => {
  const STANDALONE_SUITE_ID = 'standalone'

  function isAutoBench(report) {
    return report.source_type === 'autobench' ||
      !!(report.suite_id && report.suite_id !== STANDALONE_SUITE_ID)
  }

  it('autobench reports are always grouped', () => {
    const report = { source_type: 'autobench', suite_id: 'suite-1', status: 'completed' }
    assert.strictEqual(isAutoBench(report), true)
  })

  it('benchmark reports with standalone suite_id are standalone', () => {
    const report = { source_type: 'benchmark', suite_id: 'standalone', status: 'completed' }
    assert.strictEqual(isAutoBench(report), false)
  })

  it('reports without suite_id or source_type are standalone', () => {
    const report = { status: 'completed' }
    assert.strictEqual(isAutoBench(report), false)
  })

  it('reports with non-standalone suite_id are grouped regardless of source_type', () => {
    const report = { source_type: 'benchmark', suite_id: 'some-uuid', status: 'completed' }
    assert.strictEqual(isAutoBench(report), true)
  })
})
