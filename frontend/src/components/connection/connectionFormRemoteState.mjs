const REMOTE_TYPES = new Set(['none', 'ssh', 'winrm'])

const hasText = (value) => typeof value === 'string' && value.trim() !== ''

export function getRemoteType(connOrFormData = {}) {
  if (REMOTE_TYPES.has(connOrFormData.remote_type)) {
    return connOrFormData.remote_type
  }

  const sshEnabled = !!connOrFormData.ssh_enabled
  const winrmEnabled = !!connOrFormData.winrm_enabled

  // Historical dirty data may have both flags set. Prefer SSH so the UI stays
  // deterministic and the next save normalizes the record back to one remote type.
  if (sshEnabled) {
    return 'ssh'
  }

  if (winrmEnabled) {
    return 'winrm'
  }

  return 'none'
}

export function isRemoteTypeNone(connOrFormData = {}) {
  return getRemoteType(connOrFormData) === 'none'
}

export function isRemoteTypeSSH(connOrFormData = {}) {
  return getRemoteType(connOrFormData) === 'ssh'
}

export function isRemoteTypeWinRM(connOrFormData = {}) {
  return getRemoteType(connOrFormData) === 'winrm'
}

export function applyRemoteTypeCompatibilityFields(formData = {}) {
  const remoteType = getRemoteType(formData)

  return {
    ...formData,
    remote_type: remoteType,
    ssh_enabled: remoteType === 'ssh',
    winrm_enabled: remoteType === 'winrm'
  }
}

export function getDefaultWinRMPortByScheme(scheme = 'http') {
  return scheme === 'https' ? 5986 : 5985
}

export function shouldAutoUpdateWinRMPort(formData = {}) {
  return !formData.remote_port_user_overridden
}

export function syncRemoteHostFromGeneral(formData = {}) {
  const remoteType = getRemoteType(formData)

  if (remoteType === 'ssh') {
    return {
      ...formData,
      ssh_host: formData.host || ''
    }
  }

  if (remoteType === 'winrm') {
    return {
      ...formData,
      winrm_host: formData.host || ''
    }
  }

  return { ...formData }
}

export function getVisibleRemoteBlockingErrors(formData = {}) {
  const remoteType = getRemoteType(formData)
  const errors = {}

  if (remoteType === 'ssh') {
    if (!hasText(formData.ssh_host)) {
      errors.ssh_host = '请填写必填项：SSH 主机'
    }
    if (!formData.ssh_port || formData.ssh_port < 1 || formData.ssh_port > 65535) {
      errors.ssh_port = '请填写必填项：SSH 端口'
    }
    if (!hasText(formData.ssh_username)) {
      errors.ssh_username = '请填写必填项：SSH 用户名'
    }
  }

  if (remoteType === 'winrm') {
    if (!hasText(formData.winrm_host)) {
      errors.winrm_host = '请填写必填项：WinRM 主机'
    }
    if (!formData.winrm_port || formData.winrm_port < 1 || formData.winrm_port > 65535) {
      errors.winrm_port = '请填写必填项：WinRM 端口'
    }
    if (!hasText(formData.winrm_username)) {
      errors.winrm_username = '请填写必填项：WinRM 用户名'
    }
  }

  return errors
}

export function buildRemoteBlockingErrorMessage(blockingErrors = {}) {
  const remoteLabels = []

  if (blockingErrors.ssh_host) remoteLabels.push('SSH 主机')
  if (blockingErrors.ssh_port) remoteLabels.push('SSH 端口')
  if (blockingErrors.ssh_username) remoteLabels.push('SSH 用户名')
  if (blockingErrors.winrm_host) remoteLabels.push('WinRM 主机')
  if (blockingErrors.winrm_port) remoteLabels.push('WinRM 端口')
  if (blockingErrors.winrm_username) remoteLabels.push('WinRM 用户名')

  if (remoteLabels.length === 0) {
    return ''
  }

  return `请填写必填项：${remoteLabels.join('、')}`
}
