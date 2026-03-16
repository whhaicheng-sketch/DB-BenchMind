import { chromium } from 'playwright'

const baseUrl = process.env.TASKS_MONITOR_BASE_URL || 'http://127.0.0.1:4173'
const screenshotPath = process.env.TASKS_MONITOR_SCREENSHOT || 'artifacts/tasks-monitor-layout-fixed.png'
const annotatedScreenshotPath = process.env.TASKS_MONITOR_ANNOTATED_SCREENSHOT || ''
const viewportWidth = Number(process.env.TASKS_MONITOR_VIEWPORT_WIDTH || 1200)
const viewportHeight = Number(process.env.TASKS_MONITOR_VIEWPORT_HEIGHT || 1000)

const sampleSeries = (values) => values.map((value, index) => ({ ts: index + 1, value }))

const taskPayload = {
  id: 'task-layout-check',
  name: 'layout-check',
  status: 'running',
  current_phase: 'run',
  benchmark_tool: 'sysbench',
  action: 'run',
  started_at: '2026-03-13T10:00:00Z',
  created_at: '2026-03-13T10:00:00Z',
  readiness: {
    db_valid: true,
    db_message: 'ready',
    ssh_available: true,
    ssh_message: 'ssh metrics enabled'
  },
  template_snapshot: {
    id: 'tpl-1',
    name: 'sysbench oltp',
    tool: 'sysbench',
    db_family: 'mysql'
  },
  connection_snapshot: {
    id: 'conn-1',
    name: 'bench oracle',
    type: 'oracle',
    host: '127.0.0.1'
  },
  run_log_paths: {
    run: '/tmp/task.log'
  },
  log_tail: [],
  metrics: {
    system_enabled: true,
    system_message: '',
    tps: {
      current: 600.8,
      avg: 598.1,
      max: 634.6,
      cv: 0.024,
      series: sampleSeries([588, 602, 596, 614, 601, 620, 598, 612, 600, 607, 595, 601, 593, 609, 600, 611, 597, 604, 599, 600.8])
    },
    tpm: {
      current: 36048,
      avg: 35882,
      max: 38076,
      cv: 0.028,
      series: sampleSeries([35200, 35880, 35440, 36700, 35920, 37280, 35790, 36550, 36080, 36440, 35680, 36120, 35510, 36820, 36010, 36940, 35720, 36280, 35940, 36048])
    },
    cpu_user: { current: 42.1, series: sampleSeries([31, 42, 38, 44, 40, 46, 43, 48, 45, 42]) },
    cpu_sys: { current: 15.4, series: sampleSeries([11, 14, 12, 15, 13, 17, 15, 16, 14, 15]) },
    cpu_iowait: { current: 8.8, series: sampleSeries([7, 6, 8, 7, 9, 8, 9, 10, 8, 9]) },
    cpu_steal: { current: 3.1, series: sampleSeries([1.2, 1.8, 1.5, 2.2, 2.0, 2.6, 2.4, 2.8, 2.9, 3.1]) },
    disk_read_bps: { current: 16777216, series: sampleSeries([8e6, 12e6, 10e6, 13e6, 11e6, 14e6, 12e6, 16e6, 14e6, 16777216]) },
    disk_write_bps: { current: 8388608, series: sampleSeries([4e6, 7e6, 5e6, 7.5e6, 6e6, 8e6, 7.2e6, 7.8e6, 8.1e6, 8388608]) },
    disk_read_latency_ms: { current: 2.4, series: sampleSeries([1.1, 1.6, 1.3, 1.8, 1.4, 2.0, 1.7, 2.2, 2.0, 2.4]) },
    disk_write_latency_ms: { current: 3.2, series: sampleSeries([1.4, 1.8, 1.6, 2.1, 1.9, 2.5, 2.3, 2.7, 2.9, 3.2]) }
  }
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message)
  }
}

async function outlineForScreenshot(page) {
  await page.evaluate(() => {
    const addBadge = (target, text, styles = {}) => {
      if (!target) return
      const badge = document.createElement('div')
      badge.textContent = text
      badge.style.position = 'absolute'
      badge.style.padding = styles.padding || '2px 6px'
      badge.style.borderRadius = '999px'
      badge.style.font = styles.font || '600 9px/1 sans-serif'
      badge.style.color = styles.color || '#081018'
      badge.style.background = styles.background || '#f8fafc'
      badge.style.zIndex = '6'
      badge.style.whiteSpace = 'nowrap'
      Object.entries(styles.position || {}).forEach(([key, value]) => {
        badge.style[key] = value
      })
      target.appendChild(badge)
    }

    const markers = [
      { card: 'TPS', color: '#ff9f43', label: 'TPS boundary' },
      { card: 'TPM', color: '#34d399', label: 'TPM boundary' }
    ]
    markers.forEach(({ card: cardLabel, color, label }) => {
      const el = document.querySelector(`[data-metric-card="${cardLabel}"]`)
      if (!el) return
      el.style.position = 'relative'
      el.style.outline = `2px solid ${color}`
      el.style.outlineOffset = '-2px'
      const badge = document.createElement('div')
      badge.textContent = label
      badge.style.position = 'absolute'
      badge.style.top = '8px'
      badge.style.left = '8px'
      badge.style.padding = '2px 6px'
      badge.style.borderRadius = '999px'
      badge.style.font = '600 10px/1 sans-serif'
      badge.style.color = '#081018'
      badge.style.background = color
      badge.style.zIndex = '5'
      el.appendChild(badge)

      const value = el.querySelector('[data-testid="metric-current-value"]')
      if (value) {
        value.style.outline = `1px dashed ${color}`
        value.style.outlineOffset = '2px'
        addBadge(value, 'current value in-chart', {
          background: color,
          position: {
            top: '50%',
            left: 'calc(100% + 6px)',
            transform: 'translateY(-50%)'
          }
        })
      }

      const topTick = [...(el?.querySelectorAll('.chart-axis-left span') || [])]
        .filter((tick) => !tick.classList.contains('axis-unit'))
        .sort((a, b) => a.getBoundingClientRect().top - b.getBoundingClientRect().top)[0]
      if (topTick) {
        topTick.style.outline = `1px dashed ${color}`
        topTick.style.outlineOffset = '2px'
        addBadge(el, `${cardLabel} top tick`, {
          background: color,
          position: {
            top: `${Math.max(8, topTick.offsetTop - 10)}px`,
            left: '54px'
          }
        })
      }
    })

    const cpu = document.querySelector('[data-system-card="CPU"]')
    const disk = document.querySelector('[data-system-card="Disk IO"]')
    const cpuWrap = document.querySelector('[data-system-chart-wrap="CPU"]')
    const cpuShell = document.querySelector('[data-system-chart-shell="CPU"]')
    const diskWrap = document.querySelector('[data-system-chart-wrap="Disk IO"]')
    const diskShell = document.querySelector('[data-system-chart-shell="Disk IO"]')
    ;[cpu, disk].forEach((el, index) => {
      if (!el) return
      el.style.position = 'relative'
      el.style.outline = '2px dashed rgba(96, 165, 250, 0.95)'
      el.style.outlineOffset = '-2px'
      const badge = document.createElement('div')
      badge.textContent = index === 0 ? 'CPU clear below TPS' : 'DISK clear below TPM'
      badge.style.position = 'absolute'
      badge.style.top = '8px'
      badge.style.left = '8px'
      badge.style.padding = '2px 6px'
      badge.style.borderRadius = '999px'
      badge.style.font = '600 10px/1 sans-serif'
      badge.style.color = '#081018'
      badge.style.background = 'rgba(96, 165, 250, 0.95)'
      badge.style.zIndex = '5'
      el.appendChild(badge)
    })

    ;[
      { wrap: cpuWrap, shell: cpuShell, color: '#f59e0b', label: 'CPU inner chart frame' },
      { wrap: diskWrap, shell: diskShell, color: '#34d399', label: 'DISK IO inner chart frame' }
    ].forEach(({ wrap, shell, color, label }) => {
      if (!wrap || !shell) return
      wrap.style.position = 'relative'
      wrap.style.outline = `1px solid ${color}`
      wrap.style.outlineOffset = '-1px'
      shell.style.outline = `1px dashed ${color}`
      shell.style.outlineOffset = '-1px'
      addBadge(wrap, label, {
        background: color,
        position: {
          top: '6px',
          left: '8px'
        }
      })
    })

    const cpuTopTick = [...(cpu?.querySelectorAll('.chart-axis-left span') || [])].find((tick) => tick.textContent?.trim() === '100%')
    if (cpuTopTick) {
      cpuTopTick.style.outline = '1px dashed #34d399'
      cpuTopTick.style.outlineOffset = '2px'
      addBadge(cpu, 'CPU 100% top tick', {
        background: '#34d399',
        position: {
          top: `${Math.max(8, cpuTopTick.offsetTop + 32)}px`,
          left: '54px'
        }
      })
    }

    const cpuLegend = cpu?.querySelector('.chart-legend')
    if (cpuLegend) {
      cpuLegend.style.outline = '1px dashed #f8b4b4'
      cpuLegend.style.outlineOffset = '2px'
      addBadge(cpu, 'USER / SYS / IOWAIT / ST single row', {
        background: '#f8b4b4',
        position: {
          top: '8px',
          right: '8px'
        }
      })
    }

    const cpuSummary = cpu?.querySelector('.system-summary-lines')
    if (cpuSummary) {
      cpuSummary.style.outline = '1px dashed #facc15'
      cpuSummary.style.outlineOffset = '2px'
      addBadge(cpu, 'CPU summary includes ST', {
        background: '#fde68a',
        position: {
          top: '30px',
          right: '8px'
        }
      })
    }

    const diskLeftTopTick = [...(disk?.querySelectorAll('.chart-axis-left span') || [])]
      .filter((tick) => !tick.classList.contains('axis-unit'))
      .sort((a, b) => a.getBoundingClientRect().top - b.getBoundingClientRect().top)[0]
    const diskRightTopTick = [...(disk?.querySelectorAll('.chart-axis-right span') || [])]
      .filter((tick) => !tick.classList.contains('axis-unit'))
      .sort((a, b) => a.getBoundingClientRect().top - b.getBoundingClientRect().top)[0]
    if (diskLeftTopTick) {
      diskLeftTopTick.style.outline = '1px dashed #60a5fa'
      diskLeftTopTick.style.outlineOffset = '2px'
      addBadge(disk, 'Disk left tick has unit', {
        background: '#60a5fa',
        position: {
          top: `${Math.max(8, diskLeftTopTick.offsetTop + 32)}px`,
          left: '56px'
        }
      })
    }

    if (diskRightTopTick) {
      diskRightTopTick.style.outline = '1px dashed #86efac'
      diskRightTopTick.style.outlineOffset = '2px'
      addBadge(disk, 'Disk right tick has unit', {
        background: '#86efac',
        position: {
          top: `${Math.max(8, diskRightTopTick.offsetTop + 32)}px`,
          right: '56px'
        }
      })
    }

    const createTask = document.querySelector('[data-layout-section="create-task"]')
    if (createTask) {
      createTask.style.position = 'relative'
      createTask.style.outline = '2px solid rgba(248, 250, 252, 0.9)'
      createTask.style.outlineOffset = '-2px'
      addBadge(createTask, 'Create Task 250px', {
        background: '#f8fafc',
        position: {
          top: '8px',
          left: '8px'
        }
      })

      const connectionSelect = createTask.querySelectorAll('select')[1]
      if (connectionSelect) {
        connectionSelect.style.outline = '1px dashed #fde68a'
        connectionSelect.style.outlineOffset = '2px'
        addBadge(createTask, 'Connection name only', {
          background: '#fde68a',
          position: {
            top: `${Math.max(8, connectionSelect.offsetTop - 16)}px`,
            left: '8px'
          }
        })
      }
    }

    const monitorBoard = document.querySelector('.monitor-board')
    if (monitorBoard) {
      addBadge(monitorBoard, 'TPS / TPM unchanged', {
        background: '#c7d2fe',
        position: {
          top: '8px',
          right: '8px'
        }
      })
      addBadge(monitorBoard, 'Capacity removed', {
        background: '#bbf7d0',
        position: {
          bottom: '8px',
          right: '8px'
        }
      })
    }
  })
}

async function main() {
  const browser = await chromium.launch({ headless: true })
  const page = await browser.newPage({ viewport: { width: viewportWidth, height: viewportHeight }, deviceScaleFactor: 1 })

  await page.addInitScript((task) => {
    window.runtime = buildRuntimeStub()
    window.go = {
      bindings: {
        TaskBinding: {
          ValidateDraft: async () => ({ task: { readiness: task.readiness } }),
          CreateTask: async () => ({ task }),
          ListTasks: async () => ({ tasks: [task] }),
          StopTask: async () => ({ success: true }),
          GetTaskLogs: async () => ({ lines: [] })
        },
        ConnectionBinding: {
          ListConnections: async () => ({
            connections: [
              { id: 'conn-1', name: 'bench oracle', type: 'oracle', host: '127.0.0.1' },
              { id: 'conn-2', name: 'mysql load', type: 'mysql', host: '192.168.0.10' }
            ]
          })
        },
        TemplateBinding: {
          ListTemplates: async () => ({
            templates: [
              {
                id: 'tpl-1',
                name: 'sysbench oltp',
                description: '',
                tool: 'sysbench',
                db_family: 'oracle',
                database_types: ['oracle'],
                workload_family: 'oltp',
                tags: [],
                phases: {
                  prepare: { enabled: true },
                  run: { enabled: true },
                  cleanup: { enabled: true }
                }
              }
            ]
          })
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
    window.__TASKS_MONITOR_LAYOUT_TASK__ = task

    function buildRuntimeStub() {
      return new Proxy({
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
    }

  }, taskPayload)

  await page.goto(baseUrl, { waitUntil: 'load' })
  await page.getByRole('button', { name: '📊 Tasks & Monitor' }).click()
  await page.waitForSelector('[data-metric-card="TPS"]')
  await page.locator('[data-layout-section="create-task"] select').first().selectOption('oracle')
  await page.locator('[data-layout-section="create-task"] select').nth(1).selectOption('conn-1')

  const result = await page.evaluate(() => {
    const metricCards = ['TPS', 'TPM'].map((label) => {
      const card = document.querySelector(`[data-metric-card="${label}"]`)
      const head = card?.querySelector('.metric-card-head')
      const stats = card?.querySelector('.metric-stats')
      const chart = card?.querySelector('[data-testid="metric-chart-shell"]')
      const canvas = card?.querySelector('[data-testid="metric-chart-canvas"]')
      const value = card?.querySelector('[data-testid="metric-current-value"]')
      const svg = card?.querySelector('svg')
      const mainLine = card?.querySelector('.metric-line-main')
      const glowLine = card?.querySelector('.metric-line-glow')
      const ticks = [...(card?.querySelectorAll('.chart-axis-left span') || [])].map((tick) => ({
        text: tick.textContent?.trim() || '',
        rect: tick.getBoundingClientRect().toJSON()
      }))
      return {
        label,
        hasCard: !!card,
        headerText: head?.innerText?.replace(/\s+/g, ' ').trim() || '',
        hasChart: !!chart,
        hasCanvas: !!canvas,
        hasValue: !!value,
        card: card?.getBoundingClientRect().toJSON(),
        head: head?.getBoundingClientRect().toJSON(),
        stats: stats?.getBoundingClientRect().toJSON(),
        chart: chart?.getBoundingClientRect().toJSON(),
        canvas: canvas?.getBoundingClientRect().toJSON(),
        value: value?.getBoundingClientRect().toJSON(),
        svg: svg?.getBoundingClientRect().toJSON(),
        valueParentTestId: value?.parentElement?.dataset?.testid || null,
        mainStrokeWidth: Number(mainLine?.getAttribute('stroke-width') || 0),
        glowStrokeWidth: Number(glowLine?.getAttribute('stroke-width') || 0),
        ticks
      }
    })

    const cpuCard = document.querySelector('[data-system-card="CPU"]')
    const diskCard = document.querySelector('[data-system-card="Disk IO"]')
    const cpuHead = cpuCard?.querySelector('.system-card-head')
    const diskHead = diskCard?.querySelector('.system-card-head')
    const cpuWrap = cpuCard?.querySelector('[data-system-chart-wrap="CPU"]')
    const cpuShell = cpuCard?.querySelector('[data-system-chart-shell="CPU"]')
    const cpuCaption = cpuCard?.querySelector('.system-caption')
    const diskWrap = diskCard?.querySelector('[data-system-chart-wrap="Disk IO"]')
    const diskShell = diskCard?.querySelector('[data-system-chart-shell="Disk IO"]')
    const diskCaption = diskCard?.querySelector('.system-caption')
    const cpuSummary = cpuCard?.querySelector('.system-summary-lines')
    const diskSummary = diskCard?.querySelector('.system-summary-lines')
    const cpuLegend = cpuCard?.querySelector('.chart-legend')
    const cpuLegendItems = [...(cpuLegend?.querySelectorAll('.legend-item') || [])].map((item) => ({
      text: item.textContent?.replace(/\s+/g, ' ').trim() || '',
      rect: item.getBoundingClientRect().toJSON()
    }))
    const cpuLeftTicks = [...(cpuCard?.querySelectorAll('.chart-axis-left span') || [])].map((tick) => ({
      text: tick.textContent?.trim() || '',
      rect: tick.getBoundingClientRect().toJSON()
    }))
    const diskLeftTicks = [...(diskCard?.querySelectorAll('.chart-axis-left span') || [])].map((tick) => ({
      text: tick.textContent?.trim() || '',
      rect: tick.getBoundingClientRect().toJSON()
    }))
    const diskRightTicks = [...(diskCard?.querySelectorAll('.chart-axis-right span') || [])].map((tick) => ({
      text: tick.textContent?.trim() || '',
      rect: tick.getBoundingClientRect().toJSON()
    }))
    const diskLeftUnit = diskCard?.querySelector('[data-axis-unit="disk-left"]')
    const diskRightUnit = diskCard?.querySelector('[data-axis-unit="disk-right"]')
    const cpuAxisUnit = cpuCard?.querySelector('[data-axis-unit="cpu-left"]')
    const cpuLeftAxis = cpuCard?.querySelector('.chart-axis-left')
    const diskLeftAxis = diskCard?.querySelector('.chart-axis-left')
    const diskRightAxis = diskCard?.querySelector('.chart-axis-right')
    const cpuPolylines = [...(cpuCard?.querySelectorAll('polyline') || [])].map((line) => Number(line.getAttribute('stroke-width') || 0))
    const diskPolylines = [...(diskCard?.querySelectorAll('polyline') || [])].map((line) => Number(line.getAttribute('stroke-width') || 0))
    const createTask = document.querySelector('[data-layout-section="create-task"]')
    const leftColumn = document.querySelector('.left-column')
    const createTaskSelects = createTask?.querySelectorAll('select') || []
    const connectionSelect = createTaskSelects[1] || null
    const connectionOptions = [...(connectionSelect?.querySelectorAll('option') || [])].map((option) => ({
      value: option.value,
      text: option.textContent?.replace(/\s+/g, ' ').trim() || ''
    }))
    const selectedConnectionOption = connectionSelect?.selectedOptions?.[0] || null
    return {
      metricCards,
      viewport: {
        innerWidth: window.innerWidth,
        scrollWidth: document.documentElement.scrollWidth
      },
      cpu: {
        card: cpuCard?.getBoundingClientRect().toJSON(),
        head: cpuHead?.getBoundingClientRect().toJSON(),
        wrap: cpuWrap?.getBoundingClientRect().toJSON(),
        shell: cpuShell?.getBoundingClientRect().toJSON(),
        caption: cpuCaption?.getBoundingClientRect().toJSON(),
        summary: cpuSummary?.getBoundingClientRect().toJSON(),
        summaryText: cpuSummary?.innerText?.replace(/\s+/g, ' ').trim() || '',
        lineWidths: cpuPolylines,
        legend: cpuLegend?.getBoundingClientRect().toJSON(),
        legendItems: cpuLegendItems,
        polylineCount: cpuCard?.querySelectorAll('polyline').length || 0,
        leftTicks: cpuLeftTicks,
        leftAxis: cpuLeftAxis?.getBoundingClientRect().toJSON(),
        axisUnit: cpuAxisUnit?.getBoundingClientRect().toJSON(),
        axisUnitText: cpuAxisUnit?.textContent?.trim() || ''
      },
      disk: {
        card: diskCard?.getBoundingClientRect().toJSON(),
        head: diskHead?.getBoundingClientRect().toJSON(),
        wrap: diskWrap?.getBoundingClientRect().toJSON(),
        shell: diskShell?.getBoundingClientRect().toJSON(),
        caption: diskCaption?.getBoundingClientRect().toJSON(),
        summary: diskSummary?.getBoundingClientRect().toJSON(),
        summaryText: diskSummary?.innerText?.replace(/\s+/g, ' ').trim() || '',
        summaryLineCount: diskSummary?.querySelectorAll('.summary-line').length || 0,
        lineWidths: diskPolylines,
        leftTicks: diskLeftTicks,
        rightTicks: diskRightTicks,
        leftAxis: diskLeftAxis?.getBoundingClientRect().toJSON(),
        rightAxis: diskRightAxis?.getBoundingClientRect().toJSON(),
        leftUnit: diskLeftUnit?.getBoundingClientRect().toJSON(),
        rightUnit: diskRightUnit?.getBoundingClientRect().toJSON(),
        leftUnitText: diskLeftUnit?.textContent?.trim() || '',
        rightUnitText: diskRightUnit?.textContent?.trim() || ''
      },
      layout: {
        createTask: createTask?.getBoundingClientRect().toJSON(),
        leftColumn: leftColumn?.getBoundingClientRect().toJSON(),
        rightColumn: document.querySelector('.right-column')?.getBoundingClientRect().toJSON() || null,
        monitorBoard: document.querySelector('.monitor-board')?.getBoundingClientRect().toJSON() || null,
        monitorBoardHead: document.querySelector('.monitor-board .card-head')?.getBoundingClientRect().toJSON() || null,
        monitorBoardGrid: document.querySelector('.monitor-board-grid')?.getBoundingClientRect().toJSON() || null,
        capacityPanelPresent: !!document.querySelector('[data-capacity-panel]'),
        connectionOptions,
        selectedConnectionText: selectedConnectionOption?.textContent?.replace(/\s+/g, ' ').trim() || ''
      }
    }
  })

  for (const metric of result.metricCards) {
    assert(metric.hasCard, `${metric.label} card missing`)
    assert(metric.hasChart, `${metric.label} chart shell missing`)
    assert(metric.hasCanvas, `${metric.label} chart canvas missing`)
    assert(metric.hasValue, `${metric.label} current value missing`)
    assert(/AVG/i.test(metric.headerText) && /MAX/i.test(metric.headerText), `${metric.label} header is missing inline AVG/MAX`)
    assert(metric.stats.top <= metric.head.bottom, `${metric.label} AVG/MAX block still sits below header row`)
    assert(metric.valueParentTestId === 'metric-chart-canvas', `${metric.label} current value is not inside chart canvas`)
    assert(metric.mainStrokeWidth > 0 && metric.mainStrokeWidth <= 1.6, `${metric.label} main line is not thin enough`)
    assert(metric.glowStrokeWidth > 0 && metric.glowStrokeWidth <= 2.4, `${metric.label} glow line is too thick`)
    assert(metric.chart.top >= metric.card.top - 0.5 && metric.chart.bottom <= metric.card.bottom + 0.5, `${metric.label} chart shell escapes card`)
    assert(metric.svg.top >= metric.canvas.top - 0.5 && metric.svg.bottom <= metric.canvas.bottom + 0.5, `${metric.label} svg escapes chart canvas`)
    assert(metric.value.left >= metric.canvas.left && metric.value.right <= metric.canvas.right, `${metric.label} current value exceeds chart canvas horizontally`)
    assert(metric.value.top >= metric.canvas.top && metric.value.bottom <= metric.canvas.bottom, `${metric.label} current value exceeds chart canvas vertically`)
    assert(metric.ticks.length >= 4, `${metric.label} lost top-axis density`)
    const topTick = metric.ticks.reduce((best, tick) => (best === null || tick.rect.top < best.rect.top ? tick : best), null)
    assert(topTick.rect.top >= metric.chart.top, `${metric.label} top tick is clipped above chart shell`)
    assert(topTick.rect.bottom <= metric.chart.bottom, `${metric.label} top tick escapes chart shell`)
  }

  const tps = result.metricCards.find((item) => item.label === 'TPS')
  const tpm = result.metricCards.find((item) => item.label === 'TPM')
  assert(tps.card.bottom <= result.cpu.card.top + 0.5, 'TPS card overlaps CPU card')
  assert(tpm.card.bottom <= result.disk.card.top + 0.5, 'TPM card overlaps Disk IO card')
  assert(result.cpu.head.height <= 64, 'CPU header block is still too tall')
  assert(result.disk.head.height <= 64, 'Disk IO header block is still too tall')
  assert(Math.abs(result.cpu.head.height - result.disk.head.height) <= 4, 'CPU and Disk IO header heights are misaligned')
  assert(result.cpu.wrap.top <= result.cpu.head.bottom + 6, 'CPU chart frame still sits too low under the header')
  assert(result.disk.wrap.top <= result.disk.head.bottom + 6, 'Disk IO chart frame still sits too low under the header')
  assert(result.cpu.wrap.top < result.cpu.card.top + 72, 'CPU chart frame did not move upward enough')
  assert(result.disk.wrap.top < result.disk.card.top + 72, 'Disk IO chart frame did not move upward enough')
  assert(result.cpu.wrap.bottom <= result.cpu.card.bottom - 12, 'CPU chart frame still leaks below the card')
  assert(result.disk.wrap.bottom <= result.disk.card.bottom - 12, 'Disk IO chart frame still leaks below the card')
  assert(result.cpu.shell.bottom <= result.cpu.caption.top + 0.5, 'CPU chart shell still collides with the caption area')
  assert(result.disk.shell.bottom <= result.disk.caption.top + 0.5, 'Disk IO chart shell still collides with the caption area')
  assert(result.disk.summaryLineCount <= 2, 'Disk IO summary still expands beyond compact layout')
  assert(result.disk.summaryText.includes('read') && result.disk.summaryText.includes('write') && result.disk.summaryText.includes('r_lat') && result.disk.summaryText.includes('w_lat'), 'Disk IO summary lost required fields')
  assert(!/\n/.test(result.disk.summaryText), 'Disk IO summary is not kept compact enough at 1200px width')
  assert(result.cpu.lineWidths.length > 0 && result.cpu.lineWidths.every((width) => width > 0 && width <= 1.6), 'CPU lines are not uniformly thin enough')
  assert(result.disk.lineWidths.length > 0 && result.disk.lineWidths.every((width) => width > 0 && width <= 1.6), 'Disk IO lines are not uniformly thin enough')
  assert(result.cpu.legendItems.map((item) => item.text).join('|') === 'USER|SYS|IOWAIT|ST', 'CPU legend text changed or order regressed')
  assert(result.cpu.legendItems.length === 4, 'CPU legend item count changed')
  assert(result.cpu.legend && result.cpu.legend.height <= 18, 'CPU legend wrapped to multiple rows')
  assert(result.cpu.legend.right <= result.cpu.card.right + 0.5, 'CPU legend overflows card width')
  assert(result.cpu.polylineCount === 4, `CPU chart should render 4 lines including ST, got ${result.cpu.polylineCount}`)
  assert(result.cpu.summaryText.includes('ST 3.1%'), `CPU summary lost ST, got ${result.cpu.summaryText}`)
  const cpuTopTick = result.cpu.leftTicks.find((tick) => tick.text === '100%')
  assert(cpuTopTick, 'CPU top 100% tick missing')
  assert(cpuTopTick.rect.top >= result.cpu.card.top, 'CPU 100% tick is clipped')
  assert(cpuTopTick.rect.left >= result.cpu.leftAxis.left - 0.5, 'CPU 100% tick is clipped on the left edge')
  assert(cpuTopTick.rect.right <= result.cpu.leftAxis.right + 0.5, 'CPU 100% tick escapes CPU left axis on the right edge')
  assert(!result.cpu.axisUnit, 'CPU top standalone % marker should be removed')
  assert(result.cpu.leftTicks.every((tick) => tick.text === '' || tick.text.endsWith('%')), `CPU left axis ticks must include %, got ${result.cpu.leftTicks.map((tick) => tick.text).join(', ')}`)
  const diskTopLeftTick = result.disk.leftTicks.reduce((best, tick) => (best === null || tick.rect.top < best.rect.top ? tick : best), null)
  const diskTopRightTick = result.disk.rightTicks.reduce((best, tick) => (best === null || tick.rect.top < best.rect.top ? tick : best), null)
  assert(diskTopLeftTick && diskTopLeftTick.rect.top >= result.disk.card.top, 'Disk IO left top tick is clipped')
  assert(diskTopRightTick && diskTopRightTick.rect.top >= result.disk.card.top, 'Disk IO right top tick is clipped')
  assert(diskTopLeftTick.rect.left >= result.disk.leftAxis.left - 0.5, 'Disk IO left top tick is clipped on the left edge')
  assert(diskTopLeftTick.rect.right <= result.disk.leftAxis.right + 0.5, 'Disk IO left top tick escapes left axis')
  assert(diskTopRightTick.rect.left >= result.disk.rightAxis.left - 0.5, 'Disk IO right top tick escapes right axis')
  assert(diskTopRightTick.rect.right <= result.disk.rightAxis.right + 0.5, 'Disk IO right top tick is clipped on the right edge')
  assert(!result.disk.leftUnit, 'Disk IO top standalone B/S marker should be removed')
  assert(!result.disk.rightUnit, 'Disk IO top standalone MS marker should be removed')
  assert(result.disk.leftTicks.every((tick) => tick.text === '' || /(?:B\/s|KB\/s|MB\/s|GB\/s)$/.test(tick.text)), `Disk IO left axis ticks must include bandwidth units, got ${result.disk.leftTicks.map((tick) => tick.text).join(', ')}`)
  assert(result.disk.rightTicks.every((tick) => tick.text === '' || tick.text.endsWith(' ms')), `Disk IO right axis ticks must include ms, got ${result.disk.rightTicks.map((tick) => tick.text).join(', ')}`)
  assert(result.viewport.scrollWidth <= result.viewport.innerWidth, 'Page introduces horizontal scrolling at 1200px width')
  assert(result.layout.createTask, 'Create Task card missing layout hook')
  assert(Math.abs(result.layout.leftColumn.width - 250) <= 2, `Create Task column width is ${result.layout.leftColumn.width}, expected 250px`)
  assert(!result.layout.capacityPanelPresent, 'Capacity panel should be removed from Tasks & Monitor')
  const actualConnectionOptions = result.layout.connectionOptions.filter((option) => option.value)
  assert(actualConnectionOptions.some((option) => option.text === 'bench oracle'), `Connection option text should show only name, got ${actualConnectionOptions.map((option) => option.text).join(', ')}`)
  assert(actualConnectionOptions.every((option) => !/[·]/.test(option.text)), `Connection options still append database type: ${actualConnectionOptions.map((option) => option.text).join(', ')}`)
  assert(result.layout.selectedConnectionText === 'bench oracle', `Selected connection text should show only name, got ${result.layout.selectedConnectionText}`)

  if (viewportWidth === 1200 && viewportHeight === 800) {
    assert(result.layout.monitorBoard.height >= result.layout.rightColumn.height - 2, `Monitor Overview should fill the right column at 1200x800, got ${result.layout.monitorBoard.height}px of ${result.layout.rightColumn.height}px`)
    assert(result.layout.monitorBoardGrid.height >= 550, `Monitor Overview content area is still too short at 1200x800: ${result.layout.monitorBoardGrid.height}px`)
    assert(tps.card.height >= 244, `TPS card is still too short at 1200x800: ${tps.card.height}px`)
    assert(tpm.card.height >= 244, `TPM card is still too short at 1200x800: ${tpm.card.height}px`)
    assert(result.cpu.card.height >= 278, `CPU card is still too short at 1200x800: ${result.cpu.card.height}px`)
    assert(result.disk.card.height >= 278, `Disk IO card is still too short at 1200x800: ${result.disk.card.height}px`)
    assert(tps.chart.height >= 170, `TPS chart shell is still too flat at 1200x800: ${tps.chart.height}px`)
    assert(tpm.chart.height >= 170, `TPM chart shell is still too flat at 1200x800: ${tpm.chart.height}px`)
    assert(result.cpu.wrap.height >= 146, `CPU chart frame is still too short at 1200x800: ${result.cpu.wrap.height}px`)
    assert(result.disk.wrap.height >= 146, `Disk IO chart frame is still too short at 1200x800: ${result.disk.wrap.height}px`)
    assert(result.cpu.shell.height >= 132, `CPU chart shell is still too flat at 1200x800: ${result.cpu.shell.height}px`)
    assert(result.disk.shell.height >= 132, `Disk IO chart shell is still too flat at 1200x800: ${result.disk.shell.height}px`)
  }

  await page.screenshot({ path: screenshotPath, fullPage: true })
  if (annotatedScreenshotPath) {
    await outlineForScreenshot(page)
    await page.screenshot({ path: annotatedScreenshotPath, fullPage: true })
  }
  console.log(JSON.stringify({
    screenshotPath,
    annotatedScreenshotPath: annotatedScreenshotPath || null,
    metrics: result.metricCards.map((metric) => ({
      label: metric.label,
      valueInsideChart: metric.valueParentTestId === 'metric-chart-canvas',
      mainStrokeWidth: metric.mainStrokeWidth,
      glowStrokeWidth: metric.glowStrokeWidth,
      chartWithinCard: metric.chart.top >= metric.card.top - 0.5 && metric.chart.bottom <= metric.card.bottom + 0.5
    })),
    cpuClearOfMetrics: tps.card.bottom <= result.cpu.card.top + 0.5,
    diskClearOfMetrics: tpm.card.bottom <= result.disk.card.top + 0.5,
    cpuHeaderHeight: result.cpu.head.height,
    diskHeaderHeight: result.disk.head.height,
    diskSummaryText: result.disk.summaryText,
    cpuLegendTexts: result.cpu.legendItems.map((item) => item.text),
    viewport: result.viewport,
    cpuSummary: result.cpu.summaryText,
    createTaskWidth: result.layout.leftColumn.width,
    monitorBoardHeight: result.layout.monitorBoard.height,
    monitorBoardGridHeight: result.layout.monitorBoardGrid.height,
    rightColumnHeight: result.layout.rightColumn.height,
    tpsCardHeight: tps.card.height,
    tpmCardHeight: tpm.card.height,
    cpuCardHeight: result.cpu.card.height,
    diskCardHeight: result.disk.card.height,
    tpsShellHeight: tps.chart.height,
    tpmShellHeight: tpm.chart.height,
    cpuWrapHeight: result.cpu.wrap.height,
    diskWrapHeight: result.disk.wrap.height,
    cpuShellHeight: result.cpu.shell.height,
    diskShellHeight: result.disk.shell.height
  }, null, 2))
  await browser.close()
}

main().catch((error) => {
  console.error(error.stack || error.message)
  process.exit(1)
})
