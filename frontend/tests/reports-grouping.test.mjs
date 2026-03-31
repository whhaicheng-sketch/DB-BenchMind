import { describe, it } from 'node:test'
import assert from 'node:assert/strict'

/**
 * Reports grouping tests
 *
 * Tests the reportGroups computed property in ReportsTab.vue which groups
 * suite items by suite_id using source_type for partitioning.
 * Also tests status mapping and progress computation.
 */

// ===========================================================================
// Pure functions matching ReportsTab.vue logic
// ===========================================================================

const STANDALONE_SUITE_ID = 'standalone'

/**
 * isAutoBench: determines if a report belongs to an AutoBench suite group
 */
function isAutoBench(report) {
  return report.source_type === 'autobench' ||
    (report.suite_id && report.suite_id !== STANDALONE_SUITE_ID)
}

/**
 * computeReportGroups: mirrors the reportGroups computed property in ReportsTab.vue
 *
 * Input: reports array, suites array
 * Output: groups array
 */
function computeReportGroups(reports, suites) {
  const groups = []
  const bySuite = {}

  // Partition reports into AutoBench (grouped) vs standalone
  for (const report of reports) {
    if (isAutoBench(report)) {
      const sid = report.suite_id || 'unknown'
      if (!bySuite[sid]) {
        bySuite[sid] = []
      }
      bySuite[sid].push(report)
    } else {
      if (!bySuite[STANDALONE_SUITE_ID]) {
        bySuite[STANDALONE_SUITE_ID] = []
      }
      bySuite[STANDALONE_SUITE_ID].push(report)
    }
  }

  // Build suite map from suite data
  const suiteMap = {}
  for (const s of suites) {
    suiteMap[s.id] = s
  }

  // Suites first (ordered by suites array)
  for (const suite of suites) {
    const suiteReports = bySuite[suite.id] || []
    if (suiteReports.length > 0) {
      groups.push({
        key: suite.id,
        isSuite: true,
        name: suite.name || `AutoBench Suite ${suite.started_at}`,
        status: deriveGroupStatus(suiteReports),
        startedAt: suite.started_at,
        reports: suiteReports,
        progress: computeSuiteProgress(suiteReports)
      })
      delete bySuite[suite.id]
    }
  }

  // Remaining suite groups (suites not in store but reports have suite_id)
  for (const [sid, suiteReports] of Object.entries(bySuite)) {
    if (sid === STANDALONE_SUITE_ID) continue
    groups.push({
      key: sid,
      isSuite: true,
      name: suiteReports[0]?.template_name || `Suite ${sid.slice(0, 8)}`,
      status: deriveGroupStatus(suiteReports),
      startedAt: suiteReports[0]?.started_at,
      reports: suiteReports,
      progress: computeSuiteProgress(suiteReports)
    })
  }

  // Standalone reports
  const standalone = bySuite[STANDALONE_SUITE_ID] || []
  for (const report of standalone) {
    groups.push({
      key: report.id,
      isSuite: false,
      reports: [report]
    })
  }

  return groups
}

// ===========================================================================
// deriveGroupStatus: mirrors ReportsTab.vue deriveGroupStatus logic
// ===========================================================================
function deriveGroupStatus(reports) {
  if (reports.length === 0) return 'success'
  const statuses = reports.map(r => r.status)
  if (statuses.some(s => s === 'running' || s === 'pending' || s === 'draft' || s === 'ready')) return 'running'
  if (statuses.some(s => s === 'failed')) return 'failed'
  if (statuses.some(s => s === 'cancelled' || s === 'stopped' || s === 'interrupted')) return 'cancelled'
  if (statuses.some(s => s === 'partial_success')) return 'partial_success'
  if (statuses.every(s => s === 'success' || s === 'completed')) return 'success'
  return 'success'
}

// ===========================================================================
// computeSuiteProgress: mirrors ReportsTab.vue computeSuiteProgress logic
// ===========================================================================
function computeSuiteProgress(reports) {
  const total = reports.length
  let completed = 0
  let running = 0
  let failed = 0
  for (const r of reports) {
    if (r.status === 'success' || r.status === 'completed') completed++
    else if (r.status === 'running' || r.status === 'pending' || r.status === 'draft' || r.status === 'ready') running++
    else if (r.status === 'failed') failed++
  }
  return { total, completed, running, failed }
}

// ===========================================================================
// canViewReport: mirrors ReportsTab.vue canViewReport logic
// ===========================================================================
function canViewReport(status) {
  return status === 'completed' || status === 'success'
}

// ===========================================================================
// getStatusText: mirrors ReportsTab.vue getStatusText logic
// ===========================================================================
function getStatusText(status) {
  const textMap = {
    completed: '\u6210\u529f',
    success: '\u6210\u529f',
    failed: '\u5931\u8d25',
    cancelled: 'Stop',
    stopped: 'Stop',
    interrupted: 'Stop',
    running: '\u8fd0\u884c\u4e2d',
    pending: '\u7b49\u5f85\u4e2d',
    partial_success: '\u90e8\u5206\u6210\u529f',
    draft: '\u8349\u7a3f',
    ready: '\u5c31\u7eea'
  }
  return textMap[status] || status
}

// ===========================================================================
// Tests
// ===========================================================================

describe('computeReportGroups', () => {
  it('should group reports by same suite_id with source_type autobench', () => {
    const reports = [
      { id: 'r1', suite_id: 'suite-1', source_type: 'autobench', status: 'completed', started_at: '2025-01-01' },
      { id: 'r2', suite_id: 'suite-1', source_type: 'autobench', status: 'completed', started_at: '2025-01-02' },
      { id: 'r3', suite_id: 'suite-1', source_type: 'autobench', status: 'failed', started_at: '2025-01-03' }
    ]

    const suites = [
      { id: 'suite-1', name: 'Test Suite', status: 'success', started_at: '2025-01-01' }
    ]

    const result = computeReportGroups(reports, suites)

    assert.strictEqual(result.length, 1)
    assert.strictEqual(result[0].isSuite, true)
    assert.strictEqual(result[0].key, 'suite-1')
    assert.strictEqual(result[0].name, 'Test Suite')
    assert.strictEqual(result[0].reports.length, 3)
    assert.strictEqual(result[0].reports[0].id, 'r1')
    assert.strictEqual(result[0].reports[2].id, 'r3')
  })

  it('should never show autobench reports as standalone', () => {
    const reports = [
      { id: 'r1', suite_id: 'suite-1', source_type: 'autobench', status: 'completed', started_at: '2025-01-01' },
      // Even without a matching suite, autobench reports are grouped
      { id: 'r2', suite_id: 'suite-2', source_type: 'autobench', status: 'completed', started_at: '2025-01-02' }
    ]

    const result = computeReportGroups(reports, [])

    assert.strictEqual(result.length, 2)
    assert.strictEqual(result[0].isSuite, true)
    assert.strictEqual(result[0].key, 'suite-1')
    assert.strictEqual(result[1].isSuite, true)
    assert.strictEqual(result[1].key, 'suite-2')
    // No standalone groups
  })

  it('should produce standalone groups for benchmark source_type with standalone suite_id', () => {
    const reports = [
      { id: 'r1', suite_id: 'standalone', source_type: 'benchmark', status: 'completed', started_at: '2025-01-01' },
      { id: 'r2', source_type: 'benchmark', status: 'failed', started_at: '2025-01-02' }
    ]

    const result = computeReportGroups(reports, [])

    assert.strictEqual(result.length, 2)
    assert.strictEqual(result[0].isSuite, false)
    assert.strictEqual(result[0].key, 'r1')
    assert.strictEqual(result[0].reports.length, 1)
    assert.strictEqual(result[1].isSuite, false)
    assert.strictEqual(result[1].key, 'r2')
  })

  it('should produce standalone groups for reports with no suite_id and no source_type', () => {
    const reports = [
      { id: 'r1', status: 'completed', started_at: '2025-01-01' },
      { id: 'r2', status: 'failed', started_at: '2025-01-02' }
    ]

    const result = computeReportGroups(reports, [])

    assert.strictEqual(result.length, 2)
    assert.strictEqual(result[0].isSuite, false)
    assert.strictEqual(result[0].key, 'r1')
    assert.strictEqual(result[0].reports.length, 1)
    assert.strictEqual(result[1].isSuite, false)
    assert.strictEqual(result[1].key, 'r2')
  })

  it('should treat empty suite_id reports as standalone', () => {
    const reports = [
      { id: 'r-empty', suite_id: '', source_type: 'benchmark', status: 'completed', started_at: '2025-01-01' }
    ]
    const result = computeReportGroups(reports, [])
    assert.strictEqual(result.length, 1)
    assert.strictEqual(result[0].isSuite, false)
  })

  it('should sort suite groups before standalone reports', () => {
    const reports = [
      { id: 'r1', suite_id: 'suite-1', source_type: 'autobench', status: 'completed', started_at: '2025-01-01' },
      { id: 'r2', source_type: 'benchmark', status: 'completed', started_at: '2025-01-02' },
      { id: 'r3', suite_id: 'suite-1', source_type: 'autobench', status: 'failed', started_at: '2025-01-03' }
    ]

    const suites = [
      { id: 'suite-1', name: 'My Suite', status: 'failed', started_at: '2025-01-01' }
    ]

    const result = computeReportGroups(reports, suites)

    // Suite group first, then standalone
    assert.strictEqual(result.length, 2)
    assert.strictEqual(result[0].isSuite, true)
    assert.strictEqual(result[0].key, 'suite-1')
    assert.strictEqual(result[1].isSuite, false)
    assert.strictEqual(result[1].key, 'r2')
  })

  it('should return empty array for empty reports', () => {
    const result = computeReportGroups([], [])
    assert.strictEqual(result.length, 0)
  })

  it('should handle reports with suite_id not in suites array (orphan group)', () => {
    const reports = [
      { id: 'r1', suite_id: 'orphan-suite', source_type: 'autobench', status: 'completed', started_at: '2025-01-01', template_name: 'Orphan Test' }
    ]

    const result = computeReportGroups(reports, [])

    assert.strictEqual(result.length, 1)
    assert.strictEqual(result[0].isSuite, true)
    assert.strictEqual(result[0].key, 'orphan-suite')
    assert.strictEqual(result[0].name, 'Orphan Test')
    assert.strictEqual(result[0].status, 'success')
  })

  it('should handle mixed suite and standalone reports', () => {
    const reports = [
      { id: 'r1', suite_id: 'suite-a', source_type: 'autobench', status: 'completed', started_at: '2025-01-01' },
      { id: 'r2', suite_id: 'suite-a', source_type: 'autobench', status: 'failed', started_at: '2025-01-02' },
      { id: 'r3', source_type: 'benchmark', status: 'completed', started_at: '2025-01-03' },
      { id: 'r4', suite_id: 'suite-b', source_type: 'autobench', status: 'completed', started_at: '2025-01-04' }
    ]

    const suites = [
      { id: 'suite-a', name: 'Suite A', status: 'failed', started_at: '2025-01-01' },
      { id: 'suite-b', name: 'Suite B', status: 'success', started_at: '2025-01-04' }
    ]

    const result = computeReportGroups(reports, suites)

    assert.strictEqual(result.length, 3)
    // Suites first
    assert.strictEqual(result[0].isSuite, true)
    assert.strictEqual(result[0].key, 'suite-a')
    assert.strictEqual(result[0].reports.length, 2)
    assert.strictEqual(result[1].isSuite, true)
    assert.strictEqual(result[1].key, 'suite-b')
    assert.strictEqual(result[1].reports.length, 1)
    // Standalone last
    assert.strictEqual(result[2].isSuite, false)
    assert.strictEqual(result[2].key, 'r3')
  })

  it('should use source_type=autobench with non-standalone suite_id for grouping even without matching suite', () => {
    const reports = [
      { id: 'r1', suite_id: 'auto-uuid-123', source_type: 'autobench', status: 'running', started_at: '2025-01-01' },
      { id: 'r2', suite_id: 'auto-uuid-123', source_type: 'autobench', status: 'completed', started_at: '2025-01-02' }
    ]

    const result = computeReportGroups(reports, [])

    // No suite record, but still grouped by suite_id
    assert.strictEqual(result.length, 1)
    assert.strictEqual(result[0].isSuite, true)
    assert.strictEqual(result[0].key, 'auto-uuid-123')
    assert.strictEqual(result[0].reports.length, 2)
    assert.strictEqual(result[0].status, 'running')
  })

  it('should not leak autobench reports into standalone when suite record exists', () => {
    const reports = [
      { id: 'r1', suite_id: 'suite-1', source_type: 'autobench', status: 'completed', started_at: '2025-01-01' },
      { id: 'r2', source_type: 'benchmark', status: 'completed', started_at: '2025-01-02' }
    ]

    const suites = [
      { id: 'suite-1', name: 'Test Suite', status: 'success', started_at: '2025-01-01' }
    ]

    const result = computeReportGroups(reports, suites)

    // 1 suite group + 1 standalone, NOT 1 suite + 2 standalone
    assert.strictEqual(result.length, 2)
    assert.strictEqual(result[0].isSuite, true)
    assert.strictEqual(result[0].reports.length, 1)
    assert.strictEqual(result[1].isSuite, false)
    assert.strictEqual(result[1].reports[0].id, 'r2')
  })
})

describe('deriveGroupStatus', () => {
  it('should return success for empty reports', () => {
    assert.strictEqual(deriveGroupStatus([]), 'success')
  })

  it('should return running if any report is running', () => {
    const reports = [
      { status: 'completed' },
      { status: 'running' }
    ]
    assert.strictEqual(deriveGroupStatus(reports), 'running')
  })

  it('should return running if any report is pending', () => {
    const reports = [
      { status: 'completed' },
      { status: 'pending' }
    ]
    assert.strictEqual(deriveGroupStatus(reports), 'running')
  })

  it('should return failed if any report is failed', () => {
    const reports = [
      { status: 'completed' },
      { status: 'failed' }
    ]
    assert.strictEqual(deriveGroupStatus(reports), 'failed')
  })

  it('should return cancelled if any report is cancelled', () => {
    const reports = [
      { status: 'completed' },
      { status: 'cancelled' }
    ]
    assert.strictEqual(deriveGroupStatus(reports), 'cancelled')
  })

  it('should return cancelled if any report is stopped', () => {
    const reports = [
      { status: 'completed' },
      { status: 'stopped' }
    ]
    assert.strictEqual(deriveGroupStatus(reports), 'cancelled')
  })

  it('should return cancelled if any report is interrupted', () => {
    const reports = [
      { status: 'completed' },
      { status: 'interrupted' }
    ]
    assert.strictEqual(deriveGroupStatus(reports), 'cancelled')
  })

  it('should return partial_success if any report is partial_success', () => {
    const reports = [
      { status: 'completed' },
      { status: 'partial_success' }
    ]
    assert.strictEqual(deriveGroupStatus(reports), 'partial_success')
  })

  it('should return success if all reports are completed', () => {
    const reports = [
      { status: 'completed' },
      { status: 'completed' }
    ]
    assert.strictEqual(deriveGroupStatus(reports), 'success')
  })

  it('should return success if all reports are success', () => {
    const reports = [
      { status: 'success' },
      { status: 'success' }
    ]
    assert.strictEqual(deriveGroupStatus(reports), 'success')
  })

  it('should return success for mixed success and completed', () => {
    const reports = [
      { status: 'success' },
      { status: 'completed' }
    ]
    assert.strictEqual(deriveGroupStatus(reports), 'success')
  })

  it('should prioritize running over failed', () => {
    const reports = [
      { status: 'failed' },
      { status: 'running' }
    ]
    assert.strictEqual(deriveGroupStatus(reports), 'running')
  })

  it('should prioritize failed over cancelled', () => {
    const reports = [
      { status: 'cancelled' },
      { status: 'failed' }
    ]
    assert.strictEqual(deriveGroupStatus(reports), 'failed')
  })

  it('3 completed + 1 running = suite status running', () => {
    const reports = [
      { status: 'completed' },
      { status: 'completed' },
      { status: 'completed' },
      { status: 'running' }
    ]
    assert.strictEqual(deriveGroupStatus(reports), 'running')
  })
})

describe('computeSuiteProgress', () => {
  it('should count completed reports', () => {
    const reports = [
      { status: 'completed' },
      { status: 'success' },
      { status: 'completed' }
    ]
    const progress = computeSuiteProgress(reports)
    assert.deepStrictEqual(progress, { total: 3, completed: 3, running: 0, failed: 0 })
  })

  it('should count running and pending reports', () => {
    const reports = [
      { status: 'completed' },
      { status: 'running' },
      { status: 'pending' },
      { status: 'draft' }
    ]
    const progress = computeSuiteProgress(reports)
    assert.deepStrictEqual(progress, { total: 4, completed: 1, running: 3, failed: 0 })
  })

  it('should count failed reports', () => {
    const reports = [
      { status: 'completed' },
      { status: 'failed' },
      { status: 'failed' }
    ]
    const progress = computeSuiteProgress(reports)
    assert.deepStrictEqual(progress, { total: 3, completed: 1, running: 0, failed: 2 })
  })

  it('should handle 3 completed + 1 running', () => {
    const reports = [
      { status: 'success' },
      { status: 'success' },
      { status: 'success' },
      { status: 'running' }
    ]
    const progress = computeSuiteProgress(reports)
    assert.deepStrictEqual(progress, { total: 4, completed: 3, running: 1, failed: 0 })
  })

  it('should return zeros for empty array', () => {
    const progress = computeSuiteProgress([])
    assert.deepStrictEqual(progress, { total: 0, completed: 0, running: 0, failed: 0 })
  })
})

describe('canViewReport', () => {
  it('should allow viewing completed reports', () => {
    assert.strictEqual(canViewReport('completed'), true)
  })

  it('should allow viewing success reports', () => {
    assert.strictEqual(canViewReport('success'), true)
  })

  it('should not allow viewing running reports', () => {
    assert.strictEqual(canViewReport('running'), false)
  })

  it('should not allow viewing failed reports', () => {
    assert.strictEqual(canViewReport('failed'), false)
  })

  it('should not allow viewing pending reports', () => {
    assert.strictEqual(canViewReport('pending'), false)
  })
})

describe('getStatusText', () => {
  const testCases = [
    { status: 'completed', expected: '\u6210\u529f' },
    { status: 'success', expected: '\u6210\u529f' },
    { status: 'failed', expected: '\u5931\u8d25' },
    { status: 'cancelled', expected: 'Stop' },
    { status: 'stopped', expected: 'Stop' },
    { status: 'interrupted', expected: 'Stop' },
    { status: 'running', expected: '\u8fd0\u884c\u4e2d' },
    { status: 'pending', expected: '\u7b49\u5f85\u4e2d' },
    { status: 'partial_success', expected: '\u90e8\u5206\u6210\u529f' },
    { status: 'draft', expected: '\u8349\u7a3f' },
    { status: 'ready', expected: '\u5c31\u7eea' }
  ]

  for (const tt of testCases) {
    it(`should map "${tt.status}" to "${tt.expected}"`, () => {
      assert.strictEqual(getStatusText(tt.status), tt.expected)
    })
  }

  it('should return the status itself for unknown values', () => {
    assert.strictEqual(getStatusText('unknown'), 'unknown')
  })
})
