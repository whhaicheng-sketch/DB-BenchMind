<script setup>
/**
 * ConnectionForm.vue
 * Navicat-style database connection editor with tabs: General / Remote / AI Assistant
 * Supports MySQL, PostgreSQL, Oracle, and SQL Server with database-specific fields.
 */
import { ref, computed, watch, nextTick } from 'vue'
import { useConnectionStore } from '../../stores/connection'
import { TestConnectionDirect, TestSSHConnection, TestWinRMConnection, TestAIConnection, QueryAIModels } from '../../../wailsjs/go/bindings/ConnectionBinding'
import AiTestDialog from './AiTestDialog.vue'
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

// Props
const props = defineProps({
  connectionId: {
    type: String,
    default: null
  },
  mode: {
    type: String,
    default: 'create',
    validator: (value) => ['create', 'edit'].includes(value)
  },
  defaultType: {
    type: String,
    default: ''
  }
})

// Emits
const emit = defineEmits(['saved', 'cancelled', 'tested'])

// Store
const connectionStore = useConnectionStore()

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
// Tab State
// ============================================================
const activeTab = ref('general')
const tabs = [
  { id: 'general', label: 'General' },
  { id: 'remote', label: 'Remote' },
  { id: 'ai', label: 'AI 助手' }
]

// ============================================================
// Type Selection State (for new connection)
// ============================================================
const showTypeSelection = ref(true)
const selectedType = ref('')

const typeOptions = [
  { value: 'mysql', label: 'MySQL', icon: '🐬' },
  { value: 'postgresql', label: 'PostgreSQL', icon: '🐘' },
  { value: 'oracle', label: 'Oracle', icon: '🔴' },
  { value: 'sqlserver', label: 'SQL Server', icon: '🔷' }
]

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
// Watchers
// ============================================================
const isLoadingEditData = ref(false)

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
const showProviderDropdown = ref(false)

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

// Provider dropdown position for fixed positioning
const dropdownPosition = ref({ top: 0, left: 0 })
const addBtnRef = ref(null)

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

// Model query state
const modelQuerying = ref(false)
const modelQueryError = ref('')
const availableModels = ref([])
const showModelSelector = ref(false)
const pendingModelSelection = ref('')

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
</script>

<template>
  <div class="conn-editor">
    <!-- Header -->
    <div class="conn-editor__header">
      <h2 class="conn-editor__title">{{ title }}</h2>
    </div>

    <!-- Type Selection (only for new connection) -->
    <div v-if="showTypeSelection && !isEditing" class="conn-editor__type-select">
      <div class="conn-editor__type-label">选择数据库类型</div>
      <div class="conn-editor__type-grid">
        <button
          v-for="opt in typeOptions"
          :key="opt.value"
          class="conn-editor__type-btn"
          :class="{ 'conn-editor__type-btn--active': selectedType === opt.value }"
          @click="selectedType = opt.value"
        >
          <span class="conn-editor__type-icon">{{ opt.icon }}</span>
          <span class="conn-editor__type-name">{{ opt.label }}</span>
        </button>
      </div>
    </div>

    <!-- Tabbed Editor (shown after type selection or in edit mode) -->
    <div v-else class="conn-editor__content">
      <!-- Error Banner -->
      <div v-if="formError" class="conn-editor__error">
        {{ formError }}
      </div>

      <!-- Tabs -->
      <div class="conn-editor__tabs">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          class="conn-editor__tab"
          :class="{ 'conn-editor__tab--active': activeTab === tab.id }"
          @click="activeTab = tab.id"
        >
          {{ tab.label }}
        </button>
      </div>

      <!-- Tab Content -->
      <div class="conn-editor__tab-content">
        <!-- ==================== General Tab ==================== -->
        <div v-show="activeTab === 'general'" class="conn-editor__panel">
          <div class="conn-form">
            <!-- Connection Name -->
            <div class="conn-form__row">
              <label class="conn-form__label">连接名称 <span class="required">*</span></label>
              <input
                v-model="formData.name"
                type="text"
                class="conn-form__input"
                :class="{ 'conn-form__input--error': fieldErrors.name }"
                placeholder="输入连接名称"
                @blur="validateField('name')"
              />
            </div>
            <div v-if="fieldErrors.name" class="conn-form__error-text">{{ fieldErrors.name }}</div>

            <!-- Host & Port -->
            <div v-if="shouldShowHostPort" class="conn-form__row conn-form__row--inline">
              <div class="conn-form__field">
                <label class="conn-form__label">主机 <span class="required">*</span></label>
                <input
                  v-model="formData.host"
                  type="text"
                  class="conn-form__input"
                  :class="{ 'conn-form__input--error': fieldErrors.host }"
                  placeholder="localhost 或 IP 地址"
                  @blur="validateField('host')"
                />
              </div>
              <div class="conn-form__field conn-form__field--port">
                <label class="conn-form__label">端口 <span class="required">*</span></label>
                <input
                  v-model.number="formData.port"
                  type="number"
                  class="conn-form__input"
                  :class="{ 'conn-form__input--error': fieldErrors.port }"
                  min="1"
                  max="65535"
                  @blur="validateField('port')"
                />
              </div>
            </div>
            <div v-if="shouldShowHostPort && fieldErrors.host" class="conn-form__error-text">{{ fieldErrors.host }}</div>
            <div v-if="shouldShowHostPort && fieldErrors.port" class="conn-form__error-text">{{ fieldErrors.port }}</div>

            <!-- SQL Server: Authentication Type -->
            <div v-if="formData.type === 'sqlserver'" class="conn-form__row">
              <label class="conn-form__label">身份验证</label>
              <select v-model="formData.auth_type" class="conn-form__select">
                <option value="sql">SQL Server 身份验证</option>
                <option value="windows">Windows 身份验证</option>
              </select>
            </div>

            <!-- Oracle: Connection Type -->
            <div v-if="formData.type === 'oracle'" class="conn-form__row">
              <label class="conn-form__label">连接类型</label>
              <select v-model="formData.oracle_connect_mode" class="conn-form__select">
                <option value="basic">Basic</option>
                <option value="tns">TNS</option>
              </select>
            </div>

            <template v-if="isOracleBasicMode">
              <div class="conn-form__row">
                <label class="conn-form__label">类型</label>
                <div class="conn-form__radio-group">
                  <label class="conn-form__radio">
                    <input v-model="formData.oracle_basic_identifier_type" type="radio" value="service_name">
                    <span>Service Name</span>
                  </label>
                  <label class="conn-form__radio">
                    <input v-model="formData.oracle_basic_identifier_type" type="radio" value="sid">
                    <span>SID</span>
                  </label>
                </div>
              </div>

              <div class="conn-form__row">
                <label class="conn-form__label">{{ formData.oracle_basic_identifier_type === 'sid' ? 'SID' : 'Service Name' }} <span class="required">*</span></label>
                <input
                  v-model="formData.oracle_basic_value"
                  type="text"
                  class="conn-form__input"
                  :class="{ 'conn-form__input--error': getOracleModeFieldError('oracle_basic_value') }"
                  :placeholder="formData.oracle_basic_identifier_type === 'sid' ? 'ORCL' : 'ORCLPDB1'"
                  @blur="validateField('oracle_basic_value')"
                />
              </div>
              <div v-if="getOracleModeFieldError('oracle_basic_value')" class="conn-form__error-text">{{ getOracleModeFieldError('oracle_basic_value') }}</div>
            </template>

            <template v-if="isOracleTNSMode">
              <div class="conn-form__row" v-if="formData.type === 'oracle' && formData.oracle_connect_mode === 'tns'">
                <label class="conn-form__label">TNS <span class="required">*</span></label>
                <input
                  v-model="formData.oracle_tns_name"
                  type="text"
                  class="conn-form__input"
                  :class="{ 'conn-form__input--error': getOracleModeFieldError('oracle_tns_name') }"
                  placeholder="ORCLCDB_HIGH"
                  @blur="validateField('oracle_tns_name')"
                />
              </div>
              <div v-if="getOracleModeFieldError('oracle_tns_name')" class="conn-form__error-text">{{ getOracleModeFieldError('oracle_tns_name') }}</div>
            </template>

            <!-- Username & Password (hide for Windows auth) -->
            <template v-if="formData.type !== 'sqlserver' || formData.auth_type === 'sql'">
              <div class="conn-form__row">
                <label class="conn-form__label">用户名 <span class="required">*</span></label>
                <input
                  v-model="formData.username"
                  type="text"
                  class="conn-form__input"
                  :class="{ 'conn-form__input--error': fieldErrors.username }"
                  :placeholder="currentSchema.defaultUsername"
                  @blur="validateField('username')"
                />
              </div>
              <div v-if="fieldErrors.username" class="conn-form__error-text">{{ fieldErrors.username }}</div>

              <div class="conn-form__row">
                <label class="conn-form__label">密码</label>
                <div class="conn-form__password">
                  <input
                    v-model="formData.password"
                    :type="showPassword ? 'text' : 'password'"
                    class="conn-form__input"
                    placeholder="输入密码"
                  />
                  <button type="button" class="conn-form__password-toggle" @click="showPassword = !showPassword">
                    {{ showPassword ? '隐藏' : '显示' }}
                  </button>
                </div>
              </div>
            </template>

            <!-- Database (PostgreSQL, SQL Server, Oracle) -->
            <div v-if="shouldShowDatabaseField" class="conn-form__row">
              <label class="conn-form__label">
                {{ currentSchema.databaseLabel }}
                <span v-if="currentSchema.databaseRequired" class="required">*</span>
              </label>
              <input
                v-model="formData.database"
                type="text"
                class="conn-form__input"
                :class="{ 'conn-form__input--error': fieldErrors.database }"
                :placeholder="currentSchema.databasePlaceholder"
                @blur="validateField('database')"
              />
            </div>
            <div v-if="shouldShowDatabaseField && fieldErrors.database" class="conn-form__error-text">{{ fieldErrors.database }}</div>

            <!-- Test Result -->
            <div v-if="dbTestResult" class="conn-form__test-result" :class="dbTestResult.success ? 'conn-form__test-result--success' : 'conn-form__test-result--error'">
              <span v-if="dbTestResult.success">连接成功 ({{ dbTestResult.latency_ms }}ms)</span>
              <span v-else>{{ dbTestResult.error }}</span>
            </div>
          </div>
        </div>

        <!-- ==================== Remote Tab ==================== -->
        <div v-show="activeTab === 'remote'" class="conn-editor__panel">
          <div class="conn-form">
            <div class="conn-form__row">
              <label class="conn-form__label">远程连接类型</label>
              <div class="conn-form__radio-group">
                <label class="conn-form__radio">
                  <input :checked="isRemoteTypeNoneSelected" type="radio" @change="setRemoteType('none')">
                  <span>不使用远程连接</span>
                </label>
                <label class="conn-form__radio">
                  <input :checked="isRemoteTypeSSHSelected" type="radio" @change="setRemoteType('ssh')">
                  <span>SSH</span>
                </label>
                <label class="conn-form__radio">
                  <input :checked="isRemoteTypeWinRMSelected" type="radio" @change="setRemoteType('winrm')">
                  <span>WinRM</span>
                </label>
              </div>
            </div>

            <div v-if="isRemoteTypeNoneSelected" class="conn-form__hint">未配置远程连接方式。</div>

            <template v-if="isRemoteTypeSSHSelected">
              <div class="conn-form__row conn-form__row--inline">
                <div class="conn-form__field">
                  <label class="conn-form__label">SSH 主机 <span class="required">*</span></label>
                  <input
                    v-model="formData.ssh_host"
                    type="text"
                    class="conn-form__input"
                    :class="{ 'conn-form__input--error': fieldErrors.ssh_host }"
                    placeholder="SSH 服务器地址"
                    @input="onSshHostChange"
                  />
                </div>
                <button type="button" class="conn-form__sync-btn" @click="syncCurrentRemoteHostFromGeneral" title="从数据库主机同步">
                  同步
                </button>
              </div>
              <div v-if="fieldErrors.ssh_host" class="conn-form__error-text">{{ fieldErrors.ssh_host }}</div>

              <div class="conn-form__row conn-form__row--inline">
                <div class="conn-form__field">
                  <label class="conn-form__label">SSH 端口 <span class="required">*</span></label>
                  <input
                    v-model.number="formData.ssh_port"
                    type="number"
                    class="conn-form__input"
                    :class="{ 'conn-form__input--error': fieldErrors.ssh_port }"
                    placeholder="22"
                    min="1"
                    max="65535"
                  />
                </div>
              </div>
              <div v-if="fieldErrors.ssh_port" class="conn-form__error-text">{{ fieldErrors.ssh_port }}</div>

              <div class="conn-form__row">
                <label class="conn-form__label">SSH 用户名 <span class="required">*</span></label>
                <input
                  v-model="formData.ssh_username"
                  type="text"
                  class="conn-form__input"
                  :class="{ 'conn-form__input--error': fieldErrors.ssh_username }"
                  placeholder="SSH 用户名"
                />
              </div>
              <div v-if="fieldErrors.ssh_username" class="conn-form__error-text">{{ fieldErrors.ssh_username }}</div>

              <div class="conn-form__row">
                <label class="conn-form__label">SSH 密码</label>
                <div class="conn-form__password">
                  <input
                    v-model="formData.ssh_password"
                    :type="showSshPassword ? 'text' : 'password'"
                    class="conn-form__input"
                    placeholder="SSH 密码"
                  />
                  <button type="button" class="conn-form__password-toggle" @click="showSshPassword = !showSshPassword">
                    {{ showSshPassword ? '隐藏' : '显示' }}
                  </button>
                </div>
              </div>

              <div v-if="sshTestResult" class="conn-form__test-result" :class="sshTestResult.success ? 'conn-form__test-result--success' : 'conn-form__test-result--error'">
                <span v-if="sshTestResult.success">SSH 连接成功 ({{ sshTestResult.latency_ms }}ms)</span>
                <span v-else>{{ sshTestResult.error }}</span>
              </div>
            </template>

            <template v-if="isRemoteTypeWinRMSelected">
              <div class="conn-form__row conn-form__row--inline">
                <div class="conn-form__field">
                  <label class="conn-form__label">WinRM 主机 <span class="required">*</span></label>
                  <input
                    v-model="formData.winrm_host"
                    type="text"
                    class="conn-form__input"
                    :class="{ 'conn-form__input--error': fieldErrors.winrm_host }"
                    placeholder="WinRM 主机地址"
                    @input="onWinRMHostChange"
                  />
                </div>
                <button type="button" class="conn-form__sync-btn" @click="syncCurrentRemoteHostFromGeneral" title="从数据库主机同步">
                  同步
                </button>
              </div>
              <div v-if="fieldErrors.winrm_host" class="conn-form__error-text">{{ fieldErrors.winrm_host }}</div>

              <div class="conn-form__row conn-form__row--inline">
                <div class="conn-form__field">
                  <label class="conn-form__label">WinRM 端口 <span class="required">*</span></label>
                  <input
                    v-model.number="formData.winrm_port"
                    type="number"
                    class="conn-form__input"
                    :class="{ 'conn-form__input--error': fieldErrors.winrm_port }"
                    placeholder="5985"
                    min="1"
                    max="65535"
                    @input="onWinRMPortChange"
                  />
                </div>
              </div>
              <div v-if="fieldErrors.winrm_port" class="conn-form__error-text">{{ fieldErrors.winrm_port }}</div>

              <div class="conn-form__row">
                <label class="conn-form__label">WinRM 用户名 <span class="required">*</span></label>
                <input
                  v-model="formData.winrm_username"
                  type="text"
                  class="conn-form__input"
                  :class="{ 'conn-form__input--error': fieldErrors.winrm_username }"
                  placeholder="WinRM 用户名"
                />
              </div>
              <div v-if="fieldErrors.winrm_username" class="conn-form__error-text">{{ fieldErrors.winrm_username }}</div>

              <div class="conn-form__row">
                <label class="conn-form__label">WinRM 密码</label>
                <div class="conn-form__password">
                  <input
                    v-model="formData.winrm_password"
                    :type="showWinrmPassword ? 'text' : 'password'"
                    class="conn-form__input"
                    placeholder="WinRM 密码"
                  />
                  <button type="button" class="conn-form__password-toggle" @click="showWinrmPassword = !showWinrmPassword">
                    {{ showWinrmPassword ? '隐藏' : '显示' }}
                  </button>
                </div>
              </div>

              <div class="conn-form__row">
                <label class="conn-form__label">协议</label>
                <div class="conn-form__radio-group">
                  <label class="conn-form__radio">
                    <input v-model="formData.winrm_scheme" type="radio" value="http">
                    <span>HTTP</span>
                  </label>
                  <label class="conn-form__radio">
                    <input v-model="formData.winrm_scheme" type="radio" value="https">
                    <span>HTTPS</span>
                  </label>
                </div>
              </div>

              <div class="conn-form__row">
                <label class="conn-form__label">认证方式</label>
                <input
                  v-model="formData.winrm_auth_type"
                  type="text"
                  class="conn-form__input"
                  readonly
                />
              </div>

              <div v-if="winrmTestResult" class="conn-form__test-result" :class="winrmTestResult.success ? 'conn-form__test-result--success' : 'conn-form__test-result--error'">
                <span v-if="winrmTestResult.success">WinRM 连接成功 ({{ winrmTestResult.latency_ms }}ms)</span>
                <span v-else>{{ winrmTestResult.error }}</span>
              </div>
            </template>
          </div>
        </div>

        <!-- ==================== AI Assistant Tab ==================== -->
        <div v-show="activeTab === 'ai'" class="conn-editor__panel conn-editor__panel--ai">
          <div class="ai-config">
            <!-- Left: Assistant List -->
            <div class="ai-config__sidebar">
              <div class="ai-config__sidebar-header">AI 助手</div>
              <div class="ai-config__list">
                <div
                  v-for="assistant in formData.ai_assistants"
                  :key="assistant.id"
                  class="ai-config__list-item"
                  :class="{ 'ai-config__list-item--active': formData.selectedAssistantId === assistant.id }"
                  @click="selectAssistant(assistant.id)"
                  :title="assistant.name"
                >
                  {{ assistant.name }}
                </div>
              </div>
              <div class="ai-config__sidebar-actions">
                <button class="ai-config__action-btn" @click="toggleProviderDropdown($event)" title="添加助手">+</button>
                <button
                  class="ai-config__action-btn"
                  :class="{ 'ai-config__action-btn--disabled': !selectedAssistant }"
                  @click="removeAssistant(selectedAssistant?.id)"
                  :disabled="!selectedAssistant"
                  title="删除助手"
                >−</button>
              </div>
            </div>

            <!-- Provider Dropdown - Fixed position using Teleport -->
            <Teleport to="body">
              <div
                v-if="showProviderDropdown"
                class="ai-config__provider-dropdown"
                :style="{ top: dropdownPosition.top + 'px', left: dropdownPosition.left + 'px' }"
              >
                <div
                  v-for="p in aiProviders"
                  :key="p.value"
                  class="ai-config__provider-option"
                  @click="addAssistant(p.value)"
                >
                  {{ p.label }}
                </div>
              </div>
            </Teleport>

            <!-- Right: Config Form -->
            <div class="ai-config__main" @click="showProviderDropdown = false">
              <!-- Empty state when no assistants -->
              <div v-if="formData.ai_assistants.length === 0" class="ai-config__empty">
                <div class="ai-config__empty-icon">🤖</div>
                <div class="ai-config__empty-title">暂无 AI 助手</div>
                <div class="ai-config__empty-hint">请点击左下角 + 添加 AI 助手</div>
              </div>

              <template v-else-if="selectedAssistant">
                <!-- Basic Config Section -->
                <div class="ai-config__section">
                  <div class="ai-config__row">
                    <label class="ai-config__label">AI 助手名称</label>
                    <input v-model="selectedAssistant.name" type="text" class="ai-config__input" />
                  </div>

                  <div class="ai-config__row">
                    <label class="ai-config__label">AI 提供商</label>
                    <div class="ai-config__readonly">{{ getProviderInfo(selectedAssistant.provider).label }}</div>
                  </div>

                  <div class="ai-config__row">
                    <label class="ai-config__label">API 主机</label>
                    <input
                      v-model="selectedAssistant.api_host"
                      type="text"
                      class="ai-config__input"
                      :class="{ 'ai-config__input--error': getAIFieldError(selectedAssistant.id, 'api_host') }"
                      placeholder="https://api.example.com"
                    />
                  </div>
                  <div v-if="getAIFieldError(selectedAssistant.id, 'api_host')" class="ai-config__error-text">
                    {{ getAIFieldError(selectedAssistant.id, 'api_host') }}
                  </div>

                  <!-- API Endpoint is readonly and linked to provider -->
                  <div class="ai-config__row">
                    <label class="ai-config__label">API 端点</label>
                    <div class="ai-config__readonly">{{ getProviderInfo(selectedAssistant.provider).endpoint }}</div>
                  </div>

                  <div v-if="shouldShowAssistantApiKeyField(selectedAssistant)" class="ai-config__row">
                    <label class="ai-config__label">
                      API 密钥
                      <span class="ai-config__help" title="API 密钥用于身份验证，请妥善保管">?</span>
                    </label>
                    <div class="ai-config__key-field">
                      <input
                        v-model="selectedAssistant.api_key"
                        :type="showApiKey[selectedAssistant.id] ? 'text' : 'password'"
                        class="ai-config__input ai-config__key-input"
                        :class="{ 'ai-config__input--error': getAIFieldError(selectedAssistant.id, 'api_key') }"
                        placeholder="sk-..."
                      />
                      <button
                        type="button"
                        class="ai-config__key-toggle"
                        @click="toggleApiKeyVisibility(selectedAssistant.id)"
                        :title="showApiKey[selectedAssistant.id] ? '隐藏密钥' : '显示密钥'"
                      >
                        <span v-if="showApiKey[selectedAssistant.id]">🔒</span>
                        <span v-else>👁</span>
                      </button>
                    </div>
                  </div>
                  <div v-if="shouldShowAssistantApiKeyField(selectedAssistant) && getAIFieldError(selectedAssistant.id, 'api_key')" class="ai-config__error-text">
                    {{ getAIFieldError(selectedAssistant.id, 'api_key') }}
                  </div>

                  <div class="ai-config__row">
                    <label class="ai-config__label">模型</label>
                    <div class="ai-config__model-field">
                      <input
                        v-model="selectedAssistant.model"
                        type="text"
                        class="ai-config__input ai-config__model-input"
                        :class="{ 'ai-config__input--error': getAIFieldError(selectedAssistant.id, 'model') }"
                        placeholder="模型名称"
                      />
                      <button
                        type="button"
                        class="ai-config__model-btn"
                        :class="{ 'ai-config__model-btn--loading': modelQuerying }"
                        :disabled="modelQuerying"
                        @click="handleQueryModels"
                        title="查询可用模型"
                      >{{ modelQuerying ? '查询中' : '查询' }}</button>
                    </div>
                  </div>
                  <div v-if="getAIFieldError(selectedAssistant.id, 'model')" class="ai-config__error-text">
                    {{ getAIFieldError(selectedAssistant.id, 'model') }}
                  </div>
                  <!-- Model query error -->
                  <div v-if="modelQueryError" class="ai-config__error-text">{{ modelQueryError }}</div>

                  <div class="ai-config__row ai-config__row--temp">
                    <label class="ai-config__label">温度</label>
                    <div class="ai-config__temp-control">
                      <div class="ai-config__temp-labels">
                        <span>0.0 更有确定性</span>
                        <span>1.0 平衡</span>
                        <span>2.0 更有创造性</span>
                      </div>
                      <div class="ai-config__temp-slider">
                        <input
                          v-model.number="selectedAssistant.temperature"
                          type="range"
                          min="0"
                          max="2"
                          step="0.1"
                          class="ai-config__range"
                        />
                        <span class="ai-config__temp-value">{{ selectedAssistant.temperature.toFixed(1) }}</span>
                      </div>
                    </div>
                  </div>

                  <div class="ai-config__row">
                    <label class="ai-config__label">说明</label>
                    <textarea
                      v-model="selectedAssistant.description"
                      class="ai-config__textarea"
                      rows="2"
                      placeholder="（可选）"
                    ></textarea>
                  </div>
                </div>

                <!-- Test Button -->
                <div class="ai-config__test-area">
                  <button
                    type="button"
                    class="ai-config__test-btn"
                    :class="{
                      'ai-config__test-btn--testing': aiTestStatus === 'testing',
                      'ai-config__test-btn--success': aiTestStatus === 'success',
                      'ai-config__test-btn--error': aiTestStatus === 'error'
                    }"
                    :disabled="aiTesting"
                    @click="handleTestAI"
                  >
                    <span v-if="aiTestStatus === 'testing'" class="spinner"></span>
                    <span v-else-if="aiTestStatus === 'success'">✓</span>
                    <span v-else-if="aiTestStatus === 'error'">✗</span>
                    测试连接
                  </button>
                  <button
                    type="button"
                    class="ai-config__test-btn ai-config__test-btn--primary"
                    @click="openAiTestDialog"
                  >
                    测试模型
                  </button>
                  <span v-if="aiTestResult" class="ai-config__test-result" :class="aiTestResult.success ? 'ai-config__test-result--success' : 'ai-config__test-result--error'">
                    {{ aiTestResult.success ? (aiTestResult.message || '连接成功') : aiTestResult.error }}
                  </span>
                </div>
              </template>
            </div>
          </div>
        </div>
      </div>
    </div>

    <Teleport to="body">
      <div v-if="showModelSelector" class="ai-config__model-selector-overlay" @click.self="closeModelSelector">
        <div class="ai-config__model-selector-dialog" role="dialog" aria-modal="true" aria-label="选择模型">
          <div class="ai-config__model-selector-header">
            <span>选择模型</span>
            <button type="button" class="ai-config__model-selector-close" @click="closeModelSelector">×</button>
          </div>
          <div class="ai-config__model-selector-list">
            <button
              v-for="model in availableModels"
              :key="model.id"
              type="button"
              class="ai-config__model-option"
              :class="{ 'ai-config__model-option--selected': pendingModelSelection === model.id }"
              @click="selectModel(model.id)"
            >
              {{ model.name }}
            </button>
          </div>
          <div class="ai-config__model-selector-actions">
            <button type="button" class="ai-config__selector-btn ai-config__selector-btn--secondary" @click="closeModelSelector">
              取消
            </button>
            <button
              type="button"
              class="ai-config__selector-btn ai-config__selector-btn--primary"
              :disabled="!pendingModelSelection"
              @click="confirmModelSelection"
            >
              确定
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <AiTestDialog
      :visible="showAiTestDialog"
      :assistant="selectedAssistant"
      :provider-label="selectedAssistant ? getProviderInfo(selectedAssistant.provider).label : ''"
      @close="closeAiTestDialog"
    />

    <!-- Footer Actions -->
    <div class="conn-editor__footer">
      <div class="conn-editor__footer-left">
        <!-- General tab: Test DB button (hidden during type selection) -->
        <button
          v-if="activeTab === 'general' && !showTypeSelection"
          type="button"
          class="conn-editor__btn conn-editor__btn--test"
          :class="{
            'conn-editor__btn--testing': dbTesting,
            'conn-editor__btn--success': dbTestStatus === 'success',
            'conn-editor__btn--error': dbTestStatus === 'error'
          }"
          :disabled="dbTesting || saving"
          @click="handleTestDB"
        >
          <span v-if="dbTesting" class="spinner"></span>
          <span v-else-if="dbTestStatus === 'success'" class="icon-success">✓</span>
          <span v-else-if="dbTestStatus === 'error'" class="icon-error">✗</span>
          {{ dbTesting ? '测试中...' : 'Test DB' }}
        </button>

        <!-- Remote tab: Test SSH / Test WinRM button -->
        <button
          v-if="activeTab === 'remote' && isRemoteTypeSSHSelected"
          type="button"
          class="conn-editor__btn conn-editor__btn--test"
          :class="{
            'conn-editor__btn--testing': sshTesting,
            'conn-editor__btn--success': sshTestStatus === 'success',
            'conn-editor__btn--error': sshTestStatus === 'error'
          }"
          :disabled="sshTesting || saving"
          @click="handleTestSSH"
        >
          <span v-if="sshTesting" class="spinner"></span>
          <span v-else-if="sshTestStatus === 'success'" class="icon-success">✓</span>
          <span v-else-if="sshTestStatus === 'error'" class="icon-error">✗</span>
          {{ sshTesting ? '测试中...' : 'Test SSH' }}
        </button>
        <button
          v-if="activeTab === 'remote' && isRemoteTypeWinRMSelected"
          type="button"
          class="conn-editor__btn conn-editor__btn--test"
          :class="{
            'conn-editor__btn--testing': winrmTesting,
            'conn-editor__btn--success': winrmTestStatus === 'success',
            'conn-editor__btn--error': winrmTestStatus === 'error'
          }"
          :disabled="winrmTesting || saving"
          @click="handleTestWinRM"
        >
          <span v-if="winrmTesting" class="spinner"></span>
          <span v-else-if="winrmTestStatus === 'success'" class="icon-success">✓</span>
          <span v-else-if="winrmTestStatus === 'error'" class="icon-error">✗</span>
          {{ winrmTesting ? '测试中...' : 'Test WinRM' }}
        </button>
      </div>

      <div class="conn-editor__footer-right">
        <button type="button" class="conn-editor__btn conn-editor__btn--cancel" @click="handleCancel" :disabled="saving">
          Cancel
        </button>
        <button
          v-if="!showTypeSelection || isEditing"
          type="button"
          class="conn-editor__btn conn-editor__btn--save"
          @click="handleSave"
          :disabled="saving"
        >
          {{ saving ? 'Saving...' : 'Save' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ============================================================
   Navicat-style Light Theme Connection Editor
   ============================================================ */

.conn-editor {
  --bg-primary: #ffffff;
  --bg-secondary: #f5f7fa;
  --bg-input: #ffffff;
  --border-color: #dcdfe6;
  --border-focus: #409eff;
  --text-primary: #303133;
  --text-secondary: #606266;
  --text-muted: #909399;
  --success: #67c23a;
  --success-bg: #f0f9eb;
  --error: #f56c6c;
  --error-bg: #fef0f0;
  --primary: #409eff;
  --primary-hover: #66b1ff;
  --primary-light: #ecf5ff;

  background-color: var(--bg-primary);
  border-radius: 4px;
  width: 100%;
  min-width: 600px;
  max-width: 720px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

/* Header */
.conn-editor__header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
  background-color: var(--bg-primary);
}

.conn-editor__title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

/* Type Selection */
.conn-editor__type-select {
  padding: 24px;
}

.conn-editor__type-label {
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 16px;
}

.conn-editor__type-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.conn-editor__type-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px 16px;
  background-color: var(--bg-secondary);
  border: 2px solid var(--border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.conn-editor__type-btn:hover {
  border-color: var(--primary);
  background-color: #ecf5ff;
}

.conn-editor__type-btn--active {
  border-color: var(--primary);
  background-color: #ecf5ff;
}

.conn-editor__type-icon {
  font-size: 28px;
  margin-bottom: 8px;
}

.conn-editor__type-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}

/* Content */
.conn-editor__content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

/* Error Banner */
.conn-editor__error {
  padding: 12px 20px;
  background-color: var(--error-bg);
  color: var(--error);
  font-size: 13px;
  border-bottom: 1px solid var(--error);
}

/* Tabs */
.conn-editor__tabs {
  display: flex;
  border-bottom: 1px solid var(--border-color);
  background-color: var(--bg-secondary);
}

.conn-editor__tab {
  padding: 12px 24px;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.conn-editor__tab:hover {
  color: var(--primary);
}

.conn-editor__tab--active {
  color: var(--primary);
  border-bottom-color: var(--primary);
  background-color: var(--bg-primary);
}

/* Tab Content */
.conn-editor__tab-content {
  flex: 1;
  overflow-y: auto;
  min-height: 300px;
}

.conn-editor__panel {
  padding: 20px;
}

.conn-editor__panel--ai {
  padding: 0;
}

/* Form */
.conn-form__row {
  margin-bottom: 16px;
  display: flex;
  align-items: center;
}

.conn-form__row--inline {
  gap: 12px;
}

.conn-form__row--inline .conn-form__field {
  flex: 1;
}

.conn-form__row--checkbox {
  margin-top: 8px;
}

.conn-form__row--test {
  margin-top: 20px;
}

.conn-form__label {
  width: 140px;
  min-width: 140px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  text-align: right;
  padding-right: 12px;
}

.conn-form__label .required {
  color: var(--error);
  margin-left: 2px;
}

.conn-form__field {
  flex: 1;
  display: flex;
  align-items: center;
}

.conn-form__field--port {
  width: 100px;
  flex: none;
}

.conn-form__radio-group {
  flex: 1;
  display: flex;
  gap: 16px;
}

.conn-form__radio {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-primary);
}

.conn-form__input,
.conn-form__select,
.conn-form__textarea {
  flex: 1;
  padding: 8px 12px;
  background-color: var(--bg-input);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 13px;
  color: var(--text-primary);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.conn-form__input:focus,
.conn-form__select:focus,
.conn-form__textarea:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.conn-form__input--error {
  border-color: var(--error);
}

.conn-form__input--error:focus {
  box-shadow: 0 0 0 2px rgba(245, 108, 108, 0.2);
}

.conn-form__select {
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%23606266' d='M6 8L1 3h10z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 10px center;
  padding-right: 30px;
}

.conn-form__textarea {
  resize: vertical;
  min-height: 60px;
}

.conn-form__password {
  flex: 1;
  display: flex;
  align-items: center;
  position: relative;
}

.conn-form__password .conn-form__input {
  padding-right: 60px;
}

.conn-form__password-toggle {
  position: absolute;
  right: 8px;
  background: none;
  border: none;
  font-size: 12px;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px 8px;
}

.conn-form__password-toggle:hover {
  color: var(--primary);
}

.conn-form__sync-btn {
  padding: 6px 12px;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 12px;
  color: var(--text-secondary);
  cursor: pointer;
}

.conn-form__sync-btn:hover {
  background-color: #e9ecf0;
}

.conn-form__model-btn {
  padding: 8px 12px;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 14px;
  color: var(--text-secondary);
  cursor: pointer;
}

.conn-form__model-btn:hover {
  background-color: #e9ecf0;
}

.conn-form__slider {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
}

.conn-form__range {
  flex: 1;
  height: 4px;
  -webkit-appearance: none;
  appearance: none;
  background: var(--border-color);
  border-radius: 2px;
  cursor: pointer;
}

.conn-form__range::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 16px;
  height: 16px;
  background: var(--primary);
  border-radius: 50%;
  cursor: pointer;
}

.conn-form__slider-value {
  min-width: 36px;
  font-size: 13px;
  color: var(--text-primary);
  text-align: center;
}

.conn-form__checkbox {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-primary);
}

.conn-form__checkbox input {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: var(--primary);
}

.conn-form__hint {
  margin-left: 152px;
  margin-top: -12px;
  margin-bottom: 16px;
  font-size: 11px;
  color: var(--text-muted);
}

.conn-form__error-text {
  margin-left: 152px;
  margin-top: -12px;
  margin-bottom: 16px;
  font-size: 11px;
  color: var(--error);
}

.conn-form__section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 20px 0 16px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color);
}

.conn-form__test-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 13px;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.conn-form__test-btn:hover:not(:disabled) {
  background-color: #e9ecf0;
}

.conn-form__test-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.conn-form__test-btn--testing {
  color: var(--primary);
  border-color: var(--primary);
}

.conn-form__test-btn--success {
  color: var(--success);
  border-color: var(--success);
  background-color: var(--success-bg);
}

.conn-form__test-btn--error {
  color: var(--error);
  border-color: var(--error);
  background-color: var(--error-bg);
}

.conn-form__test-result {
  margin-left: 152px;
  margin-top: 8px;
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 12px;
}

.conn-form__test-result--success {
  background-color: var(--success-bg);
  color: var(--success);
  border: 1px solid var(--success);
}

.conn-form__test-result--error {
  background-color: var(--error-bg);
  color: var(--error);
  border: 1px solid var(--error);
}

/* ============================================================
   AI Config - Desktop Tool Style
   ============================================================ */

.ai-config {
  display: flex;
  min-height: 420px;
  background-color: var(--bg-primary);
}

/* Left Sidebar - Assistant List */
.ai-config__sidebar {
  width: 140px;
  border-right: 1px solid var(--border-color);
  background-color: var(--bg-secondary);
  display: flex;
  flex-direction: column;
}

.ai-config__sidebar-header {
  padding: 8px 10px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 1px solid var(--border-color);
}

.ai-config__list {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
}

.ai-config__list-item {
  padding: 6px 12px;
  font-size: 12px;
  color: var(--text-primary);
  cursor: pointer;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  border-left: 2px solid transparent;
  transition: none;
}

.ai-config__list-item:hover {
  background-color: #e8eaed;
}

.ai-config__list-item--active {
  background-color: #fff;
  border-left-color: var(--primary);
  color: var(--primary);
}

.ai-config__sidebar-actions {
  display: flex;
  border-top: 1px solid var(--border-color);
  padding: 4px;
  gap: 4px;
}

.ai-config__action-btn {
  flex: 1;
  height: 26px;
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 2px;
  color: var(--text-secondary);
  font-size: 14px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ai-config__action-btn:hover:not(.ai-config__action-btn--disabled) {
  background-color: #e8eaed;
  color: var(--text-primary);
}

.ai-config__action-btn--disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* Provider Dropdown */
.ai-config__add-wrapper {
  position: relative;
  flex: 1;
}

.ai-config__provider-dropdown {
  position: fixed;
  width: 190px;
  max-height: 204px;
  overflow-y: auto;
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  z-index: 1000;
}

.ai-config__provider-option {
  padding: 8px 12px;
  font-size: 12px;
  color: var(--text-primary);
  cursor: pointer;
  white-space: nowrap;
  overflow: visible;
}

.ai-config__provider-option:hover {
  background-color: #ecf5ff;
  color: var(--primary);
}

/* Readonly Field */
.ai-config__readonly {
  flex: 1;
  height: 26px;
  padding: 0 8px;
  font-size: 12px;
  line-height: 26px;
  color: var(--text-secondary);
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ai-config__readonly--hint {
  font-style: italic;
  color: var(--text-muted);
}

/* Model Field with Query Button */
.ai-config__model-field {
  flex: 1;
  display: flex;
  gap: 4px;
}

.ai-config__model-input {
  flex: 1;
}

.ai-config__model-btn {
  width: 32px;
  height: 26px;
  padding: 0;
  border: 1px solid var(--border-color);
  border-radius: 2px;
  background-color: var(--bg-secondary);
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.ai-config__model-btn:hover:not(:disabled) {
  background-color: var(--primary);
  border-color: var(--primary);
  color: #fff;
}

.ai-config__model-btn--loading,
.ai-config__model-btn:disabled {
  opacity: 0.6;
  cursor: wait;
}

/* Model Query Error */
.ai-config__model-error {
  width: 100%;
  margin-top: 4px;
  font-size: 11px;
  color: var(--error);
}

/* Model Selector Dialog */
.ai-config__model-selector-overlay {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: rgba(0, 0, 0, 0.22);
  z-index: 1200;
}

.ai-config__model-selector-dialog {
  width: min(420px, calc(100vw - 32px));
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.2);
  overflow: hidden;
}

.ai-config__model-selector-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-color);
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
}

.ai-config__model-selector-close {
  width: 20px;
  height: 20px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 16px;
  line-height: 1;
}

.ai-config__model-selector-close:hover {
  color: var(--text-primary);
}

.ai-config__model-selector-list {
  max-height: 240px;
  overflow-y: auto;
  padding: 8px 0;
}

.ai-config__model-option {
  display: block;
  width: 100%;
  padding: 8px 12px;
  font-size: 12px;
  color: var(--text-secondary);
  text-align: left;
  border: none;
  background: transparent;
  cursor: pointer;
  transition: background-color 0.15s;
}

.ai-config__model-option:hover {
  background-color: var(--bg-secondary);
}

.ai-config__model-option--selected {
  background-color: var(--primary-light);
  color: var(--primary);
}

.ai-config__model-selector-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 12px 16px 16px;
  border-top: 1px solid var(--border-color);
}

.ai-config__selector-btn {
  min-width: 72px;
  padding: 8px 14px;
  border-radius: 4px;
  border: 1px solid var(--border-color);
  font-size: 12px;
  cursor: pointer;
}

.ai-config__selector-btn--secondary {
  background-color: var(--bg-primary);
  color: var(--text-secondary);
}

.ai-config__selector-btn--primary {
  background-color: var(--primary);
  border-color: var(--primary);
  color: #fff;
}

.ai-config__selector-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Right Main Config Area */
.ai-config__main {
  flex: 1;
  padding: 12px 16px;
  overflow-y: auto;
}

/* Section */
.ai-config__section {
  margin-bottom: 12px;
}

.ai-config__section--bordered {
  padding-top: 10px;
  border-top: 1px solid var(--border-color);
}

.ai-config__section-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* Form Row */
.ai-config__row {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
  min-height: 26px;
}

.ai-config__row--temp {
  align-items: flex-start;
  flex-direction: column;
}

.ai-config__row--checkbox {
  padding-left: 100px;
}

/* Label */
.ai-config__label {
  width: 100px;
  min-width: 100px;
  font-size: 12px;
  color: var(--text-secondary);
  flex-shrink: 0;
  line-height: 26px;
}

.ai-config__help {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  margin-left: 4px;
  font-size: 10px;
  font-weight: 600;
  color: var(--text-muted);
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 50%;
  cursor: help;
}

/* Input */
.ai-config__input {
  flex: 1;
  height: 26px;
  padding: 0 8px;
  font-size: 12px;
  border: 1px solid var(--border-color);
  border-radius: 2px;
  background-color: var(--bg-input);
  color: var(--text-primary);
}

.ai-config__input:focus {
  outline: none;
  border-color: var(--primary);
}

.ai-config__input--error {
  border-color: var(--error);
}

.ai-config__input--error:focus {
  box-shadow: 0 0 0 2px rgba(245, 108, 108, 0.2);
}

.ai-config__input::placeholder {
  color: var(--text-muted);
}

/* Field Error Text */
.ai-config__error-text {
  width: 100%;
  margin-top: 2px;
  margin-bottom: 4px;
  font-size: 11px;
  color: var(--error);
  padding-left: 100px;
}

/* Select */
.ai-config__select {
  flex: 1;
  height: 26px;
  padding: 0 6px;
  font-size: 12px;
  border: 1px solid var(--border-color);
  border-radius: 2px;
  background-color: var(--bg-input);
  color: var(--text-primary);
  cursor: pointer;
}

.ai-config__select:focus {
  outline: none;
  border-color: var(--primary);
}

/* Textarea */
.ai-config__textarea {
  flex: 1;
  padding: 4px 8px;
  font-size: 12px;
  border: 1px solid var(--border-color);
  border-radius: 2px;
  background-color: var(--bg-input);
  color: var(--text-primary);
  resize: none;
  font-family: inherit;
  line-height: 1.4;
}

.ai-config__textarea:focus {
  outline: none;
  border-color: var(--primary);
}

/* API Key Field */
.ai-config__key-field {
  flex: 1;
  display: flex;
  align-items: center;
}

.ai-config__key-input {
  border-radius: 2px 0 0 2px;
  flex: 1;
}

.ai-config__key-toggle {
  width: 32px;
  height: 26px;
  padding: 0;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-left: none;
  border-radius: 0 2px 2px 0;
  color: var(--text-muted);
  font-size: 12px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ai-config__key-toggle:hover {
  background-color: #e8eaed;
  color: var(--text-primary);
}

/* Temperature Control */
.ai-config__temp-control {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.ai-config__temp-labels {
  display: flex;
  justify-content: space-between;
  font-size: 10px;
  color: var(--text-muted);
  padding: 0 2px;
}

.ai-config__temp-slider {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ai-config__range {
  flex: 1;
  height: 4px;
  background: var(--border-color);
  border-radius: 2px;
  appearance: none;
  cursor: pointer;
}

.ai-config__range::-webkit-slider-thumb {
  appearance: none;
  width: 14px;
  height: 14px;
  background: var(--primary);
  border-radius: 50%;
  cursor: pointer;
}

.ai-config__range::-webkit-slider-thumb:hover {
  background: var(--primary-hover);
}

.ai-config__temp-value {
  min-width: 28px;
  font-size: 11px;
  color: var(--text-secondary);
  text-align: right;
  font-family: monospace;
}

/* Checkbox */
.ai-config__checkbox {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-primary);
  cursor: pointer;
}

.ai-config__checkbox input[type="checkbox"] {
  width: 14px;
  height: 14px;
  margin: 0;
  cursor: pointer;
}

/* Test Button Area */
.ai-config__test-area {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
}

.ai-config__test-btn {
  padding: 6px 16px;
  font-size: 12px;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 2px;
  color: var(--text-primary);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: none;
}

.ai-config__test-btn:hover:not(:disabled) {
  background-color: #e8eaed;
}

.ai-config__test-btn--primary {
  background-color: var(--primary);
  border-color: var(--primary);
  color: #fff;
}

.ai-config__test-btn--primary:hover:not(:disabled) {
  background-color: var(--primary-hover);
}

.ai-config__test-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.ai-config__test-btn--testing {
  color: var(--primary);
}

.ai-config__test-btn--success {
  color: var(--success);
  border-color: var(--success);
}

.ai-config__test-btn--error {
  color: var(--error);
  border-color: var(--error);
}

.ai-config__test-result {
  font-size: 11px;
}

.ai-config__test-result--success {
  color: var(--success);
}

.ai-config__test-result--error {
  color: var(--error);
}

/* Footer */
.conn-editor__footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-top: 1px solid var(--border-color);
  background-color: var(--bg-primary);
}

.conn-editor__footer-left,
.conn-editor__footer-right {
  display: flex;
  gap: 12px;
}

.conn-editor__btn {
  padding: 8px 20px;
  border-radius: 4px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.conn-editor__btn--test {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
}

.conn-editor__btn--test:hover:not(:disabled) {
  background-color: #e9ecf0;
}

.conn-editor__btn--test:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.conn-editor__btn--testing {
  color: var(--primary);
  border-color: var(--primary);
}

.conn-editor__btn--success {
  color: var(--success);
  border-color: var(--success);
  background-color: var(--success-bg);
}

.conn-editor__btn--error {
  color: var(--error);
  border-color: var(--error);
  background-color: var(--error-bg);
}

.conn-editor__btn--cancel {
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
}

.conn-editor__btn--cancel:hover {
  background-color: #e9ecf0;
}

.conn-editor__btn--save {
  background-color: var(--primary);
  border: 1px solid var(--primary);
  color: white;
}

.conn-editor__btn--save:hover {
  background-color: var(--primary-hover);
}

.conn-editor__btn--save:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Spinner */
.spinner {
  width: 14px;
  height: 14px;
  border: 2px solid transparent;
  border-top-color: currentColor;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.icon-success {
  font-weight: 600;
}

.icon-error {
  font-weight: 600;
}
</style>
