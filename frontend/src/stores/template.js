/**
 * Template Pinia Store
 * Manages benchmark template state for DB-BenchMind Wails frontend.
 */
import { defineStore } from 'pinia'
import {
  ListTemplates,
  ListTemplatesByType,
  GetTemplate,
  GetTemplateParams,
  ValidateTemplateForDB
} from '../../wailsjs/go/bindings/TemplateBinding'

export const useTemplateStore = defineStore('template', {
  state: () => ({
    // Template list
    templates: [],
    // Currently selected template ID
    selectedTemplateId: null,
    // Selected template details
    selectedTemplate: null,
    // Template parameters
    templateParams: [],
    // Parameter values (user input)
    paramValues: {},
    // Loading state
    loading: false,
    // Error message
    error: null
  }),

  getters: {
    // Get tool label
    toolLabels: () => ({
      sysbench: 'Sysbench',
      hammerdb: 'HammerDB',
      swingbench: 'SwingBench'
    }),

    // Get selected template tool type
    selectedTool: (state) => {
      return state.selectedTemplate?.tool || null
    },

    // Check if template supports a database type
    supportsDatabase: (state) => (dbType) => {
      if (!state.selectedTemplate) return false
      return state.selectedTemplate.database_types?.includes(dbType) || false
    },

    // Group templates by tool
    templatesByTool: (state) => {
      const grouped = {}
      for (const tmpl of state.templates) {
        if (!grouped[tmpl.tool]) {
          grouped[tmpl.tool] = []
        }
        grouped[tmpl.tool].push(tmpl)
      }
      return grouped
    },

    // Get templates filtered by database type
    getTemplatesForDbType: (state) => (dbType) => {
      if (!dbType) return state.templates
      return state.templates.filter(t =>
        t.database_types?.includes(dbType)
      )
    }
  },

  actions: {
    /**
     * Fetch all templates from backend
     */
    async fetchTemplates() {
      this.loading = true
      this.error = null

      try {
        const result = await ListTemplates()
        if (result.error) {
          this.error = result.error
          console.error('Failed to fetch templates:', result.error)
        } else {
          this.templates = result.templates || []
        }
      } catch (err) {
        this.error = err.message || 'Failed to fetch templates'
        console.error('fetchTemplates error:', err)
      } finally {
        this.loading = false
      }
    },

    /**
     * Fetch templates filtered by database type
     */
    async fetchTemplatesByType(dbType) {
      this.loading = true
      this.error = null

      try {
        const result = await ListTemplatesByType(dbType)
        if (result.error) {
          this.error = result.error
          console.error('Failed to fetch templates by type:', result.error)
        } else {
          this.templates = result.templates || []
        }
      } catch (err) {
        this.error = err.message || 'Failed to fetch templates by type'
        console.error('fetchTemplatesByType error:', err)
      } finally {
        this.loading = false
      }
    },

    /**
     * Select a template by ID and load its details
     */
    async selectTemplate(id) {
      this.selectedTemplateId = id

      if (!id) {
        this.selectedTemplate = null
        this.templateParams = []
        this.paramValues = {}
        return
      }

      this.loading = true
      this.error = null

      try {
        // Get template details
        const tmpl = await GetTemplate(id)
        if (tmpl) {
          this.selectedTemplate = tmpl

          // Get template parameters
          const paramsResult = await GetTemplateParams(id)
          if (paramsResult.error) {
            console.error('Failed to get template params:', paramsResult.error)
          } else {
            this.templateParams = paramsResult.params || []

            // Initialize param values with defaults
            this.paramValues = {}
            for (const param of this.templateParams) {
              this.paramValues[param.name] = param.default
            }
          }
        } else {
          this.error = 'Template not found'
        }
      } catch (err) {
        this.error = err.message || 'Failed to load template'
        console.error('selectTemplate error:', err)
      } finally {
        this.loading = false
      }
    },

    /**
     * Validate template compatibility with database type
     */
    async validateForDatabase(templateId, dbType) {
      try {
        return await ValidateTemplateForDB(templateId, dbType)
      } catch (err) {
        console.error('validateForDatabase error:', err)
        return false
      }
    },

    /**
     * Update a parameter value
     */
    setParamValue(name, value) {
      this.paramValues[name] = value
    },

    /**
     * Reset parameter values to defaults
     */
    resetParamValues() {
      this.paramValues = {}
      for (const param of this.templateParams) {
        this.paramValues[param.name] = param.default
      }
    },

    /**
     * Clear template selection
     */
    clearSelection() {
      this.selectedTemplateId = null
      this.selectedTemplate = null
      this.templateParams = []
      this.paramValues = {}
    },

    /**
     * Clear error
     */
    clearError() {
      this.error = null
    }
  }
})
