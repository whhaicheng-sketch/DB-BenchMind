import { createDefaultTemplate } from '../models/template'

export const templateMocks = [
  createDefaultTemplate({
    id: 'tpl_sys_mysql_rw',
    name: 'Sysbench-MySQL-OLTP_RW-10x100k-32th-300s',
    description: 'Built-in MySQL OLTP read/write baseline for fast smoke validation.',
    tool: 'sysbench',
    dbFamily: 'mysql',
    workloadFamily: 'oltp-read-write',
    scope: 'builtin',
    status: 'ready',
    tags: ['mysql', 'baseline', 'oltp'],
    version: '1.0.0',
    compatibility: {
      supportedDatabases: ['mysql'],
      supportedVersions: ['MySQL 8.0+', 'MariaDB 10.6+'],
      compatibilityNotes: 'Prepared for transactional OLTP verification and regression checks.',
      requiresPrivileges: ['CREATE', 'DROP'],
      constraints: ['Connection binding happens in Tasks & Monitor.']
    },
    phases: {
      prepare: { enabled: true, required: false, params: {} },
      warmup: { enabled: true, required: false, params: {} },
      run: { enabled: true, required: true, params: {} },
      cleanup: { enabled: true, required: false, params: {} }
    },
    runtime: {
      concurrency: { mode: 'threads', value: 32 },
      durationSeconds: 300,
      warmupSeconds: 30,
      rampUpSeconds: 15,
      reportIntervalSeconds: 10,
      percentile: 95,
      iterations: 0,
      rateLimit: 0,
      validationEnabled: true,
      notes: 'Recommended as default baseline before environment-specific tuning.'
    },
    toolConfig: {
      sysbench: {
        dbDriver: 'mysql',
        scriptType: 'oltp_read_write',
        tables: 10,
        tableSize: 100000,
        reportChecks: true,
        extraCliArgs: '--db-ps-mode=disable'
      }
    }
  }),
  createDefaultTemplate({
    id: 'tpl_sys_mysql_ro',
    name: 'Sysbench-MySQL-OLTP_RO-20x500k-48th-600s',
    description: 'Read-heavy MySQL profile for throughput observation on warmed datasets.',
    tool: 'sysbench',
    dbFamily: 'mysql',
    workloadFamily: 'oltp-read-only',
    scope: 'builtin',
    status: 'ready',
    tags: ['mysql', 'read-only', 'throughput'],
    version: '1.0.0',
    runtime: {
      concurrency: { mode: 'threads', value: 48 },
      durationSeconds: 600,
      warmupSeconds: 45,
      rampUpSeconds: 30,
      reportIntervalSeconds: 15,
      percentile: 99,
      iterations: 0,
      rateLimit: 0,
      validationEnabled: false,
      notes: 'Useful for read pool sizing and cache hit evaluation.'
    },
    toolConfig: {
      sysbench: {
        dbDriver: 'mysql',
        scriptType: 'oltp_read_only',
        tables: 20,
        tableSize: 500000,
        extraCliArgs: '--rand-type=uniform'
      }
    }
  }),
  createDefaultTemplate({
    id: 'tpl_sys_pg_rw',
    name: 'Sysbench-PostgreSQL-OLTP_RW-16x200k-24th-420s',
    description: 'PostgreSQL OLTP mixed workload prepared for stage-by-stage validation.',
    tool: 'sysbench',
    dbFamily: 'postgresql',
    workloadFamily: 'oltp-read-write',
    scope: 'user',
    status: 'draft',
    tags: ['postgresql', 'custom', 'rw'],
    version: '0.9.0',
    phases: {
      prepare: { enabled: true, required: false, params: {} },
      warmup: { enabled: true, required: false, params: {} },
      run: { enabled: true, required: true, params: {} },
      verify: { enabled: true, required: false, params: {} }
    },
    runtime: {
      concurrency: { mode: 'threads', value: 24 },
      durationSeconds: 420,
      warmupSeconds: 60,
      rampUpSeconds: 20,
      reportIntervalSeconds: 10,
      percentile: 95,
      iterations: 0,
      rateLimit: 3000,
      validationEnabled: true,
      notes: 'Team draft for PG transactional tuning.'
    },
    toolConfig: {
      sysbench: {
        dbDriver: 'pgsql',
        scriptType: 'oltp_read_write',
        tables: 16,
        tableSize: 200000,
        extraCliArgs: '--pgsql-variant=redshift'
      }
    }
  }),
  createDefaultTemplate({
    id: 'tpl_sys_mysql_smoke',
    name: 'Sysbench-MySQL-Minimal-Test-1x1k-4th-60s',
    description: 'Minimal MySQL sysbench smoke template where prepare rebuilds the environment, run can be repeated, and cleanup fully removes the benchmark state.',
    tool: 'sysbench',
    dbFamily: 'mysql',
    workloadFamily: 'oltp-read-write',
    scope: 'test',
    status: 'ready',
    tags: ['test', 'smoke', 'minimal'],
    version: '0.1.0',
    phases: {
      prepare: { enabled: true, required: false, params: {} },
      run: { enabled: true, required: true, params: {} },
      cleanup: { enabled: true, required: false, params: {} }
    },
    runtime: {
      concurrency: { mode: 'threads', value: 4 },
      durationSeconds: 60,
      warmupSeconds: 5,
      rampUpSeconds: 5,
      reportIntervalSeconds: 10,
      percentile: 95,
      iterations: 0,
      rateLimit: 0,
      validationEnabled: false,
      notes: 'Low-volume smoke run for fast verification, conservative setup and minimal initialization rather than performance accuracy.'
    },
    toolConfig: {
      sysbench: {
        dbDriver: 'mysql',
        scriptType: 'oltp_read_write',
        tables: 1,
        tableSize: 1000,
        reportChecks: true,
        extraCliArgs: '--time=60'
      }
    }
  }),
  createDefaultTemplate({
    id: 'tpl_test_postgresql_sysbench',
    name: 'PostgreSQL - Sysbench Test',
    description: 'Minimal PostgreSQL sysbench smoke template where prepare rebuilds the environment, run can be repeated, and cleanup fully removes the benchmark state.',
    tool: 'sysbench',
    dbFamily: 'postgresql',
    workloadFamily: 'oltp-read-write',
    scope: 'test',
    status: 'ready',
    tags: ['test', 'postgresql', 'sysbench', 'smoke'],
    version: '1.0.0',
    phases: {
      prepare: { enabled: true, required: false, params: {} },
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
      validationEnabled: true,
      notes: 'Lowest-volume PostgreSQL sysbench workflow intended to finish quickly and exercise the full chain.'
    },
    toolConfig: {
      sysbench: {
        dbDriver: 'pgsql',
        scriptType: 'oltp_read_write',
        tables: 1,
        tableSize: 1000,
        reportChecks: true,
        extraCliArgs: ''
      }
    }
  }),
  createDefaultTemplate({
    id: 'tpl_swing_oe',
    name: 'Swingbench-Oracle-OE-Medium-64u-30m',
    description: 'Built-in OrderEntry scenario aligned to Oracle 11g compatible target.',
    tool: 'swingbench',
    dbFamily: 'oracle',
    workloadFamily: 'order-entry',
    scope: 'builtin',
    status: 'ready',
    tags: ['oracle', 'orderentry', 'baseline'],
    version: '1.0.0',
    phases: {
      prepare: { enabled: true, required: false, params: {} },
      warmup: { enabled: true, required: false, params: {} },
      run: { enabled: true, required: true, params: {} },
      cleanup: { enabled: true, required: false, params: {} }
    },
    runtime: {
      concurrency: { mode: 'users', value: 64 },
      durationSeconds: 1800,
      warmupSeconds: 120,
      rampUpSeconds: 60,
      reportIntervalSeconds: 30,
      percentile: 95,
      iterations: 0,
      rateLimit: 0,
      validationEnabled: false,
      notes: 'Managed configuration, suitable for Oracle staging validation.'
    },
    toolConfig: {
      swingbench: {
        benchmark: 'orderEntry',
        frontend: 'charbench',
        configMode: 'managed',
        wizardOperation: 'generate',
        userCount: 64,
        runTimeSeconds: 1800,
        minThinkTime: 0,
        maxThinkTime: 2,
        xmlOverrides: '<override key=\"soe.do_not_exit\">true</override>'
      }
    }
  }),
  createDefaultTemplate({
    id: 'tpl_test_oracle_swingbench',
    name: 'Oracle - Swingbench Test',
    description: 'Minimal Oracle Swingbench smoke template where prepare rebuilds SOE from scratch, run can be repeated, and cleanup fully removes the benchmark state.',
    tool: 'swingbench',
    dbFamily: 'oracle',
    workloadFamily: 'order-entry',
    scope: 'test',
    status: 'ready',
    tags: ['test', 'oracle', 'swingbench', 'smoke'],
    version: '1.0.0',
    phases: {
      prepare: { enabled: true, required: false, params: {} },
      run: { enabled: true, required: true, params: {} },
      cleanup: { enabled: true, required: false, params: {} }
    },
    runtime: {
      concurrency: { mode: 'users', value: 1 },
      durationSeconds: 60,
      warmupSeconds: 0,
      rampUpSeconds: 0,
      reportIntervalSeconds: 5,
      percentile: 95,
      validationEnabled: true,
      notes: 'One-user Oracle smoke profile to validate schema build and charbench execution.'
    },
    toolConfig: {
      swingbench: {
        benchmark: 'orderEntry',
        frontend: 'charbench',
        configMode: 'managed',
        wizardOperation: 'generate',
        scale: 0.1,
        userCount: 1,
        runTimeSeconds: 60,
        minThinkTime: 0,
        maxThinkTime: 0,
        xmlOverrides: ''
      }
    }
  }),
  createDefaultTemplate({
    id: 'tpl_swing_sales',
    name: 'Swingbench-Oracle-SalesHistory-Analytics-48u-20m',
    description: 'SalesHistory profile with longer analytics bursts and report pacing.',
    tool: 'swingbench',
    dbFamily: 'oracle',
    workloadFamily: 'sales-history',
    scope: 'builtin',
    status: 'ready',
    tags: ['oracle', 'analytic', 'saleshistory'],
    version: '1.0.0',
    runtime: {
      concurrency: { mode: 'users', value: 48 },
      durationSeconds: 1200,
      warmupSeconds: 90,
      rampUpSeconds: 45,
      reportIntervalSeconds: 20,
      percentile: 95,
      iterations: 0,
      rateLimit: 0,
      validationEnabled: false,
      notes: 'Oriented to mixed report style interaction and think-time validation.'
    },
    toolConfig: {
      swingbench: {
        benchmark: 'salesHistory',
        frontend: 'swingbench',
        configMode: 'managed',
        wizardOperation: 'generate',
        userCount: 48,
        runTimeSeconds: 1200,
        minThinkTime: 1,
        maxThinkTime: 4,
        xmlOverrides: ''
      }
    }
  }),
  createDefaultTemplate({
    id: 'tpl_swing_oracle_smoke',
    name: 'Swingbench-Oracle-Minimal-Test-8u-10m',
    description: 'Minimal Oracle smoke test template for quick validation of the Swingbench execution path.',
    tool: 'swingbench',
    dbFamily: 'oracle',
    workloadFamily: 'order-entry',
    scope: 'test',
    status: 'ready',
    tags: ['test', 'smoke', 'oracle'],
    version: '0.1.0',
    phases: {
      prepare: { enabled: true, required: false, params: {} },
      run: { enabled: true, required: true, params: {} },
      cleanup: { enabled: true, required: false, params: {} }
    },
    runtime: {
      concurrency: { mode: 'users', value: 8 },
      durationSeconds: 600,
      warmupSeconds: 15,
      rampUpSeconds: 10,
      reportIntervalSeconds: 15,
      percentile: 95,
      iterations: 0,
      rateLimit: 0,
      validationEnabled: false,
      notes: 'Quick validation shell with a conservative Oracle workload for smoke runs and fast chain verification.'
    },
    toolConfig: {
      swingbench: {
        benchmark: 'orderEntry',
        frontend: 'charbench',
        configMode: 'managed',
        wizardOperation: 'generate',
        userCount: 8,
        runTimeSeconds: 600,
        minThinkTime: 0,
        maxThinkTime: 1,
        xmlOverrides: ''
      }
    }
  }),
  createDefaultTemplate({
    id: 'tpl_hammer_oracle_c',
    name: 'HammerDB-Oracle-TPROC-C-1000W-96vu-30m',
    description: 'Built-in Oracle TPROC-C workflow with build, run and cleanup stages.',
    tool: 'hammerdb',
    dbFamily: 'oracle',
    workloadFamily: 'tproc-c',
    scope: 'builtin',
    status: 'ready',
    tags: ['oracle', 'tproc-c', 'workflow'],
    version: '1.0.0',
    phases: {
      build: { enabled: true, required: false, params: {} },
      prepare: { enabled: true, required: false, params: {} },
      run: { enabled: true, required: true, params: {} },
      cleanup: { enabled: true, required: false, params: {} },
      delete: { enabled: true, required: false, params: {} }
    },
    runtime: {
      concurrency: { mode: 'virtualUsers', value: 96 },
      durationSeconds: 1800,
      warmupSeconds: 120,
      rampUpSeconds: 60,
      reportIntervalSeconds: 20,
      percentile: 95,
      iterations: 1,
      rateLimit: 0,
      validationEnabled: false,
      notes: 'Workflow model kept explicit for future driver/script mapping.'
    },
    toolConfig: {
      hammerdb: {
        benchmark: 'tproc-c',
        virtualUsers: 96,
        warehouses: 1000,
        scaleFactor: 10,
        timeProfile: true,
        stepTesting: true,
        xmlConnectPool: false,
        advancedNotes: 'Use for Oracle driver path validation only in this phase.'
      }
    }
  }),
  createDefaultTemplate({
    id: 'tpl_test_sqlserver_hammerdb',
    name: 'SQL Server - HammerDB Test',
    description: 'Minimal SQL Server HammerDB smoke template where prepare rebuilds the environment, run can be repeated, and cleanup fully removes the benchmark state.',
    tool: 'hammerdb',
    dbFamily: 'sqlserver',
    workloadFamily: 'tproc-c',
    scope: 'test',
    status: 'ready',
    tags: ['test', 'sqlserver', 'hammerdb', 'smoke'],
    version: '1.0.0',
    phases: {
      prepare: { enabled: true, required: false, params: {} },
      run: { enabled: true, required: true, params: {} },
      cleanup: { enabled: true, required: false, params: {} }
    },
    runtime: {
      concurrency: { mode: 'virtualUsers', value: 1 },
      durationSeconds: 60,
      warmupSeconds: 0,
      rampUpSeconds: 0,
      reportIntervalSeconds: 1,
      percentile: 95,
      iterations: 1,
      validationEnabled: true,
      notes: 'Timed single-user SQL Server smoke run intended to verify HammerDB task flow quickly.'
    },
    toolConfig: {
      hammerdb: {
        benchmark: 'tproc-c',
        virtualUsers: 1,
        warehouses: 1,
        scaleFactor: 1,
        timeProfile: true,
        stepTesting: false,
        xmlConnectPool: false,
        advancedNotes: ''
      }
    }
  }),
  createDefaultTemplate({
    id: 'tpl_swing_oracle_shared',
    name: 'Swingbench-Oracle-Shared-OrderEntry-24u-15m',
    description: 'Readonly shared Oracle template kept as a reference copy for controlled reuse.',
    tool: 'swingbench',
    dbFamily: 'oracle',
    workloadFamily: 'order-entry',
    scope: 'readonlyShared',
    status: 'ready',
    tags: ['oracle', 'shared', 'orderentry'],
    version: '1.0.0',
    phases: {
      prepare: { enabled: true, required: false, params: {} },
      warmup: { enabled: true, required: false, params: {} },
      run: { enabled: true, required: true, params: {} },
      cleanup: { enabled: true, required: false, params: {} }
    },
    runtime: {
      concurrency: { mode: 'users', value: 24 },
      durationSeconds: 900,
      warmupSeconds: 60,
      rampUpSeconds: 30,
      reportIntervalSeconds: 20,
      percentile: 95,
      iterations: 0,
      rateLimit: 0,
      validationEnabled: false,
      notes: 'Reference-only shared template. Save As before any customization.'
    },
    toolConfig: {
      swingbench: {
        benchmark: 'orderEntry',
        frontend: 'charbench',
        configMode: 'managed',
        wizardOperation: 'generate',
        userCount: 24,
        runTimeSeconds: 900,
        minThinkTime: 0,
        maxThinkTime: 2,
        xmlOverrides: ''
      }
    }
  }),
  createDefaultTemplate({
    id: 'tpl_hammer_pg_c',
    name: 'HammerDB-PostgreSQL-TPROC-C-300W-64vu-20m',
    description: 'User-maintained PostgreSQL TPROC-C profile for task creation demos.',
    tool: 'hammerdb',
    dbFamily: 'postgresql',
    workloadFamily: 'tproc-c',
    scope: 'user',
    status: 'draft',
    tags: ['postgresql', 'tproc-c', 'user'],
    version: '0.4.0',
    runtime: {
      concurrency: { mode: 'virtualUsers', value: 64 },
      durationSeconds: 1200,
      warmupSeconds: 60,
      rampUpSeconds: 45,
      reportIntervalSeconds: 15,
      percentile: 95,
      iterations: 1,
      rateLimit: 0,
      validationEnabled: false,
      notes: 'Cloned from baseline and adjusted for customer staging.'
    },
    toolConfig: {
      hammerdb: {
        benchmark: 'tproc-c',
        virtualUsers: 64,
        warehouses: 300,
        scaleFactor: 10,
        timeProfile: true,
        stepTesting: false,
        xmlConnectPool: true,
        advancedNotes: 'Enable metrics bridge after backend API is wired.'
      }
    }
  }),
  createDefaultTemplate({
    id: 'tpl_hammer_pg_h',
    name: 'HammerDB-PostgreSQL-TPROC-H-100SF-32vu-25m',
    description: 'TPROC-H modeling placeholder that stays editable for later backend mapping.',
    tool: 'hammerdb',
    dbFamily: 'postgresql',
    workloadFamily: 'tproc-h',
    scope: 'builtin',
    status: 'ready',
    tags: ['postgresql', 'tproc-h', 'analytic'],
    version: '1.0.0',
    runtime: {
      concurrency: { mode: 'virtualUsers', value: 32 },
      durationSeconds: 1500,
      warmupSeconds: 30,
      rampUpSeconds: 20,
      reportIntervalSeconds: 30,
      percentile: 95,
      iterations: 1,
      rateLimit: 0,
      validationEnabled: false,
      notes: 'Front-end model only, no execution binding in this phase.'
    },
    toolConfig: {
      hammerdb: {
        benchmark: 'tproc-h',
        virtualUsers: 32,
        warehouses: 10,
        scaleFactor: 100,
        timeProfile: false,
        stepTesting: false,
        xmlConnectPool: false,
        advancedNotes: 'Reserve for power and throughput test mapping.'
      }
    }
  }),
  createDefaultTemplate({
    id: 'tpl_hammer_pg_smoke',
    name: 'HammerDB-PostgreSQL-TPROC-C-Minimal-Test-10W-6vu-8m',
    description: 'Minimal HammerDB smoke test template for quick validation of the TPROC-C workflow chain.',
    tool: 'hammerdb',
    dbFamily: 'postgresql',
    workloadFamily: 'tproc-c',
    scope: 'test',
    status: 'ready',
    tags: ['test', 'minimal', 'quick-validation'],
    version: '0.1.0',
    phases: {
      build: { enabled: true, required: false, params: {} },
      run: { enabled: true, required: true, params: {} },
      cleanup: { enabled: true, required: false, params: {} }
    },
    runtime: {
      concurrency: { mode: 'virtualUsers', value: 6 },
      durationSeconds: 480,
      warmupSeconds: 10,
      rampUpSeconds: 10,
      reportIntervalSeconds: 20,
      percentile: 95,
      iterations: 1,
      rateLimit: 0,
      validationEnabled: false,
      notes: 'Reduced dataset intended for smoke runs, fast verification and easy workflow handoff checks instead of benchmark baselining.'
    },
    toolConfig: {
      hammerdb: {
        benchmark: 'tproc-c',
        virtualUsers: 6,
        warehouses: 10,
        scaleFactor: 1,
        timeProfile: false,
        stepTesting: false,
        xmlConnectPool: false,
        advancedNotes: 'Smoke template only.'
      }
    }
  }),
  createDefaultTemplate({
    id: 'tpl_sys_mysql_project',
    name: 'Sysbench-MySQL-Project-Checkout-12x50k-16th-240s',
    description: 'Project-scoped MySQL OLTP profile for shared team tuning within the current workspace.',
    tool: 'sysbench',
    dbFamily: 'mysql',
    workloadFamily: 'oltp-read-write',
    scope: 'project',
    status: 'ready',
    tags: ['project', 'mysql', 'team'],
    version: '0.8.0',
    phases: {
      prepare: { enabled: true, required: false, params: {} },
      warmup: { enabled: true, required: false, params: {} },
      run: { enabled: true, required: true, params: {} },
      verify: { enabled: true, required: false, params: {} }
    },
    runtime: {
      concurrency: { mode: 'threads', value: 16 },
      durationSeconds: 240,
      warmupSeconds: 20,
      rampUpSeconds: 10,
      reportIntervalSeconds: 10,
      percentile: 95,
      iterations: 0,
      rateLimit: 0,
      validationEnabled: true,
      notes: 'Project template kept editable in local mock state for team-level tuning.'
    },
    toolConfig: {
      sysbench: {
        dbDriver: 'mysql',
        scriptType: 'oltp_read_write',
        tables: 12,
        tableSize: 50000,
        reportChecks: true,
        extraCliArgs: '--db-ps-mode=disable'
      }
    }
  })
]
