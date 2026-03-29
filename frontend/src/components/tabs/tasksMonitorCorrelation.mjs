/**
 * Cross-chart correlation and anomaly detection for Monitor.
 * Computes unified hover index, anomaly markers, and correlation data.
 */

import { evaluateThreshold } from './tasksMonitorThresholds.mjs'

/**
 * Given multiple metric series, find anomaly points across all of them.
 * Returns an array of { index, timestamp, anomalies: [{ metric, value, level }] }
 * Only returns points that have at least one warning or critical anomaly.
 *
 * @param {Object} allSeries - { tps: [...], cpu_user: [...], disk_read: [...], ... }
 * @param {Object} avgValues - { tps: 100, cpu_user: 45, ... } average values for each metric
 * @returns {Array} Anomaly records sorted by index
 */
export function detectAnomalies(allSeries, avgValues) {
  if (!allSeries || !avgValues) return []

  const maxLen = Math.max(...Object.values(allSeries).map(s => s?.length || 0))
  if (maxLen === 0) return []

  const anomalies = []

  for (let i = 0; i < maxLen; i++) {
    const pointAnomalies = []

    for (const [metricKey, series] of Object.entries(allSeries)) {
      if (!Array.isArray(series) || i >= series.length) continue
      const value = series[i]?.value
      if (value == null || value === undefined) continue

      const avg = avgValues[metricKey] || 0
      const level = evaluateThreshold(metricKey, value, avg)
      if (level !== 'normal') {
        pointAnomalies.push({ metric: metricKey, value, level, avg })
      }
    }

    if (pointAnomalies.length > 0) {
      anomalies.push({
        index: i,
        anomalies: pointAnomalies,
        worstLevel: pointAnomalies.some(a => a.level === 'critical') ? 'critical' : 'warning'
      })
    }
  }

  return anomalies
}

/**
 * Build a unified hover index that maps to the same relative position
 * across all charts. Used for cross-chart hover synchronization.
 *
 * @param {number} hoverIndex - The index in the reference series
 * @param {number} referenceLength - Length of the reference series
 * @param {number} targetLength - Length of the target series
 * @returns {number} The corresponding index in the target series
 */
export function mapHoverIndex(hoverIndex, referenceLength, targetLength) {
  if (referenceLength <= 1 || targetLength <= 0) return -1
  const ratio = hoverIndex / (referenceLength - 1)
  return Math.round(ratio * (targetLength - 1))
}

/**
 * Extract values at a unified hover position across all metrics.
 * Returns { metricKey: value } or null if hoverIndex is out of range.
 *
 * @param {Object} allSeries - { tps: [...], cpu_user: [...], ... }
 * @param {number} hoverIndex - Index in the primary (longest) series
 * @returns {Object|null} Values at hover point
 */
export function getValuesAtHover(allSeries, hoverIndex) {
  if (hoverIndex < 0 || !allSeries) return null

  const values = {}
  const maxLen = Math.max(...Object.values(allSeries).map(s => s?.length || 0))

  for (const [key, series] of Object.entries(allSeries)) {
    if (!Array.isArray(series)) continue
    const targetIndex = mapHoverIndex(hoverIndex, maxLen, series.length)
    if (targetIndex >= 0 && targetIndex < series.length) {
      values[key] = series[targetIndex]?.value ?? null
    }
  }

  return values
}

/**
 * Build avgValues map from metric objects that have .avg property.
 */
export function extractAvgValues(metrics) {
  const avgs = {}
  if (!metrics) return avgs

  // Business metrics (tps, tpm)
  for (const key of ['tps', 'tpm']) {
    if (metrics[key]?.avg != null) {
      avgs[key] = metrics[key].avg
    }
  }

  // System metrics are nested differently - extract from raw series
  return avgs
}

/**
 * Compute system metric averages from series data.
 */
export function computeSystemAvgs(systemMetrics) {
  const avgs = {}
  if (!systemMetrics) return avgs

  for (const [key, data] of Object.entries(systemMetrics)) {
    if (Array.isArray(data)) {
      const values = data.filter(d => d?.value != null).map(d => d.value)
      if (values.length > 0) {
        avgs[key] = values.reduce((a, b) => a + b, 0) / values.length
      }
    }
  }

  return avgs
}
