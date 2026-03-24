import { createDefaultTemplate } from '../models/template'

export const templateMocks = [
  createDefaultTemplate({
    id: 'tpl_test_mysql_sysbench',
    name: 'MySQL - Sysbench Test',
    description: 'Minimal MySQL sysbench smoke template where prepare rebuilds the environment, run can be repeated, and cleanup fully removes the benchmark state.',
    tool: 'sysbench',
    dbFamily: 'mysql',
    workloadFamily: 'oltp-read-write',
    is_builtin: true,
    version: '1.0.0',
    phases: {
      prepare: { enabled: true, required: false, params: {} },
      warmup: { enabled: true, required: false, params: {} },
      run: { enabled: true, required: true, params: {} },
      cleanup: { enabled: true, required: false, params: {} }
    },
    runtime: {
      concurrency: { mode: 'threads', value: 1 },
      durationSeconds: 30,
      warmupSeconds: 0,
      rampUpSeconds: 3,
      reportIntervalSeconds: 1,
      percentile: 95,
      iterations: 0,
      rateLimit: 0,
      validationEnabled: true,
      notes: 'Lowest-volume sysbench workflow intended to finish quickly and exercise the full chain.'
    },
    compatibility: {
      supportedDatabases: ['mysql'],
      supportedVersions: ['MySQL 8.0+', 'MariaDB 10.6+'],
      compatibilityNotes: 'Minimal dataset for functional validation, not performance baseline testing.',
      requiresPrivileges: ['CREATE', 'DROP'],
      constraints: ['Use for smoke validation and task-chain verification.']
    },
    toolConfig: {
      sysbench: {
        dbDriver: 'mysql',
        scriptType: 'oltp_read_write',
        tables: 1,
        tableSize: 1000,
        reportChecks: true,
        extraCliArgs: ''
      }
    }
  })
]
