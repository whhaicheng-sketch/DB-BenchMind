function normalizeItems(items) {
  return Array.isArray(items) ? items : []
}

function pickCurrentItem(items) {
  return items.find((item) => item.status === 'running')
    || items.find((item) => item.status === 'validating')
    || items.find((item) => item.status === 'preparing')
    || items.find((item) => item.status === 'cleaning')
    || null
}

function toItemRows(items) {
  return items.map((item) => ({
    id: item.id,
    connectionId: item.connectionId || '',
    profileType: item.profileType || '',
    status: item.status || 'pending',
    phaseLabel: item.phaseStatus || item.phase || 'none',
    logLabel: 'Logs available later',
    resultSummary: item.resultSummary || ''
  }))
}

export function buildAutoBenchMonitorState(statusSnapshot) {
  if (!statusSnapshot) {
    return {
      statusLabel: 'idle',
      progressPercent: 0,
      completedLabel: '0 / 0 completed',
      currentItem: null,
      itemRows: [],
      emptyMessage: 'No AutoBench suite has been started yet.'
    }
  }

  const totalItems = Number(statusSnapshot.totalItems || 0)
  const completedItems = Number(statusSnapshot.completedItems || 0)
  const items = normalizeItems(statusSnapshot.items)

  return {
    statusLabel: statusSnapshot.status || 'idle',
    progressPercent: totalItems > 0 ? Math.round((completedItems / totalItems) * 100) : 0,
    completedLabel: `${completedItems} / ${totalItems} completed`,
    currentItem: pickCurrentItem(items),
    itemRows: toItemRows(items),
    emptyMessage: ''
  }
}
