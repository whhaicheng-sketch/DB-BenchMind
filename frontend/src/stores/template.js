import { defineStore } from 'pinia'
import { TEMPLATE_CAPABILITIES } from '../constants/templateCapabilities'
import {
  cloneTemplate,
  createDefaultTemplate,
  createTemplateId,
  DB_FAMILY_LABELS,
  TEMPLATE_SCOPE_LABELS,
  TEMPLATE_STATUS_LABELS,
  TEMPLATE_TOOL_LABELS,
  WORKLOAD_LABELS
} from '../models/template'
import { templateMocks } from '../mock/templates'

function filterTemplate(template, filters) {
  const search = filters.search.trim().toLowerCase()
  const inSearch = !search || [
    template.name,
    template.description,
    template.dbFamily,
    template.tool,
    template.workloadFamily,
    ...(template.tags || [])
  ].join(' ').toLowerCase().includes(search)

  const inDb = !filters.dbFamily || template.dbFamily === filters.dbFamily
  const inTool = !filters.tool || template.tool === filters.tool
  const inScope = !filters.scope || template.scope === filters.scope
  const inTag = !filters.tag || template.tags.includes(filters.tag)

  return inSearch && inDb && inTool && inScope && inTag
}

export const useTemplateStore = defineStore('template', {
  state: () => ({
    templates: [],
    selectedTemplateId: '',
    editingTemplateDraft: null,
    editorState: 'view',
    editorMode: 'standard',
    loading: false,
    error: null,
    notice: null,
    filters: {
      search: '',
      dbFamily: '',
      tool: '',
      scope: '',
      tag: ''
    },
    templateParams: [],
    paramValues: {},
    pendingApi: {
      loadTemplates: null,
      createTemplate: null,
      updateTemplate: null,
      deleteTemplate: null,
      duplicateTemplate: null,
      exportTemplate: null,
      importTemplate: null,
      createTaskFromTemplate: null
    },
    isDirty: false
  }),

  getters: {
    toolLabels: () => TEMPLATE_TOOL_LABELS,
    dbFamilyLabels: () => DB_FAMILY_LABELS,
    scopeLabels: () => TEMPLATE_SCOPE_LABELS,
    statusLabels: () => TEMPLATE_STATUS_LABELS,
    workloadLabels: () => WORKLOAD_LABELS,
    selectedTemplate: (state) => state.templates.find((template) => template.id === state.selectedTemplateId) || null,
    activeTemplate: (state) => {
      if (state.editorState === 'creating' || state.editorState === 'editing') {
        return state.editingTemplateDraft
      }
      return state.templates.find((template) => template.id === state.selectedTemplateId) || null
    },
    displayTemplates: (state) => {
      if (state.editorState === 'creating' && state.editingTemplateDraft) {
        return [state.editingTemplateDraft, ...state.templates]
      }
      return state.templates
    },
    filteredTemplates() {
      return this.displayTemplates.filter((template) => filterTemplate(template, this.filters))
    },
    hasActiveFilters: (state) => Object.values(state.filters).some(Boolean),
    hasSearchOnly: (state) => !!state.filters.search.trim() &&
      !state.filters.dbFamily &&
      !state.filters.tool &&
      !state.filters.scope &&
      !state.filters.tag,
    allTags() {
      return [...new Set(this.displayTemplates.flatMap((template) => template.tags || []))].sort()
    },
    canEditSelected(state) {
      const selected = state.templates.find((template) => template.id === state.selectedTemplateId)
      return !!selected && selected.scope === 'user'
    },
    supportsDatabase: (state) => (dbType) => {
      const template = state.editingTemplateDraft || state.templates.find((item) => item.id === state.selectedTemplateId)
      return !!template && template.dbFamily === dbType
    },
    templatesByTool(state) {
      return state.templates.reduce((grouped, template) => {
        if (!grouped[template.tool]) {
          grouped[template.tool] = []
        }
        grouped[template.tool].push(template)
        return grouped
      }, {})
    },
    selectedTool(state) {
      const selected = state.editingTemplateDraft || state.templates.find((template) => template.id === state.selectedTemplateId)
      return selected?.tool || null
    }
  },

  actions: {
    async fetchTemplates() {
      this.loading = true
      this.error = null

      try {
        this.templates = templateMocks.map((template) => cloneTemplate(template))
      } catch (err) {
        this.error = err.message || 'Failed to load templates'
      } finally {
        this.loading = false
      }
    },

    async initializeTemplates() {
      if (this.templates.length === 0) {
        await this.fetchTemplates()
      }
    },

    setFilter(key, value) {
      this.filters[key] = value
    },

    resetFilters() {
      this.filters = {
        search: '',
        dbFamily: '',
        tool: '',
        scope: '',
        tag: ''
      }
    },

    selectTemplate(id) {
      if (this.isDirty && this.selectedTemplateId && this.selectedTemplateId !== id) {
        this.showNotice('Unsaved changes were discarded when switching templates.', 'warning')
      }

      this.selectedTemplateId = id
      this.editorState = 'view'
      this.isDirty = false
      this.editingTemplateDraft = null
    },

    clearSelection() {
      this.selectedTemplateId = ''
      this.editingTemplateDraft = null
      this.editorState = 'view'
      this.isDirty = false
    },

    startEditing() {
      if (!this.selectedTemplate || this.selectedTemplate.scope !== 'user') return
      this.editingTemplateDraft = cloneTemplate(this.selectedTemplate)
      this.editorState = 'editing'
      this.isDirty = false
    },

    cancelEditing() {
      if (this.editorState === 'creating') {
        this.clearSelection()
      } else {
        this.editorState = 'view'
        this.editingTemplateDraft = null
        this.isDirty = false
      }
    },

    createTemplate() {
      const draft = createDefaultTemplate({
        id: createTemplateId(),
        name: `New Template ${this.templates.filter((template) => template.scope === 'user').length + 1}`
      })

      this.selectedTemplateId = draft.id
      this.editingTemplateDraft = draft
      this.editorState = 'creating'
      this.isDirty = true
      this.showNotice('New template draft created. Configure it and save locally.', 'info')
    },

    markDirty() {
      if (this.editorState === 'editing' || this.editorState === 'creating') {
        this.isDirty = true
      }
    },

    updateDraftForTool(tool) {
      if (!this.editingTemplateDraft) return
      const capability = TEMPLATE_CAPABILITIES[tool]
      if (!capability) return

      this.editingTemplateDraft.tool = tool
      this.editingTemplateDraft.dbFamily = capability.dbFamilies[0]
      this.editingTemplateDraft.workloadFamily = capability.workloads[0]
      this.editingTemplateDraft.runtime.concurrency.mode = capability.concurrencyModes[0]
      this.markDirty()
    },

    updateDraftDbFamily(dbFamily) {
      if (!this.editingTemplateDraft) return
      this.editingTemplateDraft.dbFamily = dbFamily
      this.editingTemplateDraft.compatibility.supportedDatabases = [dbFamily]
      this.markDirty()
    },

    updateDraftWorkload(workloadFamily) {
      if (!this.editingTemplateDraft) return
      this.editingTemplateDraft.workloadFamily = workloadFamily
      if (this.editingTemplateDraft.tool === 'hammerdb') {
        this.editingTemplateDraft.toolConfig.hammerdb.benchmark = workloadFamily
      }
      this.markDirty()
    },

    saveTemplate() {
      const draft = this.editingTemplateDraft
      if (!draft) return

      draft.updatedAt = new Date().toISOString()
      draft.status = draft.status === 'deprecated' ? 'deprecated' : 'ready'

      if (this.editorState === 'creating') {
        this.templates.unshift(cloneTemplate(draft))
      } else {
        this.templates = this.templates.map((template) => (
          template.id === draft.id ? cloneTemplate(draft) : template
        ))
      }

      this.selectedTemplateId = draft.id
      this.editorState = 'view'
      this.editingTemplateDraft = null
      this.isDirty = false
      this.showNotice('Template saved to local mock state.', 'success')
    },

    duplicateTemplate(id = this.selectedTemplateId) {
      const source = this.displayTemplates.find((template) => template.id === id)
      if (!source) return

      const copy = cloneTemplate(source)
      copy.id = createTemplateId()
      copy.scope = 'user'
      copy.status = 'draft'
      copy.name = `${source.name} Copy`
      copy.version = '0.1.0'
      copy.createdAt = new Date().toISOString()
      copy.updatedAt = copy.createdAt

      this.templates.unshift(copy)
      this.selectedTemplateId = copy.id
      this.editorState = 'editing'
      this.editingTemplateDraft = cloneTemplate(copy)
      this.isDirty = false
      this.showNotice('Template duplicated as a user draft.', 'success')
    },

    saveAsTemplate() {
      const source = this.activeTemplate
      if (!source) return

      const copy = cloneTemplate(source)
      copy.id = createTemplateId()
      copy.scope = 'user'
      copy.status = 'draft'
      copy.name = `${source.name} Save As`
      copy.version = '0.1.0'
      copy.createdAt = new Date().toISOString()
      copy.updatedAt = copy.createdAt

      this.selectedTemplateId = copy.id
      this.editorState = 'creating'
      this.editingTemplateDraft = copy
      this.isDirty = true
      this.showNotice('Save As created a new user template draft.', 'info')
    },

    deleteTemplate(id = this.selectedTemplateId) {
      const template = this.templates.find((item) => item.id === id)
      if (!template || template.scope !== 'user') {
        this.showNotice('Built-in templates cannot be deleted in this phase.', 'warning')
        return
      }

      this.templates = this.templates.filter((item) => item.id !== id)

      if (this.selectedTemplateId === id) {
        this.clearSelection()
      }

      this.showNotice('User template removed from local mock state.', 'success')
    },

    placeholderAction(action) {
      const messages = {
        import: 'Import is a placeholder in this phase. Keep the button for later backend/parser wiring.',
        export: 'Export is a placeholder in this phase. Data stays in local mock state.',
        createTask: 'Create Task from Template is reserved for Tasks & Monitor integration.',
        save: 'Save placeholder executed.',
        unsupportedEdit: 'Built-in templates are read-only. Use Save As to create a user copy.'
      }

      this.showNotice(messages[action] || 'This action is reserved for a later phase.', 'info')
    },

    clearNotice() {
      this.notice = null
    },

    showNotice(message, tone = 'info') {
      this.notice = { message, tone, timestamp: Date.now() }
    }
  }
})
