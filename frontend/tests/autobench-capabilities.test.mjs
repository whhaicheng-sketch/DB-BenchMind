import { describe, it } from 'node:test'
import assert from 'node:assert/strict'

/**
 * AutoBench capability flags tests
 *
 * Tests that connCapabilitiesMap correctly determines SSH, WinRM, and AI
 * capability flags based on connection properties.
 * Also tests that capabilities are *not* shown when flags are absent.
 */

// ===========================================================================
// connCapabilitiesMap (mirrors AutoBenchTab.vue connCapabilitiesMap computed)
// ===========================================================================

function connCapabilitiesMap(connections) {
  const map = {}
  for (const conn of connections) {
    const caps = []
    if (conn.ssh_enabled) caps.push('SSH')
    if (conn.winrm_enabled) caps.push('WinRM')
    // Only show AI if there are assistants with a configured provider
    if (conn.ai_assistants && conn.ai_assistants.some(a => a.provider)) caps.push('AI')
    map[conn.id] = caps
  }
  return map
}

// ===========================================================================
// Tests
// ===========================================================================

describe('connCapabilitiesMap', () => {
  it('should include SSH when ssh_enabled is true', () => {
    const connections = [
      { id: 'conn-1', ssh_enabled: true, winrm_enabled: false }
    ]
    const result = connCapabilitiesMap(connections)
    assert.deepStrictEqual(result['conn-1'], ['SSH'])
  })

  it('should include WinRM when winrm_enabled is true', () => {
    const connections = [
      { id: 'conn-2', ssh_enabled: false, winrm_enabled: true }
    ]
    const result = connCapabilitiesMap(connections)
    assert.deepStrictEqual(result['conn-2'], ['WinRM'])
  })

  it('should include AI when ai_assistants has an entry with a provider', () => {
    const connections = [
      { id: 'conn-3', ssh_enabled: false, winrm_enabled: false, ai_assistants: [{ provider: 'openai' }] }
    ]
    const result = connCapabilitiesMap(connections)
    assert.deepStrictEqual(result['conn-3'], ['AI'])
  })

  it('should not include AI when ai_assistants is empty', () => {
    const connections = [
      { id: 'conn-4', ssh_enabled: false, winrm_enabled: false, ai_assistants: [] }
    ]
    const result = connCapabilitiesMap(connections)
    assert.deepStrictEqual(result['conn-4'], [])
  })

  it('should not include AI when ai_assistants entries have no provider', () => {
    const connections = [
      { id: 'conn-5', ssh_enabled: false, winrm_enabled: false, ai_assistants: [{ provider: '' }] }
    ]
    const result = connCapabilitiesMap(connections)
    assert.deepStrictEqual(result['conn-5'], [])
  })

  it('should not include AI when ai_assistants is undefined', () => {
    const connections = [
      { id: 'conn-6', ssh_enabled: false, winrm_enabled: false }
    ]
    const result = connCapabilitiesMap(connections)
    assert.deepStrictEqual(result['conn-6'], [])
  })

  it('should include all three capabilities when all flags are set', () => {
    const connections = [
      { id: 'conn-7', ssh_enabled: true, winrm_enabled: true, ai_assistants: [{ provider: 'openai' }] }
    ]
    const result = connCapabilitiesMap(connections)
    assert.deepStrictEqual(result['conn-7'], ['SSH', 'WinRM', 'AI'])
  })

  it('should return empty array when no capabilities are enabled', () => {
    const connections = [
      { id: 'conn-8', ssh_enabled: false, winrm_enabled: false, ai_assistants: [] }
    ]
    const result = connCapabilitiesMap(connections)
    assert.deepStrictEqual(result['conn-8'], [])
  })

  it('should handle multiple connections independently', () => {
    const connections = [
      { id: 'conn-a', ssh_enabled: true, winrm_enabled: false },
      { id: 'conn-b', ssh_enabled: false, winrm_enabled: true },
      { id: 'conn-c', ssh_enabled: false, winrm_enabled: false, ai_assistants: [{ provider: 'anthropic' }] }
    ]
    const result = connCapabilitiesMap(connections)
    assert.deepStrictEqual(result['conn-a'], ['SSH'])
    assert.deepStrictEqual(result['conn-b'], ['WinRM'])
    assert.deepStrictEqual(result['conn-c'], ['AI'])
  })

  it('should return empty object for empty connections array', () => {
    const result = connCapabilitiesMap([])
    assert.deepStrictEqual(result, {})
  })
})
