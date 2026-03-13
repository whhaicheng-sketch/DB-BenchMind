function binding() {
  return window?.go?.bindings?.TaskBinding
}

export function validateDraft(payload) {
  return binding().ValidateDraft(payload)
}

export function createTask(payload) {
  return binding().CreateTask(payload)
}

export function listTasks() {
  return binding().ListTasks()
}

export function stopTask(taskId) {
  return binding().StopTask(taskId)
}

export function getTaskLogs(payload) {
  return binding().GetTaskLogs(payload)
}
