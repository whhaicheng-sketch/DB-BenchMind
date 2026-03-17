export function queueTemplateForTaskState(state, payload) {
  return {
    ...state,
    pendingTaskTemplate: payload,
    activeTab: 'tasks'
  }
}

export function clearPendingTaskTemplateState(state) {
  return {
    ...state,
    pendingTaskTemplate: null
  }
}
