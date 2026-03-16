const ACTIVE_STATUSES = new Set(['starting', 'preparing', 'running', 'cleaning', 'stopping'])

export function getActiveTask(tasks = []) {
  return normalizeTasks(tasks).find((task) => ACTIVE_STATUSES.has(task?.status)) || null
}

export function getCurrentTask(tasks = []) {
  const normalizedTasks = normalizeTasks(tasks)
  return getActiveTask(normalizedTasks) || normalizedTasks[0] || null
}

export function resolveTasksMonitorBinding({ tasks = [], draft = null, pendingStopTaskId = '' } = {}) {
  void draft

  const activeTask = getActiveTask(tasks)
  const currentTask = activeTask || getCurrentTask(tasks)
  const normalizedPendingStopTaskId = normalizePendingStopTaskId(activeTask, pendingStopTaskId)
  const stopInFlight = Boolean(activeTask?.id) && activeTask.id === normalizedPendingStopTaskId
  const startBlocked = Boolean(activeTask)

  return {
    activeTask,
    currentTask,
    stopEnabled: Boolean(activeTask) && !stopInFlight,
    stopInFlight,
    stopTaskId: activeTask?.id || '',
    pendingStopTaskId: normalizedPendingStopTaskId,
    startBlocked,
    logViewerTask: activeTask || currentTask || null
  }
}

function normalizeTasks(tasks) {
  return Array.isArray(tasks) ? tasks : []
}

function normalizePendingStopTaskId(activeTask, pendingStopTaskId) {
  if (!activeTask?.id) {
    return ''
  }
  return activeTask.id === pendingStopTaskId ? pendingStopTaskId : ''
}
