/**
 * Impact Analysis Types
 * 数据模型定义 - 分析会话、事件、趋势点、集群状态等
 *
 * 注意：项目使用 JavaScript，此文件定义工厂函数和模型结构
 */

import {
  AnalysisStatus,
  PageState,
  ConnectionMode,
  WorkloadType,
  EventLevel,
  EventType,
  ConsistencyResult,
  NodeStatus,
  NodeRole
} from '../constants'

/**
 * 生成简单 UUID
 */
function generateId() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0
    const v = c === 'x' ? r : (r & 0x3 | 0x8)
    return v.toString(16)
  })
}

/**
 * 创建 MySQL Cluster Connection 配置
 */
export function createMySQLClusterConfig(partial = {}) {
  return {
    // 基础字段
    connectionId: partial.connectionId || '',
    connectionName: partial.connectionName || '',

    // 集群配置
    vip: partial.vip || '',
    primaryNodeIp: partial.primaryNodeIp || '',
    secondaryNodeIp: partial.secondaryNodeIp || '',
    nodes: partial.nodes || [],

    // 分析默认配置
    defaultConnectionMode: partial.defaultConnectionMode || ConnectionMode.LONG,
    defaultWorkloadType: partial.defaultWorkloadType || WorkloadType.INSERT,
    defaultWriteRate: partial.defaultWriteRate || 100,
    consistencyCheckEnabled: partial.consistencyCheckEnabled ?? true
  }
}

/**
 * 创建分析会话
 */
export function createAnalysisSession(partial = {}) {
  return {
    sessionId: partial.sessionId || generateId(),
    connectionId: partial.connectionId || '',
    status: partial.status || AnalysisStatus.IDLE,

    // 时间信息
    startTime: partial.startTime || null,
    endTime: partial.endTime || null,

    // 配置
    workloadType: partial.workloadType || WorkloadType.INSERT,
    connectionMode: partial.connectionMode || ConnectionMode.LONG,
    writeRate: partial.writeRate || 100,

    // 统计数据
    successCommitCount: partial.successCommitCount || 0,
    selectSuccessCount: partial.selectSuccessCount || 0,
    errorCount: partial.errorCount || 0,

    // 核心指标
    interruptionDurationMs: partial.interruptionDurationMs || 0,
    rtoMs: partial.rtoMs || 0,
    consistencyResult: partial.consistencyResult || ConsistencyResult.PENDING,

    // 运行时数据
    currentTps: partial.currentTps || 0,
    currentErrorRate: partial.currentErrorRate || 0
  }
}

/**
 * 创建运行时事件
 */
export function createRuntimeEvent(partial = {}) {
  return {
    eventId: partial.eventId || generateId(),
    timestamp: partial.timestamp || Date.now(),
    level: partial.level || EventLevel.INFO,
    type: partial.type || EventType.ANALYSIS_START,
    message: partial.message || '',
    details: partial.details || null
  }
}

/**
 * 创建趋势图数据点
 */
export function createTrendPoint(partial = {}) {
  return {
    timestamp: partial.timestamp || Date.now(),
    successTps: partial.successTps || 0,
    errorCount: partial.errorCount || 0,
    eventMarkers: partial.eventMarkers || []
  }
}

/**
 * 创建集群状态
 */
export function createClusterStatus(partial = {}) {
  return {
    vip: partial.vip || '',
    currentPrimary: partial.currentPrimary || '',
    currentSecondary: partial.currentSecondary || '',
    primaryStatus: partial.primaryStatus || NodeStatus.UNKNOWN,
    secondaryStatus: partial.secondaryStatus || NodeStatus.UNKNOWN,
    lastRoleSwitchTime: partial.lastRoleSwitchTime || null,
    nodes: partial.nodes || []
  }
}

/**
 * 创建节点信息
 */
export function createNodeInfo(partial = {}) {
  return {
    nodeId: partial.nodeId || generateId(),
    ip: partial.ip || '',
    port: partial.port || 3306,
    role: partial.role || NodeRole.UNKNOWN,
    status: partial.status || NodeStatus.UNKNOWN,
    lastCheckTime: partial.lastCheckTime || null
  }
}

/**
 * 创建摘要卡片数据
 */
export function createSummaryData(partial = {}) {
  return {
    businessInterruption: {
      value: partial.businessInterruption?.value || 0,
      unit: partial.businessInterruption?.unit || 'ms',
      label: 'Business Interruption'
    },
    rto: {
      value: partial.rto?.value || 0,
      unit: partial.rto?.unit || 'ms',
      label: 'RTO'
    },
    consistency: {
      result: partial.consistency?.result || ConsistencyResult.PENDING,
      label: 'Consistency'
    },
    commitStats: {
      successCount: partial.commitStats?.successCount || 0,
      errorCount: partial.commitStats?.errorCount || 0,
      label: 'Commit Stats'
    }
  }
}

/**
 * 格式化持续时间
 */
export function formatDuration(ms) {
  if (ms === null || ms === undefined) return '--'
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(2)}s`
  return `${(ms / 60000).toFixed(2)}m`
}

/**
 * 格式化时间戳
 */
export function formatTimestamp(timestamp) {
  if (!timestamp) return '--'
  const date = new Date(timestamp)
  return date.toLocaleTimeString('en-US', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

/**
 * 格式化完整日期时间
 */
export function formatDateTime(timestamp) {
  if (!timestamp) return '--'
  const date = new Date(timestamp)
  return date.toLocaleString('en-US', {
    hour12: false,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}
