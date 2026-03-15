import { defineStore } from 'pinia'
import { createTask, getTaskLogs, listTasks, stopTask, validateDraft } from '../lib/taskBinding'
import { getActiveTask, getCurrentTask } from '../components/tabs/tasksMonitorTaskState.mjs'

export const useTaskStore = defineStore('task', {
  state: () => ({
    tasks: [],
    readiness: null,
    loading: false,
    error: null,
    pollingTimer: null,
    logLines: [],
    selectedLogTaskId: null
  }),

  getters: {
    currentTask: (state) => getCurrentTask(state.tasks),
    activeTask: (state) => getActiveTask(state.tasks),
    hasActiveTask: (state) => Boolean(getActiveTask(state.tasks))
  },

  actions: {
    async validateDraft(draft) {
      this.loading = true
      this.error = null
      try {
        const result = await validateDraft(draft)
        this.readiness = result.task?.readiness || null
        if (result.error) {
          this.error = result.error
        }
        return result
      } catch (error) {
        this.error = error.message || 'Failed to validate draft'
        return { error: this.error }
      } finally {
        this.loading = false
      }
    },

    async createTask(draft) {
      this.loading = true
      this.error = null
      try {
        const result = await createTask(draft)
        if (result.error) {
          this.error = result.error
          return result
        }
        await this.fetchTasks()
        this.startPolling()
        return result
      } catch (error) {
        this.error = error.message || 'Failed to create task'
        return { error: this.error }
      } finally {
        this.loading = false
      }
    },

    async fetchTasks() {
      try {
        const result = await listTasks()
        if (result.error) {
          this.error = result.error
          return []
        }
        this.tasks = result.tasks || []
        return this.tasks
      } catch (error) {
        this.error = error.message || 'Failed to fetch tasks'
        return []
      }
    },

    async stopTask(taskId) {
      const result = await stopTask(taskId)
      await this.fetchTasks()
      return result
    },

    async fetchLogs({ taskId, limit = 500, query = '', phase = '' }) {
      this.selectedLogTaskId = taskId
      try {
        const result = await getTaskLogs({
          task_id: taskId,
          limit,
          query,
          phase
        })
        if (result.error) {
          this.error = result.error
          return []
        }
        this.logLines = result.lines || []
        return this.logLines
      } catch (error) {
        this.error = error.message || 'Failed to fetch task logs'
        return []
      }
    },

    startPolling() {
      this.stopPolling()
      this.pollingTimer = setInterval(async () => {
        await this.fetchTasks()
        if (this.selectedLogTaskId) {
          await this.fetchLogs({ taskId: this.selectedLogTaskId })
        }
      }, 1000)
    },

    stopPolling() {
      if (this.pollingTimer) {
        clearInterval(this.pollingTimer)
        this.pollingTimer = null
      }
    }
  }
})
