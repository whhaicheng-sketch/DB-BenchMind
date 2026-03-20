import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import {
  countConfiguredAiAssistants,
  getAiBadgeTooltip,
  hasConfiguredAiAssistants,
  isConfiguredAiAssistant
} from '../src/components/tabs/connectionCardBadges.mjs'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const connectionsTabSource = fs.readFileSync(path.resolve(__dirname, '../src/components/tabs/ConnectionsTab.vue'), 'utf8')

test('connection cards show no SSH or AI badge when neither is configured', () => {
  const connection = {
    ssh_enabled: false,
    ai_assistants: []
  }

  assert.equal(connection.ssh_enabled, false)
  assert.equal(hasConfiguredAiAssistants(connection), false)
})

test('connection cards show SSH only when only SSH is configured', () => {
  const connection = {
    ssh_enabled: true,
    ai_assistants: []
  }

  assert.equal(connection.ssh_enabled, true)
  assert.equal(hasConfiguredAiAssistants(connection), false)
})

test('connection cards show AI only when at least one configured AI assistant exists', () => {
  const connection = {
    ssh_enabled: false,
    ai_assistants: [
      {
        id: 'assistant_1',
        provider: 'deepseek',
        api_key: 'sk-live',
        model: ''
      }
    ]
  }

  assert.equal(connection.ssh_enabled, false)
  assert.equal(hasConfiguredAiAssistants(connection), true)
  assert.equal(countConfiguredAiAssistants(connection), 1)
})

test('connection cards show both SSH and AI badges when both are configured', () => {
  const connection = {
    ssh_enabled: true,
    ai_assistants: [
      {
        id: 'assistant_1',
        provider: 'minimax',
        api_key: '',
        model: 'MiniMax-M2.7'
      }
    ]
  }

  assert.equal(connection.ssh_enabled, true)
  assert.equal(hasConfiguredAiAssistants(connection), true)
})

test('AI badge is not hidden by AI test failure state', () => {
  const connection = {
    ssh_enabled: false,
    ai_test_result: { success: false, error: 'invalid api key' },
    ai_assistants: [
      {
        id: 'assistant_1',
        provider: 'glm',
        api_key: 'sk-live',
        model: ''
      }
    ]
  }

  assert.equal(hasConfiguredAiAssistants(connection), true)
})

test('adding or removing AI assistants updates AI badge state', () => {
  const connection = {
    ssh_enabled: false,
    ai_assistants: []
  }

  assert.equal(hasConfiguredAiAssistants(connection), false)

  connection.ai_assistants.push({
    id: 'assistant_1',
    provider: 'ollama',
    api_key: '',
    model: 'llama3'
  })

  assert.equal(hasConfiguredAiAssistants(connection), true)
  assert.equal(countConfiguredAiAssistants(connection), 1)

  connection.ai_assistants.length = 0

  assert.equal(hasConfiguredAiAssistants(connection), false)
})

test('default empty assistant shell does not count as configured AI', () => {
  assert.equal(
    isConfiguredAiAssistant({
      id: 'default',
      provider: 'deepseek',
      api_key: '',
      model: ''
    }),
    false
  )
})

test('AI badge tooltip reflects configured assistant count', () => {
  const connection = {
    ai_assistants: [
      {
        id: 'assistant_1',
        provider: 'deepseek',
        api_key: 'sk-live',
        model: ''
      },
      {
        id: 'assistant_2',
        provider: 'ollama',
        api_key: '',
        model: 'qwen2.5-coder'
      }
    ]
  }

  assert.equal(getAiBadgeTooltip(connection), '已配置 2 个 AI 助手')
})

test('ConnectionsTab renders AI badge with the SSH badge family and helper-based condition', () => {
  assert.match(connectionsTabSource, /<span v-if="conn\.ssh_enabled" class="tag tag-ssh">SSH<\/span>/)
  assert.match(connectionsTabSource, /<span v-if="hasConfiguredAiAssistants\(conn\)" class="tag tag-ai" :title="getAiBadgeTooltip\(conn\)">AI<\/span>/)
  assert.match(connectionsTabSource, /import\s*\{\s*getAiBadgeTooltip,\s*hasConfiguredAiAssistants\s*\}\s*from '\.\/connectionCardBadges'/)
})
