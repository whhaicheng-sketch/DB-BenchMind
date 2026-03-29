<script setup>
/**
 * ConnectionGeneralSection.vue
 * General tab content for database connection configuration.
 * Receives all state via props from parent ConnectionForm.
 */
const props = defineProps({
  formData: { type: Object, required: true },
  fieldErrors: { type: Object, required: true },
  currentSchema: { type: Object, required: true },
  shouldShowHostPort: { type: Boolean, required: true },
  shouldShowDatabaseField: { type: Boolean, required: true },
  isOracleBasicMode: { type: Boolean, required: true },
  isOracleTNSMode: { type: Boolean, required: true },
  showPassword: { type: Boolean, required: true },
  dbTestResult: { type: Object, default: null },
  validateField: { type: Function, required: true },
  getOracleModeFieldError: { type: Function, required: true }
})

const emit = defineEmits([
  'update:showPassword'
])
</script>

<template>
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
          <button type="button" class="conn-form__password-toggle" @click="emit('update:showPassword', !showPassword)">
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
</template>
