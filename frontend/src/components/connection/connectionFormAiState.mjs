export const DEFAULT_AI_TEMPERATURE = 0.1
const AI_FIELD_ERROR_PREFIX = 'ai_'

export function shouldShowApiKeyField(providerValue, isLocalProvider = () => false) {
  return !isLocalProvider(providerValue)
}

export function isAiFieldErrorKey(fieldName = '') {
  return fieldName.startsWith(AI_FIELD_ERROR_PREFIX)
}

export function normalizeModelOptions(models = []) {
  return models
    .map((model) => {
      if (typeof model === 'string') {
        const trimmed = model.trim()
        if (!trimmed) {
          return null
        }

        return {
          id: trimmed,
          name: trimmed
        }
      }

      if (!model || typeof model !== 'object') {
        return null
      }

      const id = `${model.id ?? model.name ?? ''}`.trim()
      if (!id) {
        return null
      }

      const name = `${model.name ?? model.id}`.trim() || id
      return { id, name }
    })
    .filter(Boolean)
}

export function getBlockingValidationResult(formData, currentSchema) {
  const errors = {}

  if (!formData.name?.trim()) {
    errors.name = '连接名称不能为空'
  }

  if (!formData.host?.trim()) {
    errors.host = '主机地址不能为空'
  }

  if (!formData.port || formData.port < 1 || formData.port > 65535) {
    errors.port = '端口必须在 1-65535 之间'
  }

  if (formData.type !== 'sqlserver' || formData.auth_type === 'sql') {
    if (!formData.username?.trim()) {
      errors.username = '用户名不能为空'
    }
  }

  if (currentSchema?.databaseRequired && !formData.database?.trim()) {
    errors.database = `${currentSchema.databaseLabel}不能为空`
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors
  }
}

export function collectAiFieldErrors(aiAssistants = [], isLocalProvider = () => false) {
  const errors = {}

  for (const assistant of aiAssistants) {
    const prefix = `ai_${assistant.id}`

    if (!assistant.api_host?.trim()) {
      errors[`${prefix}_api_host`] = 'API 主机不能为空'
    }

    if (!assistant.model?.trim()) {
      errors[`${prefix}_model`] = '模型不能为空'
    }

    if (!isLocalProvider(assistant.provider) && !assistant.api_key?.trim()) {
      errors[`${prefix}_api_key`] = '云端模型需要 API 密钥'
    }
  }

  return errors
}

export function collectAdvisoryAiErrors(aiAssistants = [], isLocalProvider = () => false) {
  return collectAiFieldErrors(aiAssistants, isLocalProvider)
}

export function createSaveValidationSnapshot(formData, currentSchema, isLocalProvider = () => false) {
  const blockingResult = getBlockingValidationResult(formData, currentSchema)

  return {
    isValid: blockingResult.isValid,
    blockingErrors: blockingResult.errors,
    advisoryErrors: collectAdvisoryAiErrors(formData.ai_assistants, isLocalProvider)
  }
}

export function applyBlockingFieldValidation(existingErrors = {}, blockingErrors = {}) {
  const nextErrors = {}

  for (const [fieldName, message] of Object.entries(existingErrors)) {
    if (isAiFieldErrorKey(fieldName)) {
      nextErrors[fieldName] = message
    }
  }

  return {
    ...nextErrors,
    ...blockingErrors
  }
}

export function pruneInactiveAiErrors(existingErrors = {}, aiAssistants = [], isLocalProvider = () => false) {
  const nextErrors = {}

  for (const [fieldName, message] of Object.entries(existingErrors)) {
    if (!isAiFieldErrorKey(fieldName)) {
      nextErrors[fieldName] = message
    }
  }

  return {
    ...nextErrors,
    ...collectAdvisoryAiErrors(aiAssistants, isLocalProvider)
  }
}
