import { DEFAULT_AI_TEMPERATURE } from './connectionFormAiState.mjs'

export const DEFAULT_AI_TEST_PROMPT = '你是谁？请简单介绍你自己。'

export function createAiTestDialogState() {
  return {
    prompt: DEFAULT_AI_TEST_PROMPT,
    status: 'idle',
    responseText: '',
    errorText: '',
    latencyMs: null
  }
}

export function buildAiChatTestRequest(assistant, prompt) {
  return {
    provider: assistant.provider,
    api_host: assistant.api_host,
    api_endpoint: assistant.api_endpoint,
    api_key: assistant.api_key || '',
    model: assistant.model,
    prompt,
    temperature: assistant.temperature ?? DEFAULT_AI_TEMPERATURE
  }
}
