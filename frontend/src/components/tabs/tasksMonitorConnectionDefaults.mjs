export function getPreferredConnectionId(connections = [], currentConnectionId = '') {
  const normalizedConnections = Array.isArray(connections) ? connections : []

  if (currentConnectionId && normalizedConnections.some((connection) => connection?.id === currentConnectionId)) {
    return currentConnectionId
  }

  return normalizedConnections[0]?.id || ''
}
