import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import {
  applyRemoteTypeCompatibilityFields,
  getDefaultWinRMPortByScheme,
  getRemoteType,
  getVisibleRemoteBlockingErrors,
  isRemoteTypeNone,
  isRemoteTypeSSH,
  isRemoteTypeWinRM,
  shouldAutoUpdateWinRMPort,
  syncRemoteHostFromGeneral
} from '../src/components/connection/connectionFormRemoteState.mjs'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const connectionFormSource = fs.readFileSync(path.resolve(__dirname, '../src/components/connection/ConnectionForm.vue'), 'utf8')
const remoteSectionSource = fs.readFileSync(path.resolve(__dirname, '../src/components/connection/ConnectionRemoteSection.vue'), 'utf8')
const stateSource = fs.readFileSync(path.resolve(__dirname, '../src/components/connection/useConnectionFormState.mjs'), 'utf8')
const connectionsTabSource = fs.readFileSync(path.resolve(__dirname, '../src/components/tabs/ConnectionsTab.vue'), 'utf8')
const connectionStoreSource = fs.readFileSync(path.resolve(__dirname, '../src/stores/connection.js'), 'utf8')

test('remote type compatibility prefers SSH for dirty historical data and normalizes mapping fields', () => {
  assert.equal(getRemoteType({ ssh_enabled: false, winrm_enabled: false }), 'none')
  assert.equal(getRemoteType({ ssh_enabled: true, winrm_enabled: false }), 'ssh')
  assert.equal(getRemoteType({ ssh_enabled: false, winrm_enabled: true }), 'winrm')
  assert.equal(getRemoteType({ ssh_enabled: true, winrm_enabled: true }), 'ssh')

  assert.deepEqual(applyRemoteTypeCompatibilityFields({ remote_type: 'none' }), {
    remote_type: 'none',
    ssh_enabled: false,
    winrm_enabled: false
  })
  assert.deepEqual(applyRemoteTypeCompatibilityFields({ remote_type: 'ssh' }), {
    remote_type: 'ssh',
    ssh_enabled: true,
    winrm_enabled: false
  })
  assert.deepEqual(applyRemoteTypeCompatibilityFields({ remote_type: 'winrm' }), {
    remote_type: 'winrm',
    ssh_enabled: false,
    winrm_enabled: true
  })
})

test('remote type helpers expose mutually exclusive semantic checks', () => {
  assert.equal(isRemoteTypeNone({ remote_type: 'none' }), true)
  assert.equal(isRemoteTypeSSH({ remote_type: 'ssh' }), true)
  assert.equal(isRemoteTypeWinRM({ remote_type: 'winrm' }), true)
  assert.equal(isRemoteTypeSSH({ remote_type: 'winrm' }), false)
  assert.equal(isRemoteTypeWinRM({ remote_type: 'ssh' }), false)
})

test('remote host sync updates the active remote host field only', () => {
  assert.deepEqual(
    syncRemoteHostFromGeneral(
      {
        remote_type: 'ssh',
        host: '10.0.0.8',
        ssh_host: '',
        winrm_host: 'old-winrm'
      }
    ),
    {
      remote_type: 'ssh',
      host: '10.0.0.8',
      ssh_host: '10.0.0.8',
      winrm_host: 'old-winrm'
    }
  )

  assert.deepEqual(
    syncRemoteHostFromGeneral(
      {
        remote_type: 'winrm',
        host: '10.0.0.9',
        ssh_host: 'old-ssh',
        winrm_host: ''
      }
    ),
    {
      remote_type: 'winrm',
      host: '10.0.0.9',
      ssh_host: 'old-ssh',
      winrm_host: '10.0.0.9'
    }
  )
})

test('WinRM default port helpers only auto-update when the user has not overridden the port', () => {
  assert.equal(getDefaultWinRMPortByScheme('http'), 5985)
  assert.equal(getDefaultWinRMPortByScheme('https'), 5986)
  assert.equal(shouldAutoUpdateWinRMPort({ remote_port_user_overridden: false }), true)
  assert.equal(shouldAutoUpdateWinRMPort({ remote_port_user_overridden: true }), false)
})

test('remote blocking validation only checks the active remote type fields', () => {
  assert.deepEqual(
    getVisibleRemoteBlockingErrors({
      remote_type: 'none',
      ssh_host: '',
      ssh_port: 0,
      ssh_username: '',
      winrm_host: '',
      winrm_port: 0,
      winrm_username: ''
    }),
    {}
  )

  assert.deepEqual(
    getVisibleRemoteBlockingErrors({
      remote_type: 'ssh',
      ssh_host: '',
      ssh_port: 0,
      ssh_username: '',
      winrm_host: '',
      winrm_port: 5985,
      winrm_username: 'Administrator'
    }),
    {
      ssh_host: '请填写必填项：SSH 主机',
      ssh_port: '请填写必填项：SSH 端口',
      ssh_username: '请填写必填项：SSH 用户名'
    }
  )

  assert.deepEqual(
    getVisibleRemoteBlockingErrors({
      remote_type: 'winrm',
      ssh_host: 'ignored',
      ssh_port: 22,
      ssh_username: 'ignored',
      winrm_host: '',
      winrm_port: 0,
      winrm_username: ''
    }),
    {
      winrm_host: '请填写必填项：WinRM 主机',
      winrm_port: '请填写必填项：WinRM 端口',
      winrm_username: '请填写必填项：WinRM 用户名'
    }
  )
})

test('ConnectionForm renders Remote tab semantics instead of the legacy SSH-only tab', () => {
  assert.match(stateSource, /\{ id: 'remote', label: 'Remote' \}/)
  assert.match(stateSource, /remote_type/)
  assert.match(connectionFormSource, /Test WinRM/)
  assert.match(remoteSectionSource, /不使用远程连接/)
  assert.match(remoteSectionSource, /WinRM/)
  assert.doesNotMatch(stateSource, /\{ id: 'ssh', label: 'SSH' \}/)
})

test('ConnectionsTab derives remote badges and remote status from remote_type rather than raw enabled flags', () => {
  assert.match(connectionsTabSource, /getRemoteType\(conn\)/)
  assert.doesNotMatch(connectionsTabSource, /<span v-if="conn\.ssh_enabled" class="tag tag-ssh">SSH<\/span>/)
  assert.doesNotMatch(connectionsTabSource, /<span v-if="conn\.winrm_enabled" class="tag tag-winrm">WinRM<\/span>/)
})

test('connection store aggregated test routes remote checks by remote_type and keeps AI test intact', () => {
  assert.match(connectionStoreSource, /getRemoteType\(connection\)/)
  assert.match(connectionStoreSource, /if \(remoteType === 'winrm'\)/)
  assert.match(connectionStoreSource, /testWinRMConnection|TestWinRMConnection/)
  assert.match(connectionStoreSource, /TestAIConnection/)
})
