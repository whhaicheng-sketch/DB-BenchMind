import { defineStore } from 'pinia'
import {
  CreateTemplate as CreateTemplateApi,
  DeleteTemplate as DeleteTemplateApi,
  DuplicateTemplate as DuplicateTemplateApi,
  ListTemplates as ListTemplatesApi,
  UpdateTemplate as UpdateTemplateApi
} from '../../wailsjs/go/bindings/TemplateBinding'
import { getCapabilityForTool, getDefaultToolForDbFamily, getToolsForDbFamily } from '../constants/templateCapabilities'
import {
  cloneTemplate,
  createDefaultTemplate,
  createTemplateId,
  normalizeTemplateRecord,
  DB_FAMILY_LABELS,
  createPhaseState,
  PHASE_KEYS,
  TEMPLATE_SCOPE_LABELS,
  TEMPLATE_STATUS_LABELS,
  TEMPLATE_TOOL_LABELS,
  WORKLOAD_LABELS
} from '../models/template'
import { templateMocks } from '../mock/templates'

const ENABLE_TEMPLATE_BACKEND = typeof window !== 'undefined' && !!window.go?.bindings?.TemplateBinding

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

function getValueByPath(obj, path) {
  return path.split('.').reduce((value, segment) => value?.[segment], obj)
}

function buildValidationErrorKey(fieldPath) {
  const map = {
    'toolConfig.sysbench.scriptType': 'sysbenchScriptType',
    'toolConfig.swingbench.benchmark': 'swingbenchBenchmark',
    'toolConfig.hammerdb.benchmark': 'hammerdbBenchmark',
    'toolConfig.hammerdb.warehouses': 'hammerdbWarehouses',
    'toolConfig.hammerdb.scaleFactor': 'hammerdbScaleFactor'
  }

  return map[fieldPath] || fieldPath.replace(/\./g, '_')
}

function canEditScope(scope) {
  return ['user', 'project', 'test'].includes(scope)
}

function canDeleteScope(scope) {
  return ['user', 'test'].includes(scope)
}

function applyPhaseRules(template, capability, changeLabels) {
  const nextPhases = createPhaseState(template.phases)

  PHASE_KEYS.forEach((phase) => {
    const allowed = capability.allowedPhases.includes(phase)
    const wasEnabled = !!nextPhases[phase].enabled
    nextPhases[phase].enabled = allowed ? nextPhases[phase].enabled : false
    nextPhases[phase].required = capability.requiredPhases.includes(phase)

    if (!allowed && wasEnabled) {
      changeLabels.push(`phase:${phase}`)
    }

    if (nextPhases[phase].required) {
      nextPhases[phase].enabled = true
    }
  })

  return nextPhases
}

function normalizeDraftForCapability(template) {
  const normalized = cloneTemplate(template)
  const changeLabels = []
  const allowedTools = getToolsForDbFamily(normalized.dbFamily)

  if (!allowedTools.includes(normalized.tool)) {
    normalized.tool = getDefaultToolForDbFamily(normalized.dbFamily, normalized.tool)
    changeLabels.push('Benchmark Tool')
  }

  const capability = getCapabilityForTool(normalized.tool)

  if (!capability.workloads.includes(normalized.workloadFamily)) {
    normalized.workloadFamily = capability.workloads[0]
    changeLabels.push('Workload Family')
  }

  if (!capability.concurrencyModes.includes(normalized.runtime.concurrency.mode)) {
    normalized.runtime.concurrency.mode = capability.concurrencyModes[0]
    changeLabels.push('Concurrency Mode')
  }

  normalized.compatibility.supportedDatabases = [normalized.dbFamily]
  normalized.database_types = [normalized.dbFamily]
  normalized.phases = applyPhaseRules(normalized, capability, changeLabels)

  const workloadDefaults = capability.workloadFieldMap?.[normalized.workloadFamily] || {}
  Object.entries(workloadDefaults).forEach(([key, value]) => {
    if (normalized.toolConfig[capability.toolConfigKey][key] !== value) {
      changeLabels.push(key)
      normalized.toolConfig[capability.toolConfigKey][key] = value
    }
  })

  if (normalized.tool === 'sysbench') {
    const nextDriver = normalized.dbFamily === 'postgresql' ? 'pgsql' : 'mysql'
    if (normalized.toolConfig.sysbench.dbDriver !== nextDriver) {
      changeLabels.push('dbDriver')
    }
    normalized.toolConfig.sysbench.dbDriver = nextDriver
  }

  if (normalized.tool === 'swingbench') {
    if (normalized.toolConfig.swingbench.userCount !== normalized.runtime.concurrency.value) {
      changeLabels.push('userCount')
    }
    normalized.toolConfig.swingbench.userCount = normalized.runtime.concurrency.value
  }

  if (normalized.tool === 'hammerdb') {
    if (normalized.toolConfig.hammerdb.benchmark !== normalized.workloadFamily) {
      changeLabels.push('benchmark')
    }
    normalized.toolConfig.hammerdb.benchmark = normalized.workloadFamily
    if (normalized.toolConfig.hammerdb.virtualUsers !== normalized.runtime.concurrency.value) {
      changeLabels.push('virtualUsers')
    }
    normalized.toolConfig.hammerdb.virtualUsers = normalized.runtime.concurrency.value

    if (normalized.workloadFamily === 'tproc-c') {
      const nextScaleFactor = Math.max(1, normalized.toolConfig.hammerdb.scaleFactor || 10)
      if (normalized.toolConfig.hammerdb.scaleFactor !== nextScaleFactor) {
        changeLabels.push('scaleFactor')
      }
      normalized.toolConfig.hammerdb.scaleFactor = nextScaleFactor
    }

    if (normalized.workloadFamily === 'tproc-h') {
      const nextWarehouses = Math.max(1, normalized.toolConfig.hammerdb.warehouses || 10)
      if (normalized.toolConfig.hammerdb.warehouses !== nextWarehouses) {
        changeLabels.push('warehouses')
      }
      normalized.toolConfig.hammerdb.warehouses = nextWarehouses
    }
  }

  return {
    normalized,
    changedFields: [...new Set(changeLabels)]
  }
}

function sortTemplates(templates) {
  return [...templates].sort((a, b) => a.name.localeCompare(b.name))
}

function normalizeTemplatesFromBackend(templates = []) {
  return sortTemplates(templates.map((template) => normalizeTemplateRecord(template)))
}

export const useTemplateStore = defineStore('template', {
  state: () => ({
    templates: [],
    selectedTemplateId: '',
    isEditorOpen: false,
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
      createTaskFromTemplate: null
    },
    isDirty: false,
    validationErrors: {},
    deleteCandidateId: ''
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
      return !!selected && canEditScope(selected.scope)
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
    },
    templatesByDatabase: (state) => (dbFamily) => {
      if (!dbFamily) return []
      return state.templates.filter((template) => template.dbFamily === dbFamily || template.database_types?.includes(dbFamily))
    },
    hasValidationErrors: (state) => Object.keys(state.validationErrors).length > 0,
    deleteCandidate(state) {
      return state.displayTemplates.find((template) => template.id === state.deleteCandidateId) || null
    }
  },

  actions: {
    async fetchTemplates() {
      this.loading = true
      this.error = null

      try {
        if (ENABLE_TEMPLATE_BACKEND) {
          const result = await ListTemplatesApi()
          if (result.error) {
            throw new Error(result.error)
          }
          this.templates = normalizeTemplatesFromBackend(result.templates || [])
        } else {
          this.templates = templateMocks.map((template) => normalizeTemplateRecord(cloneTemplate(template)))
        }
      } catch (err) {
        this.error = err.message || 'Failed to load templates'
        this.templates = templateMocks.map((template) => normalizeTemplateRecord(cloneTemplate(template)))
        this.showNotice(`Failed to load templates from backend. Falling back to mock data. ${this.error}`, 'warning')
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
      this.isEditorOpen = false
      this.editorState = 'view'
      this.isDirty = false
      this.editingTemplateDraft = null
      this.validationErrors = {}
    },

    openTemplate(id) {
      if (this.isDirty && this.selectedTemplateId && this.selectedTemplateId !== id) {
        this.showNotice('Unsaved changes were discarded when switching templates.', 'warning')
      }

      this.selectedTemplateId = id
      this.isEditorOpen = true
      this.editorState = 'view'
      this.isDirty = false
      this.editingTemplateDraft = null
      this.validationErrors = {}
    },

    clearSelection() {
      this.selectedTemplateId = ''
      this.isEditorOpen = false
      this.editingTemplateDraft = null
      this.editorState = 'view'
      this.isDirty = false
      this.validationErrors = {}
    },

    closeEditor() {
      if (this.isDirty && (this.editorState === 'editing' || this.editorState === 'creating')) {
        this.showNotice('Unsaved changes were discarded when closing the template editor.', 'warning')
      }

      this.isEditorOpen = false

      if (this.editorState === 'creating') {
        this.selectedTemplateId = ''
      }

      this.editingTemplateDraft = null
      this.editorState = 'view'
      this.isDirty = false
      this.validationErrors = {}
    },

    startEditing() {
      if (!this.selectedTemplate || !canEditScope(this.selectedTemplate.scope)) return
      this.editingTemplateDraft = cloneTemplate(this.selectedTemplate)
      this.isEditorOpen = true
      this.editorState = 'editing'
      this.isDirty = false
      this.validationErrors = {}
    },

    cancelEditing() {
      if (this.editorState === 'creating') {
        this.clearSelection()
      } else {
        this.editorState = 'view'
        this.editingTemplateDraft = null
        this.isDirty = false
        this.validationErrors = {}
      }
    },

    createTemplate() {
      const draft = createDefaultTemplate({
        id: createTemplateId(),
        name: `New Template ${this.templates.filter((template) => template.scope === 'user').length + 1}`
      })

      this.selectedTemplateId = draft.id
      this.isEditorOpen = true
      this.editingTemplateDraft = draft
      this.editorState = 'creating'
      this.isDirty = true
      this.validationErrors = {}
      this.showNotice('New template draft created. Configure it and save locally.', 'info')
    },

    markDirty() {
      if (this.editorState === 'editing' || this.editorState === 'creating') {
        this.isDirty = true
      }
    },

    applyNormalization(template, reason = 'updated') {
      const { normalized, changedFields } = normalizeDraftForCapability(template)

      if (changedFields.length > 0) {
        const preview = changedFields.slice(0, 3).join(', ')
        const suffix = changedFields.length > 3 ? ' and more' : ''
        this.showNotice(`Adjusted incompatible fields after ${reason}: ${preview}${suffix}.`, 'info')
      }

      return normalized
    },

    validateTemplate(template) {
      const targetTemplate = template || this.editingTemplateDraft || this.activeTemplate
      const errors = {}
      const capability = targetTemplate ? getCapabilityForTool(targetTemplate.tool) : null

      if (!targetTemplate) {
        this.validationErrors = { form: 'No template selected.' }
        return false
      }

      if (!targetTemplate.name?.trim()) {
        errors.name = 'Template name is required.'
      }

      if (!targetTemplate.tool) {
        errors.tool = 'Benchmark tool is required.'
      }

      if (!targetTemplate.dbFamily) {
        errors.dbFamily = 'Database type is required.'
      }

      if (!targetTemplate.workloadFamily) {
        errors.workloadFamily = 'Workload family is required.'
      }

      const allowedTools = getToolsForDbFamily(targetTemplate.dbFamily)
      if (targetTemplate.dbFamily && targetTemplate.tool && !allowedTools.includes(targetTemplate.tool)) {
        errors.tool = `${DB_FAMILY_LABELS[targetTemplate.dbFamily] || targetTemplate.dbFamily} templates can only use ${allowedTools.map((tool) => TEMPLATE_TOOL_LABELS[tool] || tool).join(', ')}.`
      }

      if (capability && !capability.workloads.includes(targetTemplate.workloadFamily)) {
        errors.workloadFamily = `Choose a workload family supported by ${TEMPLATE_TOOL_LABELS[targetTemplate.tool]}.`
      }

      if (!targetTemplate.runtime?.concurrency?.mode) {
        errors.concurrencyMode = 'Concurrency mode is required.'
      }

      if (!Number.isFinite(Number(targetTemplate.runtime?.concurrency?.value)) || Number(targetTemplate.runtime?.concurrency?.value) < 1) {
        errors.concurrencyValue = 'Concurrency must be at least 1.'
      }

      if (capability && targetTemplate.runtime?.concurrency?.mode && !capability.concurrencyModes.includes(targetTemplate.runtime.concurrency.mode)) {
        errors.concurrencyMode = `${TEMPLATE_TOOL_LABELS[targetTemplate.tool]} only supports ${capability.concurrencyModes.join(', ')} mode here.`
      }

      if (!Number.isFinite(Number(targetTemplate.runtime?.durationSeconds)) || Number(targetTemplate.runtime?.durationSeconds) < 1) {
        errors.durationSeconds = 'Duration must be at least 1 second.'
      }

      const enabledPhases = Object.entries(targetTemplate.phases || {})
        .filter(([, config]) => config.enabled)
        .map(([phase]) => phase)

      if (!enabledPhases.includes('run')) {
        errors.phaseRun = 'Run phase is mandatory for every template.'
      }

      const invalidPhases = enabledPhases.filter((phase) => capability && !capability.allowedPhases.includes(phase))
      if (invalidPhases.length > 0) {
        errors.phaseCombination = `${TEMPLATE_TOOL_LABELS[targetTemplate.tool]} does not use ${invalidPhases.join(', ')} in this workflow.`
      }

      if (targetTemplate.tool === 'sysbench' && targetTemplate.toolConfig.sysbench?.scriptType && !Object.values(capability.workloadFieldMap).some((entry) => entry.scriptType === targetTemplate.toolConfig.sysbench.scriptType)) {
        errors.sysbenchScriptType = 'Sysbench script is locked to the selected OLTP workload.'
      }

      if (targetTemplate.tool === 'swingbench' && targetTemplate.dbFamily !== 'oracle') {
        errors.tool = 'Selected database type does not support Swingbench.'
      }

      if (targetTemplate.tool === 'swingbench') {
        const benchmark = targetTemplate.toolConfig.swingbench?.benchmark
        const expectedBenchmark = capability.workloadFieldMap?.[targetTemplate.workloadFamily]?.benchmark
        if (!benchmark) {
          errors.swingbenchBenchmark = 'Swingbench benchmark is required.'
        } else if (expectedBenchmark && benchmark !== expectedBenchmark) {
          errors.swingbenchBenchmark = 'Swingbench benchmark follows the selected workload family.'
        }
      }

      if (targetTemplate.tool === 'hammerdb') {
        const benchmark = targetTemplate.toolConfig.hammerdb?.benchmark
        if (benchmark !== targetTemplate.workloadFamily) {
          errors.hammerdbBenchmark = 'HammerDB profile follows the selected workload family.'
        }

        if (targetTemplate.workloadFamily === 'tproc-c') {
          if (!Number.isFinite(Number(targetTemplate.toolConfig.hammerdb?.warehouses)) || Number(targetTemplate.toolConfig.hammerdb?.warehouses) < 1) {
            errors.hammerdbWarehouses = 'TPROC-C requires warehouses >= 1.'
          }
        }

        if (targetTemplate.workloadFamily === 'tproc-h') {
          if (!Number.isFinite(Number(targetTemplate.toolConfig.hammerdb?.scaleFactor)) || Number(targetTemplate.toolConfig.hammerdb?.scaleFactor) < 1) {
            errors.hammerdbScaleFactor = 'TPROC-H requires scale factor >= 1.'
          }
        }
      }

      const dynamicRequiredFields = [
        ...(capability?.requiredFields || []),
        ...((capability?.requiredFieldsByWorkload?.[targetTemplate.workloadFamily]) || [])
      ]

      dynamicRequiredFields.forEach((fieldPath) => {
        const value = getValueByPath(targetTemplate, fieldPath)
        const isEmpty = value === undefined || value === null || value === '' || (typeof value === 'number' && Number(value) < 1)
        if (isEmpty) {
          const key = buildValidationErrorKey(fieldPath)
          if (!errors[key]) {
            errors[key] = `${fieldPath.split('.').slice(-1)[0]} is required for the current tool/workload.`
          }
        }
      })

      this.validationErrors = errors

      if (Object.keys(errors).length > 0) {
        this.showNotice('Please fix validation errors before saving.', 'warning')
        return false
      }

      return true
    },

    updateDraftForTool(tool) {
      if (!this.editingTemplateDraft) return
      const nextTool = getDefaultToolForDbFamily(this.editingTemplateDraft.dbFamily, tool)
      this.editingTemplateDraft.tool = nextTool
      this.editingTemplateDraft = this.applyNormalization(this.editingTemplateDraft, 'changing tool')
      this.markDirty()
      this.validateTemplate(this.editingTemplateDraft)
    },

    updateDraftDbFamily(dbFamily) {
      if (!this.editingTemplateDraft) return
      this.editingTemplateDraft.dbFamily = dbFamily
      this.editingTemplateDraft = this.applyNormalization(this.editingTemplateDraft, 'changing database type')
      this.markDirty()
      this.validateTemplate(this.editingTemplateDraft)
    },

    updateDraftWorkload(workloadFamily) {
      if (!this.editingTemplateDraft) return
      this.editingTemplateDraft.workloadFamily = workloadFamily
      this.editingTemplateDraft = this.applyNormalization(this.editingTemplateDraft, 'changing workload family')
      this.markDirty()
      this.validateTemplate(this.editingTemplateDraft)
    },

    updateDraftConcurrencyMode(mode) {
      if (!this.editingTemplateDraft) return
      this.editingTemplateDraft.runtime.concurrency.mode = mode
      this.editingTemplateDraft = this.applyNormalization(this.editingTemplateDraft, 'changing concurrency mode')
      this.markDirty()
      this.validateTemplate(this.editingTemplateDraft)
    },

    updateDraftConcurrencyValue(value) {
      if (!this.editingTemplateDraft) return
      this.editingTemplateDraft.runtime.concurrency.value = value

      if (this.editingTemplateDraft.tool === 'swingbench') {
        this.editingTemplateDraft.toolConfig.swingbench.userCount = value
      }

      if (this.editingTemplateDraft.tool === 'hammerdb') {
        this.editingTemplateDraft.toolConfig.hammerdb.virtualUsers = value
      }

      this.markDirty()
      this.validateTemplate(this.editingTemplateDraft)
    },

    updateDraftPhase(phase, enabled) {
      if (!this.editingTemplateDraft) return
      this.editingTemplateDraft.phases[phase].enabled = enabled
      this.editingTemplateDraft = this.applyNormalization(this.editingTemplateDraft, 'changing phases')
      this.markDirty()
      this.validateTemplate(this.editingTemplateDraft)
    },

    async saveTemplate() {
      const draft = this.editingTemplateDraft
      if (!draft) return

      if (!this.validateTemplate(draft)) {
        return
      }

      draft.updatedAt = new Date().toISOString()
      draft.status = draft.status === 'deprecated' ? 'deprecated' : 'ready'

      try {
        let savedTemplate = cloneTemplate(draft)

        if (ENABLE_TEMPLATE_BACKEND) {
          const result = this.editorState === 'creating'
            ? await CreateTemplateApi(savedTemplate)
            : await UpdateTemplateApi(savedTemplate)

          if (result.error) {
            throw new Error(result.error)
          }

          savedTemplate = normalizeTemplateRecord(result.template || savedTemplate)
        } else if (this.editorState === 'creating') {
          savedTemplate.id = savedTemplate.id || createTemplateId()
        }

        if (this.editorState === 'creating') {
          this.templates.unshift(savedTemplate)
        } else {
          this.templates = this.templates.map((template) => (
            template.id === savedTemplate.id ? savedTemplate : template
          ))
        }

        this.templates = sortTemplates(this.templates)
        this.selectedTemplateId = savedTemplate.id
        this.editorState = 'view'
        this.editingTemplateDraft = null
        this.isDirty = false
        this.validationErrors = {}
        this.showNotice(ENABLE_TEMPLATE_BACKEND ? 'Template saved successfully.' : 'Template saved to local mock state.', 'success')
      } catch (err) {
        this.error = err.message || 'Failed to save template'
        this.showNotice(this.error, 'warning')
      }
    },

    async duplicateTemplate(id = this.selectedTemplateId) {
      const source = this.displayTemplates.find((template) => template.id === id)
      if (!source) return

      try {
        let copy

        if (ENABLE_TEMPLATE_BACKEND) {
          const result = await DuplicateTemplateApi(id)
          if (result.error) {
            throw new Error(result.error)
          }
          copy = normalizeTemplateRecord(result.template)
        } else {
          copy = cloneTemplate(source)
          copy.id = createTemplateId()
          copy.scope = 'user'
          copy.status = 'draft'
          copy.name = `${source.name} Copy`
          copy.version = '0.1.0'
          copy.createdAt = new Date().toISOString()
          copy.updatedAt = copy.createdAt
        }

        this.templates = sortTemplates([copy, ...this.templates.filter((template) => template.id !== copy.id)])
        this.selectedTemplateId = copy.id
        this.isEditorOpen = true
        this.editorState = 'editing'
        this.editingTemplateDraft = cloneTemplate(copy)
        this.isDirty = false
        this.validationErrors = {}
        this.showNotice(ENABLE_TEMPLATE_BACKEND ? 'Template duplicated successfully.' : 'Template duplicated as a user draft.', 'success')
      } catch (err) {
        this.error = err.message || 'Failed to duplicate template'
        this.showNotice(this.error, 'warning')
      }
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
      this.isEditorOpen = true
      this.editorState = 'creating'
      this.editingTemplateDraft = copy
      this.isDirty = true
      this.validationErrors = {}
      this.showNotice('Save As created a new user template draft.', 'info')
    },

    requestDeleteTemplate(id = this.selectedTemplateId) {
      const template = this.templates.find((item) => item.id === id)
      if (!template || !canDeleteScope(template.scope)) {
        this.showNotice('Only editable user or test templates can be deleted in this phase.', 'warning')
        return
      }

      this.deleteCandidateId = id
    },

    cancelDeleteTemplate() {
      this.deleteCandidateId = ''
    },

    async confirmDeleteTemplate() {
      const id = this.deleteCandidateId
      const template = this.templates.find((item) => item.id === id)
      if (!template || !canDeleteScope(template.scope)) {
        this.deleteCandidateId = ''
        this.showNotice('Only editable user or test templates can be deleted in this phase.', 'warning')
        return
      }

      try {
        if (ENABLE_TEMPLATE_BACKEND) {
          const result = await DeleteTemplateApi(id)
          if (result.error) {
            throw new Error(result.error)
          }
        }

        this.templates = this.templates.filter((item) => item.id !== id)

        if (this.selectedTemplateId === id) {
          this.clearSelection()
        }

        this.deleteCandidateId = ''
        this.showNotice(ENABLE_TEMPLATE_BACKEND ? 'Template deleted successfully.' : 'User template removed from local mock state.', 'success')
      } catch (err) {
        this.error = err.message || 'Failed to delete template'
        this.deleteCandidateId = ''
        this.showNotice(this.error, 'warning')
      }
    },

    createTaskFromTemplate() {
      const template = this.activeTemplate || this.selectedTemplate
      if (!template) {
        this.showNotice('Select a template before creating a task shell.', 'warning')
        return null
      }

      return {
        templateId: template.id,
        templateName: template.name,
        tool: template.tool,
        dbFamily: template.dbFamily,
        workloadFamily: template.workloadFamily,
        createdAt: new Date().toISOString(),
        source: 'templates'
      }
    },

    placeholderAction(action) {
      const messages = {
        createTask: 'Create Task from Template is reserved for Tasks & Monitor integration.',
        save: 'Save placeholder executed.',
        unsupportedEdit: 'This template is read-only. Use Save As to create an editable copy.',
        readonlySaveAs: 'This template remains read-only. Save As creates a user-editable copy.'
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
