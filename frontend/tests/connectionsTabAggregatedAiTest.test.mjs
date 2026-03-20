import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import {
  buildAiTestRequest,
  selectPreferredAiAssistant,
  shouldTestAiForConnection
} from '../src/stores/connectionAiAggregation.mjs'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const connectionStoreSource = fs.readFileSync(path.resolve(__dirname, '../src/stores/connection.js'), 'utf8')
const connectionsTabSource = fs.readFileSync(path.resolve(__dirname, '../src/components/tabs/ConnectionsTab.vue'), 'utf8')

test('aggregated connection test only enables AI testing when at least one valid assistant exists', () => {
  assert.equal(shouldTestAiForConnection({ ai_assistants: [] }), false)
  assert.equal(
    shouldTestAiForConnection({
      ai_assistants: [{ id: 'draft', provider: 'deepseek', api_key: '', model: '' }]
    }),
    false
  )
  assert.equal(
    shouldTestAiForConnection({
      ai_assistants: [{ id: 'ready', provider: 'ollama', api_key: '', model: 'llama3' }]
    }),
    true
  )
})

test('aggregated connection test selects the first valid AI assistant when multiple assistants exist', () => {
  const selected = selectPreferredAiAssistant({
    ai_assistants: [
      { id: 'draft', provider: 'deepseek', api_key: '', model: '' },
      { id: 'primary', provider: 'deepseek', api_key: 'sk-live', model: '' },
      { id: 'secondary', provider: 'ollama', api_key: '', model: 'llama3' }
    ]
  })

  assert.deepEqual(selected, { id: 'primary', provider: 'deepseek', api_key: 'sk-live', model: '' })
})

test('aggregated connection test builds AI connection requests from the selected assistant', () => {
  const request = buildAiTestRequest({
    provider: 'deepseek',
    api_host: 'https://api.deepseek.com',
    api_endpoint: '/v1/chat/completions',
    api_key: 'sk-live',
    model: 'deepseek-chat'
  })

  assert.deepEqual(request, {
    provider: 'deepseek',
    api_host: 'https://api.deepseek.com',
    api_endpoint: '/v1/chat/completions',
    api_key: 'sk-live',
    model: 'deepseek-chat'
  })
})

test('connection store keeps AI test results separate from DB and SSH test state and calls TestAIConnection', () => {
  assert.match(connectionStoreSource, /TestAIConnection/)
  assert.match(connectionStoreSource, /aiTestResultById:\s*\{\}/)
  assert.match(connectionStoreSource, /getAITestResultById/)
  assert.match(connectionStoreSource, /state\.aiTestResultById\[id\]/)
  assert.match(connectionStoreSource, /state\.testResultById\[id\]\s*=\s*\{/)
  assert.match(connectionStoreSource, /state\.sshTestResultById\[id\]\s*=\s*result\.ssh_result/)
  assert.match(connectionStoreSource, /state\.aiTestResultById\[id\]\s*=\s*aiResult/)
})

test('ConnectionsTab renders independent AI status text only when AI is configured', () => {
  assert.match(connectionsTabSource, /AI: \{\{ getAiTestStatusText\(conn\) \}\}/)
  assert.match(connectionsTabSource, /hasConfiguredAiAssistants\(conn\) && connectionStore\.getAITestResultById\(conn\.id\)/)
  assert.match(connectionsTabSource, /const getAiTestStatusText = \(conn\) =>/)
  assert.match(connectionsTabSource, /const getAiTestStatusClass = \(conn\) =>/)
})
