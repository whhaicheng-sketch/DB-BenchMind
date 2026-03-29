/**
 * Threshold configuration for Monitor metrics.
 * Defines warning/critical thresholds for each metric type.
 * Used to compute status badges, card colors, and anomaly markers.
 */

export const THRESHOLDS = {
  tps: {
    label: 'TPS',
    unit: 'ops/s',
    warning: { dropPercent: 30 },  // 30% drop from avg = warning
    critical: { dropPercent: 60 }, // 60% drop from avg = critical
    description: 'TPS drop > 30% from average'
  },
  tpm: {
    label: 'TPM',
    unit: 'ops/m',
    warning: { dropPercent: 30 },
    critical: { dropPercent: 60 },
    description: 'TPM drop > 30% from average'
  },
  cpu_user: {
    label: 'CPU User',
    unit: '%',
    warning: { value: 70 },
    critical: { value: 85 },
    description: 'CPU User > 70% = Warning, > 85% = Critical'
  },
  cpu_sys: {
    label: 'CPU System',
    unit: '%',
    warning: { value: 50 },
    critical: { value: 75 },
    description: 'CPU System > 50% = Warning, > 75% = Critical'
  },
  cpu_iowait: {
    label: 'CPU I/O Wait',
    unit: '%',
    warning: { value: 20 },
    critical: { value: 40 },
    description: 'I/O Wait > 20% = Warning, > 40% = Critical'
  },
  cpu_st: {
    label: 'CPU Steal',
    unit: '%',
    warning: { value: 10 },
    critical: { value: 25 },
    description: 'Steal > 10% = Warning, > 25% = Critical'
  },
  disk_read: {
    label: 'Disk Read',
    unit: 'MB/s',
    warning: { spikePercent: 200 },
    critical: { spikePercent: 400 },
    description: 'Disk read spike > 200% from avg'
  },
  disk_write: {
    label: 'Disk Write',
    unit: 'MB/s',
    warning: { spikePercent: 200 },
    critical: { spikePercent: 400 },
    description: 'Disk write spike > 200% from avg'
  },
  disk_latency: {
    label: 'Disk Latency',
    unit: 'ms',
    warning: { value: 10 },
    critical: { value: 50 },
    description: 'Disk latency > 10ms = Warning, > 50ms = Critical'
  }
}

/**
 * Evaluate a metric value against its threshold configuration.
 * Returns 'normal', 'warning', or 'critical'.
 */
export function evaluateThreshold(metricKey, value, avgValue) {
  const config = THRESHOLDS[metricKey]
  if (!config) return 'normal'

  // Value-based thresholds
  if (config.critical?.value != null && value >= config.critical.value) return 'critical'
  if (config.warning?.value != null && value >= config.warning.value) return 'warning'

  // Drop-percentage thresholds (TPS, TPM)
  if (config.critical?.dropPercent && avgValue > 0) {
    const dropPct = ((avgValue - value) / avgValue) * 100
    if (dropPct >= config.critical.dropPercent) return 'critical'
  }
  if (config.warning?.dropPercent && avgValue > 0) {
    const dropPct = ((avgValue - value) / avgValue) * 100
    if (dropPct >= config.warning.dropPercent) return 'warning'
  }

  // Spike-percentage thresholds (disk)
  if (config.critical?.spikePercent && avgValue > 0) {
    const spikePct = ((value - avgValue) / avgValue) * 100
    if (spikePct >= config.critical.spikePercent) return 'critical'
  }
  if (config.warning?.spikePercent && avgValue > 0) {
    const spikePct = ((value - avgValue) / avgValue) * 100
    if (spikePct >= config.warning.spikePercent) return 'warning'
  }

  return 'normal'
}

/**
 * Get the status icon for a threshold level (non-color accessibility indicator).
 */
export function thresholdIcon(level) {
  const icons = {
    normal: '✓',
    warning: '⚡',
    critical: '🔴'
  }
  return icons[level] || ''
}

/**
 * Get human-readable description for all configured thresholds.
 */
export function getThresholdDescriptions() {
  return Object.entries(THRESHOLDS).map(([key, config]) => ({
    key,
    label: config.label,
    unit: config.unit,
    description: config.description
  }))
}
