/**
 * Impact Analysis Mock Data
 * Mock 数据提供器 - 用于前端演示和开发测试
 */

import {
  AnalysisStatus,
  EventLevel,
  EventType,
  ConsistencyResult,
  NodeStatus,
  NodeRole,
  ConnectionMode,
  WorkloadType
} from '../constants'
import {
  createAnalysisSession,
  createRuntimeEvent,
  createTrendPoint,
  createClusterStatus,
  createNodeInfo,
  createMySQLClusterConfig,
  createSummaryData
} from '../types'

/**
 * Mock MySQL Cluster Connection 列表
 */
export function getMockMySQLClusterConnections() {
  return [
    {
      id: 'conn-001',
      name: 'MySQL HA Cluster - Production',
      type: 'mysql',
      host: '192.168.1.100',
      port: 3306,
      database: 'benchmind_test',
      username: 'root',
      vip: '192.168.1.200',
      primaryNodeIp: '192.168.1.101',
      secondaryNodeIp: '192.168.1.102',
      defaultConnectionMode: ConnectionMode.LONG,
      defaultWorkloadType: WorkloadType.INSERT,
      defaultWriteRate: 100,
      consistencyCheckEnabled: true
    },
    {
      id: 'conn-002',
      name: 'MySQL MHA - Demo',
      type: 'mysql',
      host: '10.0.0.100',
      port: 3306,
      database: 'test_db',
      username: 'root',
      vip: '10.0.0.200',
      primaryNodeIp: '10.0.0.101',
      secondaryNodeIp: '10.0.0.102',
      defaultConnectionMode: ConnectionMode.SHORT,
      defaultWorkloadType: WorkloadType.SELECT,
      defaultWriteRate: 50,
      consistencyCheckEnabled: true
    }
  ]
}

/**
 * 生成 Mock 分析会话
 */
export function getMockAnalysisSession(status = AnalysisStatus.RUNNING) {
  return createAnalysisSession({
    sessionId: 'session-demo-001',
    connectionId: 'conn-001',
    status,
    startTime: Date.now() - 60000,
    workloadType: WorkloadType.INSERT,
    connectionMode: ConnectionMode.LONG,
    writeRate: 100,
    successCommitCount: 5420,
    selectSuccessCount: 0,
    errorCount: 23,
    interruptionDurationMs: 8500,
    rtoMs: 12000,
    consistencyResult: ConsistencyResult.PASSED,
    currentTps: 98,
    currentErrorRate: 0
  })
}

/**
 * 生成 Mock 事件流
 */
export function getMockEvents() {
  const now = Date.now()
  return [
    createRuntimeEvent({
      eventId: 'evt-001',
      timestamp: now - 60000,
      level: EventLevel.INFO,
      type: EventType.ANALYSIS_START,
      message: 'Impact analysis started'
    }),
    createRuntimeEvent({
      eventId: 'evt-002',
      timestamp: now - 59000,
      level: EventLevel.INFO,
      type: EventType.WORKLOAD_STARTED,
      message: 'Insert workload started with rate 100 TPS'
    }),
    createRuntimeEvent({
      eventId: 'evt-003',
      timestamp: now - 30000,
      level: EventLevel.WARN,
      type: EventType.ERROR_SPIKE,
      message: 'Error rate spike detected: 15 errors in 1 second'
    }),
    createRuntimeEvent({
      eventId: 'evt-004',
      timestamp: now - 28000,
      level: EventLevel.ERROR,
      type: EventType.SUCCESS_TPS_ZERO,
      message: 'Success TPS dropped to zero'
    }),
    createRuntimeEvent({
      eventId: 'evt-005',
      timestamp: now - 27000,
      level: EventLevel.ERROR,
      type: EventType.CONNECTION_FAILURE,
      message: 'Connection failure detected to VIP 192.168.1.200'
    }),
    createRuntimeEvent({
      eventId: 'evt-006',
      timestamp: now - 25000,
      level: EventLevel.WARN,
      type: EventType.VIP_SWITCH_DETECTED,
      message: 'VIP switch detected: 192.168.1.200 -> 192.168.1.101'
    }),
    createRuntimeEvent({
      eventId: 'evt-007',
      timestamp: now - 24000,
      level: EventLevel.WARN,
      type: EventType.PRIMARY_ROLE_CHANGED,
      message: 'Primary role changed from node-1 to node-2'
    }),
    createRuntimeEvent({
      eventId: 'evt-008',
      timestamp: now - 23000,
      level: EventLevel.WARN,
      type: EventType.FAILOVER_DETECTED,
      message: 'Failover completed successfully'
    }),
    createRuntimeEvent({
      eventId: 'evt-009',
      timestamp: now - 19500,
      level: EventLevel.SUCCESS,
      type: EventType.RECOVERY_DETECTED,
      message: 'Service recovered, TPS back to normal'
    }),
    createRuntimeEvent({
      eventId: 'evt-010',
      timestamp: now - 5000,
      level: EventLevel.INFO,
      type: EventType.CONSISTENCY_CHECK_START,
      message: 'Consistency check started'
    }),
    createRuntimeEvent({
      eventId: 'evt-011',
      timestamp: now - 1000,
      level: EventLevel.SUCCESS,
      type: EventType.CONSISTENCY_CHECK_PASSED,
      message: 'Consistency check passed: all committed data verified'
    })
  ]
}

/**
 * 生成 Mock 趋势图数据
 */
export function getMockTrendData() {
  const points = []
  const now = Date.now()
  const totalPoints = 120

  // 阶段定义
  const phase1End = 30    // 正常运行
  const phase2Start = 35  // 故障开始
  const phase2End = 55    // 故障期间
  const phase3Start = 60  // 恢复

  for (let i = 0; i < totalPoints; i++) {
    const timestamp = now - (totalPoints - i) * 1000
    let successTps = 0
    let errorCount = 0
    const eventMarkers = []

    if (i < phase1End) {
      // 正常运行阶段
      successTps = 95 + Math.random() * 10
      errorCount = Math.random() < 0.1 ? 1 : 0
    } else if (i >= phase2Start && i <= phase2End) {
      // 故障期间
      successTps = 0
      errorCount = 5 + Math.floor(Math.random() * 10)

      if (i === phase2Start) {
        eventMarkers.push(EventType.ERROR_SPIKE)
      }
      if (i === phase2Start + 2) {
        eventMarkers.push(EventType.SUCCESS_TPS_ZERO)
        eventMarkers.push(EventType.CONNECTION_FAILURE)
      }
      if (i === phase2Start + 5) {
        eventMarkers.push(EventType.VIP_SWITCH_DETECTED)
      }
      if (i === phase2Start + 7) {
        eventMarkers.push(EventType.FAILOVER_DETECTED)
      }
    } else if (i > phase2End && i < phase3Start) {
      // 恢复过渡
      successTps = (i - phase2End) * 10
      errorCount = Math.floor(Math.random() * 3)
    } else {
      // 恢复后正常运行
      successTps = 95 + Math.random() * 10
      errorCount = Math.random() < 0.1 ? 1 : 0
    }

    points.push(createTrendPoint({
      timestamp,
      successTps: Math.round(successTps),
      errorCount,
      eventMarkers
    }))
  }

  return points
}

/**
 * 生成 Mock 集群状态
 */
export function getMockClusterStatus() {
  return createClusterStatus({
    vip: '192.168.1.200',
    currentPrimary: '192.168.1.102',
    currentSecondary: '192.168.1.101',
    primaryStatus: NodeStatus.ONLINE,
    secondaryStatus: NodeStatus.ONLINE,
    lastRoleSwitchTime: Date.now() - 35000,
    nodes: [
      createNodeInfo({
        nodeId: 'node-1',
        ip: '192.168.1.101',
        port: 3306,
        role: NodeRole.SECONDARY,
        status: NodeStatus.ONLINE,
        lastCheckTime: Date.now()
      }),
      createNodeInfo({
        nodeId: 'node-2',
        ip: '192.168.1.102',
        port: 3306,
        role: NodeRole.PRIMARY,
        status: NodeStatus.ONLINE,
        lastCheckTime: Date.now()
      })
    ]
  })
}

/**
 * 生成 Mock 摘要数据
 */
export function getMockSummaryData() {
  return createSummaryData({
    businessInterruption: {
      value: 8500,
      unit: 'ms',
      label: 'Business Interruption'
    },
    rto: {
      value: 12000,
      unit: 'ms',
      label: 'RTO'
    },
    consistency: {
      result: ConsistencyResult.PASSED,
      label: 'Consistency'
    },
    commitStats: {
      successCount: 5420,
      errorCount: 23,
      label: 'Commit Stats'
    }
  })
}

/**
 * 模拟实时数据更新
 * @param {number} currentTps 当前 TPS
 * @param {boolean} isRunning 是否正在运行
 */
export function generateRealtimeUpdate(currentTps, isRunning = true) {
  const baseTps = isRunning ? currentTps : 0
  const variance = isRunning ? Math.random() * 10 - 5 : 0

  return {
    timestamp: Date.now(),
    successTps: Math.max(0, Math.round(baseTps + variance)),
    errorCount: isRunning && Math.random() < 0.05 ? 1 : 0
  }
}
