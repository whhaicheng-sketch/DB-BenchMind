import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const connectionFormSource = fs.readFileSync(path.resolve(__dirname, '../src/components/connection/ConnectionForm.vue'), 'utf8')

test('Oracle connection type selector only offers Basic and TNS', () => {
  assert.match(connectionFormSource, /<option value="basic">Basic<\/option>/)
  assert.match(connectionFormSource, /<option value="tns">TNS<\/option>/)
  assert.doesNotMatch(connectionFormSource, /<option value="service_name">Service Name<\/option>/)
  assert.doesNotMatch(connectionFormSource, /<option value="sid">SID<\/option>/)
})

test('Oracle defaults initialize connect type as Basic and identifier type as service name', () => {
  assert.match(connectionFormSource, /oracle_connect_mode:\s*'basic'/)
  assert.match(connectionFormSource, /oracle_basic_identifier_type:\s*'service_name'/)
})

test('Oracle Basic mode renders Service Name and SID radio choices', () => {
  assert.match(connectionFormSource, /value="service_name"/)
  assert.match(connectionFormSource, /value="sid"/)
  assert.match(connectionFormSource, /Service Name/)
  assert.match(connectionFormSource, /SID/)
})

test('Oracle TNS mode renders a dedicated TNS field', () => {
  assert.match(connectionFormSource, /v-if="formData\.type === 'oracle' && formData\.oracle_connect_mode === 'tns'"/)
  assert.match(connectionFormSource, /TNS/)
  assert.match(connectionFormSource, /oracle_tns_name/)
})

test('Oracle mode switch uses mode-aware validation helpers', () => {
  assert.match(connectionFormSource, /getOracleModeFieldError/)
  assert.match(connectionFormSource, /clearOracleModeSpecificFields/)
})
