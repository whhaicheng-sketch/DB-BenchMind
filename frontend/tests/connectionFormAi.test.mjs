import test from 'node:test'
import assert from 'node:assert/strict'

import {
  DEFAULT_AI_TEMPERATURE,
  collectAiFieldErrors,
  getBlockingValidationResult,
  normalizeModelOptions
} from '../src/components/connection/connectionFormAiState.mjs'

test('normalizeModelOptions accepts string and object model payloads', () => {
  assert.deepEqual(
    normalizeModelOptions(['deepseek-chat', 'deepseek-reasoner']),
    [
      { id: 'deepseek-chat', name: 'deepseek-chat' },
      { id: 'deepseek-reasoner', name: 'deepseek-reasoner' }
    ]
  )

  assert.deepEqual(
    normalizeModelOptions([
      { id: 'deepseek-chat', name: 'DeepSeek Chat' },
      { id: 'deepseek-reasoner' }
    ]),
    [
      { id: 'deepseek-chat', name: 'DeepSeek Chat' },
      { id: 'deepseek-reasoner', name: 'deepseek-reasoner' }
    ]
  )
})

test('AI assistant validation remains advisory and does not block save', () => {
  const formData = {
    name: 'prod-db',
    host: '127.0.0.1',
    port: 3306,
    username: 'root',
    database: '',
    type: 'mysql',
    auth_type: 'sql',
    ai_assistants: [
      {
        id: 'default',
        provider: 'deepseek',
        api_host: 'https://api.deepseek.com',
        api_key: '',
        model: ''
      }
    ]
  }

  const schema = {
    databaseRequired: false,
    databaseLabel: 'Database'
  }

  const blockingResult = getBlockingValidationResult(formData, schema)
  const aiErrors = collectAiFieldErrors(formData.ai_assistants, (provider) => provider === 'ollama')

  assert.equal(blockingResult.isValid, true)
  assert.equal(aiErrors.ai_default_model, '模型不能为空')
  assert.equal(aiErrors.ai_default_api_key, '云端模型需要 API 密钥')
})

test('default AI temperature stays at 0.1', () => {
  assert.equal(DEFAULT_AI_TEMPERATURE, 0.1)
})
