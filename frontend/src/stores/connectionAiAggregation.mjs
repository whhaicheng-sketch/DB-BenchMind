const hasText = (value) => typeof value === 'string' && value.trim() !== ''

export const isConfiguredAiAssistant = (assistant) => {
  if (!assistant || !hasText(assistant.provider)) {
    return false
  }

  return hasText(assistant.api_key) || hasText(assistant.model)
}

export const selectPreferredAiAssistant = (connection) => {
  if (!Array.isArray(connection?.ai_assistants)) {
    return null
  }

  return connection.ai_assistants.find(isConfiguredAiAssistant) || null
}

export const shouldTestAiForConnection = (connection) => {
  return selectPreferredAiAssistant(connection) !== null
}

export const buildAiTestRequest = (assistant) => {
  if (!assistant) {
    return null
  }

  return {
    provider: assistant.provider,
    api_host: assistant.api_host || '',
    api_endpoint: assistant.api_endpoint || '',
    api_key: assistant.api_key || '',
    model: assistant.model || ''
  }
}
