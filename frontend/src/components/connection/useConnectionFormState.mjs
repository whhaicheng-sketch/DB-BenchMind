/**
 * useConnectionFormState.mjs
 * Composable that encapsulates all reactive state, computed properties, watchers,
 * validation logic, and action handlers for ConnectionForm.
 */
import { ref, computed, watch, nextTick } from 'vue'
import { useConnectionStore } from '../../stores/connection'
import { TestConnectionDirect, TestSSHConnection, TestWinRMConnection, TestAIConnection, QueryAIModels } from '../../../wailsjs/go/bindings/ConnectionBinding'
import {
  applyBlockingFieldValidation,
  createSaveValidationSnapshot,
  pruneInactiveAiErrors,
  DEFAULT_AI_TEMPERATURE,
  normalizeModelOptions,
  shouldShowApiKeyField
} from './connectionFormAiState.mjs'
import {
  applyRemoteTypeCompatibilityFields,
  buildRemoteBlockingErrorMessage,
  getDefaultWinRMPortByScheme,
  getRemoteType,
  isRemoteTypeNone,
  isRemoteTypeSSH,
  isRemoteTypeWinRM,
  shouldAutoUpdateWinRMPort,
  syncRemoteHostFromGeneral
} from './connectionFormRemoteState.mjs'

// ============================================================
// Database Schema Configuration - Data Driven Design
// ============================================================
const DB_SCHEMA = {
  mysql: {
    label: 'MySQL',
    icon: '🐬',
    defaultPort: 3306,
    defaultUsername: 'root',
    defaultDatabase: '',
    databaseLabel: 'Database',
    databaseRequired: false,
    databasePlaceholder: 'mydb (optional)',
    showDatabase: false
  },
  postgresql: {
    label: 'PostgreSQL',
    icon: '🐘',
    defaultPort: 5432,
    defaultUsername: 'postgres',
    defaultDatabase: 'postgres',
    databaseLabel: 'Database',
    databaseRequired: true,
    databasePlaceholder: 'postgres',
    showDatabase: true
  },
  oracle: {
    label: 'Oracle',
    icon: '🔴',
    defaultPort: 1521,
    defaultUsername: 'system',
    defaultDatabase: '',
    databaseLabel: 'Service Name / SID',
    databaseRequired: true,
    databasePlaceholder: 'ORCL',
    showDatabase: true
  },
  sqlserver: {
    label: 'SQL Server',
    icon: '🔷',
    defaultPort: 1433,
    defaultUsername: 'sa',
    defaultDatabase: '',
    databaseLabel: 'Database',
    databaseRequired: false,
    databasePlaceholder: 'mydb (optional)',
    showDatabase: true
  }
}

// ============================================================
// AI Provider Options
// ============================================================
const aiProviders = [
  { value: 'deepseek', label: 'DeepSeek', host: 'https://api.deepseek.com', endpoint: '/v1/chat/completions', model: '' },
  { value: 'qwen', label: '阿里云 通义千问', host: 'https://dashscope.aliyuncs.com/compatible-mode', endpoint: '/v1/chat/completions', model: '' },
  { value: 'doubao', label: '字节跳动 豆包', host: 'https://ark.cn-beijing.volces.com/api/v3', endpoint: '/chat/completions', model: '' },
  { value: 'glm', label: '智谱 GLM', host: 'https://open.bigmodel.cn/api/paas/v4', endpoint: '/chat/completions', model: '' },
  { value: 'minimax', label: 'MiniMax', host: 'https://api.minimaxi.com', endpoint: '/v1/chat/completions', model: 'MiniMax-M2.7' },
  { value: 'moonshot', label: 'Moonshot Kimi', host: 'https://api.moonshot.cn', endpoint: '/v1/chat/completions', model: '' },
  { value: 'openai', label: 'OpenAI ChatGPT', host: 'https://api.openai.com', endpoint: '/v1/chat/completions', model: '' },
  { value: 'gemini', label: 'Google Gemini', host: 'https://generativelanguage.googleapis.com', endpoint: '/v1beta/models/gemini-pro:generateContent', model: '' },
  { value: 'anthropic', label: 'Anthropic Claude', host: 'https://api.anthropic.com', endpoint: '/v1/messages', model: '' },
  { value: 'xai', label: 'xAI Grok', host: 'https://api.x.ai', endpoint: '/v1/chat/completions', model: '' },
  { value: 'ollama', label: 'Ollama (本地)', host: 'http://localhost:11434', endpoint: '/api/chat', model: '' }
]

// ============================================================
// Type Selection Options
// ============================================================
const typeOptions = [
  { value: 'mysql', label: 'MySQL', icon: '🐬' },
  { value: 'postgresql', label: 'PostgreSQL', icon: '🐘' },
  { value: 'oracle', label: 'Oracle', icon: '🔴' },
  { value: 'sqlserver', label: 'SQL Server', icon: '🔷' }
]

// ============================================================
// Tab Definitions
// ============================================================
const tabs = [
  { id: 'general', label: 'General' },
  { id: 'remote', label: 'Remote' },
  { id: 'ai', label: 'AI 助手' }
]

export function useConnectionFormState(props, emit) {
  // Store
  const connectionStore = useConnectionStore()

  // ============================================================
  // Tab State
  // ============================================================
  const activeTab = ref('general')

  // ============================================================
  // Type Selection State (for new connection)
  // ============================================================
  const showTypeSelection = ref(true)
  const selectedType = ref('')

  // ============================================================
  // Form State
  // ============================================================
  const formData = ref({
    name: '',
    type: 'mysql',
    host: '',
    port: 3306,
    database: '',
    username: 'root',
    password: '',
    // Oracle specific
    oracle_connect_mode: 'basic',
    oracle_basic_identifier_type: 'service_name',
    oracle_basic_value: '',
    oracle_tns_name: '',
    // SQL Server specific
    auth_type: 'sql', // sql or windows
    remote_type: 'none',
    remote_port_user_overridden: false,
    ssh_enabled: false,
    winrm_enabled: false,
    // SSH Configuration
    ssh_host: '',
    ssh_port: 22,
    ssh_username: '',
    ssh_password: '',
    // WinRM Configuration
    winrm_host: '',
    winrm_port: 5985,
    winrm_username: 'Administrator',
    winrm_password: '',
    winrm_scheme: 'http',
    winrm_auth_type: 'basic',
    // AI Assistant Configuration
    ai_assistants: [
      {
        id: 'default',
        name: 'DeepSeek',
        provider: 'deepseek',
        api_host: 'https://api.deepseek.com',
        api_endpoint: '/v1/chat/completions',
        api_key: '',
        model: '',
        temperature: DEFAULT_AI_TEMPERATURE,
        description: ''
      }
    ],
    selectedAssistantId: 'default'
  })

  // UI State
  const showPassword = ref(false)
  const showSshPassword = ref(false)
  const showWinrmPassword = ref(false)
  const showApiKey = ref({})
  const saving = ref(false)
  const formError = ref(null)
  const fieldErrors = ref({})
  const aiAdvisoryErrors = ref({})

  // Test States
  const dbTesting = ref(false)
  const dbTestResult = ref(null)
  const dbTestStatus = ref('idle')

  const sshTesting = ref(false)
  const sshTestResult = ref(null)
  const sshTestStatus = ref('idle')

  const winrmTesting = ref(false)
  const winrmTestResult = ref(null)
  const winrmTestStatus = ref('idle')

  const aiTesting = ref(false)
  const aiTestResult = ref(null)
  const aiTestStatus = ref('idle')
  const showAiTestDialog = ref(false)

  // Flag to track if SSH host was manually modified
  const sshHostManuallyModified = ref(false)
  const winrmHostManuallyModified = ref(false)

  // AI model query state
  const showProviderDropdown = ref(false)
  const modelQuerying = ref(false)
  const modelQueryError = ref('')
  const availableModels = ref([])
  const showModelSelector = ref(false)
  const pendingModelSelection = ref('')

  // Provider dropdown position for fixed positioning
  const dropdownPosition = ref({ top: 0, left: 0 })
  const addBtnRef = ref(null)

  // ============================================================
  // Computed Properties
  // ============================================================
  const isEditing = computed(() => props.mode === 'edit' && props.connectionId)
  const title = computed(() => isEditing.value ? '编辑连接' : '新建连接')
  const currentSchema = computed(() => DB_SCHEMA[formData.value.type])
  const selectedAssistant = computed(() => {
    return formData.value.ai_assistants.find(a => a.id === formData.value.selectedAssistantId) || formData.value.ai_assistants[0]
  })
  const isOracleBasicMode = computed(() => formData.value.type === 'oracle' && formData.value.oracle_connect_mode === 'basic')
  const isOracleTNSMode = computed(() => formData.value.type === 'oracle' && formData.value.oracle_connect_mode === 'tns')
  const shouldShowHostPort = computed(() => formData.value.type !== 'oracle' || isOracleBasicMode.value)
  const shouldShowDatabaseField = computed(() => {
    if (formData.value.type !== 'oracle') {
      return currentSchema.value.showDatabase
    }
    return false
  })
  const isRemoteTypeNoneSelected = computed(() => isRemoteTypeNone(formData.value))
  const isRemoteTypeSSHSelected = computed(() => isRemoteTypeSSH(formData.value))
  const isRemoteTypeWinRMSelected = computed(() => isRemoteTypeWinRM(formData.value))

  // ============================================================
  // isLoadingEditData
  // ============================================================
  const isLoadingEditData = ref(false)

  // ============================================================
  // Watchers
  // ============================================================

  // Handle type selection
  watch(selectedType, (newType) => {
    if (newType && showTypeSelection.value && !isEditing.value) {
      formData.value.type = newType
      const schema = DB_SCHEMA[newType]
      formData.value.port = schema.defaultPort
      formData.value.username = schema.defaultUsername
      formData.value.database = schema.defaultDatabase
      showTypeSelection.value = false
    }
  })

  // Close provider dropdown when switching tabs
  watch(activeTab, () => {
    showProviderDropdown.value = false
  })

  // Watch for host changes to sync SSH host (only if not manually modified)
  watch(() => formData.value.host, (newHost) => {
    if (!sshHostManuallyModified.value) {
      formData.value.ssh_host = newHost
    }
    if (!winrmHostManuallyModified.value) {
      formData.value.winrm_host = newHost
    }
  })

  watch(() => formData.value.remote_type, (remoteType) => {
    const normalized = applyRemoteTypeCompatibilityFields(formData.value)
    formData.value.ssh_enabled = normalized.ssh_enabled
    formData.value.winrm_enabled = normalized.winrm_enabled

    if (remoteType === 'ssh' && !formData.value.ssh_port) {
      formData.value.ssh_port = 22
    }

    if (remoteType === 'winrm') {
      if (!formData.value.winrm_scheme) {
        formData.value.winrm_scheme = 'http'
      }
      if (!formData.value.winrm_auth_type) {
        formData.value.winrm_auth_type = 'basic'
      }
      if (!formData.value.winrm_port) {
        formData.value.winrm_port = getDefaultWinRMPortByScheme(formData.value.winrm_scheme)
        formData.value.remote_port_user_overridden = false
      }
    }
  }, { immediate: true })

  watch(() => formData.value.winrm_scheme, (scheme) => {
    if (!isRemoteTypeWinRM(formData.value)) {
      return
    }

    if (shouldAutoUpdateWinRMPort(formData.value)) {
      formData.value.winrm_port = getDefaultWinRMPortByScheme(scheme)
    }
  })

  watch(() => formData.value.oracle_connect_mode, (mode) => {
    if (formData.value.type !== 'oracle') {
      return
    }
    clearOracleModeSpecificFields(mode)
  })

  // Watch for connectionId to load edit data
  watch(() => props.connectionId, async (newId) => {
    if (newId && props.mode === 'edit') {
      const conn = connectionStore.connections.find(c => c.id === newId)
      if (conn) {
        isLoadingEditData.value = true
        showTypeSelection.value = false

        formData.value = {
          name: conn.name,
          type: conn.type,
          host: conn.host,
          port: conn.port,
          database: conn.database || '',
          username: conn.username,
          password: conn.password || '',
          oracle_connect_mode: conn.connect_type || 'basic',
          oracle_basic_identifier_type: conn.identifier_type || (conn.sid ? 'sid' : 'service_name'),
          oracle_basic_value: conn.service_name || conn.sid || '',
          oracle_tns_name: conn.tns_name || '',
          auth_type: conn.auth_type || 'sql',
          remote_type: getRemoteType(conn),
          remote_port_user_overridden: false,
          ssh_enabled: !!conn.ssh_enabled,
          winrm_enabled: !!conn.winrm_enabled,
          ssh_host: conn.host, // Will be overwritten if SSH was configured
          ssh_port: conn.ssh_port || 22,
          ssh_username: conn.ssh_username || '',
          ssh_password: conn.ssh_password || '',
          winrm_host: conn.host,
          winrm_port: conn.winrm_port || getDefaultWinRMPortByScheme(conn.winrm_use_https ? 'https' : 'http'),
          winrm_username: conn.winrm_username || 'Administrator',
          winrm_password: conn.winrm_password || '',
          winrm_scheme: conn.winrm_use_https ? 'https' : 'http',
          winrm_auth_type: 'basic',
          ai_assistants: conn.ai_assistants || formData.value.ai_assistants,
          selectedAssistantId: conn.ai_assistants?.[0]?.id || 'default'
        }

        // If SSH was configured with different host, use it
        if (conn.ssh_enabled && conn.ssh_username) {
          formData.value.ssh_host = conn.host
          sshHostManuallyModified.value = false
        }
        if (conn.winrm_enabled && conn.winrm_username) {
          formData.value.winrm_host = conn.host
          winrmHostManuallyModified.value = false
        }

        nextTick(() => {
          isLoadingEditData.value = false
        })
      }
    }
  }, { immediate: true })

  // Handle defaultType prop for new connections
  watch(() => props.defaultType, (newType) => {
    if (newType && !isEditing.value) {
      selectedType.value = newType
    }
  }, { immediate: true })

  watch(
    () => formData.value.ai_assistants.map((assistant) => assistant.provider).join('|'),
    () => {
      syncAllAssistantProviderDefaults()
      refreshVisibleAiFieldErrors()
      resetAiInteractionState()
    },
    { immediate: true }
  )

  watch(
    () => [
      formData.value.selectedAssistantId,
      selectedAssistant.value?.api_key || '',
      selectedAssistant.value?.model || ''
    ],
    () => {
      refreshVisibleAiFieldErrors()
      resetAiInteractionState()
    }
  )

  // ============================================================
  // Validation
  // ============================================================
  const validateField = (field) => {
    const nextErrors = { ...fieldErrors.value }
    const errors = {}

    delete nextErrors[field]

    switch (field) {
      case 'name':
        if (!formData.value.name.trim()) {
          errors.name = '连接名称不能为空'
        }
        break
      case 'host':
        if (shouldShowHostPort.value && !formData.value.host.trim()) {
          errors.host = '主机地址不能为空'
        }
        break
      case 'port':
        if (shouldShowHostPort.value && (!formData.value.port || formData.value.port < 1 || formData.value.port > 65535)) {
          errors.port = '端口必须在 1-65535 之间'
        }
        break
      case 'username':
        if (formData.value.type !== 'sqlserver' || formData.value.auth_type === 'sql') {
          if (!formData.value.username.trim()) {
            errors.username = '用户名不能为空'
          }
        }
        break
      case 'database':
        if (shouldShowDatabaseField.value && currentSchema.value.databaseRequired && !formData.value.database.trim()) {
          errors.database = `${currentSchema.value.databaseLabel}不能为空`
        }
        break
      case 'oracle_basic_value':
        if (isOracleBasicMode.value && !formData.value.oracle_basic_value.trim()) {
          errors.oracle_basic_value = formData.value.oracle_basic_identifier_type === 'sid' ? 'SID不能为空' : 'Service Name不能为空'
        }
        break
      case 'oracle_tns_name':
        if (isOracleTNSMode.value && !formData.value.oracle_tns_name.trim()) {
          errors.oracle_tns_name = 'TNS不能为空'
        }
        break
    }

    fieldErrors.value = { ...nextErrors, ...errors }
    return Object.keys(errors).length === 0
  }

  const buildBlockingErrorMessage = (blockingErrors) => {
    const remoteMessage = buildRemoteBlockingErrorMessage(blockingErrors)
    if (remoteMessage) {
      return remoteMessage
    }

    const messages = Object.values(blockingErrors).filter(Boolean)
    return messages.join('；')
  }

  const clearOracleModeSpecificFields = (mode) => {
    if (mode === 'tns') {
      formData.value.host = ''
      formData.value.port = DB_SCHEMA.oracle.defaultPort
      formData.value.oracle_basic_value = ''
      delete fieldErrors.value.host
      delete fieldErrors.value.port
      delete fieldErrors.value.oracle_basic_value
    } else {
      formData.value.oracle_tns_name = ''
      delete fieldErrors.value.oracle_tns_name
    }
  }

  const getOracleModeFieldError = (fieldName) => {
    return fieldErrors.value[fieldName] || ''
  }

  const validateForm = () => {
    const snapshot = createSaveValidationSnapshot(
      {
        ...formData.value,
        ai_assistants: getVisibleAiAssistants()
      },
      currentSchema.value,
      isLocalProvider
    )

    fieldErrors.value = pruneInactiveAiErrors(
      applyBlockingFieldValidation(fieldErrors.value, snapshot.blockingErrors),
      [],
      isLocalProvider
    )
    aiAdvisoryErrors.value = pruneInactiveAiErrors(
      snapshot.advisoryErrors,
      getVisibleAiAssistants(),
      isLocalProvider
    )

    return snapshot.isValid
  }

  // Get AI field error
  const getAIFieldError = (assistantId, fieldName) => {
    return aiAdvisoryErrors.value[`ai_${assistantId}_${fieldName}`] || ''
  }

  // ============================================================
  // Actions
  // ============================================================
  const handleSave = async () => {
    formError.value = null
    if (!validateForm()) {
      const snapshot = createSaveValidationSnapshot(
        {
          ...formData.value,
          ai_assistants: getVisibleAiAssistants()
        },
        currentSchema.value,
        isLocalProvider
      )
      formError.value = buildBlockingErrorMessage(snapshot.blockingErrors)
      return
    }

    refreshVisibleAiFieldErrors()

    saving.value = true

    try {
      // Sync provider-linked values to assistants before saving
      const syncedAssistants = formData.value.ai_assistants.map(assistant => {
        const providerInfo = getProviderInfo(assistant.provider)
        return {
          ...assistant,
          api_host: providerInfo.host,
          api_endpoint: providerInfo.endpoint,
          model: assistant.model || providerInfo.model
        }
      })

      const payload = {
        ...applyRemoteTypeCompatibilityFields(formData.value),
        ai_assistants: syncedAssistants,
        // Map Oracle connect_type to appropriate field
        connect_type: formData.value.type === 'oracle' ? formData.value.oracle_connect_mode : '',
        identifier_type: formData.value.type === 'oracle' ? formData.value.oracle_basic_identifier_type : '',
        service_name: formData.value.type === 'oracle' && formData.value.oracle_connect_mode === 'basic' && formData.value.oracle_basic_identifier_type === 'service_name'
          ? formData.value.oracle_basic_value : '',
        sid: formData.value.type === 'oracle' && formData.value.oracle_connect_mode === 'basic' && formData.value.oracle_basic_identifier_type === 'sid'
          ? formData.value.oracle_basic_value : '',
        tns_name: formData.value.type === 'oracle' && formData.value.oracle_connect_mode === 'tns'
          ? formData.value.oracle_tns_name : '',
        // SSH configuration
        ssh_port: formData.value.ssh_port || 22,
        ssh_username: formData.value.ssh_username || '',
        ssh_password: formData.value.ssh_password || '',
        // WinRM configuration
        winrm_port: formData.value.winrm_port || getDefaultWinRMPortByScheme(formData.value.winrm_scheme),
        winrm_use_https: formData.value.winrm_scheme === 'https',
        winrm_username: formData.value.winrm_username || '',
        winrm_password: formData.value.winrm_password || ''
      }

      if (isEditing.value) {
        const updated = await connectionStore.updateConnection({
          id: props.connectionId,
          ...payload
        })
        if (updated) {
          emit('saved', updated)
        } else {
          formError.value = connectionStore.error || '更新连接失败'
        }
      } else {
        const created = await connectionStore.createConnection(payload)
        if (created) {
          emit('saved', created)
          resetForm()
        } else {
          formError.value = connectionStore.error || '创建连接失败'
        }
      }
    } finally {
      saving.value = false
    }
  }

  const handleCancel = () => {
    emit('cancelled')
    resetForm()
  }

  const resetForm = () => {
    formData.value = {
      name: '',
      type: 'mysql',
      host: '',
      port: 3306,
      database: '',
      username: 'root',
      password: '',
      oracle_connect_mode: 'basic',
      oracle_basic_identifier_type: 'service_name',
      oracle_basic_value: '',
      oracle_tns_name: '',
      auth_type: 'sql',
      remote_type: 'none',
      remote_port_user_overridden: false,
      ssh_enabled: false,
      winrm_enabled: false,
      ssh_host: '',
      ssh_port: 22,
      ssh_username: '',
      ssh_password: '',
      winrm_host: '',
      winrm_port: 5985,
      winrm_username: 'Administrator',
      winrm_password: '',
      winrm_scheme: 'http',
      winrm_auth_type: 'basic',
      ai_assistants: [
        {
          id: 'default',
          name: 'DeepSeek',
          provider: 'deepseek',
          api_host: 'https://api.deepseek.com',
          api_endpoint: '/v1/chat/completions',
          api_key: '',
          model: '',
          temperature: DEFAULT_AI_TEMPERATURE,
          description: ''
        }
      ],
      selectedAssistantId: 'default'
    }
    showTypeSelection.value = true
    selectedType.value = ''
    formError.value = null
    fieldErrors.value = {}
    aiAdvisoryErrors.value = {}
    dbTestResult.value = null
    dbTestStatus.value = 'idle'
    sshTestResult.value = null
    sshTestStatus.value = 'idle'
    winrmTestResult.value = null
    winrmTestStatus.value = 'idle'
    aiTestResult.value = null
    aiTestStatus.value = 'idle'
    sshHostManuallyModified.value = false
    winrmHostManuallyModified.value = false
    showSshPassword.value = false
    showWinrmPassword.value = false
    showProviderDropdown.value = false
  }

  // ============================================================
  // DB Test - Direct connection test ONLY
  // ============================================================
  const handleTestDB = async () => {
    const requiresHost = !(formData.value.type === 'oracle' && formData.value.oracle_connect_mode === 'tns')
    if ((!formData.value.username) || (requiresHost && !formData.value.host)) {
      dbTestStatus.value = 'error'
      dbTestResult.value = {
        success: false,
        error: requiresHost ? '请填写主机和用户名' : '请填写用户名'
      }
      return
    }

    dbTesting.value = true
    dbTestStatus.value = 'testing'
    dbTestResult.value = null

    try {
      const testRequest = {
        name: formData.value.name || 'Test Connection',
        type: formData.value.type,
        host: formData.value.host,
        port: formData.value.port,
        database: formData.value.database || '',
        username: formData.value.username,
        password: formData.value.password || '',
        ssl_mode: ''
      }

      // Add Oracle-specific fields
      if (formData.value.type === 'oracle') {
        testRequest.connect_type = formData.value.oracle_connect_mode || 'basic'
        testRequest.identifier_type = formData.value.oracle_basic_identifier_type || 'service_name'
        testRequest.sid = formData.value.oracle_connect_mode === 'basic' && formData.value.oracle_basic_identifier_type === 'sid' ? formData.value.oracle_basic_value : ''
        testRequest.service_name = formData.value.oracle_connect_mode === 'basic' && formData.value.oracle_basic_identifier_type === 'service_name' ? formData.value.oracle_basic_value : ''
        testRequest.tns_name = formData.value.oracle_connect_mode === 'tns' ? formData.value.oracle_tns_name : ''
      }

      // IMPORTANT: Use TestConnectionDirect - direct DB test without SSH tunnel
      const result = await TestConnectionDirect(testRequest)

      dbTestResult.value = result
      dbTestStatus.value = result.success ? 'success' : 'error'
    } catch (err) {
      dbTestStatus.value = 'error'
      dbTestResult.value = {
        success: false,
        error: err.message || '测试失败'
      }
    } finally {
      dbTesting.value = false
    }
  }

  // ============================================================
  // SSH Test - SSH connection test ONLY
  // ============================================================
  const handleTestSSH = async () => {
    if (!formData.value.ssh_host || !formData.value.ssh_port || !formData.value.ssh_username) {
      sshTestStatus.value = 'error'
      sshTestResult.value = {
        success: false,
        error: '请填写 SSH 主机、SSH 端口和 SSH 用户名'
      }
      return
    }

    sshTesting.value = true
    sshTestStatus.value = 'testing'
    sshTestResult.value = null

    try {
      const sshConfig = {
        host: formData.value.ssh_host,
        port: formData.value.ssh_port || 22,
        username: formData.value.ssh_username,
        password: formData.value.ssh_password || ''
      }

      // IMPORTANT: Use TestSSHConnection - standalone SSH test, NOT DB connection
      const result = await TestSSHConnection(sshConfig)

      sshTestResult.value = result
      sshTestStatus.value = result.success ? 'success' : 'error'
    } catch (err) {
      sshTestStatus.value = 'error'
      sshTestResult.value = {
        success: false,
        error: err.message || 'SSH 测试失败'
      }
    } finally {
      sshTesting.value = false
    }
  }

  const handleTestWinRM = async () => {
    if (!formData.value.winrm_host || !formData.value.winrm_port || !formData.value.winrm_username) {
      winrmTestStatus.value = 'error'
      winrmTestResult.value = {
        success: false,
        error: '请填写 WinRM 主机、WinRM 端口和 WinRM 用户名'
      }
      return
    }

    winrmTesting.value = true
    winrmTestStatus.value = 'testing'
    winrmTestResult.value = null

    try {
      const result = await TestWinRMConnection({
        host: formData.value.winrm_host,
        port: formData.value.winrm_port || getDefaultWinRMPortByScheme(formData.value.winrm_scheme),
        username: formData.value.winrm_username,
        password: formData.value.winrm_password || '',
        use_https: formData.value.winrm_scheme === 'https'
      })

      winrmTestResult.value = result
      winrmTestStatus.value = result.success ? 'success' : 'error'
    } catch (err) {
      winrmTestStatus.value = 'error'
      winrmTestResult.value = {
        success: false,
        error: err.message || 'WinRM 测试失败'
      }
    } finally {
      winrmTesting.value = false
    }
  }

  // ============================================================
  // AI Test - AI API connection test ONLY
  // ============================================================
  const handleTestAI = async () => {
    const assistant = selectedAssistant.value
    const providerInfo = getProviderInfo(assistant.provider)

    if (!providerInfo.host || (shouldShowAssistantApiKeyField(assistant) && !assistant.api_key)) {
      aiTestStatus.value = 'error'
      aiTestResult.value = {
        success: false,
        error: '请填写 API 密钥'
      }
      return
    }

    aiTesting.value = true
    aiTestStatus.value = 'testing'
    aiTestResult.value = null

    try {
      // Build AI test request using provider-linked values
      const testRequest = {
        provider: assistant.provider,
        api_host: assistant.api_host || providerInfo.host,
        api_endpoint: assistant.api_endpoint || providerInfo.endpoint,
        api_key: assistant.api_key,
        model: assistant.model || providerInfo.model
      }

      // Call real backend AI test - completely independent from DB/SSH tests
      const result = await TestAIConnection(testRequest)

      aiTestResult.value = result
      aiTestStatus.value = result.success ? 'success' : 'error'
    } catch (err) {
      aiTestStatus.value = 'error'
      aiTestResult.value = {
        success: false,
        error: err.message || 'AI API 测试失败'
      }
    } finally {
      aiTesting.value = false
    }
  }

  // ============================================================
  // AI Assistant Management
  // ============================================================
  const addAssistant = (providerValue) => {
    const provider = aiProviders.find(p => p.value === providerValue)
    if (!provider) return

    const newId = `assistant_${Date.now()}`
    formData.value.ai_assistants.push({
      id: newId,
      name: provider.label,
      provider: providerValue,
      api_host: provider.host,
      api_endpoint: provider.endpoint,
      api_key: '',
      model: provider.model,
      temperature: DEFAULT_AI_TEMPERATURE,
      description: ''
    })
    formData.value.selectedAssistantId = newId
    showProviderDropdown.value = false
  }

  const toggleProviderDropdown = (event) => {
    showProviderDropdown.value = !showProviderDropdown.value
    if (showProviderDropdown.value) {
      // Calculate position relative to viewport for fixed positioning
      const rect = event.target.getBoundingClientRect()
      dropdownPosition.value = {
        top: rect.top - 4, // 4px margin from button top
        left: rect.left + rect.width + 4 // 4px to the right of the button
      }
    }
  }

  // Get provider display info
  const getProviderInfo = (providerValue) => {
    return aiProviders.find(p => p.value === providerValue) || aiProviders[0]
  }

  // Check if provider is local (Ollama) - no API key required
  const isLocalProvider = (providerValue) => {
    return providerValue === 'ollama'
  }

  const shouldShowAssistantApiKeyField = (assistant) => {
    if (!assistant) {
      return false
    }
    return shouldShowApiKeyField(assistant.provider, isLocalProvider)
  }

  // Cloud providers require API key to query models
  // Local providers (Ollama) can query models without API key
  const requiresApiKeyForModelQuery = (providerValue) => {
    return !isLocalProvider(providerValue)
  }

  const removeAssistant = (id) => {
    const index = formData.value.ai_assistants.findIndex(a => a.id === id)
    if (index > -1) {
      formData.value.ai_assistants.splice(index, 1)
      // If deleted the selected one
      if (formData.value.selectedAssistantId === id) {
        // If still have assistants, select the first one
        if (formData.value.ai_assistants.length > 0) {
          formData.value.selectedAssistantId = formData.value.ai_assistants[0].id
        } else {
          // No assistants left, clear selection
          formData.value.selectedAssistantId = null
        }
      }
    }
  }

  const selectAssistant = (id) => {
    formData.value.selectedAssistantId = id
  }

  const toggleApiKeyVisibility = (id) => {
    showApiKey.value[id] = !showApiKey.value[id]
  }

  const getVisibleAiAssistants = () => {
    return selectedAssistant.value ? [selectedAssistant.value] : []
  }

  const syncAssistantProviderDefaults = (assistant) => {
    if (!assistant) {
      return
    }

    const providerInfo = getProviderInfo(assistant.provider)
    assistant.name = providerInfo.label
    assistant.api_host = providerInfo.host
    assistant.api_endpoint = providerInfo.endpoint
    if (!assistant.model) {
      assistant.model = providerInfo.model
    }
  }

  const syncAllAssistantProviderDefaults = () => {
    for (const assistant of formData.value.ai_assistants) {
      syncAssistantProviderDefaults(assistant)
    }
  }

  const refreshVisibleAiFieldErrors = () => {
    aiAdvisoryErrors.value = pruneInactiveAiErrors(aiAdvisoryErrors.value, getVisibleAiAssistants(), isLocalProvider)
  }

  const resetAiInteractionState = () => {
    modelQueryError.value = ''
    availableModels.value = []
    showModelSelector.value = false
    pendingModelSelection.value = selectedAssistant.value?.model || ''
    aiTestResult.value = null
    aiTestStatus.value = 'idle'
    showAiTestDialog.value = false
  }

  // Model query handler - calls real backend
  const handleQueryModels = async () => {
    if (!selectedAssistant.value) return

    syncAssistantProviderDefaults(selectedAssistant.value)

    const provider = selectedAssistant.value.provider
    const apiHost = selectedAssistant.value.api_host
    const apiKey = selectedAssistant.value.api_key

    // Cloud providers require API key for model query
    if (requiresApiKeyForModelQuery(provider) && !apiKey) {
      modelQueryError.value = '云端模型需要先填写 API 密钥'
      return
    }

    modelQuerying.value = true
    modelQueryError.value = ''
    availableModels.value = []

    try {
      const result = await QueryAIModels({
        provider: provider,
        api_host: apiHost,
        api_key: apiKey
      })

      const normalizedModels = normalizeModelOptions(result.models)

      if (result.success && normalizedModels.length > 0) {
        availableModels.value = normalizedModels
        pendingModelSelection.value = selectedAssistant.value.model || normalizedModels[0].id
        showModelSelector.value = true
      } else if (result.error) {
        modelQueryError.value = result.error
      } else {
        modelQueryError.value = '未找到可用模型'
      }
    } catch (err) {
      modelQueryError.value = `查询失败: ${err.message || err}`
    } finally {
      modelQuerying.value = false
    }
  }

  // Select model from list
  const selectModel = (modelId) => {
    pendingModelSelection.value = modelId
  }

  const confirmModelSelection = () => {
    if (selectedAssistant.value && pendingModelSelection.value) {
      selectedAssistant.value.model = pendingModelSelection.value
    }
    showModelSelector.value = false
  }

  // Close model selector
  const closeModelSelector = () => {
    showModelSelector.value = false
    pendingModelSelection.value = selectedAssistant.value?.model || ''
  }

  const openAiTestDialog = () => {
    if (!selectedAssistant.value) {
      return
    }
    showAiTestDialog.value = true
  }

  const closeAiTestDialog = () => {
    showAiTestDialog.value = false
  }

  // ============================================================
  // SSH Host sync control
  // ============================================================
  const onSshHostChange = () => {
    sshHostManuallyModified.value = true
  }

  const onWinRMHostChange = () => {
    winrmHostManuallyModified.value = true
  }

  const onWinRMPortChange = () => {
    formData.value.remote_port_user_overridden = true
  }

  const setRemoteType = (remoteType) => {
    formData.value.remote_type = remoteType
  }

  const syncCurrentRemoteHostFromGeneral = () => {
    formData.value = syncRemoteHostFromGeneral(formData.value)
    if (isRemoteTypeSSH(formData.value)) {
      sshHostManuallyModified.value = false
    }
    if (isRemoteTypeWinRM(formData.value)) {
      winrmHostManuallyModified.value = false
    }
  }

  return {
    // Constants
    DB_SCHEMA,
    tabs,
    typeOptions,
    aiProviders,

    // Tab State
    activeTab,

    // Type Selection
    showTypeSelection,
    selectedType,

    // Form State
    formData,

    // UI State
    showPassword,
    showSshPassword,
    showWinrmPassword,
    showApiKey,
    saving,
    formError,
    fieldErrors,
    aiAdvisoryErrors,
    isLoadingEditData,

    // Test States
    dbTesting,
    dbTestResult,
    dbTestStatus,
    sshTesting,
    sshTestResult,
    sshTestStatus,
    winrmTesting,
    winrmTestResult,
    winrmTestStatus,
    aiTesting,
    aiTestResult,
    aiTestStatus,
    showAiTestDialog,

    // AI model query state
    showProviderDropdown,
    modelQuerying,
    modelQueryError,
    availableModels,
    showModelSelector,
    pendingModelSelection,
    dropdownPosition,
    addBtnRef,

    // Host sync flags
    sshHostManuallyModified,
    winrmHostManuallyModified,

    // Computed
    isEditing,
    title,
    currentSchema,
    selectedAssistant,
    isOracleBasicMode,
    isOracleTNSMode,
    shouldShowHostPort,
    shouldShowDatabaseField,
    isRemoteTypeNoneSelected,
    isRemoteTypeSSHSelected,
    isRemoteTypeWinRMSelected,

    // Validation
    validateField,
    getOracleModeFieldError,
    getAIFieldError,

    // Actions
    handleSave,
    handleCancel,
    resetForm,
    handleTestDB,
    handleTestSSH,
    handleTestWinRM,
    handleTestAI,

    // AI Management
    addAssistant,
    toggleProviderDropdown,
    getProviderInfo,
    isLocalProvider,
    shouldShowAssistantApiKeyField,
    removeAssistant,
    selectAssistant,
    toggleApiKeyVisibility,
    handleQueryModels,
    selectModel,
    confirmModelSelection,
    closeModelSelector,
    openAiTestDialog,
    closeAiTestDialog,

    // Remote Host sync
    onSshHostChange,
    onWinRMHostChange,
    onWinRMPortChange,
    setRemoteType,
    syncCurrentRemoteHostFromGeneral,

    // Provider dropdown close helper
    closeProviderDropdown: () => { showProviderDropdown.value = false }
  }
}
