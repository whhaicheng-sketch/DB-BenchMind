<script setup>
/**
 * ConnectionRemoteSection.vue
 * Remote tab content for SSH/WinRM connection configuration.
 * Receives all state via props from parent ConnectionForm.
 */
const props = defineProps({
  formData: { type: Object, required: true },
  fieldErrors: { type: Object, required: true },
  isRemoteTypeNoneSelected: { type: Boolean, required: true },
  isRemoteTypeSSHSelected: { type: Boolean, required: true },
  isRemoteTypeWinRMSelected: { type: Boolean, required: true },
  showSshPassword: { type: Boolean, required: true },
  showWinrmPassword: { type: Boolean, required: true },
  sshTestResult: { type: Object, default: null },
  winrmTestResult: { type: Object, default: null }
})

const emit = defineEmits([
  'set-remote-type',
  'on-ssh-host-change',
  'on-winrm-host-change',
  'on-winrm-port-change',
  'sync-remote-host-from-general',
  'update:showSshPassword',
  'update:showWinrmPassword'
])
</script>

<template>
  <div class="conn-form">
    <div class="conn-form__row">
      <label class="conn-form__label">远程连接类型</label>
      <div class="conn-form__radio-group">
        <label class="conn-form__radio">
          <input :checked="isRemoteTypeNoneSelected" type="radio" @change="emit('set-remote-type', 'none')">
          <span>不使用远程连接</span>
        </label>
        <label class="conn-form__radio">
          <input :checked="isRemoteTypeSSHSelected" type="radio" @change="emit('set-remote-type', 'ssh')">
          <span>SSH</span>
        </label>
        <label class="conn-form__radio">
          <input :checked="isRemoteTypeWinRMSelected" type="radio" @change="emit('set-remote-type', 'winrm')">
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
            @input="emit('on-ssh-host-change')"
          />
        </div>
        <button type="button" class="conn-form__sync-btn" @click="emit('sync-remote-host-from-general')" title="从数据库主机同步">
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
          <button type="button" class="conn-form__password-toggle" @click="emit('update:showSshPassword', !showSshPassword)">
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
            @input="emit('on-winrm-host-change')"
          />
        </div>
        <button type="button" class="conn-form__sync-btn" @click="emit('sync-remote-host-from-general')" title="从数据库主机同步">
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
            @input="emit('on-winrm-port-change')"
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
          <button type="button" class="conn-form__password-toggle" @click="emit('update:showWinrmPassword', !showWinrmPassword)">
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
</template>
