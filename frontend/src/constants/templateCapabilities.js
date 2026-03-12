import { DB_FAMILY_LABELS, TEMPLATE_TOOL_LABELS, WORKLOAD_LABELS } from '../models/template'

export const TEMPLATE_CAPABILITIES = {
  sysbench: {
    dbFamilies: ['mysql', 'postgresql'],
    workloads: ['oltp-read-write', 'oltp-read-only', 'oltp-write-only', 'oltp-point-select'],
    concurrencyModes: ['threads'],
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
    dbFamilies: ['oracle'],
    workloads: ['order-entry', 'sales-history', 'stress-test'],
    concurrencyModes: ['users'],
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
    dbFamilies: ['oracle', 'sqlserver', 'db2', 'postgresql', 'mysql', 'mariadb'],
    workloads: ['tproc-c', 'tproc-h'],
    concurrencyModes: ['virtualUsers'],
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
