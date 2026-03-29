import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import {
  DEFAULT_AI_TEMPERATURE,
  applyBlockingFieldValidation,
  collectAdvisoryAiErrors,
  createSaveValidationSnapshot,
  collectAiFieldErrors,
  getBlockingValidationResult,
  normalizeModelOptions,
  pruneInactiveAiErrors,
  shouldShowApiKeyField
} from '../src/components/connection/connectionFormAiState.mjs'
import {
  DEFAULT_AI_TEST_PROMPT,
  buildAiChatTestRequest,
  createAiTestDialogState
} from '../src/components/connection/connectionFormAiTestState.mjs'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const connectionFormSource = fs.readFileSync(path.resolve(__dirname, '../src/components/connection/ConnectionForm.vue'), 'utf8')
const aiSectionSource = fs.readFileSync(path.resolve(__dirname, '../src/components/connection/ConnectionAISection.vue'), 'utf8')
const stateSource = fs.readFileSync(path.resolve(__dirname, '../src/components/connection/useConnectionFormState.mjs'), 'utf8')

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

test('createSaveValidationSnapshot keeps AI errors advisory while blocking only on database fields', () => {
  const snapshot = createSaveValidationSnapshot(
    {
      name: 'prod-db',
      host: '127.0.0.1',
      port: 3306,
      username: 'root',
      database: '',
      type: 'mysql',
      auth_type: 'sql',
      ai_assistants: [
        {
          id: 'glm',
          provider: 'glm',
          api_host: 'https://open.bigmodel.cn/api/paas/v4',
          api_key: '',
          model: ''
        }
      ]
    },
    { databaseRequired: false, databaseLabel: 'Database' },
    (provider) => provider === 'ollama'
  )

  assert.equal(snapshot.isValid, true)
  assert.deepEqual(snapshot.blockingErrors, {})
  assert.equal(snapshot.advisoryErrors.ai_glm_model, '模型不能为空')
  assert.equal(snapshot.advisoryErrors.ai_glm_api_key, '云端模型需要 API 密钥')
})

test('createSaveValidationSnapshot blocks save when database type is missing even if AI fields are present', () => {
  const snapshot = createSaveValidationSnapshot(
    {
      name: 'prod-db',
      host: '127.0.0.1',
      port: 3306,
      username: 'root',
      database: '',
      type: '',
      auth_type: 'sql',
      ai_assistants: [
        {
          id: 'minimax',
          provider: 'minimax',
          api_host: 'https://api.minimaxi.com',
          api_key: 'sk-live',
          model: 'MiniMax-M2.7'
        }
      ]
    },
    { databaseRequired: false, databaseLabel: 'Database' },
    (provider) => provider === 'ollama'
  )

  assert.equal(snapshot.isValid, false)
  assert.equal(snapshot.blockingErrors.type, '数据库类型不能为空')
})

test('applyBlockingFieldValidation removes stale blocking errors while preserving AI advisory hints', () => {
  const nextErrors = applyBlockingFieldValidation(
    {
      host: '主机地址不能为空',
      ai_default_model: '模型不能为空'
    },
    {
      name: '连接名称不能为空'
    }
  )

  assert.deepEqual(nextErrors, {
    name: '连接名称不能为空',
    ai_default_model: '模型不能为空'
  })
})

test('pruneInactiveAiErrors removes hidden assistant errors after provider or assistant switch', () => {
  const nextErrors = pruneInactiveAiErrors(
    {
      ai_default_model: '模型不能为空',
      ai_removed_api_key: '云端模型需要 API 密钥',
      host: '主机地址不能为空'
    },
    [
      {
        id: 'default',
        provider: 'ollama',
        api_host: 'http://localhost:11434',
        api_key: '',
        model: 'llama3'
      }
    ],
    (provider) => provider === 'ollama'
  )

  assert.deepEqual(nextErrors, {
    host: '主机地址不能为空'
  })
})

test('collectAdvisoryAiErrors skips hidden local-provider API key errors', () => {
  const nextErrors = collectAdvisoryAiErrors(
    [
      {
        id: 'local',
        provider: 'ollama',
        api_host: 'http://localhost:11434',
        api_key: '',
        model: ''
      }
    ],
    (provider) => provider === 'ollama'
  )

  assert.deepEqual(nextErrors, {
    ai_local_model: '模型不能为空'
  })
})

test('default AI temperature stays at 0.1', () => {
  assert.equal(DEFAULT_AI_TEMPERATURE, 0.1)
})

test('API key field is only rendered for remote providers', () => {
  assert.equal(shouldShowApiKeyField('deepseek', (provider) => provider === 'ollama'), true)
  assert.equal(shouldShowApiKeyField('ollama', (provider) => provider === 'ollama'), false)
})

test('AI test dialog state opens with default prompt and no stale result', () => {
  const state = createAiTestDialogState()

  assert.equal(state.prompt, DEFAULT_AI_TEST_PROMPT)
  assert.equal(state.status, 'idle')
  assert.equal(state.responseText, '')
  assert.equal(state.errorText, '')
})

test('AI test request uses unsaved assistant form values directly', () => {
  const request = buildAiChatTestRequest(
    {
      provider: 'deepseek',
      api_host: 'https://api.deepseek.com',
      api_endpoint: '/v1/chat/completions',
      api_key: 'sk-live',
      model: 'deepseek-chat',
      temperature: 0.3
    },
    DEFAULT_AI_TEST_PROMPT
  )

  assert.deepEqual(request, {
    provider: 'deepseek',
    api_host: 'https://api.deepseek.com',
    api_endpoint: '/v1/chat/completions',
    api_key: 'sk-live',
    model: 'deepseek-chat',
    prompt: DEFAULT_AI_TEST_PROMPT,
    temperature: 0.3
  })
})

test('ConnectionForm wires the visible AI test dialog entry points and removes the local API key placeholder row', () => {
  assert.match(aiSectionSource, /测试模型/)
  assert.match(aiSectionSource, /<AiTestDialog/)
  assert.doesNotMatch(aiSectionSource, /本地模型无需 API 密钥/)
})

test('ConnectionForm save flow only uses blocking validation for the top banner', () => {
  assert.match(stateSource, /if \(!validateForm\(\)\)/)
  assert.doesNotMatch(stateSource, /aiTestStatus[\s\S]{0,200}请修正表单中的错误/)
  assert.doesNotMatch(stateSource, /modelQueryError[\s\S]{0,200}请修正表单中的错误/)
})

test('ConnectionForm save flow does not gate save on AI test failures or model query failures', () => {
  // Extract the handleSave function body using a brace-matching approach
  const startIdx = stateSource.indexOf('const handleSave = async () => {')
  assert.ok(startIdx >= 0, 'expected handleSave function')

  // Find the matching closing brace
  let depth = 0
  let endIdx = startIdx
  for (let i = startIdx; i < stateSource.length; i++) {
    if (stateSource[i] === '{') depth++
    else if (stateSource[i] === '}') {
      depth--
      if (depth === 0) { endIdx = i; break }
    }
  }

  const handleSaveSource = stateSource.slice(startIdx, endIdx)
  assert.doesNotMatch(handleSaveSource, /aiTestStatus|aiTestResult|modelQueryError|availableModels|QueryAIModels|TestAIConnection/)
})

test('ConnectionForm clears stale AI interaction state when switching providers or assistants', () => {
  assert.match(stateSource, /resetAiInteractionState\(\)/)
  assert.match(stateSource, /formData\.value\.ai_assistants\.map\(\(assistant\) => assistant\.provider\)\.join\('\|'\)/)
  assert.match(stateSource, /selectedAssistant\.value\?\.api_key \|\| ''/)
  assert.match(stateSource, /selectedAssistant\.value\?\.model \|\| ''/)
})

test('ConnectionForm uses China-region MiniMax defaults in provider options and initial assistant state', () => {
  assert.match(stateSource, /\{ value: 'minimax', label: 'MiniMax', host: 'https:\/\/api\.minimaxi\.com', endpoint: '\/v1\/chat\/completions', model: 'MiniMax-M2\.7' \}/)
  assert.doesNotMatch(stateSource, /\{ value: 'minimax', label: 'MiniMax', host: 'https:\/\/api\.minimax\.io'/)
})

test('ConnectionForm renders AI test buttons in the required left-to-right order', () => {
  const testAreaMatch = aiSectionSource.match(/<div class="ai-config__test-area">([\s\S]*?)<\/div>\s*<\/template>/)
  assert.ok(testAreaMatch, 'expected AI test area markup')

  const testAreaSource = testAreaMatch[1]
  const connectIndex = testAreaSource.indexOf('测试连接')
  const modelIndex = testAreaSource.indexOf('测试模型')

  assert.ok(connectIndex >= 0, 'expected 测试连接 button')
  assert.ok(modelIndex >= 0, 'expected 测试模型 button')
  assert.ok(connectIndex < modelIndex, 'expected 测试连接 button before 测试模型 button')
})
