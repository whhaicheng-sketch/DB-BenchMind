import { chromium } from 'playwright'

const baseUrl = process.env.TEMPLATE_VERIFY_BASE_URL || 'http://127.0.0.1:4173'
const screenshotPath = process.env.TEMPLATE_VERIFY_SCREENSHOT || 'artifacts/templates-db-tool-linkage.png'
const annotatedScreenshotPath = process.env.TEMPLATE_VERIFY_ANNOTATED_SCREENSHOT || ''

const templates = [
  {
    id: 'tpl-user-mysql',
    name: 'Editable MySQL Template',
    description: 'User-owned template for edit verification.',
    tool: 'sysbench',
    dbFamily: 'mysql',
    workloadFamily: 'oltp-read-write',
    scope: 'user',
    tags: ['editable'],
    status: 'ready',
    version: '0.1.0',
    compatibility: {
      supportedDatabases: ['mysql'],
      supportedVersions: ['test'],
      compatibilityNotes: '',
      requiresPrivileges: [],
      constraints: []
    },
    phases: {
      build: { enabled: false, required: false, params: {} },
      prepare: { enabled: true, required: false, params: {} },
      generate: { enabled: false, required: false, params: {} },
      warmup: { enabled: true, required: false, params: {} },
      run: { enabled: true, required: true, params: {} },
      verify: { enabled: true, required: false, params: {} },
      cleanup: { enabled: true, required: false, params: {} },
      delete: { enabled: false, required: false, params: {} }
    },
    runtime: {
      concurrency: { mode: 'threads', value: 16 },
      durationSeconds: 300,
      warmupSeconds: 30,
      rampUpSeconds: 15,
      reportIntervalSeconds: 10,
      percentile: 95,
      iterations: 0,
      rateLimit: 0,
      validationEnabled: false,
      notes: ''
    },
    toolConfig: {
      sysbench: { dbDriver: 'mysql', scriptType: 'oltp_read_write', tables: 10, tableSize: 100000, reportChecks: true, extraCliArgs: '' },
      swingbench: { benchmark: 'orderEntry', frontend: 'charbench', configMode: 'managed', wizardOperation: 'generate', userCount: 64, runTimeSeconds: 1800, minThinkTime: 0, maxThinkTime: 2, xmlOverrides: '' },
      hammerdb: { benchmark: 'tproc-c', virtualUsers: 64, warehouses: 100, scaleFactor: 10, timeProfile: true, stepTesting: false, xmlConnectPool: false, advancedNotes: '' }
    },
    createdAt: '2026-03-14T10:00:00Z',
    updatedAt: '2026-03-14T10:00:00Z',
    database_types: ['mysql']
  }
]

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

async function annotate(page) {
  await page.evaluate(() => {
    const addBadge = (target, text, styles = {}) => {
      if (!target) return
      const badge = document.createElement('div')
      badge.textContent = text
      badge.style.position = 'absolute'
      badge.style.padding = styles.padding || '2px 6px'
      badge.style.borderRadius = '999px'
      badge.style.font = styles.font || '600 10px/1 sans-serif'
      badge.style.color = styles.color || '#081018'
      badge.style.background = styles.background || '#f8fafc'
      badge.style.zIndex = '10'
      badge.style.whiteSpace = 'nowrap'
      Object.entries(styles.position || {}).forEach(([key, value]) => {
        badge.style[key] = value
      })
      target.appendChild(badge)
    }

    const dialog = document.querySelector('.editor-dialog')
    if (dialog) {
      dialog.style.position = 'relative'
      dialog.style.outline = '2px dashed #93c5fd'
      dialog.style.outlineOffset = '-2px'
      addBadge(dialog, 'Create Template linkage', {
        background: '#93c5fd',
        position: { top: '10px', left: '10px' }
      })
    }

    const fields = [...document.querySelectorAll('.field')]
    const dbField = fields.find((field) => field.textContent?.includes('Database Type'))
    const toolField = fields.find((field) => field.textContent?.includes('Benchmark Tool'))
    if (dbField) {
      dbField.style.position = 'relative'
      dbField.style.outline = '1px dashed #f59e0b'
      dbField.style.outlineOffset = '2px'
      addBadge(dbField, 'Database Type drives options', {
        background: '#fde68a',
        position: { top: '-14px', left: '0' }
      })
    }
    if (toolField) {
      toolField.style.position = 'relative'
      toolField.style.outline = '1px dashed #34d399'
      toolField.style.outlineOffset = '2px'
      addBadge(toolField, 'Benchmark Tool filtered', {
        background: '#86efac',
        position: { top: '-14px', left: '0' }
      })
    }

    const selectedItem = [...document.querySelectorAll('.template-item')].find((item) => item.textContent?.includes('Editable MySQL Template'))
    if (selectedItem) {
      selectedItem.style.position = 'relative'
      selectedItem.style.outline = '1px dashed #c4b5fd'
      selectedItem.style.outlineOffset = '2px'
      addBadge(selectedItem, 'Edit scenario saved', {
        background: '#c4b5fd',
        position: { top: '8px', right: '8px' }
      })
    }
  })
}

async function main() {
  const browser = await chromium.launch({ headless: true })
  const page = await browser.newPage({ viewport: { width: 1200, height: 1000 }, deviceScaleFactor: 1 })

  await page.addInitScript((seedTemplates) => {
    const state = {
      templates: JSON.parse(JSON.stringify(seedTemplates))
    }
    window.__templateTestState = state
    window.runtime = new Proxy({
      EventsOnMultiple() { return () => {} },
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

    const clone = (value) => JSON.parse(JSON.stringify(value))

    window.go = {
      bindings: {
        TemplateBinding: {
          ListTemplates: async () => ({ templates: clone(state.templates) }),
          CreateTemplate: async (template) => {
            state.templates.unshift(clone(template))
            return { template: clone(template) }
          },
          UpdateTemplate: async (template) => {
            state.templates = state.templates.map((item) => (item.id === template.id ? clone(template) : item))
            return { template: clone(template) }
          },
          DeleteTemplate: async (id) => {
            state.templates = state.templates.filter((item) => item.id !== id)
            return { success: true }
          },
          DuplicateTemplate: async (id) => {
            const source = state.templates.find((item) => item.id === id)
            return { template: clone(source) }
          }
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
        }
      }
    }
  }, templates)

  await page.goto(baseUrl, { waitUntil: 'load' })
  await page.getByRole('button', { name: /Templates/ }).click()
  await page.getByRole('button', { name: 'New Template' }).click()
  await page.waitForSelector('.editor-dialog')

  const dbSelect = page.locator('.field', { hasText: 'Database Type' }).locator('select')
  const toolSelect = page.locator('.field', { hasText: 'Benchmark Tool' }).locator('select')
  const workloadSelect = page.locator('.field', { hasText: 'Workload Family' }).locator('select')

  const readOptions = async (locator) => locator.locator('option').evaluateAll((options) => options.map((option) => ({
    value: option.value,
    text: option.textContent?.trim() || ''
  })))

  const createInitialTools = await readOptions(toolSelect)
  assert(createInitialTools.map((item) => item.value).join('|') === 'sysbench|hammerdb', `MySQL create tools should be sysbench|hammerdb, got ${createInitialTools.map((item) => item.value).join('|')}`)

  await dbSelect.selectOption('oracle')
  await page.waitForTimeout(100)
  const oracleTools = await readOptions(toolSelect)
  const oracleToolValue = await toolSelect.inputValue()
  const oracleWorkloads = await readOptions(workloadSelect)
  assert(oracleTools.map((item) => item.value).join('|') === 'swingbench|hammerdb', `Oracle tools should be swingbench|hammerdb, got ${oracleTools.map((item) => item.value).join('|')}`)
  assert(oracleToolValue === 'swingbench', `Oracle create tool should reset to swingbench, got ${oracleToolValue}`)
  assert(oracleWorkloads.length > 0, 'Oracle workload options should refresh after db change')

  await toolSelect.selectOption('hammerdb')
  await dbSelect.selectOption('mysql')
  await page.waitForTimeout(100)
  const mysqlToolsAfterSwitch = await readOptions(toolSelect)
  const mysqlToolValueAfterSwitch = await toolSelect.inputValue()
  assert(mysqlToolsAfterSwitch.map((item) => item.value).join('|') === 'sysbench|hammerdb', `MySQL tools after switch should be sysbench|hammerdb, got ${mysqlToolsAfterSwitch.map((item) => item.value).join('|')}`)
  assert(mysqlToolValueAfterSwitch === 'hammerdb', `Legal tool should be preserved after db change, got ${mysqlToolValueAfterSwitch}`)

  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await page.waitForTimeout(100)
  await page.getByRole('button', { name: 'Close', exact: true }).click()
  await page.waitForTimeout(100)

  await page.getByText('Editable MySQL Template').click()
  await page.getByRole('button', { name: 'Edit Template' }).click()
  await page.waitForTimeout(100)
  const editDbSelect = page.locator('.field', { hasText: 'Database Type' }).locator('select')
  const editToolSelect = page.locator('.field', { hasText: 'Benchmark Tool' }).locator('select')
  await editDbSelect.selectOption('oracle')
  await page.waitForTimeout(100)
  const editOracleTools = await readOptions(editToolSelect)
  const editOracleValue = await editToolSelect.inputValue()
  assert(editOracleTools.map((item) => item.value).join('|') === 'swingbench|hammerdb', `Edit oracle tools should be swingbench|hammerdb, got ${editOracleTools.map((item) => item.value).join('|')}`)
  assert(editOracleValue === 'swingbench', `Edit tool should reset to swingbench for oracle, got ${editOracleValue}`)
  await page.getByRole('button', { name: 'Save', exact: true }).click()
  await page.waitForTimeout(100)

  const storedTemplates = await page.evaluate(() => window.__templateTestState.templates)
  const editedTemplate = storedTemplates.find((item) => item.id === 'tpl-user-mysql')
  assert(editedTemplate?.dbFamily === 'oracle', `Edited template dbFamily should persist as oracle, got ${editedTemplate?.dbFamily}`)
  assert(editedTemplate?.tool === 'swingbench', `Edited template tool should persist as swingbench, got ${editedTemplate?.tool}`)
  assert(Array.isArray(editedTemplate?.database_types) && editedTemplate.database_types[0] === 'oracle', 'Edited template database_types should stay aligned with dbFamily')

  await page.screenshot({ path: screenshotPath, fullPage: true })
  if (annotatedScreenshotPath) {
    await annotate(page)
    await page.screenshot({ path: annotatedScreenshotPath, fullPage: true })
  }

  console.log(JSON.stringify({
    screenshotPath,
    annotatedScreenshotPath: annotatedScreenshotPath || null,
    createInitialTools,
    oracleTools,
    oracleToolValue,
    mysqlToolsAfterSwitch,
    mysqlToolValueAfterSwitch,
    editOracleTools,
    editOracleValue,
    savedTemplate: {
      dbFamily: editedTemplate?.dbFamily || '',
      tool: editedTemplate?.tool || '',
      databaseTypes: editedTemplate?.database_types || []
    }
  }, null, 2))

  await browser.close()
}

main().catch((error) => {
  console.error(error.message)
  process.exit(1)
})
