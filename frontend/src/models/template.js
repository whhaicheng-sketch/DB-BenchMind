export const TEMPLATE_TOOLS = ['sysbench', 'swingbench', 'hammerdb']

export const TEMPLATE_TOOL_LABELS = {
  sysbench: 'Sysbench',
  swingbench: 'Swingbench 2.7',
  hammerdb: 'HammerDB'
}

export const DB_FAMILY_LABELS = {
  mysql: 'MySQL',
  postgresql: 'PostgreSQL',
  oracle: 'Oracle',
  sqlserver: 'SQL Server',
  mariadb: 'MariaDB',
  db2: 'Db2'
}

export const WORKLOAD_LABELS = {
  'oltp-read-write': 'OLTP Read Write',
  'oltp-read-only': 'OLTP Read Only',
  'oltp-write-only': 'OLTP Write Only',
  'oltp-point-select': 'OLTP Point Select',
  'order-entry': 'OrderEntry',
  'sales-history': 'SalesHistory',
  'stress-test': 'StressTest',
  'tproc-c': 'TPROC-C',
  'tproc-h': 'TPROC-H'
}

export const PROFILE_TYPE_LABELS = {
  cpu_bound: 'CPU Bound',
  io_bound: 'IO Bound',
  test: 'Test'
}

export const SOURCE_ALIGNMENT_LABELS = {
  direct_from_doc: 'Direct From Doc',
  direct_from_doc_as_baseline: 'Direct From Doc Baseline',
  engineered_split_from_baseline: 'Engineered Split',
  engineered_minimal: 'Engineered Minimal'
}

export const CONCURRENCY_MODE_LABELS = {
  threads: 'Threads',
  users: 'Users',
  virtualUsers: 'Virtual Users'
}

export const PHASE_KEYS = [
  'prepare',
  'warmup',
  'run',
  'cleanup'
]

function createPhaseConfig(overrides, defaults) {
  return {
    enabled: defaults.enabled,
    required: defaults.required,
    params: {},
    ...(overrides || {})
  }
}

export function createPhaseState(overrides = {}) {
  return {
    prepare: createPhaseConfig(overrides.prepare, { enabled: false, required: false }),
    warmup: createPhaseConfig(overrides.warmup, { enabled: false, required: false }),
    run: createPhaseConfig(overrides.run, { enabled: true, required: true }),
    cleanup: createPhaseConfig(overrides.cleanup, { enabled: false, required: false })
  }
}

function normalizePhasesForTool(tool, phases = {}) {
  const normalized = createPhaseState({
    prepare: phases.prepare,
    warmup: phases.warmup,
    run: phases.run,
    cleanup: phases.cleanup
  })

  normalized.prepare.enabled = true
  normalized.warmup.enabled = true
  normalized.run.enabled = true
  normalized.run.required = true
  normalized.cleanup.enabled = true

  return normalized
}

export function cloneTemplate(template) {
  return JSON.parse(JSON.stringify(template))
}

export function createDefaultTemplate(partial = {}) {
  const now = new Date().toISOString()

  return {
    id: partial.id ?? createTemplateId(),
    name: partial.name ?? 'New Template',
    description: partial.description ?? '',
    tool: partial.tool ?? 'sysbench',
    profile_type: partial.profile_type ?? '',
    goal: partial.goal ?? '',
    readonly: partial.readonly ?? !!partial.is_builtin,
    source_alignment: partial.source_alignment ?? '',
    prepare_config: partial.prepare_config ?? {},
    run_config: partial.run_config ?? {},
    cleanup_config: partial.cleanup_config ?? {},
    metrics: partial.metrics ?? [],
    tags: partial.tags ?? [],
    test_position: partial.test_position ?? '',
    dbFamily: partial.dbFamily ?? 'mysql',
    workloadFamily: partial.workloadFamily ?? 'oltp-read-write',
    is_builtin: partial.is_builtin ?? false,
    version: partial.version ?? '0.1.0',
    compatibility: {
      supportedDatabases: partial.compatibility?.supportedDatabases ?? [partial.dbFamily ?? 'mysql'],
      supportedVersions: partial.compatibility?.supportedVersions ?? ['TBD'],
      compatibilityNotes: partial.compatibility?.compatibilityNotes ?? '',
      requiresPrivileges: partial.compatibility?.requiresPrivileges ?? [],
      constraints: partial.compatibility?.constraints ?? []
    },
    phases: normalizePhasesForTool(partial.tool ?? 'sysbench', partial.phases || {}),
    runtime: {
      concurrency: {
        mode: partial.runtime?.concurrency?.mode ?? 'threads',
        value: partial.runtime?.concurrency?.value ?? 16
      },
      durationSeconds: partial.runtime?.durationSeconds ?? 300,
      warmupSeconds: partial.runtime?.warmupSeconds ?? 30,
      rampUpSeconds: partial.runtime?.rampUpSeconds ?? 15,
      reportIntervalSeconds: partial.runtime?.reportIntervalSeconds ?? 10,
      percentile: partial.runtime?.percentile ?? 95,
      iterations: partial.runtime?.iterations ?? 0,
      rateLimit: partial.runtime?.rateLimit ?? 0,
      validationEnabled: partial.runtime?.validationEnabled ?? false,
      notes: partial.runtime?.notes ?? ''
    },
    toolConfig: {
      sysbench: {
        dbDriver: partial.toolConfig?.sysbench?.dbDriver ?? 'mysql',
        scriptType: partial.toolConfig?.sysbench?.scriptType ?? 'oltp_read_write',
        tables: partial.toolConfig?.sysbench?.tables ?? 10,
        tableSize: partial.toolConfig?.sysbench?.tableSize ?? 100000,
        reportChecks: partial.toolConfig?.sysbench?.reportChecks ?? true,
        extraCliArgs: partial.toolConfig?.sysbench?.extraCliArgs ?? '',
        ...partial.toolConfig?.sysbench
      },
      swingbench: {
        benchmark: partial.toolConfig?.swingbench?.benchmark ?? 'orderEntry',
        frontend: partial.toolConfig?.swingbench?.frontend ?? 'charbench',
        configMode: partial.toolConfig?.swingbench?.configMode ?? 'managed',
        wizardOperation: partial.toolConfig?.swingbench?.wizardOperation ?? 'generate',
        userCount: partial.toolConfig?.swingbench?.userCount ?? 64,
        runTimeSeconds: partial.toolConfig?.swingbench?.runTimeSeconds ?? 1800,
        minThinkTime: partial.toolConfig?.swingbench?.minThinkTime ?? 0,
        maxThinkTime: partial.toolConfig?.swingbench?.maxThinkTime ?? 2,
        xmlOverrides: partial.toolConfig?.swingbench?.xmlOverrides ?? '',
        ...partial.toolConfig?.swingbench
      },
      hammerdb: (() => {
        const benchmark = partial.toolConfig?.hammerdb?.benchmark ?? partial.workloadFamily ?? 'tproc-c'
        return {
          benchmark,
          virtualUsers: partial.toolConfig?.hammerdb?.virtualUsers ?? 64,
          warehouses: partial.toolConfig?.hammerdb?.warehouses ?? (benchmark === 'tproc-c' ? 100 : undefined),
          scaleFactor: partial.toolConfig?.hammerdb?.scaleFactor ?? (benchmark === 'tproc-h' ? 10 : undefined),
          timeProfile: partial.toolConfig?.hammerdb?.timeProfile ?? true,
          stepTesting: partial.toolConfig?.hammerdb?.stepTesting ?? false,
          xmlConnectPool: partial.toolConfig?.hammerdb?.xmlConnectPool ?? false,
          advancedNotes: partial.toolConfig?.hammerdb?.advancedNotes ?? '',
          ...partial.toolConfig?.hammerdb
        }
      })()
    },
    createdAt: partial.createdAt ?? now,
    database_types: partial.database_types ?? [partial.dbFamily ?? 'mysql']
  }
}

export function normalizeTemplateRecord(template = {}) {
  const normalized = createDefaultTemplate(template)
  normalized.phases = normalizePhasesForTool(normalized.tool, normalized.phases)
  normalized.database_types = template.database_types ?? [normalized.dbFamily]
  normalized.readonly = normalized.readonly ?? !!normalized.is_builtin
  return normalized
}

export function createTemplateId() {
  return `tpl_${Math.random().toString(36).slice(2, 10)}`
}
