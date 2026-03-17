import { chromium } from 'playwright'

const baseUrl = process.env.TEMPLATE_TEST_VERIFY_BASE_URL || 'http://127.0.0.1:4173'

const templates = [
  {
    id: 'tpl_builtin_mysql',
    name: 'Sysbench MySQL Baseline',
    description: 'Baseline template',
    tool: 'sysbench',
    dbFamily: 'mysql',
    workloadFamily: 'oltp-read-write',
    scope: 'builtin',
    tags: ['baseline'],
    status: 'ready',
    version: '1.0.0',
    compatibility: { supportedDatabases: ['mysql'], supportedVersions: [], compatibilityNotes: '', requiresPrivileges: [], constraints: [] },
    phases: phaseSet(),
    runtime: runtime('threads', 16, 300),
    toolConfig: toolConfig({
      sysbench: { dbDriver: 'mysql', scriptType: 'oltp_read_write', tables: 10, tableSize: 100000, reportChecks: true, extraCliArgs: '' }
    }),
    createdAt: '2026-03-15T00:00:00Z',
    updatedAt: '2026-03-15T00:00:00Z',
    database_types: ['mysql']
  },
  createSysbenchTest('tpl_test_mysql_sysbench', 'MySQL - Sysbench Test', 'mysql', 'mysql'),
  createSysbenchTest('tpl_test_postgresql_sysbench', 'PostgreSQL - Sysbench Test', 'postgresql', 'pgsql'),
  createSwingbenchTest(),
  createHammerDBTest()
]

const connections = [
  { id: 'conn-mysql', name: 'MySQL Smoke', type: 'mysql', host: '127.0.0.1', port: 3306, database: 'sbtest', username: 'root' },
  { id: 'conn-pg', name: 'PostgreSQL Smoke', type: 'postgresql', host: '127.0.0.1', port: 5432, database: 'sbtest', username: 'postgres' },
  { id: 'conn-oracle', name: 'Oracle Smoke', type: 'oracle', host: '127.0.0.1', port: 1521, database: 'ORCL', username: 'system' },
  { id: 'conn-sqlserver', name: 'SQL Server Smoke', type: 'sqlserver', host: '127.0.0.1', port: 1433, database: 'tpcc', username: 'sa' }
]

function phaseSet() {
  return {
    build: { enabled: false, required: false, params: {} },
    prepare: { enabled: true, required: false, params: {} },
    generate: { enabled: false, required: false, params: {} },
    warmup: { enabled: false, required: false, params: {} },
    run: { enabled: true, required: true, params: {} },
    verify: { enabled: false, required: false, params: {} },
    cleanup: { enabled: true, required: false, params: {} },
    delete: { enabled: false, required: false, params: {} }
  }
}

function runtime(mode, value, durationSeconds) {
  return {
    concurrency: { mode, value },
    durationSeconds,
    warmupSeconds: 0,
    rampUpSeconds: 0,
    reportIntervalSeconds: 1,
    percentile: 95,
    iterations: 1,
    rateLimit: 0,
    validationEnabled: true,
    notes: ''
  }
}

function toolConfig(overrides = {}) {
  return {
    sysbench: { dbDriver: 'mysql', scriptType: 'oltp_read_write', tables: 1, tableSize: 1000, reportChecks: true, extraCliArgs: '', ...overrides.sysbench },
    swingbench: { benchmark: 'orderEntry', frontend: 'charbench', configMode: 'managed', wizardOperation: 'generate', userCount: 1, runTimeSeconds: 60, minThinkTime: 0, maxThinkTime: 0, xmlOverrides: '', ...overrides.swingbench },
    hammerdb: { benchmark: 'tproc-c', virtualUsers: 1, warehouses: 1, scaleFactor: 1, timeProfile: true, stepTesting: false, xmlConnectPool: false, advancedNotes: '', ...overrides.hammerdb }
  }
}

function createSysbenchTest(id, name, dbFamily, dbDriver) {
  return {
    id,
    name,
    description: `${name} smoke template`,
    tool: 'sysbench',
    dbFamily,
    workloadFamily: 'oltp-read-write',
    scope: 'test',
    tags: ['test', dbFamily, 'sysbench'],
    status: 'ready',
    version: '1.0.0',
    compatibility: { supportedDatabases: [dbFamily], supportedVersions: [], compatibilityNotes: '', requiresPrivileges: [], constraints: [] },
    phases: phaseSet(),
    runtime: runtime('threads', 1, 30),
    toolConfig: toolConfig({
      sysbench: { dbDriver, scriptType: 'oltp_read_write', tables: 1, tableSize: 1000, reportChecks: true, extraCliArgs: '' }
    }),
    createdAt: '2026-03-15T00:00:00Z',
    updatedAt: '2026-03-15T00:00:00Z',
    database_types: [dbFamily]
  }
}

function createSwingbenchTest() {
  return {
    id: 'tpl_test_oracle_swingbench',
    name: 'Oracle - Swingbench Test',
    description: 'Oracle smoke template',
    tool: 'swingbench',
    dbFamily: 'oracle',
    workloadFamily: 'order-entry',
    scope: 'test',
    tags: ['test', 'oracle', 'swingbench'],
    status: 'ready',
    version: '1.0.0',
    compatibility: { supportedDatabases: ['oracle'], supportedVersions: [], compatibilityNotes: '', requiresPrivileges: [], constraints: [] },
    phases: phaseSet(),
    runtime: runtime('users', 1, 60),
    toolConfig: toolConfig({
      swingbench: { benchmark: 'orderEntry', frontend: 'charbench', configMode: 'managed', wizardOperation: 'generate', userCount: 1, runTimeSeconds: 60, minThinkTime: 0, maxThinkTime: 0, xmlOverrides: '' }
    }),
    parameters: {
      scale: { type: 'integer', label: 'Scale', default: 1 }
    },
    createdAt: '2026-03-15T00:00:00Z',
    updatedAt: '2026-03-15T00:00:00Z',
    database_types: ['oracle']
  }
}

function createHammerDBTest() {
  return {
    id: 'tpl_test_sqlserver_hammerdb',
    name: 'SQL Server - HammerDB Test',
    description: 'SQL Server smoke template',
    tool: 'hammerdb',
    dbFamily: 'sqlserver',
    workloadFamily: 'tproc-c',
    scope: 'test',
    tags: ['test', 'sqlserver', 'hammerdb'],
    status: 'ready',
    version: '1.0.0',
    compatibility: { supportedDatabases: ['sqlserver'], supportedVersions: [], compatibilityNotes: '', requiresPrivileges: [], constraints: [] },
    phases: phaseSet(),
    runtime: runtime('virtualUsers', 1, 60),
    toolConfig: toolConfig({
      hammerdb: { benchmark: 'tproc-c', virtualUsers: 1, warehouses: 1, scaleFactor: 1, timeProfile: true, stepTesting: false, xmlConnectPool: false, advancedNotes: '' }
    }),
    createdAt: '2026-03-15T00:00:00Z',
    updatedAt: '2026-03-15T00:00:00Z',
    database_types: ['sqlserver']
  }
}

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

function clone(value) {
  return JSON.parse(JSON.stringify(value))
}

async function main() {
  const browser = await chromium.launch({ headless: true })
  const page = await browser.newPage({ viewport: { width: 1440, height: 1100 } })

  await page.addInitScript(({ seedTemplates, seedConnections }) => {
    const state = {
      templates: clone(seedTemplates),
      connections: clone(seedConnections),
      tasks: [],
      preview: null
    }

    function clone(value) {
      return JSON.parse(JSON.stringify(value))
    }

    window.runtime = new Proxy({
      EventsOnMultiple() { return () => {} },
      EventsOn() { return () => {} },
      EventsOff() {},
      EventsOffAll() {},
      EventsEmit() {},
      LogPrint() {},
      LogTrace() {},
      LogDebug() {},
      LogInfo() {},
      LogWarning() {},
      LogError() {},
      LogFatal() {}
    }, {
      get(target, prop) {
        if (prop in target) return target[prop]
        return () => {}
      }
    })

    window.go = {
      bindings: {
        TemplateBinding: {
          ListTemplates: async () => ({ templates: clone(state.templates) }),
          CreateTemplate: async (template) => ({ template: clone(template) }),
          UpdateTemplate: async (template) => ({ template: clone(template) }),
          DeleteTemplate: async () => ({ success: true }),
          DuplicateTemplate: async (id) => ({ template: clone(state.templates.find((item) => item.id === id)) })
        },
        ConnectionBinding: {
          ListConnections: async () => ({ connections: clone(state.connections) }),
          GetConnection: async (id) => ({ connection: clone(state.connections.find((item) => item.id === id)) }),
          CreateConnection: async () => ({ error: 'not implemented in test' }),
          UpdateConnection: async () => ({ error: 'not implemented in test' }),
          DeleteConnection: async () => ({ success: true }),
          TestConnection: async () => ({ success: true, latency_ms: 5 }),
          TestSSHConnection: async () => ({ success: false, latency_ms: 0, error: 'ssh disabled' }),
          TestWinRMConnection: async () => ({ success: false, latency_ms: 0, error: 'winrm disabled' })
        },
        TaskBinding: {
          ValidateDraft: async (payload) => {
            state.preview = {
              id: 'preview-1',
              preview_token: 'preview-token',
              readiness: {
                template_selected: Boolean(payload.template_id),
                connection_selected: Boolean(payload.connection_id),
                action_supported: true,
                runtime_valid: true,
                db_valid: Boolean(payload.connection_id),
                db_message: payload.connection_id ? 'DB ok (5 ms)' : 'Select a connection',
                ssh_available: false,
                ssh_message: 'SSH required'
              },
              template_snapshot: {
                id: payload.template_id,
                name: state.templates.find((item) => item.id === payload.template_id)?.name || '',
                tool: state.templates.find((item) => item.id === payload.template_id)?.tool || '',
                db_family: state.templates.find((item) => item.id === payload.template_id)?.dbFamily || '',
                phases: { prepare: true, run: true, cleanup: true }
              }
            }
            return { task: clone(state.preview) }
          },
          CreateTask: async (payload) => {
            const task = {
              id: 'task-1',
              preview_token: payload.preview_token,
              name: payload.task_name || 'task-1',
              status: 'starting',
              current_phase: 'prepare',
              benchmark_tool: state.templates.find((item) => item.id === payload.template_id)?.tool || '',
              template_snapshot: {
                id: payload.template_id,
                name: state.templates.find((item) => item.id === payload.template_id)?.name || '',
                tool: state.templates.find((item) => item.id === payload.template_id)?.tool || '',
                db_family: state.templates.find((item) => item.id === payload.template_id)?.dbFamily || ''
              },
              readiness: state.preview?.readiness || null
            }
            state.tasks = [task]
            return { task: clone(task) }
          },
          ListTasks: async () => ({ tasks: clone(state.tasks) }),
          StopTask: async () => ({ success: true }),
          GetTaskLogs: async () => ({ lines: [] })
        },
        MonitorBinding: {
          StartMonitoring: async () => ({ success: true }),
          StopMonitoring: async () => ({ success: true }),
          GetMonitorState: async () => ({}),
          GetMonitorData: async () => ({}),
          ClearData: async () => ({ success: true }),
          StartSystemMonitoring: async () => ({ success: true }),
          StopSystemMonitoring: async () => ({ success: true }),
          GetSystemMetrics: async () => ({}),
          IsSystemMonitoring: async () => false
        },
        BenchmarkBinding: {
          StartBenchmark: async () => ({ run_id: 'run-1' }),
          PrepareOnly: async () => ({ run_id: 'run-prepare' }),
          RunBenchmark: async () => ({ run_id: 'run-run' }),
          CleanupOnly: async () => ({ run_id: 'run-cleanup' }),
          StopBenchmark: async () => ({ success: true }),
          GetBenchmarkStatus: async () => ({}),
          ListBenchmarks: async () => ({ runs: [] })
        }
      }
    }
  }, { seedTemplates: templates, seedConnections: connections })

  await page.goto(baseUrl, { waitUntil: 'load' })
  await page.getByRole('button', { name: /Templates/ }).click()

  const scopeSelect = page.locator('.filter-field', { hasText: 'Scope / Tag' }).locator('select').nth(0)
  const dbSelect = page.locator('.filter-field', { hasText: 'Database Type' }).locator('select')
  const toolSelect = page.locator('.filter-field', { hasText: 'Benchmark Tool' }).locator('select')
  const subtitle = page.locator('.panel-subtitle')
  const listItems = page.locator('.template-item')

  await scopeSelect.selectOption('test')
  await page.waitForTimeout(100)
  const subtitleText = await subtitle.textContent()
  const visibleTestCount = await listItems.count()
  assert(subtitleText === '4 visible / 5 total', `Expected 4 visible / 5 total, got ${subtitleText}`)
  assert(visibleTestCount === 4, `Expected 4 visible test templates, got ${visibleTestCount}`)

  const comboResults = {}

  await dbSelect.selectOption('mysql')
  await toolSelect.selectOption('')
  await page.waitForTimeout(100)
  comboResults.mysqlTest = await listItems.count()
  assert(comboResults.mysqlTest === 1, `Expected 1 MySQL test template, got ${comboResults.mysqlTest}`)
  await toolSelect.selectOption('sysbench')
  await page.waitForTimeout(100)
  comboResults.mysqlSysbenchTest = await listItems.count()
  assert(comboResults.mysqlSysbenchTest === 1, `Expected 1 MySQL Sysbench test template, got ${comboResults.mysqlSysbenchTest}`)

  await dbSelect.selectOption('postgresql')
  await toolSelect.selectOption('')
  await page.waitForTimeout(100)
  comboResults.postgresqlTest = await listItems.count()
  assert(comboResults.postgresqlTest === 1, `Expected 1 PostgreSQL test template, got ${comboResults.postgresqlTest}`)
  await toolSelect.selectOption('sysbench')
  await page.waitForTimeout(100)
  comboResults.postgresqlSysbenchTest = await listItems.count()
  assert(comboResults.postgresqlSysbenchTest === 1, `Expected 1 PostgreSQL Sysbench test template, got ${comboResults.postgresqlSysbenchTest}`)

  await dbSelect.selectOption('oracle')
  await toolSelect.selectOption('')
  await page.waitForTimeout(100)
  comboResults.oracleTest = await listItems.count()
  assert(comboResults.oracleTest === 1, `Expected 1 Oracle test template, got ${comboResults.oracleTest}`)
  await toolSelect.selectOption('swingbench')
  await page.waitForTimeout(100)
  comboResults.oracleSwingbenchTest = await listItems.count()
  assert(comboResults.oracleSwingbenchTest === 1, `Expected 1 Oracle Swingbench test template, got ${comboResults.oracleSwingbenchTest}`)

  await dbSelect.selectOption('sqlserver')
  await toolSelect.selectOption('')
  await page.waitForTimeout(100)
  await toolSelect.selectOption('hammerdb')
  await page.waitForTimeout(100)
  comboResults.sqlserverTest = await listItems.count()
  comboResults.sqlserverHammerdbTest = comboResults.sqlserverTest
  assert(comboResults.sqlserverHammerdbTest === 1, `Expected 1 SQL Server HammerDB test template, got ${comboResults.sqlserverHammerdbTest}`)
  const sqlServerText = await page.locator('.template-item').first().textContent()
  assert(sqlServerText?.includes('SQL Server - HammerDB Test'), `Expected SQL Server HammerDB template text, got ${sqlServerText}`)

  await page.locator('.template-item', { hasText: 'SQL Server - HammerDB Test' }).click()
  await page.getByRole('button', { name: 'Create Task from Template' }).click()
  await page.waitForTimeout(200)

  const tasksDbSelect = page.locator('.field', { hasText: 'Database Type' }).locator('select')
  const tasksTemplateSelect = page.locator('.field', { hasText: 'Template' }).locator('select')

  assert(await tasksDbSelect.inputValue() === 'sqlserver', `Expected tasks database_type=sqlserver, got ${await tasksDbSelect.inputValue()}`)
  assert(await tasksTemplateSelect.inputValue() === 'tpl_test_sqlserver_hammerdb', `Expected handed off template_id, got ${await tasksTemplateSelect.inputValue()}`)

  console.log(JSON.stringify({
    visibleCount: visibleTestCount,
    subtitle: subtitleText,
    combos: comboResults,
    tasksDatabaseType: await tasksDbSelect.inputValue(),
    tasksTemplateId: await tasksTemplateSelect.inputValue()
  }, null, 2))

  await browser.close()
}

main().catch((error) => {
  console.error(error)
  process.exitCode = 1
})
