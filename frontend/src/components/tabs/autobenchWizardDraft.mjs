export const placeholderConnections = [
  {
    id: 'placeholder-oracle',
    label: 'Oracle Placeholder',
    databaseType: 'oracle',
    detail: 'Local placeholder target for future suite binding.'
  },
  {
    id: 'placeholder-mysql',
    label: 'MySQL Placeholder',
    databaseType: 'mysql',
    detail: 'Static candidate item without connection-store wiring.'
  },
  {
    id: 'placeholder-postgresql',
    label: 'PostgreSQL Placeholder',
    databaseType: 'postgresql',
    detail: 'Static candidate item for Wizard selection flow.'
  },
  {
    id: 'placeholder-sqlserver',
    label: 'SQL Server Placeholder',
    databaseType: 'sqlserver',
    detail: 'Static candidate item for local filter display coverage.'
  }
]

export const connectionFilterOptions = [
  { id: 'all', label: 'All' },
  { id: 'oracle', label: 'Oracle' },
  { id: 'mysql', label: 'MySQL' },
  { id: 'postgresql', label: 'PostgreSQL' },
  { id: 'sqlserver', label: 'SQL Server' }
]

export const profileOptions = [
  {
    id: 'test',
    label: 'test',
    scope: 'smoke',
    description: 'Fast smoke profile for early validation before heavier workload coverage.'
  },
  {
    id: 'cpu_bound',
    label: 'cpu_bound',
    scope: 'throughput',
    description: 'CPU-oriented profile placeholder for sustained compute pressure coverage.'
  },
  {
    id: 'io_bound',
    label: 'io_bound',
    scope: 'storage',
    description: 'I/O-oriented profile placeholder for storage and latency coverage.'
  }
]

export const policySummaryItems = [
  { label: 'Execution mode', value: 'serial' },
  { label: 'Failure policy', value: 'continue_by_connection' },
  { label: 'Cleanup', value: 'true' },
  { label: 'Default profile order', value: 'test -> cpu_bound -> io_bound' }
]

const profileOrder = profileOptions.map((item) => item.id)

function toggleValue(list, value, order = []) {
  const hasValue = list.includes(value)
  const next = hasValue ? list.filter((item) => item !== value) : [...list, value]

  if (order.length === 0) {
    return next
  }

  return order.filter((item) => next.includes(item))
}

export function createAutoBenchWizardDraft() {
  return {
    selectedConnectionIds: [],
    selectedProfiles: [...profileOrder],
    executionMode: 'serial',
    failurePolicy: 'continue_by_connection',
    cleanupEnabled: true
  }
}

export function toggleDraftConnectionSelection(draft, connectionId) {
  return {
    ...draft,
    selectedConnectionIds: toggleValue(draft.selectedConnectionIds, connectionId)
  }
}

export function toggleDraftProfileSelection(draft, profileId) {
  return {
    ...draft,
    selectedProfiles: toggleValue(draft.selectedProfiles, profileId, profileOrder)
  }
}

export function validateAutoBenchWizardDraft(draft) {
  const connectionError = draft.selectedConnectionIds.length > 0 ? '' : 'Select at least one placeholder connection to continue.'
  const profileError = draft.selectedProfiles.length > 0 ? '' : 'Select at least one profile to continue.'

  return {
    canCreateSuite: connectionError === '' && profileError === '',
    connectionError,
    profileError
  }
}

export function filterPlaceholderConnections(connections, activeDatabaseType) {
  if (activeDatabaseType === 'all') {
    return connections
  }

  return connections.filter((connection) => connection.databaseType === activeDatabaseType)
}

export function describeSelectedProfiles(selectedProfiles) {
  if (!selectedProfiles || selectedProfiles.length === 0) {
    return 'none'
  }

  return profileOrder.filter((profileId) => selectedProfiles.includes(profileId)).join(' -> ')
}

export function buildLocalPlanPreview(draft, connections) {
  const selectedProfiles = profileOrder.filter((profileId) => draft.selectedProfiles.includes(profileId))
  const selectedConnections = connections.filter((connection) => draft.selectedConnectionIds.includes(connection.id))

  if (selectedProfiles.length === 0 || selectedConnections.length === 0) {
    return {
      totalItems: 0,
      items: [],
      summary: {
        executionMode: draft.executionMode,
        failurePolicy: draft.failurePolicy,
        cleanupEnabled: draft.cleanupEnabled
      }
    }
  }

  const items = selectedConnections.flatMap((connection) =>
    selectedProfiles.map((profileId, index) => ({
      id: `${connection.id}:${profileId}`,
      order: index + 1,
      connectionId: connection.id,
      connectionLabel: connection.label,
      databaseType: connection.databaseType,
      profileId
    }))
  )

  return {
    totalItems: items.length,
    items,
    summary: {
      executionMode: draft.executionMode,
      failurePolicy: draft.failurePolicy,
      cleanupEnabled: draft.cleanupEnabled
    }
  }
}
