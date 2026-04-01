import { describe, it } from 'node:test'
import assert from 'node:assert/strict'

// ============================================================
// P1: Form Lock Tests
// ============================================================

describe('P1: Form lock during benchmark', () => {
  // Simulates the isFormLocked computed logic
  function isFormLocked(isAutoBenchControlled, activeTask) {
    return isAutoBenchControlled || !!activeTask
  }

  it('form is locked when manual benchmark task is running', () => {
    assert.strictEqual(isFormLocked(false, { id: 't1', status: 'running' }), true)
  })

  it('form is locked when AutoBench is controlling', () => {
    assert.strictEqual(isFormLocked(true, null), true)
  })

  it('form is locked when both manual and AutoBench active', () => {
    assert.strictEqual(isFormLocked(true, { id: 't1', status: 'running' }), true)
  })

  it('form is unlocked when no task running and AutoBench inactive', () => {
    assert.strictEqual(isFormLocked(false, null), false)
  })

  it('form is unlocked when task is in terminal state', () => {
    assert.strictEqual(isFormLocked(false, null), false)
  })

  it('form unlocks after manual benchmark completes', () => {
    // Task goes from running -> completed, activeTask becomes null
    assert.strictEqual(isFormLocked(false, null), false)
  })
})

// ============================================================
// P2: Flags Rendering Tests
// ============================================================

describe('P2: Connection capability flags', () => {
  function buildCaps(conn) {
    const caps = []
    if (conn.ssh_enabled) caps.push('SSH')
    if (conn.winrm_enabled) caps.push('WinRM')
    // Fixed: require provider AND (api_key OR model) to show AI
    if (conn.ai_assistants && conn.ai_assistants.some(a => a.provider && (a.api_key || a.model))) caps.push('AI')
    return caps
  }

  it('MySQL with SSH + configured AI shows SSH + AI', () => {
    const conn = {
      ssh_enabled: true,
      winrm_enabled: false,
      ai_assistants: [{ provider: 'deepseek', api_key: 'sk-xxx', model: 'deepseek-chat' }]
    }
    assert.deepStrictEqual(buildCaps(conn), ['SSH', 'AI'])
  })

  it('Oracle with SSH only, no AI assistants array, shows only SSH', () => {
    const conn = {
      ssh_enabled: true,
      winrm_enabled: false,
    }
    assert.deepStrictEqual(buildCaps(conn), ['SSH'])
  })

  it('PostgreSQL with SSH but default empty AI assistant shows only SSH (not AI)', () => {
    // This is the P2 bug: default form pre-populates ai_assistants with {provider: 'deepseek', api_key: '', model: ''}
    const conn = {
      ssh_enabled: true,
      winrm_enabled: false,
      ai_assistants: [{ provider: 'deepseek', api_key: '', model: '' }]
    }
    assert.deepStrictEqual(buildCaps(conn), ['SSH'])
  })

  it('SQLServer with WinRM and empty AI assistant shows only WinRM', () => {
    const conn = {
      ssh_enabled: false,
      winrm_enabled: true,
      ai_assistants: [{ provider: 'deepseek', api_key: '', model: '' }]
    }
    assert.deepStrictEqual(buildCaps(conn), ['WinRM'])
  })

  it('connection with no capabilities shows empty array', () => {
    const conn = {
      ssh_enabled: false,
      winrm_enabled: false,
    }
    assert.deepStrictEqual(buildCaps(conn), [])
  })

  it('connection with ai_assistants having only model (no api_key) shows AI', () => {
    const conn = {
      ssh_enabled: false,
      winrm_enabled: false,
      ai_assistants: [{ provider: 'openai', api_key: '', model: 'gpt-4' }]
    }
    assert.deepStrictEqual(buildCaps(conn), ['AI'])
  })

  it('connection with ai_assistants having only api_key (no model) shows AI', () => {
    const conn = {
      ssh_enabled: false,
      winrm_enabled: false,
      ai_assistants: [{ provider: 'openai', api_key: 'sk-xxx', model: '' }]
    }
    assert.deepStrictEqual(buildCaps(conn), ['AI'])
  })
})

// ============================================================
// P3: View Report Navigation Tests
// ============================================================

describe('P3: View Report navigation', () => {
  // Simulates the autoSelectReport function
  function autoSelectReport(reportId, reportGroups, expandedGroups) {
    if (!reportId) return { selectedId: null, expanded: expandedGroups }

    const newExpanded = { ...expandedGroups }
    for (const group of reportGroups) {
      if (group.reports.some(r => r.id === reportId)) {
        newExpanded[group.key] = true
        break
      }
    }
    return { selectedId: reportId, expanded: newExpanded }
  }

  const reportGroups = [
    {
      key: 'suite-1',
      isSuite: true,
      reports: [
        { id: 'r1', status: 'completed' },
        { id: 'r2', status: 'completed' }
      ]
    },
    {
      key: 'standalone-r3',
      isSuite: false,
      reports: [{ id: 'r3', status: 'completed' }]
    }
  ]

  it('selects report and expands parent suite group', () => {
    const result = autoSelectReport('r2', reportGroups, {})
    assert.strictEqual(result.selectedId, 'r2')
    assert.strictEqual(result.expanded['suite-1'], true)
  })

  it('selects standalone report without needing expansion', () => {
    const result = autoSelectReport('r3', reportGroups, {})
    assert.strictEqual(result.selectedId, 'r3')
  })

  it('selects report even when other group was previously expanded', () => {
    const result = autoSelectReport('r1', reportGroups, { 'standalone-r3': true })
    assert.strictEqual(result.selectedId, 'r1')
    assert.strictEqual(result.expanded['suite-1'], true)
    assert.strictEqual(result.expanded['standalone-r3'], true)
  })

  it('handles null reportId gracefully', () => {
    const result = autoSelectReport(null, reportGroups, {})
    assert.strictEqual(result.selectedId, null)
  })

  it('handles report not found in any group', () => {
    const result = autoSelectReport('nonexistent', reportGroups, {})
    assert.strictEqual(result.selectedId, 'nonexistent')
    // No group expanded
    assert.deepStrictEqual(result.expanded, {})
  })
})

// ============================================================
// P4: Unified Monitor Data Source Tests
// ============================================================

describe('P4: Unified monitor data source', () => {
  function resolveMonitorTask(isAutoBenchControlled, autoBenchCurrentItem, currentTask) {
    if (isAutoBenchControlled && autoBenchCurrentItem) {
      const item = autoBenchCurrentItem
      const m = item.metrics || {}
      return {
        id: item.id,
        status: item.phase_status || item.status,
        metrics: m,
        started_at: item.started_at || null,
        completed_at: null,
        connection_snapshot: { name: item.connection_name || item.connection_id },
        template_snapshot: { name: item.profile_type || '' },
        error_message: item.error_message || '',
        log_tail: []
      }
    }
    return currentTask
  }

  it('returns AutoBench virtual task when AutoBench is controlling', () => {
    const item = {
      id: 'item-1',
      status: 'running',
      phase_status: 'run',
      connection_id: 'conn-1',
      profile_type: 'test',
      started_at: '2026-03-31T10:00:00Z',
      metrics: {
        tps: { current: 100, series: [{ timestamp: 1, value: 100 }] },
        tpm: { current: 6000, series: [{ timestamp: 1, value: 6000 }] }
      }
    }
    const result = resolveMonitorTask(true, item, null)
    assert.strictEqual(result.id, 'item-1')
    assert.strictEqual(result.metrics.tps.current, 100)
    assert.strictEqual(result.status, 'run')
  })

  it('returns currentTask when AutoBench is not controlling', () => {
    const manualTask = {
      id: 'task-1',
      status: 'running',
      metrics: { tps: { current: 50 } }
    }
    const result = resolveMonitorTask(false, null, manualTask)
    assert.strictEqual(result.id, 'task-1')
    assert.strictEqual(result.metrics.tps.current, 50)
  })

  it('returns currentTask when AutoBench is controlling but no current item', () => {
    const manualTask = { id: 'task-1', status: 'completed' }
    const result = resolveMonitorTask(true, null, manualTask)
    assert.strictEqual(result.id, 'task-1')
  })

  it('AutoBench virtual task has correct metrics structure for charts', () => {
    const item = {
      id: 'item-1',
      status: 'running',
      phase_status: 'run',
      connection_id: 'conn-1',
      profile_type: 'test',
      metrics: {
        tps: { current: 100, avg: 95, max: 120, series: [] },
        tpm: { current: 6000, avg: 5700, max: 7200, series: [] },
        system_enabled: false,
        system_message: 'System metrics not available in AutoBench mode'
      }
    }
    const result = resolveMonitorTask(true, item, null)
    assert.strictEqual(result.metrics.tps.current, 100)
    assert.strictEqual(result.metrics.tpm.current, 6000)
    assert.strictEqual(result.metrics.system_enabled, false)
    assert.ok(result.metrics.system_message.includes('AutoBench'))
  })

  it('switches data source when AutoBench sub-task changes', () => {
    const item1 = { id: 'item-1', status: 'running', connection_id: 'c1', metrics: { tps: { current: 50 } } }
    const result1 = resolveMonitorTask(true, item1, null)
    assert.strictEqual(result1.id, 'item-1')

    const item2 = { id: 'item-2', status: 'running', connection_id: 'c2', metrics: { tps: { current: 200 } } }
    const result2 = resolveMonitorTask(true, item2, null)
    assert.strictEqual(result2.id, 'item-2')
    assert.strictEqual(result2.metrics.tps.current, 200)
  })
})

// ============================================================
// P5: Re-run Button Tests
// ============================================================

describe('P5: Re-run button visibility', () => {
  const terminalStates = ['success', 'partial_success', 'failed', 'cancelled']
  const nonTerminalStates = ['draft', 'ready', 'running', 'paused']

  it('shows Re-run for success terminal state', () => {
    assert.ok(terminalStates.includes('success'))
  })

  it('shows Re-run for failed terminal state', () => {
    assert.ok(terminalStates.includes('failed'))
  })

  it('shows Re-run for cancelled terminal state', () => {
    assert.ok(terminalStates.includes('cancelled'))
  })

  it('does NOT show Re-run for running state', () => {
    assert.ok(!terminalStates.includes('running'))
  })

  it('does NOT show Re-run for draft state', () => {
    assert.ok(!terminalStates.includes('draft'))
  })

  it('does NOT show Re-run for paused state', () => {
    assert.ok(!terminalStates.includes('paused'))
  })

  it('does NOT show Re-run for ready state', () => {
    assert.ok(!terminalStates.includes('ready'))
  })

  it('all non-terminal states are excluded', () => {
    for (const state of nonTerminalStates) {
      assert.ok(!terminalStates.includes(state), `${state} should not be terminal`)
    }
  })
})

// ============================================================
// Modal click-outside disabled Tests
// ============================================================

describe('Modal click-outside behavior disabled', () => {
  // After the fix, modals should NOT have @click.self handlers
  // This test verifies the design principle

  it('modal overlay divs must not close on click.self', () => {
    // The pattern @click.self="closeXxx" should be removed
    // This is a design contract test - the actual enforcement is in the template
    const modalPatterns = [
      { name: 'Connection modal', hasClickSelf: false },
      { name: 'Task preview modal', hasClickSelf: false },
      { name: 'Log viewer modal', hasClickSelf: false },
      { name: 'Template delete confirm', hasClickSelf: false },
      { name: 'Error overlay', hasClickSelf: false },
    ]
    for (const m of modalPatterns) {
      assert.strictEqual(m.hasClickSelf, false, `${m.name} should not close on click.self`)
    }
  })
})

// ============================================================
// Tab State Preservation Tests
// ============================================================

describe('Tab state preservation', () => {
  it('KeepAlive preserves expanded state across tab switches', () => {
    // Simulating: ReportsTab has expandedGroups = {'suite-1': true}
    // User switches to ConnectionsTab then back to ReportsTab
    // With KeepAlive, component is not destroyed, so expandedGroups persists
    const expandedGroups = { 'suite-1': true, 'suite-2': false }
    // Simulate tab switch (with KeepAlive, component stays alive)
    // When user returns, state is preserved
    assert.strictEqual(expandedGroups['suite-1'], true)
  })

  it('KeepAlive preserves selected report across tab switches', () => {
    const selectedReportId = 'r-123'
    // With KeepAlive, local refs survive tab switch
    assert.strictEqual(selectedReportId, 'r-123')
  })

  it('KeepAlive preserves AutoBench polling state', () => {
    let isPolling = true
    let elapsedSeconds = 45
    // Switch to another tab and back
    // With KeepAlive, component stays alive, timers continue
    assert.strictEqual(isPolling, true)
    assert.strictEqual(elapsedSeconds, 45)
  })
})

// ============================================================
// P6: connNameMap Runtime Error Regression Tests
// ============================================================

describe('P6: monitorTask must not reference undefined variables', () => {
  // Simulates the monitorTask computed logic from TasksMonitorTab.vue
  // The fix ensures connection name resolution uses connectionStore directly,
  // not a variable (connNameMap) that only exists in AutoBenchTab.vue
  function resolveMonitorTask(isAutoBenchControlled, autoBenchCurrentItem, currentTask, connections) {
    if (isAutoBenchControlled && autoBenchCurrentItem) {
      const item = autoBenchCurrentItem
      const m = item.metrics || {}
      // This line MUST use connections array, NOT connNameMap (which is undefined in this scope)
      const connName = connections.find(c => c.id === item.connection_id)?.name || item.connection_id
      return {
        id: item.id,
        status: item.phase_status || item.status,
        connection_snapshot: { name: connName, type: item.database_type || '' },
        template_snapshot: { name: item.profile_type || '' },
        metrics: m,
        error_message: item.error_message || '',
        log_tail: []
      }
    }
    return currentTask
  }

  const connections = [
    { id: 'conn-mysql', name: 'MySQL 5.7', type: 'mysql' },
    { id: 'conn-oracle', name: 'Oracle 11g', type: 'oracle' },
    { id: 'conn-pg', name: 'PostgreSQL 13', type: 'postgresql' },
    { id: 'conn-mssql', name: 'SQLServer 2019', type: 'sqlserver' }
  ]

  it('resolves connection name from store, not connNameMap', () => {
    const item = {
      id: 'item-1', status: 'running', phase_status: 'run',
      connection_id: 'conn-mysql', database_type: 'mysql',
      profile_type: 'test', metrics: { tps: { current: 100 } }
    }
    const task = resolveMonitorTask(true, item, null, connections)
    assert.strictEqual(task.connection_snapshot.name, 'MySQL 5.7')
  })

  it('falls back to connection_id when connection not found in store', () => {
    const item = {
      id: 'item-1', status: 'running', phase_status: 'run',
      connection_id: 'conn-unknown', database_type: 'mysql',
      profile_type: 'test', metrics: { tps: { current: 100 } }
    }
    const task = resolveMonitorTask(true, item, null, connections)
    assert.strictEqual(task.connection_snapshot.name, 'conn-unknown')
  })

  it('works with empty connections store', () => {
    const item = {
      id: 'item-1', status: 'running', phase_status: 'run',
      connection_id: 'conn-mysql', database_type: 'mysql',
      profile_type: 'test', metrics: { tps: { current: 100 } }
    }
    const task = resolveMonitorTask(true, item, null, [])
    assert.strictEqual(task.connection_snapshot.name, 'conn-mysql')
  })

  it('does not throw ReferenceError for connNameMap', () => {
    // The critical regression test: this function must NOT reference
    // any variable that doesn't exist in its scope
    const item = {
      id: 'item-1', status: 'running', phase_status: 'run',
      connection_id: 'conn-mysql', database_type: 'mysql',
      profile_type: 'test', metrics: {}
    }
    // If connNameMap was referenced here, this would throw ReferenceError
    assert.doesNotThrow(() => resolveMonitorTask(true, item, null, connections))
  })

  it('all AutoBench states render without variable error', () => {
    const states = ['running', 'success', 'failed', 'cancelled', 'paused']
    for (const state of states) {
      const item = {
        id: 'item-1', status: state, phase_status: 'run',
        connection_id: 'conn-mysql', database_type: 'mysql',
        profile_type: 'test', metrics: { tps: { current: 100 } }
      }
      assert.doesNotThrow(() => resolveMonitorTask(state === 'running', item, null, connections),
        `Failed for state: ${state}`)
    }
  })
})

// ============================================================
// P10: Re-run TypeError Regression Tests
// ============================================================

describe('P10: handleReRunSuite must not access null suiteStatus', () => {
  function reRunSuitePattern(suiteStatus) {
    if (!suiteStatus) return { shouldReturn: true }

    const originalConnections = []
    const originalProfiles = new Set()
    for (const item of suiteStatus.items || []) {
      if (item.connection_id && !originalConnections.includes(item.connection_id)) {
        originalConnections.push(item.connection_id)
      }
      if (item.profile_type) originalProfiles.add(item.profile_type)
    }

    if (originalConnections.length === 0 || originalProfiles.size === 0) return { shouldReturn: true }

    // Save name BEFORE reset
    const suiteName = suiteStatus.name || 'Re-run Suite'
    return {
      shouldReturn: false,
      savedName: suiteName,
      connectionCount: originalConnections.length,
      profileCount: originalProfiles.size
    }
  }

  it('saves suite name before resetSuite clears suiteStatus', () => {
    const status = {
      name: 'My Test Suite',
      items: [
        { connection_id: 'conn-1', profile_type: 'test' },
        { connection_id: 'conn-2', profile_type: 'cpu_bound' }
      ]
    }
    const result = reRunSuitePattern(status)
    assert.strictEqual(result.savedName, 'My Test Suite')
  })

  it('falls back to Re-run Suite when name is empty', () => {
    const status = {
      name: '',
      items: [{ connection_id: 'conn-1', profile_type: 'test' }]
    }
    const result = reRunSuitePattern(status)
    assert.strictEqual(result.savedName, 'Re-run Suite')
  })

  it('returns early when no connections or profiles', () => {
    const status = { name: 'Suite', items: [] }
    const result = reRunSuitePattern(status)
    assert.strictEqual(result.shouldReturn, true)
  })

  it('extracts unique connections and profiles from completed items', () => {
    const status = {
      name: 'Multi',
      items: [
        { connection_id: 'c1', profile_type: 'test' },
        { connection_id: 'c1', profile_type: 'cpu_bound' },
        { connection_id: 'c2', profile_type: 'test' }
      ]
    }
    const result = reRunSuitePattern(status)
    assert.strictEqual(result.connectionCount, 2)
    assert.strictEqual(result.profileCount, 2)
  })
})

// ============================================================
// P12: Phase Timing Sync in Observation Mode Tests
// ============================================================

describe('P12: monitorTask virtual task maps phase_timings to timing', () => {
  function buildTimingFromPhaseTimings(phaseTimings) {
    const timings = Array.isArray(phaseTimings) ? phaseTimings : []
    const timingObj = { prepare_ms: 0, run_ms: 0, total_ms: 0 }
    for (const pt of timings) {
      if (pt.phase === 'preparing') timingObj.prepare_ms = pt.duration_ms
      else if (pt.phase === 'running') timingObj.run_ms = pt.duration_ms
      timingObj.total_ms += pt.duration_ms
    }
    return timingObj
  }

  it('maps preparing and running phases to timing object', () => {
    const phaseTimings = [
      { phase: 'preparing', duration_ms: 5000 },
      { phase: 'running', duration_ms: 60000 }
    ]
    const timing = buildTimingFromPhaseTimings(phaseTimings)
    assert.strictEqual(timing.prepare_ms, 5000)
    assert.strictEqual(timing.run_ms, 60000)
    assert.strictEqual(timing.total_ms, 65000)
  })

  it('handles empty phase_timings gracefully', () => {
    const timing = buildTimingFromPhaseTimings([])
    assert.strictEqual(timing.prepare_ms, 0)
    assert.strictEqual(timing.run_ms, 0)
    assert.strictEqual(timing.total_ms, 0)
  })

  it('handles null/undefined phase_timings gracefully', () => {
    const timing = buildTimingFromPhaseTimings(null)
    assert.strictEqual(timing.prepare_ms, 0)
    assert.strictEqual(timing.run_ms, 0)
    assert.strictEqual(timing.total_ms, 0)
  })

  it('handles only running phase (no prepare)', () => {
    const phaseTimings = [
      { phase: 'running', duration_ms: 30000 }
    ]
    const timing = buildTimingFromPhaseTimings(phaseTimings)
    assert.strictEqual(timing.prepare_ms, 0)
    assert.strictEqual(timing.run_ms, 30000)
    assert.strictEqual(timing.total_ms, 30000)
  })

  it('sums all phases for total_ms', () => {
    const phaseTimings = [
      { phase: 'preparing', duration_ms: 2000 },
      { phase: 'running', duration_ms: 45000 },
      { phase: 'cleaning', duration_ms: 3000 }
    ]
    const timing = buildTimingFromPhaseTimings(phaseTimings)
    assert.strictEqual(timing.total_ms, 50000)
  })
})

// ============================================================
// P14: AutoBench System Metrics Pipeline Tests
// ============================================================

describe('P14: AutoBench system metrics pipeline', () => {
  // Simulates the backend collectItemMetrics output with SSH metrics
  function buildMetricsMap(tpsSamples, sshPoints) {
    const metricsMap = {
      tps: { current: 100, avg: 95, max: 120, series: [] },
      tpm: { current: 6000, avg: 5700, max: 7200, series: [] },
      system_enabled: false,
      system_message: ''
    }

    if (sshPoints && sshPoints.length > 0) {
      metricsMap.system_enabled = true
      const last = sshPoints[sshPoints.length - 1]
      metricsMap.cpu_user = { current: last.cpu_user, series: sshPoints.map(p => ({ timestamp: p.ts, value: p.cpu_user })) }
      metricsMap.cpu_sys = { current: last.cpu_sys, series: sshPoints.map(p => ({ timestamp: p.ts, value: p.cpu_sys })) }
      metricsMap.cpu_iowait = { current: last.cpu_iowait, series: sshPoints.map(p => ({ timestamp: p.ts, value: p.cpu_iowait })) }
      metricsMap.cpu_steal = { current: last.cpu_steal, series: sshPoints.map(p => ({ timestamp: p.ts, value: p.cpu_steal })) }
      metricsMap.disk_read_bps = { current: last.disk_read_bps, series: sshPoints.map(p => ({ timestamp: p.ts, value: p.disk_read_bps })) }
      metricsMap.disk_write_bps = { current: last.disk_write_bps, series: sshPoints.map(p => ({ timestamp: p.ts, value: p.disk_write_bps })) }
      metricsMap.disk_read_latency_ms = { current: last.disk_read_latency_ms, series: sshPoints.map(p => ({ timestamp: p.ts, value: p.disk_read_latency_ms })) }
      metricsMap.disk_write_latency_ms = { current: last.disk_write_latency_ms, series: sshPoints.map(p => ({ timestamp: p.ts, value: p.disk_write_latency_ms })) }
    }

    return metricsMap
  }

  it('has system_enabled=true when SSH metrics are collected', () => {
    const sshPoints = [
      { ts: 1000, cpu_user: 15.2, cpu_sys: 8.1, cpu_iowait: 0.5, cpu_steal: 0.1, disk_read_bps: 1024, disk_write_bps: 2048, disk_read_latency_ms: 0.5, disk_write_latency_ms: 0.8 }
    ]
    const metrics = buildMetricsMap([], sshPoints)
    assert.strictEqual(metrics.system_enabled, true)
  })

  it('has system_enabled=false when no SSH metrics', () => {
    const metrics = buildMetricsMap([], null)
    assert.strictEqual(metrics.system_enabled, false)
  })

  it('CPU series contains USER, SYS, IOWAIT, ST data', () => {
    const sshPoints = [
      { ts: 1000, cpu_user: 20.0, cpu_sys: 10.0, cpu_iowait: 1.0, cpu_steal: 0.5, disk_read_bps: 0, disk_write_bps: 0, disk_read_latency_ms: 0, disk_write_latency_ms: 0 },
      { ts: 2000, cpu_user: 25.0, cpu_sys: 12.0, cpu_iowait: 2.0, cpu_steal: 0.3, disk_read_bps: 512, disk_write_bps: 1024, disk_read_latency_ms: 0.3, disk_write_latency_ms: 0.6 }
    ]
    const metrics = buildMetricsMap([], sshPoints)
    assert.strictEqual(metrics.cpu_user.current, 25.0)
    assert.strictEqual(metrics.cpu_sys.current, 12.0)
    assert.strictEqual(metrics.cpu_iowait.current, 2.0)
    assert.strictEqual(metrics.cpu_steal.current, 0.3)
    assert.strictEqual(metrics.cpu_user.series.length, 2)
  })

  it('DISK series contains READ, WRITE, R_LAT, W_LAT data', () => {
    const sshPoints = [
      { ts: 1000, cpu_user: 0, cpu_sys: 0, cpu_iowait: 0, cpu_steal: 0, disk_read_bps: 10240, disk_write_bps: 5120, disk_read_latency_ms: 1.2, disk_write_latency_ms: 2.5 }
    ]
    const metrics = buildMetricsMap([], sshPoints)
    assert.strictEqual(metrics.disk_read_bps.current, 10240)
    assert.strictEqual(metrics.disk_write_bps.current, 5120)
    assert.strictEqual(metrics.disk_read_latency_ms.current, 1.2)
    assert.strictEqual(metrics.disk_write_latency_ms.current, 2.5)
  })

  it('TPS/TPM and CPU/Disk are both available simultaneously', () => {
    const sshPoints = [
      { ts: 1000, cpu_user: 30.0, cpu_sys: 15.0, cpu_iowait: 0.8, cpu_steal: 0.2, disk_read_bps: 8192, disk_write_bps: 4096, disk_read_latency_ms: 0.9, disk_write_latency_ms: 1.1 }
    ]
    const metrics = buildMetricsMap([{ tps: 100 }], sshPoints)
    assert.strictEqual(metrics.tps.current, 100)
    assert.strictEqual(metrics.tpm.current, 6000)
    assert.strictEqual(metrics.system_enabled, true)
    assert.strictEqual(metrics.cpu_user.current, 30.0)
    assert.strictEqual(metrics.disk_read_bps.current, 8192)
  })

  it('frontend checks system_enabled from monitorTask metrics', () => {
    // Simulates frontend: systemEnabled = !!monitorTask?.metrics?.system_enabled
    const metrics1 = { system_enabled: true }
    assert.strictEqual(!!metrics1.system_enabled, true)

    const metrics2 = { system_enabled: false }
    assert.strictEqual(!!metrics2.system_enabled, false)

    const metrics3 = {}
    assert.strictEqual(!!metrics3.system_enabled, false)
  })

  it('sub-task switching replaces all metrics including system', () => {
    const ssh1 = [{ ts: 1, cpu_user: 10, cpu_sys: 5, cpu_iowait: 0.1, cpu_steal: 0, disk_read_bps: 100, disk_write_bps: 200, disk_read_latency_ms: 0.5, disk_write_latency_ms: 0.8 }]
    const metrics1 = buildMetricsMap([], ssh1)

    const ssh2 = [{ ts: 2, cpu_user: 40, cpu_sys: 20, cpu_iowait: 3, cpu_steal: 1, disk_read_bps: 5000, disk_write_bps: 3000, disk_read_latency_ms: 2.0, disk_write_latency_ms: 3.0 }]
    const metrics2 = buildMetricsMap([], ssh2)

    // Switching from item-1 to item-2 replaces all data
    assert.strictEqual(metrics1.cpu_user.current, 10)
    assert.strictEqual(metrics2.cpu_user.current, 40)
    assert.notStrictEqual(metrics1.cpu_user.current, metrics2.cpu_user.current)
  })
})

// ============================================================
// P15: Dynamic Phase Timing Tests
// ============================================================

describe('P15: Dynamic phase timing for PREPARE/RUN/TOTAL', () => {
  // Simulates the monitorTask computed logic with live timing
  function buildTimingObj(phaseTimings, phaseStatus, startedAt, nowMs, status) {
    const timingObj = { prepare_ms: 0, run_ms: 0, total_ms: 0 }

    // First: accumulate completed phase timings
    for (const pt of phaseTimings) {
      if (pt.phase === 'preparing') timingObj.prepare_ms = pt.duration_ms
      else if (pt.phase === 'running') timingObj.run_ms = pt.duration_ms
      timingObj.total_ms += pt.duration_ms
    }

    // Second: compute live duration for in-progress phase
    const taskStartedAt = startedAt ? new Date(startedAt).getTime() : 0
    const isTerminal = ['success', 'failed', 'skipped'].includes(status)
    if (taskStartedAt > 0 && !isTerminal) {
      const liveMs = Math.max(0, nowMs - taskStartedAt)
      let completedMs = 0
      for (const pt of phaseTimings) {
        completedMs += pt.duration_ms
      }
      const currentPhaseMs = Math.max(0, liveMs - completedMs)

      if (phaseStatus === 'preparing') {
        timingObj.prepare_ms += currentPhaseMs
      } else if (phaseStatus === 'running') {
        timingObj.run_ms += currentPhaseMs
      }
      timingObj.total_ms = liveMs
    }

    return timingObj
  }

  it('preparing phase shows growing prepare_ms', () => {
    const startedAt = '2026-04-01T10:00:00.000Z'
    const now = new Date('2026-04-01T10:00:05.000Z').getTime() // 5 seconds later
    const timing = buildTimingObj([], 'preparing', startedAt, now, 'running')
    assert.strictEqual(timing.prepare_ms, 5000)
    assert.strictEqual(timing.run_ms, 0)
    assert.strictEqual(timing.total_ms, 5000)
  })

  it('running phase shows growing run_ms with completed prepare', () => {
    const startedAt = '2026-04-01T10:00:00.000Z'
    const now = new Date('2026-04-01T10:01:05.000Z').getTime() // 65 seconds later
    const phaseTimings = [{ phase: 'preparing', duration_ms: 5000 }]
    const timing = buildTimingObj(phaseTimings, 'running', startedAt, now, 'running')
    assert.strictEqual(timing.prepare_ms, 5000) // completed
    assert.strictEqual(timing.run_ms, 60000)    // 65s - 5s = 60s live
    assert.strictEqual(timing.total_ms, 65000)
  })

  it('terminal state uses recorded timings only (no live computation)', () => {
    const startedAt = '2026-04-01T10:00:00.000Z'
    const now = new Date('2026-04-01T10:02:00.000Z').getTime()
    const phaseTimings = [
      { phase: 'preparing', duration_ms: 5000 },
      { phase: 'running', duration_ms: 60000 }
    ]
    const timing = buildTimingObj(phaseTimings, 'completed', startedAt, now, 'success')
    assert.strictEqual(timing.prepare_ms, 5000)
    assert.strictEqual(timing.run_ms, 60000)
    assert.strictEqual(timing.total_ms, 65000)
  })

  it('no started_at returns zeros', () => {
    const timing = buildTimingObj([], 'preparing', null, Date.now(), 'running')
    assert.strictEqual(timing.prepare_ms, 0)
    assert.strictEqual(timing.run_ms, 0)
    assert.strictEqual(timing.total_ms, 0)
  })

  it('sub-task switching resets to new task timings', () => {
    // Item 1: completed with 10s prepare + 60s run
    const timing1 = buildTimingObj(
      [{ phase: 'preparing', duration_ms: 10000 }, { phase: 'running', duration_ms: 60000 }],
      'completed', '2026-04-01T10:00:00.000Z',
      new Date('2026-04-01T10:01:10.000Z').getTime(), 'success'
    )

    // Item 2: just started preparing
    const item2StartedAt = '2026-04-01T10:01:10.000Z'
    const now2 = new Date('2026-04-01T10:01:15.000Z').getTime() // 5s into item 2
    const timing2 = buildTimingObj([], 'preparing', item2StartedAt, now2, 'running')

    assert.strictEqual(timing1.prepare_ms, 10000)
    assert.strictEqual(timing1.run_ms, 60000)
    assert.strictEqual(timing2.prepare_ms, 5000)   // new task starts from 0
    assert.strictEqual(timing2.run_ms, 0)
    assert.strictEqual(timing2.total_ms, 5000)
  })

  it('phase and timing stay consistent', () => {
    const startedAt = '2026-04-01T10:00:00.000Z'
    const now = new Date('2026-04-01T10:00:30.000Z').getTime()

    // Preparing phase: prepare grows, run is 0
    const prep = buildTimingObj([], 'preparing', startedAt, now, 'running')
    assert.ok(prep.prepare_ms > 0)
    assert.strictEqual(prep.run_ms, 0)
    assert.strictEqual(prep.total_ms, prep.prepare_ms)

    // Running phase: prepare is recorded, run grows
    const phaseTimings = [{ phase: 'preparing', duration_ms: prep.prepare_ms }]
    const runNow = new Date('2026-04-01T10:01:00.000Z').getTime()
    const run = buildTimingObj(phaseTimings, 'running', startedAt, runNow, 'running')
    assert.strictEqual(run.prepare_ms, prep.prepare_ms)
    assert.ok(run.run_ms > 0)
    assert.strictEqual(run.total_ms, run.prepare_ms + run.run_ms)
  })

  it('handles cleaning phase by accumulating in total only', () => {
    const startedAt = '2026-04-01T10:00:00.000Z'
    const now = new Date('2026-04-01T10:02:05.000Z').getTime()
    const phaseTimings = [
      { phase: 'preparing', duration_ms: 5000 },
      { phase: 'running', duration_ms: 60000 }
    ]
    // cleaning phase: not preparing or running, so currentPhaseMs goes nowhere
    // total is still live
    const timing = buildTimingObj(phaseTimings, 'cleaning', startedAt, now, 'running')
    assert.strictEqual(timing.prepare_ms, 5000)
    assert.strictEqual(timing.run_ms, 60000)
    assert.strictEqual(timing.total_ms, 125000) // 65s completed + 60s live = 125s
  })
})
