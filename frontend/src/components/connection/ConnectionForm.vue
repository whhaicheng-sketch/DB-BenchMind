<script setup>
/**
 * ConnectionForm.vue
 * Navicat-style database connection editor with tabs: General / SSH / AI Assistant
 * Supports MySQL, PostgreSQL, Oracle, and SQL Server with database-specific fields.
 */
import { ref, computed, watch, nextTick } from 'vue'
import { useConnectionStore } from '../../stores/connection'
import { TestConnectionDirect, TestSSHConnection, TestAIConnection } from '../../../wailsjs/go/bindings/ConnectionBinding'

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
  { id: 'ssh', label: 'SSH' },
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
  connect_type: 'service_name', // service_name or sid
  // SQL Server specific
  auth_type: 'sql', // sql or windows
  // SSH Configuration
  ssh_host: '',
  ssh_port: 22,
  ssh_username: '',
  ssh_password: '',
  // AI Assistant Configuration
  ai_assistants: [
    {
      id: 'default',
      name: 'DeepSeek',
      provider: 'deepseek',
      api_host: 'https://api.deepseek.com',
      api_endpoint: '/v1/chat/completions',
      api_key: '',
      model: 'deepseek-chat',
      temperature: 0.7,
      description: '',
      enter_action: 'send',
      compare_with_others: false,
      language: 'zh-CN'
    }
  ],
  selectedAssistantId: 'default'
})

// UI State
const showPassword = ref(false)
const showSshPassword = ref(false)
const showApiKey = ref({})
const saving = ref(false)
const formError = ref(null)
const fieldErrors = ref({})

// Test States
const dbTesting = ref(false)
const dbTestResult = ref(null)
const dbTestStatus = ref('idle')

const sshTesting = ref(false)
const sshTestResult = ref(null)
const sshTestStatus = ref('idle')

const aiTesting = ref(false)
const aiTestResult = ref(null)
const aiTestStatus = ref('idle')

// Flag to track if SSH host was manually modified
const sshHostManuallyModified = ref(false)

// ============================================================
// Computed Properties
// ============================================================
const isEditing = computed(() => props.mode === 'edit' && props.connectionId)
const title = computed(() => isEditing.value ? '编辑连接' : '新建连接')
const currentSchema = computed(() => DB_SCHEMA[formData.value.type])
const selectedAssistant = computed(() => {
  return formData.value.ai_assistants.find(a => a.id === formData.value.selectedAssistantId) || formData.value.ai_assistants[0]
})

// ============================================================
// AI Provider Options
// ============================================================
const aiProviders = [
  { value: 'deepseek', label: 'DeepSeek' },
  { value: 'openai', label: 'OpenAI' },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'azure', label: 'Azure OpenAI' },
  { value: 'custom', label: 'Custom' }
]

const enterActions = [
  { value: 'send', label: '发送消息' },
  { value: 'newline', label: '换行' }
]

const languages = [
  { value: 'zh-CN', label: '简体中文' },
  { value: 'en-US', label: 'English' }
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

// Watch for host changes to sync SSH host (only if not manually modified)
watch(() => formData.value.host, (newHost) => {
  if (!sshHostManuallyModified.value) {
    formData.value.ssh_host = newHost
  }
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
        connect_type: conn.connect_type || 'service_name',
        auth_type: conn.auth_type || 'sql',
        ssh_host: conn.host, // Will be overwritten if SSH was configured
        ssh_port: conn.ssh_port || 22,
        ssh_username: conn.ssh_username || '',
        ssh_password: conn.ssh_password || '',
        ai_assistants: conn.ai_assistants || formData.value.ai_assistants,
        selectedAssistantId: conn.ai_assistants?.[0]?.id || 'default'
      }

      // If SSH was configured with different host, use it
      if (conn.ssh_enabled && conn.ssh_username) {
        formData.value.ssh_host = conn.host
        sshHostManuallyModified.value = false
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
  const errors = {}

  switch (field) {
    case 'name':
      if (!formData.value.name.trim()) {
        errors.name = '连接名称不能为空'
      }
      break
    case 'host':
      if (!formData.value.host.trim()) {
        errors.host = '主机地址不能为空'
      }
      break
    case 'port':
      if (!formData.value.port || formData.value.port < 1 || formData.value.port > 65535) {
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
      if (currentSchema.value.databaseRequired && !formData.value.database.trim()) {
        errors.database = `${currentSchema.value.databaseLabel}不能为空`
      }
      break
  }

  fieldErrors.value = { ...fieldErrors.value, ...errors }
  return Object.keys(errors).length === 0
}

const validateForm = () => {
  fieldErrors.value = {}
  let isValid = true

  if (!validateField('name')) isValid = false
  if (!validateField('host')) isValid = false
  if (!validateField('port')) isValid = false
  if (!validateField('username')) isValid = false
  if (!validateField('database')) isValid = false

  return isValid
}

// ============================================================
// Actions
// ============================================================
const handleSave = async () => {
  formError.value = null

  if (!validateForm()) {
    formError.value = '请修正表单中的错误'
    return
  }

  saving.value = true

  try {
    const payload = {
      ...formData.value,
      // Map Oracle connect_type to appropriate field
      service_name: formData.value.type === 'oracle' && formData.value.connect_type === 'service_name'
        ? formData.value.database : '',
      sid: formData.value.type === 'oracle' && formData.value.connect_type === 'sid'
        ? formData.value.database : '',
      // SSH configuration
      ssh_enabled: !!(formData.value.ssh_username && formData.value.ssh_host),
      ssh_port: formData.value.ssh_port || 22,
      ssh_username: formData.value.ssh_username || '',
      ssh_password: formData.value.ssh_password || ''
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
    connect_type: 'service_name',
    auth_type: 'sql',
    ssh_host: '',
    ssh_port: 22,
    ssh_username: '',
    ssh_password: '',
    ai_assistants: [
      {
        id: 'default',
        name: 'DeepSeek',
        provider: 'deepseek',
        api_host: 'https://api.deepseek.com',
        api_endpoint: '/v1/chat/completions',
        api_key: '',
        model: 'deepseek-chat',
        temperature: 0.7,
        description: '',
        enter_action: 'send',
        compare_with_others: false,
        language: 'zh-CN'
      }
    ],
    selectedAssistantId: 'default'
  }
  showTypeSelection.value = true
  selectedType.value = ''
  formError.value = null
  fieldErrors.value = {}
  dbTestResult.value = null
  dbTestStatus.value = 'idle'
  sshTestResult.value = null
  sshTestStatus.value = 'idle'
  aiTestResult.value = null
  aiTestStatus.value = 'idle'
  sshHostManuallyModified.value = false
}

// ============================================================
// DB Test - Direct connection test ONLY
// ============================================================
const handleTestDB = async () => {
  if (!formData.value.host || !formData.value.username) {
    dbTestStatus.value = 'error'
    dbTestResult.value = {
      success: false,
      error: '请填写主机和用户名'
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
      testRequest.connect_type = formData.value.connect_type || 'service_name'
      testRequest.sid = formData.value.connect_type === 'sid' ? formData.value.database : ''
      testRequest.service_name = formData.value.connect_type === 'service_name' ? formData.value.database : ''
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
  if (!formData.value.ssh_host || !formData.value.ssh_username) {
    sshTestStatus.value = 'error'
    sshTestResult.value = {
      success: false,
      error: '请填写 SSH 主机和用户名'
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

// ============================================================
// AI Test - AI API connection test ONLY
// ============================================================
const handleTestAI = async () => {
  const assistant = selectedAssistant.value
  if (!assistant.api_host || !assistant.api_key) {
    aiTestStatus.value = 'error'
    aiTestResult.value = {
      success: false,
      error: '请填写 API 主机和 API 密钥'
    }
    return
  }

  aiTesting.value = true
  aiTestStatus.value = 'testing'
  aiTestResult.value = null

  try {
    // Build AI test request
    const testRequest = {
      provider: assistant.provider || 'deepseek',
      api_host: assistant.api_host,
      api_endpoint: assistant.api_endpoint || '/v1/chat/completions',
      api_key: assistant.api_key,
      model: assistant.model || 'deepseek-chat'
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
const addAssistant = () => {
  const newId = `assistant_${Date.now()}`
  formData.value.ai_assistants.push({
    id: newId,
    name: '新助手',
    provider: 'openai',
    api_host: 'https://api.openai.com',
    api_endpoint: '/v1/chat/completions',
    api_key: '',
    model: 'gpt-4',
    temperature: 0.7,
    description: '',
    enter_action: 'send',
    compare_with_others: false,
    language: 'zh-CN'
  })
  formData.value.selectedAssistantId = newId
}

const removeAssistant = (id) => {
  if (formData.value.ai_assistants.length <= 1) {
    return // Keep at least one assistant
  }
  const index = formData.value.ai_assistants.findIndex(a => a.id === id)
  if (index > -1) {
    formData.value.ai_assistants.splice(index, 1)
    if (formData.value.selectedAssistantId === id) {
      formData.value.selectedAssistantId = formData.value.ai_assistants[0].id
    }
  }
}

const selectAssistant = (id) => {
  formData.value.selectedAssistantId = id
}

const toggleApiKeyVisibility = (id) => {
  showApiKey.value[id] = !showApiKey.value[id]
}

// ============================================================
// SSH Host sync control
// ============================================================
const onSshHostChange = () => {
  sshHostManuallyModified.value = true
}

const syncSshHostFromDb = () => {
  formData.value.ssh_host = formData.value.host
  sshHostManuallyModified.value = false
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
            <div class="conn-form__row conn-form__row--inline">
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
            <div v-if="fieldErrors.host" class="conn-form__error-text">{{ fieldErrors.host }}</div>
            <div v-if="fieldErrors.port" class="conn-form__error-text">{{ fieldErrors.port }}</div>

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
              <select v-model="formData.connect_type" class="conn-form__select">
                <option value="service_name">Service Name</option>
                <option value="sid">SID</option>
              </select>
            </div>

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
            <div v-if="currentSchema.showDatabase" class="conn-form__row">
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
            <div v-if="fieldErrors.database" class="conn-form__error-text">{{ fieldErrors.database }}</div>

            <!-- Test Result -->
            <div v-if="dbTestResult" class="conn-form__test-result" :class="dbTestResult.success ? 'conn-form__test-result--success' : 'conn-form__test-result--error'">
              <span v-if="dbTestResult.success">连接成功 ({{ dbTestResult.latency_ms }}ms)</span>
              <span v-else>{{ dbTestResult.error }}</span>
            </div>
          </div>
        </div>

        <!-- ==================== SSH Tab ==================== -->
        <div v-show="activeTab === 'ssh'" class="conn-editor__panel">
          <div class="conn-form">
            <div class="conn-form__row conn-form__row--inline">
              <div class="conn-form__field">
                <label class="conn-form__label">SSH 主机</label>
                <input
                  v-model="formData.ssh_host"
                  type="text"
                  class="conn-form__input"
                  placeholder="SSH 服务器地址"
                  @input="onSshHostChange"
                />
              </div>
              <button type="button" class="conn-form__sync-btn" @click="syncSshHostFromDb" title="从数据库主机同步">
                同步
              </button>
            </div>
            <div class="conn-form__hint">默认继承自主机字段，可手动修改</div>

            <div class="conn-form__row conn-form__row--inline">
              <div class="conn-form__field">
                <label class="conn-form__label">SSH 端口</label>
                <input
                  v-model.number="formData.ssh_port"
                  type="number"
                  class="conn-form__input"
                  placeholder="22"
                  min="1"
                  max="65535"
                />
              </div>
            </div>

            <div class="conn-form__row">
              <label class="conn-form__label">SSH 用户名</label>
              <input
                v-model="formData.ssh_username"
                type="text"
                class="conn-form__input"
                placeholder="SSH 用户名"
              />
            </div>

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

            <!-- Test Result -->
            <div v-if="sshTestResult" class="conn-form__test-result" :class="sshTestResult.success ? 'conn-form__test-result--success' : 'conn-form__test-result--error'">
              <span v-if="sshTestResult.success">SSH 连接成功 ({{ sshTestResult.latency_ms }}ms)</span>
              <span v-else>{{ sshTestResult.error }}</span>
            </div>
          </div>
        </div>

        <!-- ==================== AI Assistant Tab ==================== -->
        <div v-show="activeTab === 'ai'" class="conn-editor__panel conn-editor__panel--ai">
          <div class="ai-assistant">
            <!-- Left: Assistant List -->
            <div class="ai-assistant__list">
              <div class="ai-assistant__list-header">
                <span>AI 助手列表</span>
                <button class="ai-assistant__add-btn" @click="addAssistant" title="添加助手">+</button>
              </div>
              <div class="ai-assistant__list-items">
                <div
                  v-for="assistant in formData.ai_assistants"
                  :key="assistant.id"
                  class="ai-assistant__item"
                  :class="{ 'ai-assistant__item--active': formData.selectedAssistantId === assistant.id }"
                  @click="selectAssistant(assistant.id)"
                >
                  <span class="ai-assistant__item-name">{{ assistant.name }}</span>
                  <button
                    v-if="formData.ai_assistants.length > 1"
                    class="ai-assistant__item-remove"
                    @click.stop="removeAssistant(assistant.id)"
                    title="删除助手"
                  >-</button>
                </div>
              </div>
            </div>

            <!-- Right: Assistant Config -->
            <div class="ai-assistant__config">
              <template v-if="selectedAssistant">
                <div class="conn-form">
                  <!-- Name -->
                  <div class="conn-form__row">
                    <label class="conn-form__label">AI 助手名称</label>
                    <input v-model="selectedAssistant.name" type="text" class="conn-form__input" placeholder="助手名称" />
                  </div>

                  <!-- Provider -->
                  <div class="conn-form__row">
                    <label class="conn-form__label">AI 提供商</label>
                    <select v-model="selectedAssistant.provider" class="conn-form__select">
                      <option v-for="p in aiProviders" :key="p.value" :value="p.value">{{ p.label }}</option>
                    </select>
                  </div>

                  <!-- API Host -->
                  <div class="conn-form__row">
                    <label class="conn-form__label">API 主机</label>
                    <input v-model="selectedAssistant.api_host" type="text" class="conn-form__input" placeholder="https://api.example.com" />
                  </div>

                  <!-- API Endpoint -->
                  <div class="conn-form__row">
                    <label class="conn-form__label">API 端点</label>
                    <input v-model="selectedAssistant.api_endpoint" type="text" class="conn-form__input" placeholder="/v1/chat/completions" />
                  </div>

                  <!-- API Key -->
                  <div class="conn-form__row">
                    <label class="conn-form__label">API 密钥</label>
                    <div class="conn-form__password">
                      <input
                        v-model="selectedAssistant.api_key"
                        :type="showApiKey[selectedAssistant.id] ? 'text' : 'password'"
                        class="conn-form__input"
                        placeholder="sk-..."
                      />
                      <button type="button" class="conn-form__password-toggle" @click="toggleApiKeyVisibility(selectedAssistant.id)">
                        {{ showApiKey[selectedAssistant.id] ? '隐藏' : '显示' }}
                      </button>
                    </div>
                  </div>

                  <!-- Model -->
                  <div class="conn-form__row conn-form__row--inline">
                    <div class="conn-form__field">
                      <label class="conn-form__label">模型</label>
                      <input v-model="selectedAssistant.model" type="text" class="conn-form__input" placeholder="gpt-4" />
                    </div>
                    <button type="button" class="conn-form__model-btn" title="选择模型">...</button>
                  </div>

                  <!-- Temperature -->
                  <div class="conn-form__row">
                    <label class="conn-form__label">温度 (Temperature)</label>
                    <div class="conn-form__slider">
                      <input
                        v-model.number="selectedAssistant.temperature"
                        type="range"
                        min="0"
                        max="2"
                        step="0.1"
                        class="conn-form__range"
                      />
                      <span class="conn-form__slider-value">{{ selectedAssistant.temperature }}</span>
                    </div>
                  </div>

                  <!-- Description -->
                  <div class="conn-form__row">
                    <label class="conn-form__label">说明</label>
                    <textarea
                      v-model="selectedAssistant.description"
                      class="conn-form__textarea"
                      placeholder="助手描述（可选）"
                      rows="2"
                    ></textarea>
                  </div>

                  <!-- AI UI Settings -->
                  <div class="conn-form__section-title">AI 助手 UI</div>

                  <div class="conn-form__row">
                    <label class="conn-form__label">按下回车键时执行的操作</label>
                    <select v-model="selectedAssistant.enter_action" class="conn-form__select">
                      <option v-for="a in enterActions" :key="a.value" :value="a.value">{{ a.label }}</option>
                    </select>
                  </div>

                  <div class="conn-form__row conn-form__row--checkbox">
                    <label class="conn-form__checkbox">
                      <input type="checkbox" v-model="selectedAssistant.compare_with_others" />
                      <span>与其他助手比较</span>
                    </label>
                  </div>

                  <div class="conn-form__section-title">询问 AI</div>

                  <div class="conn-form__row">
                    <label class="conn-form__label">语言</label>
                    <select v-model="selectedAssistant.language" class="conn-form__select">
                      <option v-for="l in languages" :key="l.value" :value="l.value">{{ l.label }}</option>
                    </select>
                  </div>

                  <!-- Test Button -->
                  <div class="conn-form__row conn-form__row--test">
                    <button
                      type="button"
                      class="conn-form__test-btn"
                      :class="{
                        'conn-form__test-btn--testing': aiTestStatus === 'testing',
                        'conn-form__test-btn--success': aiTestStatus === 'success',
                        'conn-form__test-btn--error': aiTestStatus === 'error'
                      }"
                      :disabled="aiTesting"
                      @click="handleTestAI"
                    >
                      <span v-if="aiTestStatus === 'testing'" class="spinner"></span>
                      <span v-else-if="aiTestStatus === 'success'" class="icon-success">✓</span>
                      <span v-else-if="aiTestStatus === 'error'" class="icon-error">✗</span>
                      <span>{{ aiTesting ? '测试中...' : '测试连接' }}</span>
                    </button>
                  </div>

                  <!-- AI Test Result -->
                  <div v-if="aiTestResult" class="conn-form__test-result" :class="aiTestResult.success ? 'conn-form__test-result--success' : 'conn-form__test-result--error'">
                    <span v-if="aiTestResult.success">{{ aiTestResult.message || '连接成功' }}</span>
                    <span v-else>{{ aiTestResult.error }}</span>
                  </div>
                </div>
              </template>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Footer Actions -->
    <div class="conn-editor__footer">
      <div class="conn-editor__footer-left">
        <!-- General tab: Test DB button -->
        <button
          v-if="activeTab === 'general'"
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

        <!-- SSH tab: Test SSH button -->
        <button
          v-if="activeTab === 'ssh'"
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

/* AI Assistant Layout */
.ai-assistant {
  display: flex;
  min-height: 400px;
}

.ai-assistant__list {
  width: 180px;
  border-right: 1px solid var(--border-color);
  background-color: var(--bg-secondary);
  display: flex;
  flex-direction: column;
}

.ai-assistant__list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--border-color);
}

.ai-assistant__add-btn {
  width: 24px;
  height: 24px;
  background-color: var(--primary);
  border: none;
  border-radius: 4px;
  color: white;
  font-size: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ai-assistant__add-btn:hover {
  background-color: var(--primary-hover);
}

.ai-assistant__list-items {
  flex: 1;
  overflow-y: auto;
}

.ai-assistant__item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.ai-assistant__item:hover {
  background-color: #e9ecf0;
}

.ai-assistant__item--active {
  background-color: #ecf5ff;
  border-left: 3px solid var(--primary);
}

.ai-assistant__item-name {
  font-size: 13px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ai-assistant__item-remove {
  width: 20px;
  height: 20px;
  background-color: transparent;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  color: var(--text-muted);
  font-size: 14px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ai-assistant__item-remove:hover {
  background-color: var(--error-bg);
  border-color: var(--error);
  color: var(--error);
}

.ai-assistant__config {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
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
