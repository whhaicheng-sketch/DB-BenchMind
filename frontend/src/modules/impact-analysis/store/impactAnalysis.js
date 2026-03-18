/**
 * Impact Analysis Store
 * Pinia Store - 管理分析会话、事件流、趋势数据、集群状态
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  AnalysisStatus,
  PageState,
  ConnectionMode,
  WorkloadType,
  ConsistencyResult,
  ChartConfig
} from '../constants'
import {
  createAnalysisSession,
  createRuntimeEvent,
  createTrendPoint,
  createClusterStatus,
  createSummaryData
} from '../types'
import {
  getMockMySQLClusterConnections,
  getMockAnalysisSession,
  getMockEvents,
  getMockTrendData,
  getMockClusterStatus,
  getMockSummaryData,
  generateRealtimeUpdate
} from '../mock'

export const useImpactAnalysisStore = defineStore('impactAnalysis', () => {
  // ==================== State ====================

  // 可用的 MySQL Cluster Connections
  const mysqlConnections = ref([])

  // 当前选中的连接 ID
  const selectedConnectionId = ref(null)

  // 当前分析会话
  const session = ref(null)

  // 事件流
  const events = ref([])

  // 趋势数据
  const trendData = ref([])

  // 集群状态
  const clusterStatus = ref(null)

  // 摘要数据
  const summaryData = ref(null)

  // 运行配置
  const config = ref({
    connectionMode: ConnectionMode.LONG,
    workloadType: WorkloadType.INSERT,
    writeRate: 100,
    consistencyCheckEnabled: true
  })

  // 加载状态
  const loading = ref(false)
  const error = ref(null)

  // 实时更新定时器
  let updateTimer = null

  // ==================== Getters ====================

  // 页面状态
  const pageState = computed(() => {
    if (!mysqlConnections.value || mysqlConnections.value.length === 0) {
      return PageState.NO_CONNECTION
    }
    if (!session.value || session.value.status === AnalysisStatus.IDLE) {
      return PageState.READY
    }
    if (session.value.status === AnalysisStatus.RUNNING) {
      return PageState.ANALYZING
    }
    return PageState.COMPLETED
  })

  // 当前选中的连接
  const selectedConnection = computed(() => {
    if (!selectedConnectionId.value || !mysqlConnections.value) return null
    return mysqlConnections.value.find(c => c.id === selectedConnectionId.value)
  })

  // 是否正在分析
  const isAnalyzing = computed(() => {
    return session.value?.status === AnalysisStatus.RUNNING
  })

  // 是否有可用连接
  const hasConnections = computed(() => {
    return mysqlConnections.value && mysqlConnections.value.length > 0
  })

  // 最新的事件（取最近 50 条）
  const recentEvents = computed(() => {
    return events.value.slice(-50)
  })

  // 最新的趋势数据（取最近 MAX_DATA_POINTS 个点）
  const recentTrendData = computed(() => {
    return trendData.value.slice(-ChartConfig.MAX_DATA_POINTS)
  })

  // ==================== Actions ====================

  /**
   * 初始化 - 加载 MySQL Cluster Connections
   */
  async function initialize() {
    loading.value = true
    error.value = null

    try {
      // TODO: 从真实 API 加载
      // 目前使用 Mock 数据
      await new Promise(resolve => setTimeout(resolve, 300))

      mysqlConnections.value = getMockMySQLClusterConnections()

      // 默认选中第一个连接
      if (mysqlConnections.value.length > 0 && !selectedConnectionId.value) {
        selectedConnectionId.value = mysqlConnections.value[0].id
        const conn = mysqlConnections.value[0]

        // 应用连接的默认配置
        config.value = {
          connectionMode: conn.defaultConnectionMode || ConnectionMode.LONG,
          workloadType: conn.defaultWorkloadType || WorkloadType.INSERT,
          writeRate: conn.defaultWriteRate || 100,
          consistencyCheckEnabled: conn.consistencyCheckEnabled ?? true
        }
      }

      // 初始化集群状态
      clusterStatus.value = getMockClusterStatus()
    } catch (err) {
      error.value = err.message || 'Failed to load connections'
      console.error('Failed to initialize Impact Analysis:', err)
    } finally {
      loading.value = false
    }
  }

  /**
   * 选择连接
   */
  function selectConnection(connectionId) {
    selectedConnectionId.value = connectionId

    const conn = mysqlConnections.value.find(c => c.id === connectionId)
    if (conn) {
      config.value = {
        connectionMode: conn.defaultConnectionMode || ConnectionMode.LONG,
        workloadType: conn.defaultWorkloadType || WorkloadType.INSERT,
        writeRate: conn.defaultWriteRate || 100,
        consistencyCheckEnabled: conn.consistencyCheckEnabled ?? true
      }
    }

    // 重置会话
    resetSession()
  }

  /**
   * 更新配置
   */
  function updateConfig(newConfig) {
    config.value = { ...config.value, ...newConfig }
  }

  /**
   * 开始分析
   */
  async function startAnalysis() {
    if (!selectedConnectionId.value) {
      error.value = 'Please select a MySQL cluster connection first'
      return
    }

    loading.value = true
    error.value = null

    try {
      // 创建新会话
      session.value = createAnalysisSession({
        connectionId: selectedConnectionId.value,
        status: AnalysisStatus.RUNNING,
        startTime: Date.now(),
        ...config.value
      })

      // 清空事件和趋势数据
      events.value = []
      trendData.value = []

      // 添加开始事件
      addEvent(createRuntimeEvent({
        level: 'info',
        type: 'analysis_start',
        message: `Impact analysis started on ${selectedConnection.value?.name}`
      }))

      // 添加 workload 开始事件
      addEvent(createRuntimeEvent({
        level: 'info',
        type: 'workload_started',
        message: `${config.value.workloadType.toUpperCase()} workload started with rate ${config.value.writeRate} TPS`
      }))

      // TODO: 调用真实 API 启动分析
      // 目前使用 Mock 数据模拟
      await mockAnalysisProcess()

    } catch (err) {
      error.value = err.message || 'Failed to start analysis'
      console.error('Failed to start analysis:', err)
      session.value = {
        ...session.value,
        status: AnalysisStatus.ERROR
      }
    } finally {
      loading.value = false
    }
  }

  /**
   * 停止分析
   */
  async function stopAnalysis() {
    if (!session.value) return

    try {
      // 添加停止事件
      addEvent(createRuntimeEvent({
        level: 'warn',
        type: 'analysis_stop',
        message: 'Analysis stopped by user'
      }))

      session.value = {
        ...session.value,
        status: AnalysisStatus.COMPLETED,
        endTime: Date.now()
      }

      // 停止实时更新
      stopRealtimeUpdates()

      // TODO: 调用真实 API 停止分析
    } catch (err) {
      error.value = err.message || 'Failed to stop analysis'
      console.error('Failed to stop analysis:', err)
    }
  }

  /**
   * 重置会话
   */
  function resetSession() {
    stopRealtimeUpdates()
    session.value = null
    events.value = []
    trendData.value = []
    summaryData.value = null
    error.value = null
  }

  /**
   * 添加事件
   */
  function addEvent(event) {
    events.value.push(event)
  }

  /**
   * 添加趋势数据点
   */
  function addTrendPoint(point) {
    trendData.value.push(point)
    // 保持最大数据点限制
    if (trendData.value.length > ChartConfig.MAX_DATA_POINTS) {
      trendData.value = trendData.value.slice(-ChartConfig.MAX_DATA_POINTS)
    }
  }

  /**
   * 更新集群状态
   */
  function updateClusterStatus(status) {
    clusterStatus.value = { ...clusterStatus.value, ...status }
  }

  /**
   * 更新摘要数据
   */
  function updateSummaryData(data) {
    summaryData.value = { ...summaryData.value, ...data }
  }

  /**
   * 启动实时更新
   */
  function startRealtimeUpdates() {
    if (updateTimer) return

    updateTimer = setInterval(() => {
      if (session.value?.status !== AnalysisStatus.RUNNING) {
        stopRealtimeUpdates()
        return
      }

      // 模拟实时数据
      const update = generateRealtimeUpdate(config.value.writeRate, true)
      addTrendPoint(createTrendPoint(update))

      // 更新会话统计
      session.value = {
        ...session.value,
        currentTps: update.successTps,
        successCommitCount: session.value.successCommitCount + update.successTps,
        errorCount: session.value.errorCount + update.errorCount
      }
    }, ChartConfig.REFRESH_INTERVAL_MS)
  }

  /**
   * 停止实时更新
   */
  function stopRealtimeUpdates() {
    if (updateTimer) {
      clearInterval(updateTimer)
      updateTimer = null
    }
  }

  /**
   * Mock 分析过程（仅用于演示）
   */
  async function mockAnalysisProcess() {
    // 加载 Mock 事件
    const mockEvents = getMockEvents()
    const mockTrend = getMockTrendData()

    // 模拟事件流
    for (const evt of mockEvents) {
      await new Promise(resolve => setTimeout(resolve, 500))
      addEvent(evt)
    }

    // 加载趋势数据
    trendData.value = mockTrend

    // 加载摘要数据
    summaryData.value = getMockSummaryData()

    // 更新会话状态
    session.value = {
      ...session.value,
      ...getMockAnalysisSession(AnalysisStatus.COMPLETED),
      endTime: Date.now()
    }

    // 更新集群状态
    clusterStatus.value = getMockClusterStatus()
  }

  return {
    // State
    mysqlConnections,
    selectedConnectionId,
    session,
    events,
    trendData,
    clusterStatus,
    summaryData,
    config,
    loading,
    error,

    // Getters
    pageState,
    selectedConnection,
    isAnalyzing,
    hasConnections,
    recentEvents,
    recentTrendData,

    // Actions
    initialize,
    selectConnection,
    updateConfig,
    startAnalysis,
    stopAnalysis,
    resetSession,
    addEvent,
    addTrendPoint,
    updateClusterStatus,
    updateSummaryData,
    startRealtimeUpdates,
    stopRealtimeUpdates
  }
})
