<script setup>
/**
 * ConnectionForm.vue
 * Form for creating and editing database connections.
 * Supports MySQL, PostgreSQL, Oracle, and SQL Server with database-specific fields.
 */
import { ref, computed, watch, nextTick } from 'vue'
import { useConnectionStore } from '../../stores/connection'
import { TestConnectionDirect, TestSSHConnection } from '../../../wailsjs/go/bindings/ConnectionBinding'

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
    defaultPort: 3306,
    defaultUsername: 'root',
    defaultDatabase: '',
    databaseLabel: 'Database Name',
    databaseRequired: false,
    databasePlaceholder: 'mydb (optional)',
    databaseHelper: 'If omitted, connection will be established without selecting a default schema.',
    supportsSSH: true,
    supportsWinRM: false,
    supportsAuthType: false,
    supportsConnectAs: false,
    supportsConnectionString: false,
    connectTypeLabel: null,
    connectTypeOptions: null,
    authTypeLabel: null,
    authTypeOptions: null,
    connectAsLabel: null,
    connectAsOptions: null
  },
  postgresql: {
    label: 'PostgreSQL',
    defaultPort: 5432,
    defaultUsername: 'postgres',
    defaultDatabase: 'postgres',
    databaseLabel: 'Database Name',
    databaseRequired: true,
    databasePlaceholder: 'postgres',
    databaseHelper: 'Required for PostgreSQL connections.',
    supportsSSH: true,
    supportsWinRM: false,
    supportsAuthType: false,
    supportsConnectAs: false,
    supportsConnectionString: false,
    connectTypeLabel: null,
    connectTypeOptions: null,
    authTypeLabel: null,
    authTypeOptions: null,
    connectAsLabel: null,
    connectAsOptions: null
  },
  oracle: {
    label: 'Oracle',
    defaultPort: 1521,
    defaultUsername: 'system',
    defaultDatabase: '',
    databaseLabel: 'Service Name / SID',
    databaseRequired: true,
    databasePlaceholder: 'ORCL',
    databaseHelper: 'Enter either Service Name or SID. Prefer Service Name for modern Oracle deployments.',
    supportsSSH: true,
    supportsWinRM: false,
    supportsAuthType: false,
    supportsConnectAs: true,
    supportsConnectionString: true,
    connectTypeLabel: 'Connection Type',
    connectTypeOptions: [
      { value: 'service_name', label: 'Service Name' },
      { value: 'sid', label: 'SID' },
      { value: 'connection_string', label: 'Connection String' }
    ],
    authTypeLabel: null,
    authTypeOptions: null,
    connectAsLabel: 'Connect As',
    connectAsOptions: [
      { value: 'normal', label: 'Normal' },
      { value: 'sysdba', label: 'SYSDBA' },
      { value: 'sysoper', label: 'SYSOPER' }
    ]
  },
  sqlserver: {
    label: 'SQL Server',
    defaultPort: 1433,
    defaultUsername: 'sa',
    defaultDatabase: '',
    databaseLabel: 'Database Name',
    databaseRequired: false,
    databasePlaceholder: 'mydb (optional)',
    databaseHelper: 'If omitted, connection will be established to default database.',
    supportsSSH: false,
    supportsWinRM: true,
    supportsAuthType: true,
    supportsConnectAs: false,
    supportsConnectionString: false,
    connectTypeLabel: null,
    connectTypeOptions: null,
    authTypeLabel: 'Authentication Type',
    authTypeOptions: [
      { value: 'sql', label: 'SQL Server Authentication' },
      { value: 'windows', label: 'Windows Authentication' }
    ],
    connectAsLabel: null,
    connectAsOptions: null
  }
}

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
  // Connection type (Oracle)
  connect_type: 'service_name',
  connection_string: '',
  // Auth type (SQL Server)
  auth_type: 'sql',
  // Connect As (Oracle)
  connect_as: 'normal',
  // SSH Configuration
  ssh_enabled: false,
  ssh_port: 22,
  ssh_username: '',
  ssh_password: '',
  // WinRM Configuration
  winrm_enabled: false,
  winrm_port: 5985,
  winrm_use_https: false,
  winrm_username: '',
  winrm_password: ''
})

// UI State
const showPassword = ref(false)
const showSshPassword = ref(false)
const showWinrmPassword = ref(false)
const saving = ref(false)
const testing = ref(false)
const formError = ref(null)
const fieldErrors = ref({})
const testResult = ref(null)
const testStatus = ref('idle')
const sshTesting = ref(false)
const sshTestResult = ref(null)
const sshTestStatus = ref('idle')

// ============================================================
// Computed Properties
// ============================================================
const isEditing = computed(() => props.mode === 'edit' && props.connectionId)
const title = computed(() => isEditing.value ? 'Edit Connection' : 'New Connection')
const currentSchema = computed(() => DB_SCHEMA[formData.value.type])
const canTest = computed(() => formData.value.host && formData.value.username)

// Show/hide fields based on connection type
const showConnectType = computed(() => currentSchema.value?.connectTypeOptions?.length > 0)
const showConnectionString = computed(() => formData.value.type === 'oracle' && formData.value.connect_type === 'connection_string')
const showAuthType = computed(() => currentSchema.value?.supportsAuthType)
const showConnectAs = computed(() => currentSchema.value?.supportsConnectAs)
const showUsernamePassword = computed(() => {
  if (formData.value.type === 'sqlserver' && formData.value.auth_type === 'windows') {
    return false
  }
  return true
})

// Database type options
const typeOptions = [
  { value: 'mysql', label: 'MySQL', icon: '🐬' },
  { value: 'postgresql', label: 'PostgreSQL', icon: '🐘' },
  { value: 'oracle', label: 'Oracle', icon: '🔴' },
  { value: 'sqlserver', label: 'SQL Server', icon: '🔷' }
]

// ============================================================
// Watchers
// ============================================================
// Flag to prevent type watch from resetting during edit mode data load
const isLoadingEditData = ref(false)

watch(() => formData.value.type, (newType) => {
  // Skip reset when loading edit data - the connectionId watch will set all values
  if (isLoadingEditData.value) {
    return
  }
  const schema = DB_SCHEMA[newType]
  formData.value.port = schema.defaultPort
  formData.value.username = schema.defaultUsername
  formData.value.database = schema.defaultDatabase
  // Reset fields
  formData.value.ssh_enabled = false
  formData.value.winrm_enabled = false
  formData.value.connect_type = 'service_name'
  formData.value.connection_string = ''
  formData.value.auth_type = 'sql'
  formData.value.connect_as = 'normal'
  // Clear test result
  testResult.value = null
  testStatus.value = 'idle'
  fieldErrors.value = {}
})

watch(() => props.connectionId, async (newId) => {
  if (newId && props.mode === 'edit') {
    const conn = connectionStore.connections.find(c => c.id === newId)
    if (conn) {
      // Set flag to prevent type watch from resetting values
      isLoadingEditData.value = true
      formData.value = {
        name: conn.name,
        type: conn.type,
        host: conn.host,
        port: conn.port,
        database: conn.database || '',
        username: conn.username,
        password: conn.password || '',  // Keep saved password for edit
        connect_type: conn.connect_type || 'service_name',
        connection_string: conn.connection_string || '',
        auth_type: conn.auth_type || 'sql',
        connect_as: conn.connect_as || 'normal',
        ssh_enabled: conn.ssh_enabled || false,
        ssh_port: conn.ssh_port || 22,
        ssh_username: conn.ssh_username || '',
        ssh_password: conn.ssh_password || '',  // Keep saved SSH password
        winrm_enabled: conn.winrm_enabled || false,
        winrm_port: conn.winrm_port || 5985,
        winrm_use_https: conn.winrm_use_https || false,
        winrm_username: conn.winrm_username || '',
        winrm_password: conn.winrm_password || ''  // Keep saved WinRM password
      }
      // Clear flag in next tick to ensure type watch has completed
      // IMPORTANT: Must use nextTick because Vue batches updates,
      // and the type watch might trigger again in the same tick
      nextTick(() => {
        isLoadingEditData.value = false
      })
    }
  }
}, { immediate: true })

// ============================================================
// Validation
// ============================================================
const validateField = (field) => {
  const errors = {}
  const schema = currentSchema.value

  switch (field) {
    case 'name':
      if (!formData.value.name.trim()) {
        errors.name = 'Connection name is required'
      }
      break
    case 'host':
      if (!formData.value.host.trim()) {
        errors.host = 'Host is required'
      }
      break
    case 'port':
      if (!formData.value.port || formData.value.port < 1 || formData.value.port > 65535) {
        errors.port = 'Port must be between 1 and 65535'
      }
      break
    case 'username':
      if (showUsernamePassword.value && !formData.value.username.trim()) {
        errors.username = 'Username is required'
      }
      break
    case 'database':
      if (schema.databaseRequired && !formData.value.database.trim()) {
        errors.database = `${schema.databaseLabel} is required`
      }
      break
    case 'connection_string':
      if (formData.value.type === 'oracle' && formData.value.connect_type === 'connection_string' && !formData.value.connection_string.trim()) {
        errors.connection_string = 'Connection string is required'
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
  if (!validateField('connection_string')) isValid = false

  return isValid
}

// ============================================================
// Actions
// ============================================================
const handleSave = async () => {
  formError.value = null

  if (!validateForm()) {
    formError.value = 'Please fix the validation errors before saving.'
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
        ? formData.value.database : ''
    }

    if (isEditing.value) {
      const updated = await connectionStore.updateConnection({
        id: props.connectionId,
        ...payload
      })
      if (updated) {
        emit('saved', updated)
      } else {
        formError.value = connectionStore.error || 'Failed to update connection'
      }
    } else {
      const created = await connectionStore.createConnection(payload)
      if (created) {
        emit('saved', created)
        resetForm()
      } else {
        formError.value = connectionStore.error || 'Failed to create connection'
      }
    }
  } finally {
    saving.value = false
  }
}

const handleTestConnection = async () => {
  formError.value = null
  testResult.value = null
  testStatus.value = 'testing'

  // Validate required fields first
  if (!formData.value.host || !formData.value.username) {
    testStatus.value = 'error'
    testResult.value = {
      success: false,
      error: 'Please fill in Host and Username before testing'
    }
    return
  }

  testing.value = true

  try {
    // Build test request with type-specific fields
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

    const result = await TestConnectionDirect(testRequest)

    testResult.value = result
    testStatus.value = result.success ? 'success' : 'error'
    emit('tested', result)
  } catch (err) {
    testStatus.value = 'error'
    testResult.value = {
      success: false,
      error: err.message || 'Test failed'
    }
  } finally {
    testing.value = false
  }
}

// SSH Connection Test
const handleTestSSH = async () => {
  if (!formData.value.host || !formData.value.ssh_username) {
    sshTestStatus.value = 'error'
    sshTestResult.value = {
      success: false,
      error: 'Please fill in Host and SSH Username before testing'
    }
    return
  }

  sshTesting.value = true
  sshTestStatus.value = 'testing'
  sshTestResult.value = null

  try {
    const result = await TestSSHConnection({
      host: formData.value.host,
      port: formData.value.ssh_port || 22,
      username: formData.value.ssh_username,
      password: formData.value.ssh_password || ''
    })

    sshTestResult.value = result
    sshTestStatus.value = result.success ? 'success' : 'error'
  } catch (err) {
    sshTestStatus.value = 'error'
    sshTestResult.value = {
      success: false,
      error: err.message || 'SSH test failed'
    }
  } finally {
    sshTesting.value = false
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
    connection_string: '',
    auth_type: 'sql',
    connect_as: 'normal',
    ssh_enabled: false,
    ssh_port: 22,
    ssh_username: '',
    ssh_password: '',
    winrm_enabled: false,
    winrm_port: 5985,
    winrm_use_https: false,
    winrm_username: '',
    winrm_password: ''
  }
  formError.value = null
  fieldErrors.value = {}
  testResult.value = null
  testStatus.value = 'idle'
  sshTestResult.value = null
  sshTestStatus.value = 'idle'
  showPassword.value = false
  showSshPassword.value = false
  showWinrmPassword.value = false
}
</script>

<template>
  <div class="conn-form">
    <!-- Header -->
    <div class="conn-form__header">
      <h2 class="conn-form__title">{{ title }}</h2>
      <p class="conn-form__subtitle">
        Configure database access and test connectivity before saving.
      </p>
    </div>

    <!-- Error Banner -->
    <div v-if="formError" class="conn-form__error-banner">
      <span class="conn-form__error-icon">⚠</span>
      {{ formError }}
    </div>

    <!-- Form Body -->
    <div class="conn-form__body">
      <!-- Section: Basic Information -->
      <section class="conn-form__section">
        <h3 class="conn-form__section-title">
          <span class="conn-form__section-icon">1</span>
          Basic Information
        </h3>
        <div class="conn-form__grid conn-form__grid--basic">
          <!-- Database Type -->
          <div class="conn-form__field">
            <label class="conn-form__label">
              Database Type <span class="conn-form__required">*</span>
            </label>
            <select v-model="formData.type" class="conn-form__select">
              <option v-for="opt in typeOptions" :key="opt.value" :value="opt.value">
                {{ opt.icon }} {{ opt.label }}
              </option>
            </select>
          </div>

          <!-- Connection Name -->
          <div class="conn-form__field">
            <label class="conn-form__label">
              Connection Name <span class="conn-form__required">*</span>
            </label>
            <input
              v-model="formData.name"
              type="text"
              class="conn-form__input"
              :class="{ 'conn-form__input--error': fieldErrors.name }"
              placeholder="My Database"
              @blur="validateField('name')"
            />
            <span v-if="fieldErrors.name" class="conn-form__field-error">{{ fieldErrors.name }}</span>
          </div>

          <!-- Host & Port -->
          <div class="conn-form__field conn-form__field--half">
            <label class="conn-form__label">
              Host <span class="conn-form__required">*</span>
            </label>
            <input
              v-model="formData.host"
              type="text"
              class="conn-form__input"
              :class="{ 'conn-form__input--error': fieldErrors.host }"
              placeholder="localhost or IP address"
              @blur="validateField('host')"
            />
            <span v-if="fieldErrors.host" class="conn-form__field-error">{{ fieldErrors.host }}</span>
          </div>

          <div class="conn-form__field conn-form__field--port">
            <label class="conn-form__label">
              Port <span class="conn-form__required">*</span>
            </label>
            <input
              v-model.number="formData.port"
              type="number"
              class="conn-form__input"
              :class="{ 'conn-form__input--error': fieldErrors.port }"
              min="1"
              max="65535"
              @blur="validateField('port')"
            />
            <span v-if="fieldErrors.port" class="conn-form__field-error">{{ fieldErrors.port }}</span>
          </div>
        </div>
      </section>

      <!-- Section: Authentication -->
      <section class="conn-form__section">
        <h3 class="conn-form__section-title">
          <span class="conn-form__section-icon">2</span>
          Authentication
        </h3>
        <div class="conn-form__grid conn-form__grid--auth">
          <!-- Auth Type (SQL Server) -->
          <div v-if="showAuthType" class="conn-form__field conn-form__field--full">
            <label class="conn-form__label">
              {{ currentSchema.authTypeLabel }} <span class="conn-form__required">*</span>
            </label>
            <select v-model="formData.auth_type" class="conn-form__select">
              <option v-for="opt in currentSchema.authTypeOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
            <span class="conn-form__helper">
              {{ formData.auth_type === 'windows' ? 'Uses Windows credentials for authentication.' : 'Uses SQL Server login credentials.' }}
            </span>
          </div>

          <!-- Username -->
          <div v-if="showUsernamePassword" class="conn-form__field">
            <label class="conn-form__label">
              Username <span class="conn-form__required">*</span>
            </label>
            <input
              v-model="formData.username"
              type="text"
              class="conn-form__input"
              :class="{ 'conn-form__input--error': fieldErrors.username }"
              :placeholder="currentSchema.defaultUsername"
              @blur="validateField('username')"
            />
            <span v-if="fieldErrors.username" class="conn-form__field-error">{{ fieldErrors.username }}</span>
          </div>

          <!-- Password -->
          <div v-if="showUsernamePassword" class="conn-form__field">
            <label class="conn-form__label">
              Password
              <span v-if="isEditing" class="conn-form__optional">(leave empty to keep current)</span>
            </label>
            <div class="conn-form__password-wrapper">
              <input
                v-model="formData.password"
                :type="showPassword ? 'text' : 'password'"
                class="conn-form__input"
                :placeholder="isEditing ? '••••••••' : 'Enter password'"
              />
              <button
                type="button"
                class="conn-form__password-toggle"
                @click="showPassword = !showPassword"
                :title="showPassword ? 'Hide password' : 'Show password'"
              >
                {{ showPassword ? '🙈' : '👁️' }}
              </button>
            </div>
          </div>

          <!-- Database/SID - Dynamic -->
          <div class="conn-form__field conn-form__field--full">
            <!-- Connect Type (Oracle) -->
            <div v-if="showConnectType" class="conn-form__field">
              <label class="conn-form__label">
                {{ currentSchema.connectTypeLabel }} <span class="conn-form__required">*</span>
              </label>
              <select v-model="formData.connect_type" class="conn-form__select">
                <option v-for="opt in currentSchema.connectTypeOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
            </div>

            <!-- Connection String (Oracle) -->
            <div v-if="showConnectionString" class="conn-form__field">
              <label class="conn-form__label">
                Connection String <span class="conn-form__required">*</span>
              </label>
              <input
                v-model="formData.connection_string"
                type="text"
                class="conn-form__input"
                :class="{ 'conn-form__input--error': fieldErrors.connection_string }"
                placeholder="host:port/service_name or (host:port)/service_name"
                @blur="validateField('connection_string')"
              />
              <span v-if="fieldErrors.connection_string" class="conn-form__field-error">{{ fieldErrors.connection_string }}</span>
              <span class="conn-form__helper">Full Oracle connection string (EZCONNECT or TNS format)</span>
            </div>

            <!-- Database Name / Service Name / SID -->
            <div v-if="!showConnectionString" class="conn-form__field">
              <label class="conn-form__label">
                {{ currentSchema.databaseLabel }}
                <span v-if="!currentSchema.databaseRequired" class="conn-form__optional">(optional)</span>
                <span v-else class="conn-form__required">*</span>
              </label>
              <input
                v-model="formData.database"
                type="text"
                class="conn-form__input"
                :class="{ 'conn-form__input--error': fieldErrors.database }"
                :placeholder="currentSchema.databasePlaceholder"
                @blur="validateField('database')"
              />
              <span class="conn-form__helper">{{ currentSchema.databaseHelper }}</span>
              <span v-if="fieldErrors.database" class="conn-form__field-error">{{ fieldErrors.database }}</span>
            </div>

            <!-- Connect As (Oracle) -->
            <div v-if="showConnectAs" class="conn-form__field" style="margin-top: 12px;">
              <label class="conn-form__label">
                {{ currentSchema.connectAsLabel }}
                <span class="conn-form__optional">(optional)</span>
              </label>
              <select v-model="formData.connect_as" class="conn-form__select">
                <option v-for="opt in currentSchema.connectAsOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
              <span class="conn-form__helper">
                SYSDBA/SYSOPER privileges are required for administrative tasks.
              </span>
            </div>
          </div>
        </div>
      </section>

      <!-- Section: Advanced Options -->
      <section class="conn-form__section conn-form__section--collapsible">
        <h3 class="conn-form__section-title conn-form__section-title--collapsible">
          <span class="conn-form__section-icon">3</span>
          Advanced Options
          <span class="conn-form__section-hint">(SSH Tunnel, WinRM)</span>
        </h3>

        <!-- SSH Tunnel (MySQL, PostgreSQL, Oracle) -->
        <div v-if="currentSchema.supportsSSH" class="conn-form__advanced">
          <label class="conn-form__checkbox">
            <input
              type="checkbox"
              v-model="formData.ssh_enabled"
              class="conn-form__checkbox-input"
            />
            <span class="conn-form__checkbox-label">Enable SSH Tunnel</span>
          </label>

          <div v-if="formData.ssh_enabled" class="conn-form__advanced-fields">
            <div class="conn-form__grid conn-form__grid--ssh">
              <div class="conn-form__field conn-form__field--port">
                <label class="conn-form__label">SSH Port</label>
                <input
                  v-model.number="formData.ssh_port"
                  type="number"
                  class="conn-form__input"
                  placeholder="22"
                  min="1"
                  max="65535"
                />
              </div>
              <div class="conn-form__field">
                <label class="conn-form__label">SSH Username</label>
                <input
                  v-model="formData.ssh_username"
                  type="text"
                  class="conn-form__input"
                  placeholder="root"
                />
              </div>
              <div class="conn-form__field">
                <label class="conn-form__label">SSH Password</label>
                <div class="conn-form__password-wrapper">
                  <input
                    v-model="formData.ssh_password"
                    :type="showSshPassword ? 'text' : 'password'"
                    class="conn-form__input"
                    placeholder="SSH password"
                  />
                  <button
                    type="button"
                    class="conn-form__password-toggle"
                    @click="showSshPassword = !showSshPassword"
                  >
                    {{ showSshPassword ? '🙈' : '👁️' }}
                  </button>
                </div>
              </div>
            </div>
            <div class="conn-form__helper conn-form__helper--indent">
              SSH Host: {{ formData.host || 'N/A' }}
            </div>
          </div>
        </div>

        <!-- WinRM (SQL Server only) -->
        <div v-if="currentSchema.supportsWinRM" class="conn-form__advanced">
          <label class="conn-form__checkbox">
            <input
              type="checkbox"
              v-model="formData.winrm_enabled"
              class="conn-form__checkbox-input"
            />
            <span class="conn-form__checkbox-label">Enable WinRM</span>
          </label>

          <div v-if="formData.winrm_enabled" class="conn-form__advanced-fields">
            <div class="conn-form__grid conn-form__grid--ssh">
              <div class="conn-form__field conn-form__field--port">
                <label class="conn-form__label">WinRM Port</label>
                <input
                  v-model.number="formData.winrm_port"
                  type="number"
                  class="conn-form__input"
                  :placeholder="formData.winrm_use_https ? '5986' : '5985'"
                  min="1"
                  max="65535"
                />
              </div>
              <div class="conn-form__field">
                <label class="conn-form__label">WinRM Username</label>
                <input
                  v-model="formData.winrm_username"
                  type="text"
                  class="conn-form__input"
                  placeholder="Administrator"
                />
              </div>
              <div class="conn-form__field">
                <label class="conn-form__label">WinRM Password</label>
                <div class="conn-form__password-wrapper">
                  <input
                    v-model="formData.winrm_password"
                    :type="showWinrmPassword ? 'text' : 'password'"
                    class="conn-form__input"
                    placeholder="WinRM password"
                  />
                  <button
                    type="button"
                    class="conn-form__password-toggle"
                    @click="showWinrmPassword = !showWinrmPassword"
                  >
                    {{ showWinrmPassword ? '🙈' : '👁️' }}
                  </button>
                </div>
              </div>
            </div>
            <label class="conn-form__checkbox conn-form__checkbox--sub">
              <input
                type="checkbox"
                v-model="formData.winrm_use_https"
                class="conn-form__checkbox-input"
              />
              <span class="conn-form__checkbox-label">Use HTTPS (requires port 5986)</span>
            </label>
            <div class="conn-form__helper conn-form__helper--indent">
              WinRM Host: {{ formData.host || 'N/A' }}
            </div>
          </div>
        </div>
      </section>

      <!-- Section: Connection Test -->
      <section class="conn-form__section conn-form__section--test">
        <h3 class="conn-form__section-title">
          <span class="conn-form__section-icon">4</span>
          Connection Test
        </h3>

        <div class="conn-form__test">
          <button
            class="conn-form__test-btn"
            :class="{
              'conn-form__test-btn--testing': testStatus === 'testing',
              'conn-form__test-btn--success': testStatus === 'success',
              'conn-form__test-btn--error': testStatus === 'error'
            }"
            @click="handleTestConnection"
            :disabled="testing || saving || !canTest"
          >
            <span v-if="testStatus === 'testing'" class="conn-form__spinner"></span>
            <span v-else-if="testStatus === 'success'" class="conn-form__test-icon">✓</span>
            <span v-else-if="testStatus === 'error'" class="conn-form__test-icon">✗</span>
            <span>{{ testing ? ' Testing...' : 'Test Connection' }}</span>
          </button>

          <div v-if="testResult" class="conn-form__test-result" :class="testResult.success ? 'conn-form__test-result--success' : 'conn-form__test-result--error'">
            <div class="conn-form__test-status">
              {{ testResult.success ? 'Connection successful' : 'Connection failed' }}
            </div>
            <div v-if="testResult.success" class="conn-form__test-details">
              <span>Latency: {{ testResult.latency_ms }}ms</span>
            </div>
            <div v-if="testResult.error" class="conn-form__test-error">
              {{ testResult.error }}
            </div>
          </div>
        </div>

        <!-- SSH Test Button -->
        <div v-if="formData.ssh_enabled" class="conn-form__test" style="margin-top: 12px;">
          <button
            class="conn-form__test-btn"
            :class="{
              'conn-form__test-btn--testing': sshTestStatus === 'testing',
              'conn-form__test-btn--success': sshTestStatus === 'success',
              'conn-form__test-btn--error': sshTestStatus === 'error'
            }"
            @click="handleTestSSH"
            :disabled="sshTesting || saving || !formData.ssh_username"
          >
            <span v-if="sshTestStatus === 'testing'" class="conn-form__spinner"></span>
            <span v-else-if="sshTestStatus === 'success'" class="conn-form__test-icon">✓</span>
            <span v-else-if="sshTestStatus === 'error'" class="conn-form__test-icon">✗</span>
            <span>{{ sshTesting ? ' Testing SSH...' : 'Test SSH Tunnel' }}</span>
          </button>

          <div v-if="sshTestResult" class="conn-form__test-result" :class="sshTestResult.success ? 'conn-form__test-result--success' : 'conn-form__test-result--error'">
            <div class="conn-form__test-status">
              {{ sshTestResult.success ? 'SSH connection successful' : 'SSH connection failed' }}
            </div>
            <div v-if="sshTestResult.error" class="conn-form__test-error">
              {{ sshTestResult.error }}
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- Footer Actions -->
    <div class="conn-form__footer">
      <button
        class="conn-form__btn conn-form__btn--cancel"
        @click="handleCancel"
        :disabled="saving"
      >
        Cancel
      </button>
      <button
        class="conn-form__btn conn-form__btn--save"
        @click="handleSave"
        :disabled="saving"
      >
        {{ saving ? 'Saving...' : 'Save' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
/* ============================================================
   Base Form Styles
   ============================================================ */
.conn-form {
  --form-bg: #1e2a3a;
  --form-bg-secondary: #252f3f;
  --form-border: #3a4a5a;
  --form-border-focus: #4299e1;
  --form-text: #e2e8f0;
  --form-text-secondary: #a0aec0;
  --form-text-muted: #718096;
  --form-success: #48bb78;
  --form-error: #f44336;
  --form-error-bg: rgba(244, 67, 54, 0.1);
  --form-success-bg: rgba(72, 187, 120, 0.1);

  background-color: var(--form-bg);
  border-radius: 8px;
  padding: 0;
  width: 100%;
  min-width: 0;
  max-width: 100%;
  display: flex;
  flex-direction: column;
  max-height: 90vh; /* 限制最大高度 */
}

/* ============================================================
   Header
   ============================================================ */
.conn-form__header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--form-border);
  flex-shrink: 0; /* 防止被压缩 */
}

.conn-form__title {
  font-size: 18px;
  font-weight: 600;
  color: var(--form-text);
  margin: 0 0 4px 0;
}

.conn-form__subtitle {
  font-size: 13px;
  color: var(--form-text-muted);
  margin: 0;
}

/* ============================================================
   Error Banner
   ============================================================ */
.conn-form__error-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  background-color: var(--form-error-bg);
  border-bottom: 1px solid var(--form-error);
  color: var(--form-error);
  font-size: 13px;
}

.conn-form__error-icon {
  font-size: 16px;
}

/* ============================================================
   Body - Scrollable Container
   ============================================================ */
.conn-form__body {
  padding: 20px 24px;
  overflow-y: auto;
  overflow-x: hidden;
  max-height: calc(90vh - 160px); /* 减去 header + footer 高度 */
  flex: 1;
  min-height: 0; /* 关键：允许 flex 子项收缩 */
}

/* ============================================================
   Sections
   ============================================================ */
.conn-form__section {
  margin-bottom: 24px;
}

.conn-form__section:last-child {
  margin-bottom: 0;
}

.conn-form__section-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  font-weight: 600;
  color: var(--form-text);
  margin: 0 0 16px 0;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--form-border);
}

.conn-form__section-title--collapsible {
  cursor: pointer;
}

.conn-form__section-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  background-color: var(--form-border);
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  color: var(--form-text-secondary);
}

.conn-form__section-hint {
  font-weight: 400;
  font-size: 12px;
  color: var(--form-text-muted);
  margin-left: auto;
}

/* ============================================================
   Grid Layout
   ============================================================ */
.conn-form__grid {
  display: grid;
  gap: 16px;
  width: 100%;
}

.conn-form__grid--basic {
  grid-template-columns: 1fr 1fr;
}

.conn-form__grid--auth {
  grid-template-columns: 1fr 1fr;
}

.conn-form__grid--ssh {
  grid-template-columns: 100px 1fr 1fr;
}

/* ============================================================
   Fields
   ============================================================ */
.conn-form__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.conn-form__field--half {
  flex: 1;
}

.conn-form__field--port {
  width: 100px;
  flex-shrink: 0;
}

.conn-form__field--full {
  grid-column: 1 / -1;
}

.conn-form__label {
  font-size: 12px;
  font-weight: 500;
  color: var(--form-text-secondary);
}

.conn-form__required {
  color: var(--form-error);
}

.conn-form__optional {
  font-weight: 400;
  color: var(--form-text-muted);
  font-style: italic;
}

.conn-form__helper {
  font-size: 11px;
  color: var(--form-text-muted);
  line-height: 1.4;
}

.conn-form__helper--indent {
  padding-left: 24px;
}

.conn-form__field-error {
  font-size: 11px;
  color: var(--form-error);
}

/* ============================================================
   Inputs
   ============================================================ */
.conn-form__input,
.conn-form__select {
  padding: 10px 12px;
  background-color: var(--form-bg-secondary);
  border: 1px solid var(--form-border);
  border-radius: 6px;
  color: var(--form-text);
  font-size: 13px;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.conn-form__input:focus,
.conn-form__select:focus {
  outline: none;
  border-color: var(--form-border-focus);
  box-shadow: 0 0 0 2px rgba(66, 153, 225, 0.2);
}

/* Select 暗色主题完整支持 */
.conn-form__select {
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%23a0aec0' d='M6 8L1 3h10z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 36px;
  cursor: pointer;
}

.conn-form__select:hover {
  background-color: #2a3a4a;
  border-color: #4a5a6a;
}

.conn-form__select option {
  background-color: #252f3f;
  color: var(--form-text);
  padding: 8px;
}

.conn-form__select option:hover,
.conn-form__select option:checked {
  background-color: #3a4a5a;
  color: #fff;
}

/* Firefox 选项样式 */
.conn-form__select:-moz-focusring {
  color: transparent;
  text-shadow: 0 0 0 var(--form-text);
}

.conn-form__input::placeholder {
  color: var(--form-text-muted);
}

.conn-form__input--error {
  border-color: var(--form-error);
}

.conn-form__input--error:focus {
  box-shadow: 0 0 0 2px rgba(244, 67, 54, 0.2);
}

/* ============================================================
   Password Input
   ============================================================ */
.conn-form__password-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.conn-form__password-wrapper .conn-form__input {
  padding-right: 40px;
  width: 100%;
}

.conn-form__password-toggle {
  position: absolute;
  right: 10px;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 14px;
  padding: 4px;
  opacity: 0.7;
  transition: opacity 0.2s ease;
}

.conn-form__password-toggle:hover {
  opacity: 1;
}

/* ============================================================
   Checkbox
   ============================================================ */
.conn-form__checkbox {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  padding: 8px 0;
}

.conn-form__checkbox-input {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: var(--form-border-focus);
}

.conn-form__checkbox-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--form-text);
}

.conn-form__checkbox--sub {
  padding-left: 28px;
  font-weight: 400;
}

/* ============================================================
   Advanced Fields
   ============================================================ */
.conn-form__advanced {
  margin-top: 12px;
}

.conn-form__advanced-fields {
  margin-top: 12px;
  padding: 16px;
  background-color: var(--form-bg-secondary);
  border-radius: 6px;
  border: 1px solid var(--form-border);
}

/* ============================================================
   Test Section
   ============================================================ */
.conn-form__section--test {
  background-color: var(--form-bg-secondary);
  margin: 20px -24px;
  padding: 20px 24px;
  border-top: 1px solid var(--form-border);
  border-bottom: 1px solid var(--form-border);
}

.conn-form__test {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.conn-form__test-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 24px;
  background-color: var(--form-border);
  border: none;
  border-radius: 6px;
  color: var(--form-text);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  align-self: flex-start;
}

.conn-form__test-btn:hover:not(:disabled) {
  background-color: #4a5a6a;
}

.conn-form__test-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.conn-form__test-btn--testing {
  background-color: #2d4a5a;
  color: #63b3ed;
}

.conn-form__test-btn--success {
  background-color: #1a3a2a;
  border: 1px solid var(--form-success);
  color: var(--form-success);
}

.conn-form__test-btn--error {
  background-color: #3a1a1a;
  border: 1px solid var(--form-error);
  color: var(--form-error);
}

.conn-form__spinner {
  width: 16px;
  height: 16px;
  border: 2px solid transparent;
  border-top-color: currentColor;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.conn-form__test-icon {
  font-size: 16px;
  font-weight: 600;
}

.conn-form__test-result {
  padding: 12px 16px;
  border-radius: 6px;
  font-size: 13px;
}

.conn-form__test-result--success {
  background-color: var(--form-success-bg);
  border: 1px solid var(--form-success);
}

.conn-form__test-result--error {
  background-color: var(--form-error-bg);
  border: 1px solid var(--form-error);
}

.conn-form__test-status {
  font-weight: 600;
  margin-bottom: 4px;
}

.conn-form__test-result--success .conn-form__test-status {
  color: var(--form-success);
}

.conn-form__test-result--error .conn-form__test-status {
  color: var(--form-error);
}

.conn-form__test-details {
  color: var(--form-text-secondary);
  font-size: 12px;
}

.conn-form__test-error {
  color: var(--form-error);
  font-size: 12px;
  margin-top: 4px;
  word-break: break-word;
}

/* ============================================================
   Footer - Fixed at bottom
   ============================================================ */
.conn-form__footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid var(--form-border);
  flex-shrink: 0; /* 防止被压缩 */
  background-color: var(--form-bg);
}

.conn-form__btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.conn-form__btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.conn-form__btn--cancel {
  background-color: var(--form-border);
  color: var(--form-text-secondary);
}

.conn-form__btn--cancel:hover:not(:disabled) {
  background-color: #4a5a6a;
  color: var(--form-text);
}

.conn-form__btn--save {
  background-color: var(--form-success);
  color: white;
}

.conn-form__btn--save:hover:not(:disabled) {
  background-color: #38a169;
}
</style>
