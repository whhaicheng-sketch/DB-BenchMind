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

export const TEMPLATE_SCOPE_LABELS = {
  builtin: 'Built-in',
  user: 'User'
}

export const TEMPLATE_STATUS_LABELS = {
  draft: 'Draft',
  ready: 'Ready',
  deprecated: 'Deprecated'
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

export const CONCURRENCY_MODE_LABELS = {
  threads: 'Threads',
  users: 'Users',
  virtualUsers: 'Virtual Users'
}

export const PHASE_KEYS = [
  'build',
  'prepare',
  'generate',
  'warmup',
  'run',
  'verify',
  'cleanup',
  'delete'
]

export function createPhaseState(overrides = {}) {
  return {
    build: { enabled: false, required: false, params: {} },
    prepare: { enabled: false, required: false, params: {} },
    generate: { enabled: false, required: false, params: {} },
    warmup: { enabled: false, required: false, params: {} },
    run: { enabled: true, required: true, params: {} },
    verify: { enabled: false, required: false, params: {} },
    cleanup: { enabled: false, required: false, params: {} },
    delete: { enabled: false, required: false, params: {} },
    ...overrides
  }
}

export function cloneTemplate(template) {
  return JSON.parse(JSON.stringify(template))
}

export function createDefaultTemplate(partial = {}) {
  const now = new Date().toISOString()

  return {
    id: partial.id || createTemplateId(),
    name: partial.name || 'New Template',
    description: partial.description || 'Benchmark scenario draft for future task binding.',
    tool: partial.tool || 'sysbench',
    dbFamily: partial.dbFamily || 'mysql',
    workloadFamily: partial.workloadFamily || 'oltp-read-write',
    scope: partial.scope || 'user',
    tags: partial.tags || ['draft'],
    status: partial.status || 'draft',
    version: partial.version || '0.1.0',
    compatibility: {
      supportedDatabases: partial.compatibility?.supportedDatabases || [partial.dbFamily || 'mysql'],
      supportedVersions: partial.compatibility?.supportedVersions || ['TBD'],
      compatibilityNotes: partial.compatibility?.compatibilityNotes || '',
      requiresPrivileges: partial.compatibility?.requiresPrivileges || [],
      constraints: partial.compatibility?.constraints || []
    },
    phases: createPhaseState(partial.phases || {}),
    runtime: {
      concurrency: {
        mode: partial.runtime?.concurrency?.mode || 'threads',
        value: partial.runtime?.concurrency?.value || 16
      },
      durationSeconds: partial.runtime?.durationSeconds || 300,
      warmupSeconds: partial.runtime?.warmupSeconds || 30,
      rampUpSeconds: partial.runtime?.rampUpSeconds || 15,
      reportIntervalSeconds: partial.runtime?.reportIntervalSeconds || 10,
      percentile: partial.runtime?.percentile || 95,
      iterations: partial.runtime?.iterations || 0,
      rateLimit: partial.runtime?.rateLimit || 0,
      validationEnabled: partial.runtime?.validationEnabled ?? false,
      notes: partial.runtime?.notes || ''
    },
    toolConfig: {
      sysbench: {
        dbDriver: 'mysql',
        scriptType: 'oltp_read_write',
        tables: 10,
        tableSize: 100000,
        reportChecks: true,
        extraCliArgs: '',
        ...partial.toolConfig?.sysbench
      },
      swingbench: {
        benchmark: 'orderEntry',
        frontend: 'charbench',
        configMode: 'managed',
        wizardOperation: 'generate',
        userCount: 64,
        runTimeSeconds: 1800,
        minThinkTime: 0,
        maxThinkTime: 2,
        xmlOverrides: '',
        ...partial.toolConfig?.swingbench
      },
      hammerdb: {
        benchmark: 'tproc-c',
        virtualUsers: 64,
        warehouses: 100,
        scaleFactor: 10,
        timeProfile: true,
        stepTesting: false,
        xmlConnectPool: false,
        advancedNotes: '',
        ...partial.toolConfig?.hammerdb
      }
    },
    createdAt: partial.createdAt || now,
    updatedAt: partial.updatedAt || now
  }
}

export function createTemplateId() {
  return `tpl_${Math.random().toString(36).slice(2, 10)}`
}

export function formatTemplateDate(value) {
  if (!value) return 'Unknown'

  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(new Date(value))
}
