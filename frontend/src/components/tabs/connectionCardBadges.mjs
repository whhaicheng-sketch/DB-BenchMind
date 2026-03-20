import { isConfiguredAiAssistant } from '../../stores/connectionAiAggregation.mjs'

export { isConfiguredAiAssistant }

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
