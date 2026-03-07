<script setup>
/**
 * ConnectionForm.vue
 * Form for creating and editing database connections.
 * Supports MySQL, PostgreSQL, Oracle, and SQL Server.
 */
import { ref, computed, watch } from 'vue'
import { useConnectionStore } from '../../stores/connection'

// Props
const props = defineProps({
  // For editing: pass existing connection ID
  connectionId: {
    type: String,
    default: null
  },
  // Mode: 'create' or 'edit'
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

// Form state
const formData = ref({
  name: '',
  type: 'mysql',
  host: '',
  port: 3306,
  database: '',
  username: '',
  password: '',
  ssl_mode: ''
})

const saving = ref(false)
const testing = ref(false)
const formError = ref(null)
const localTestResult = ref(null)

// Default ports by type
const defaultPorts = {
  mysql: 3306,
  postgresql: 5432,
  oracle: 1521,
  sqlserver: 1433
}

// SSL modes by type
const sslModeOptions = {
  mysql: [
    { value: 'preferred', label: 'Preferred' },
    { value: 'required', label: 'Required' },
    { value: 'disabled', label: 'Disabled' }
  ],
  postgresql: [
    { value: 'prefer', label: 'Prefer' },
    { value: 'require', label: 'Require' },
    { value: 'disable', label: 'Disable' },
    { value: 'verify-full', label: 'Verify Full' }
  ],
  oracle: [],
  sqlserver: [
    { value: '', label: 'None' },
    { value: 'encrypt', label: 'Encrypt' }
  ]
}

// Connection type options
const typeOptions = [
  { value: 'mysql', label: 'MySQL' },
  { value: 'postgresql', label: 'PostgreSQL' },
  { value: 'oracle', label: 'Oracle' },
  { value: 'sqlserver', label: 'SQL Server' }
]

// Computed
const isEditing = computed(() => props.mode === 'edit' && props.connectionId)
const title = computed(() => isEditing.value ? 'Edit Connection' : 'New Connection')
const currentSslOptions = computed(() => sslModeOptions[formData.value.type] || [])
const showPasswordField = computed(() => props.mode === 'create')

// Watch type changes to update default port
watch(() => formData.value.type, (newType) => {
  formData.value.port = defaultPorts[newType] || 3306
  formData.value.ssl_mode = ''
})

// Load existing connection for editing
watch(() => props.connectionId, async (newId) => {
  if (newId && props.mode === 'edit') {
    const conn = await connectionStore.connections.find(c => c.id === newId)
    if (conn) {
      formData.value = {
        name: conn.name,
        type: conn.type,
        host: conn.host,
        port: conn.port,
        database: conn.database || '',
        username: conn.username,
        password: '', // Don't prefill password
        ssl_mode: conn.ssl_mode || ''
      }
    }
  }
}, { immediate: true })

// Methods
const validateForm = () => {
  if (!formData.value.name.trim()) {
    formError.value = 'Connection name is required'
    return false
  }
  if (!formData.value.host.trim()) {
    formError.value = 'Host is required'
    return false
  }
  if (!formData.value.port || formData.value.port < 1 || formData.value.port > 65535) {
    formError.value = 'Port must be between 1 and 65535'
    return false
  }
  if (!formData.value.username.trim()) {
    formError.value = 'Username is required'
    return false
  }
  return true
}

const handleSave = async () => {
  formError.value = null

  if (!validateForm()) {
    return
  }

  saving.value = true

  try {
    if (isEditing.value) {
      // Update existing connection
      const updated = await connectionStore.updateConnection({
        id: props.connectionId,
        ...formData.value
      })
      if (updated) {
        emit('saved', updated)
      } else {
        formError.value = connectionStore.error || 'Failed to update connection'
      }
    } else {
      // Create new connection
      const created = await connectionStore.createConnection(formData.value)
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

const handleTest = async () => {
  // For testing, we need to save first if it's a new connection
  // or test an existing connection
  formError.value = null
  localTestResult.value = null

  if (!props.connectionId) {
    formError.value = 'Please save the connection before testing'
    return
  }

  testing.value = true

  try {
    const result = await connectionStore.testConnectionById(props.connectionId)
    localTestResult.value = result
    emit('tested', result)
  } finally {
    testing.value = false
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
    username: '',
    password: '',
    ssl_mode: ''
  }
  formError.value = null
  localTestResult.value = null
}
</script>

<template>
  <div class="connection-form">
    <h3 class="form-title">{{ title }}</h3>

    <!-- Error display -->
    <div v-if="formError" class="form-error">
      {{ formError }}
    </div>

    <!-- Form fields -->
    <div class="form-body">
      <!-- Connection Name -->
      <div class="form-group">
        <label>Connection Name *</label>
        <input
          v-model="formData.name"
          type="text"
          class="form-input"
          placeholder="My Database"
        />
      </div>

      <!-- Database Type -->
      <div class="form-group">
        <label>Database Type *</label>
        <select v-model="formData.type" class="form-select">
          <option v-for="opt in typeOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <!-- Host -->
      <div class="form-group">
        <label>Host *</label>
        <input
          v-model="formData.host"
          type="text"
          class="form-input"
          placeholder="localhost"
        />
      </div>

      <!-- Port -->
      <div class="form-group">
        <label>Port *</label>
        <input
          v-model.number="formData.port"
          type="number"
          class="form-input"
          min="1"
          max="65535"
        />
      </div>

      <!-- Database -->
      <div class="form-group">
        <label>Database</label>
        <input
          v-model="formData.database"
          type="text"
          class="form-input"
          placeholder="mydb"
        />
      </div>

      <!-- Username -->
      <div class="form-group">
        <label>Username *</label>
        <input
          v-model="formData.username"
          type="text"
          class="form-input"
          placeholder="root"
        />
      </div>

      <!-- Password -->
      <div class="form-group" v-if="showPasswordField">
        <label>Password</label>
        <input
          v-model="formData.password"
          type="password"
          class="form-input"
          placeholder="••••••••"
        />
      </div>

      <!-- Password hint for edit mode -->
      <div class="form-group" v-else>
        <label>Password</label>
        <div class="password-hint">
          Password is preserved. Enter new password to update.
        </div>
        <input
          v-model="formData.password"
          type="password"
          class="form-input"
          placeholder="Leave empty to keep current"
        />
      </div>

      <!-- SSL Mode -->
      <div class="form-group" v-if="currentSslOptions.length > 0">
        <label>SSL Mode</label>
        <select v-model="formData.ssl_mode" class="form-select">
          <option value="">Default</option>
          <option v-for="opt in currentSslOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>
    </div>

    <!-- Test Result -->
    <div v-if="localTestResult" class="test-result" :class="{ success: localTestResult.success, error: !localTestResult.success }">
      <div class="result-header">
        <span v-if="localTestResult.success" class="result-icon">✓</span>
        <span v-else class="result-icon">✗</span>
        <span class="result-status">
          {{ localTestResult.success ? 'Connection Successful' : 'Connection Failed' }}
        </span>
      </div>
      <div v-if="localTestResult.success" class="result-details">
        <div>Latency: {{ localTestResult.latency_ms }}ms</div>
        <div v-if="localTestResult.database_version">
          Version: {{ localTestResult.database_version }}
        </div>
      </div>
      <div v-if="localTestResult.error" class="result-error">
        {{ localTestResult.error }}
      </div>
    </div>

    <!-- Buttons -->
    <div class="form-actions">
      <button
        class="btn btn-test"
        @click="handleTest"
        :disabled="testing || saving || !connectionId"
      >
        {{ testing ? 'Testing...' : 'Test' }}
      </button>
      <button
        class="btn btn-cancel"
        @click="handleCancel"
        :disabled="saving"
      >
        Cancel
      </button>
      <button
        class="btn btn-save"
        @click="handleSave"
        :disabled="saving"
      >
        {{ saving ? 'Saving...' : 'Save' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.connection-form {
  background-color: #1e2a3a;
  border-radius: 8px;
  padding: 16px;
}

.form-title {
  font-size: 16px;
  font-weight: 600;
  color: #e2e8f0;
  margin-bottom: 16px;
}

.form-error {
  background-color: #742a2a;
  color: #fc8181;
  padding: 10px;
  border-radius: 6px;
  margin-bottom: 16px;
  font-size: 14px;
}

.form-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.form-group label {
  font-size: 12px;
  color: #a0aec0;
  font-weight: 500;
}

.form-input,
.form-select {
  padding: 8px 12px;
  background-color: #1a202c;
  border: 1px solid #3a4a5a;
  border-radius: 6px;
  color: #e2e8f0;
  font-size: 14px;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #4299e1;
}

.form-input::placeholder {
  color: #4a5568;
}

.password-hint {
  font-size: 11px;
  color: #718096;
  margin-bottom: 4px;
}

.test-result {
  margin-top: 16px;
  padding: 12px;
  border-radius: 6px;
}

.test-result.success {
  background-color: #1a3a2a;
  border: 1px solid #38a169;
}

.test-result.error {
  background-color: #3a1a1a;
  border: 1px solid #e53e3e;
}

.result-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.result-icon {
  font-size: 16px;
}

.test-result.success .result-icon {
  color: #68d391;
}

.test-result.error .result-icon {
  color: #fc8181;
}

.result-status {
  font-weight: 600;
  font-size: 14px;
}

.test-result.success .result-status {
  color: #68d391;
}

.test-result.error .result-status {
  color: #fc8181;
}

.result-details {
  font-size: 13px;
  color: #a0aec0;
}

.result-details div {
  margin-top: 4px;
}

.result-error {
  margin-top: 8px;
  font-size: 13px;
  color: #fc8181;
  word-break: break-all;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-test {
  background-color: #2d4a5a;
  color: #63b3ed;
}

.btn-test:hover:not(:disabled) {
  background-color: #3d5a6a;
}

.btn-cancel {
  background-color: #3a4a5a;
  color: #a0aec0;
}

.btn-cancel:hover:not(:disabled) {
  background-color: #4a5a6a;
}

.btn-save {
  background-color: #38a169;
  color: white;
}

.btn-save:hover:not(:disabled) {
  background-color: #2f855a;
}
</style>
