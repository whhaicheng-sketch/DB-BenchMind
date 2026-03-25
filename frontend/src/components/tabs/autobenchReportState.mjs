function toCount(value) {
  return Number.isFinite(value) ? value : 0
}

function buildSummaryCards(summary = {}) {
  const totalItems = toCount(summary.totalItems)
  const completedItemCount = toCount(summary.completedItemCount)
  const successItemCount = toCount(summary.successItemCount)
  const failedItemCount = toCount(summary.failedItemCount)
  const skippedItemCount = toCount(summary.skippedItemCount)
  const status = typeof summary.status === 'string' && summary.status.trim() !== ''
    ? summary.status
    : 'not_started'

  return [
    { label: 'Suite Status', value: status },
    { label: 'Completed Items', value: `${completedItemCount} / ${totalItems}` },
    { label: 'Successful Items', value: String(successItemCount) },
    { label: 'Failed Items', value: String(failedItemCount) },
    { label: 'Skipped Items', value: String(skippedItemCount) }
  ]
}

function buildFailureRows(failures) {
  if (!Array.isArray(failures)) {
    return []
  }

  return failures.map((failure, index) => ({
    id: failure?.suiteItemId || `failure-${index + 1}`,
    connectionId: failure?.connectionId || 'unknown-connection',
    profileType: failure?.profileType || 'unknown-profile',
    errorSummary: failure?.errorSummary || 'No error summary recorded.'
  }))
}

function buildExportEntries(artifactPaths = {}) {
  const htmlPath = typeof artifactPaths.html === 'string' ? artifactPaths.html : ''
  const jsonPath = typeof artifactPaths.json === 'string' ? artifactPaths.json : ''

  return [
    {
      id: 'html',
      label: 'HTML Report',
      path: htmlPath,
      available: htmlPath !== ''
    },
    {
      id: 'json',
      label: 'JSON Result',
      path: jsonPath,
      available: jsonPath !== ''
    }
  ]
}

export function buildAutoBenchReportState(reportSnapshot) {
  if (!reportSnapshot) {
    return {
      statusLabel: 'not_started',
      generatedAtLabel: 'Pending',
      summaryCards: buildSummaryCards(),
      failureRows: [],
      recommendations: ['The suite report will appear here after AutoBench produces HTML and JSON artifacts.'],
      exportEntries: buildExportEntries()
    }
  }

  const summary = reportSnapshot.summary || {}
  const recommendations = Array.isArray(reportSnapshot.recommendations) && reportSnapshot.recommendations.length > 0
    ? reportSnapshot.recommendations
    : ['No additional recommendations were generated for this suite report.']

  return {
    statusLabel: typeof summary.status === 'string' && summary.status.trim() !== ''
      ? summary.status
      : 'not_started',
    generatedAtLabel: typeof reportSnapshot.generatedAt === 'string' && reportSnapshot.generatedAt.trim() !== ''
      ? reportSnapshot.generatedAt
      : 'Pending',
    summaryCards: buildSummaryCards(summary),
    failureRows: buildFailureRows(reportSnapshot.failures),
    recommendations,
    exportEntries: buildExportEntries(reportSnapshot.artifactPaths)
  }
}
