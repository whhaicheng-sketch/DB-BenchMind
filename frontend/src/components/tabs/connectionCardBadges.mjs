const hasText = (value) => typeof value === 'string' && value.trim() !== ''

export const isConfiguredAiAssistant = (assistant) => {
  if (!assistant || !hasText(assistant.provider)) {
    return false
  }

  // Default assistant shells always have provider/host defaults, so only count
  // assistants that have at least one meaningful configured value.
  return hasText(assistant.api_key) || hasText(assistant.model)
}

export const countConfiguredAiAssistants = (connection) => {
  if (!Array.isArray(connection?.ai_assistants)) {
    return 0
  }

  return connection.ai_assistants.filter(isConfiguredAiAssistant).length
}

export const hasConfiguredAiAssistants = (connection) => {
  return countConfiguredAiAssistants(connection) > 0
}

export const getAiBadgeTooltip = (connection) => {
  const count = countConfiguredAiAssistants(connection)
  if (count <= 0) {
    return ''
  }

  return `已配置 ${count} 个 AI 助手`
}
