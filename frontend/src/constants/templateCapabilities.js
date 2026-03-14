import { DB_FAMILY_LABELS, TEMPLATE_TOOL_LABELS, TEMPLATE_TOOLS, WORKLOAD_LABELS } from '../models/template'

export const TEMPLATE_CAPABILITIES = {
  sysbench: {
    toolConfigKey: 'sysbench',
    dbFamilies: ['mysql', 'postgresql'],
    workloads: ['oltp-read-write', 'oltp-read-only', 'oltp-write-only', 'oltp-point-select'],
    concurrencyModes: ['threads'],
    allowedPhases: ['prepare', 'warmup', 'run', 'verify', 'cleanup'],
    requiredPhases: ['run'],
    defaultEnabledPhases: ['prepare', 'warmup', 'run', 'cleanup'],
    workloadFieldMap: {
      'oltp-read-write': { scriptType: 'oltp_read_write' },
      'oltp-read-only': { scriptType: 'oltp_read_only' },
      'oltp-write-only': { scriptType: 'oltp_write_only' },
      'oltp-point-select': { scriptType: 'oltp_point_select' }
    },
    requiredFields: ['toolConfig.sysbench.scriptType'],
    toolFields: [
      {
        key: 'scriptType',
        label: 'Workload Script',
        type: 'select',
        options: [
          { label: 'oltp_read_write', value: 'oltp_read_write' },
          { label: 'oltp_read_only', value: 'oltp_read_only' },
          { label: 'oltp_write_only', value: 'oltp_write_only' },
          { label: 'oltp_point_select', value: 'oltp_point_select' }
        ]
      },
      { key: 'tables', label: 'Tables', type: 'number', min: 1 },
      { key: 'tableSize', label: 'Table Size', type: 'number', min: 1 },
      { key: 'extraCliArgs', label: 'Additional Params', type: 'textarea' }
    ]
  },
  swingbench: {
    toolConfigKey: 'swingbench',
    dbFamilies: ['oracle'],
    workloads: ['order-entry', 'sales-history', 'stress-test'],
    concurrencyModes: ['users'],
    allowedPhases: ['build', 'generate', 'warmup', 'run', 'delete'],
    requiredPhases: ['run'],
    defaultEnabledPhases: ['build', 'generate', 'warmup', 'run'],
    workloadFieldMap: {
      'order-entry': { benchmark: 'orderEntry' },
      'sales-history': { benchmark: 'salesHistory' },
      'stress-test': { benchmark: 'stressTest' }
    },
    requiredFields: ['toolConfig.swingbench.benchmark'],
    toolFields: [
      {
        key: 'benchmark',
        label: 'Benchmark',
        type: 'select',
        options: [
          { label: 'OrderEntry', value: 'orderEntry' },
          { label: 'SalesHistory', value: 'salesHistory' },
          { label: 'StressTest', value: 'stressTest' }
        ]
      },
      {
        key: 'frontend',
        label: 'Frontend',
        type: 'select',
        options: [
          { label: 'Swingbench', value: 'swingbench' },
          { label: 'Charbench', value: 'charbench' },
          { label: 'Minibench', value: 'minibench' }
        ]
      },
      { key: 'userCount', label: 'Virtual Users', type: 'number', min: 1 },
      { key: 'xmlOverrides', label: 'Additional Params', type: 'textarea' }
    ]
  },
  hammerdb: {
    toolConfigKey: 'hammerdb',
    dbFamilies: ['oracle', 'sqlserver', 'db2', 'postgresql', 'mysql', 'mariadb'],
    workloads: ['tproc-c', 'tproc-h'],
    concurrencyModes: ['virtualUsers'],
    allowedPhases: ['build', 'prepare', 'run', 'verify', 'cleanup', 'delete'],
    requiredPhases: ['run'],
    defaultEnabledPhases: ['build', 'prepare', 'run', 'cleanup'],
    workloadFieldMap: {
      'tproc-c': { benchmark: 'tproc-c' },
      'tproc-h': { benchmark: 'tproc-h' }
    },
    requiredFieldsByWorkload: {
      'tproc-c': ['toolConfig.hammerdb.benchmark', 'toolConfig.hammerdb.warehouses'],
      'tproc-h': ['toolConfig.hammerdb.benchmark', 'toolConfig.hammerdb.scaleFactor']
    },
    toolFields: [
      {
        key: 'benchmark',
        label: 'Profile',
        type: 'select',
        options: [
          { label: 'TPROC-C', value: 'tproc-c' },
          { label: 'TPROC-H', value: 'tproc-h' }
        ]
      },
      { key: 'virtualUsers', label: 'Virtual Users', type: 'number', min: 1 },
      {
        key: 'warehouses',
        label: 'Warehouses',
        type: 'number',
        min: 1,
        visibleWhen: (template) => template.workloadFamily === 'tproc-c'
      },
      {
        key: 'scaleFactor',
        label: 'Scale Factor',
        type: 'number',
        min: 1,
        visibleWhen: (template) => template.workloadFamily === 'tproc-h'
      },
      { key: 'advancedNotes', label: 'Additional Params', type: 'textarea' }
    ]
  }
}

export const TOOL_OPTIONS = Object.entries(TEMPLATE_TOOL_LABELS).map(([value, label]) => ({ value, label }))
export const DB_OPTIONS = Object.entries(DB_FAMILY_LABELS).map(([value, label]) => ({ value, label }))
export const WORKLOAD_OPTIONS = Object.entries(WORKLOAD_LABELS).map(([value, label]) => ({ value, label }))

export function getCapabilityForTool(tool) {
  return TEMPLATE_CAPABILITIES[tool] || TEMPLATE_CAPABILITIES.sysbench
}

export function getToolsForDbFamily(dbFamily) {
  if (!dbFamily) return TEMPLATE_TOOLS
  return TEMPLATE_TOOLS.filter((tool) => TEMPLATE_CAPABILITIES[tool]?.dbFamilies.includes(dbFamily))
}

export function getDefaultToolForDbFamily(dbFamily, currentTool = '') {
  const tools = getToolsForDbFamily(dbFamily)
  if (currentTool && tools.includes(currentTool)) return currentTool
  if (tools.length > 0) return tools[0]
  return currentTool || TEMPLATE_TOOLS[0]
}
