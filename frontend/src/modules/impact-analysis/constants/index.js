/**
 * Impact Analysis Constants
 * 常量定义 - 事件类型、状态、连接模式、Workload 类型等
 */

// 分析会话状态
export const AnalysisStatus = {
  IDLE: 'idle',           // 空闲/未开始
  RUNNING: 'running',     // 分析中
  COMPLETED: 'completed', // 已完成
  ERROR: 'error'          // 错误
}

// 页面状态
export const PageState = {
  NO_CONNECTION: 'no_connection',     // 无可用 MySQL cluster connection
  READY: 'ready',                     // 已配置未开始
  ANALYZING: 'analyzing',             // 分析中
  COMPLETED: 'completed'              // 已完成
}

// 连接模式
export const ConnectionMode = {
  LONG: 'long',       // 长连接
  SHORT: 'short'      // 短连接
}

// Workload 类型
export const WorkloadType = {
  INSERT: 'insert',   // 写入
  SELECT: 'select'    // 查询
}

// 事件级别
export const EventLevel = {
  INFO: 'info',
  WARN: 'warn',
  ERROR: 'error',
  SUCCESS: 'success'
}

// 事件类型
export const EventType = {
  // 开始/结束
  ANALYSIS_START: 'analysis_start',
  ANALYSIS_STOP: 'analysis_stop',
  ANALYSIS_COMPLETE: 'analysis_complete',

  // Workload 事件
  WORKLOAD_STARTED: 'workload_started',
  WORKLOAD_STOPPED: 'workload_stopped',

  // 错误/恢复事件
  ERROR_SPIKE: 'error_spike',
  SUCCESS_TPS_ZERO: 'success_tps_zero',
  CONNECTION_FAILURE: 'connection_failure',
  RECOVERY_DETECTED: 'recovery_detected',

  // 切换事件
  VIP_SWITCH_DETECTED: 'vip_switch_detected',
  PRIMARY_ROLE_CHANGED: 'primary_role_changed',
  FAILOVER_DETECTED: 'failover_detected',

  // 一致性校验
  CONSISTENCY_CHECK_START: 'consistency_check_start',
  CONSISTENCY_CHECK_PASSED: 'consistency_check_passed',
  CONSISTENCY_CHECK_FAILED: 'consistency_check_failed'
}

// 事件类型显示配置
export const EventTypeConfig = {
  [EventType.ANALYSIS_START]: { label: 'Analysis Started', icon: '▶️', level: EventLevel.INFO },
  [EventType.ANALYSIS_STOP]: { label: 'Analysis Stopped', icon: '⏹️', level: EventLevel.WARN },
  [EventType.ANALYSIS_COMPLETE]: { label: 'Analysis Complete', icon: '✅', level: EventLevel.SUCCESS },
  [EventType.WORKLOAD_STARTED]: { label: 'Workload Started', icon: '🚀', level: EventLevel.INFO },
  [EventType.WORKLOAD_STOPPED]: { label: 'Workload Stopped', icon: '🛑', level: EventLevel.WARN },
  [EventType.ERROR_SPIKE]: { label: 'Error Rate Spike Detected', icon: '⚠️', level: EventLevel.WARN },
  [EventType.SUCCESS_TPS_ZERO]: { label: 'Success TPS Dropped to Zero', icon: '📉', level: EventLevel.ERROR },
  [EventType.CONNECTION_FAILURE]: { label: 'Connection Failure Detected', icon: '🔌', level: EventLevel.ERROR },
  [EventType.RECOVERY_DETECTED]: { label: 'Recovery Detected', icon: '🔄', level: EventLevel.SUCCESS },
  [EventType.VIP_SWITCH_DETECTED]: { label: 'VIP Switch Detected', icon: '🔀', level: EventLevel.WARN },
  [EventType.PRIMARY_ROLE_CHANGED]: { label: 'Primary Role Changed', icon: '👑', level: EventLevel.WARN },
  [EventType.FAILOVER_DETECTED]: { label: 'Failover Detected', icon: '🔁', level: EventLevel.WARN },
  [EventType.CONSISTENCY_CHECK_START]: { label: 'Consistency Check Started', icon: '🔍', level: EventLevel.INFO },
  [EventType.CONSISTENCY_CHECK_PASSED]: { label: 'Consistency Check Passed', icon: '✅', level: EventLevel.SUCCESS },
  [EventType.CONSISTENCY_CHECK_FAILED]: { label: 'Consistency Check Failed', icon: '❌', level: EventLevel.ERROR }
}

// 一致性结果
export const ConsistencyResult = {
  PENDING: 'pending',
  PASSED: 'passed',
  FAILED: 'failed'
}

// 节点状态
export const NodeStatus = {
  ONLINE: 'online',
  OFFLINE: 'offline',
  UNKNOWN: 'unknown'
}

// 节点角色
export const NodeRole = {
  PRIMARY: 'primary',
  SECONDARY: 'secondary',
  UNKNOWN: 'unknown'
}

// 连接模式选项
export const ConnectionModeOptions = [
  { value: ConnectionMode.LONG, label: 'Long Connection' },
  { value: ConnectionMode.SHORT, label: 'Short Connection' }
]

// Workload 类型选项
export const WorkloadTypeOptions = [
  { value: WorkloadType.INSERT, label: 'Insert (Write)' },
  { value: WorkloadType.SELECT, label: 'Select (Read)' }
]

// 默认写入速率选项
export const DefaultWriteRates = [10, 50, 100, 200, 500, 1000]

// 图表配置
export const ChartConfig = {
  MAX_DATA_POINTS: 300,
  REFRESH_INTERVAL_MS: 1000,
  COLORS: {
    SUCCESS_TPS: '#48bb78',
    ERROR_COUNT: '#f56565',
    EVENT_MARKER: '#ecc94b'
  }
}
