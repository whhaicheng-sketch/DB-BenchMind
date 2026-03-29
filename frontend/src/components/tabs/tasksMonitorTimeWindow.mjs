/**
 * Time window controls for Monitor charts.
 * Provides cropping logic so only points within the selected window are shown.
 */

export const TIME_WINDOWS = [
  { label: '30s', value: 30 },
  { label: '2m', value: 120 },
  { label: '10m', value: 600 },
  { label: 'All', value: 0 }
]

export const DEFAULT_TIME_WINDOW = 120 // 2 minutes

/**
 * Crop a series array to only include data points within the time window.
 * Points are assumed to be ordered oldest-first, sampled at 1s intervals.
 * @param {Array} series - Array of { value, ... } objects
 * @param {number} windowSeconds - Window size in seconds; 0 = show all
 * @returns {Array} Cropped series
 */
export function cropSeriesToWindow(series, windowSeconds) {
  if (!Array.isArray(series) || series.length === 0) return series
  if (!windowSeconds || windowSeconds <= 0) return series

  const maxPoints = windowSeconds
  if (series.length <= maxPoints) return series

  return series.slice(series.length - maxPoints)
}

/**
 * Crop multiple metric objects (each having a .series array) at once.
 * @param {Object} metrics - { tps: { series, ... }, tpm: { series, ... }, ... }
 * @param {number} windowSeconds - Window size
 * @returns {Object} New metrics object with cropped series
 */
export function cropMetricsToWindow(metrics, windowSeconds) {
  if (!metrics || !windowSeconds) return metrics

  const cropped = {}
  for (const [key, metric] of Object.entries(metrics)) {
    if (metric?.series) {
      cropped[key] = {
        ...metric,
        series: cropSeriesToWindow(metric.series, windowSeconds)
      }
    } else {
      cropped[key] = metric
    }
  }
  return cropped
}

/**
 * Build x-coordinate for a point index given cropped series length and plot bounds.
 * Maps point index to x position within the chart's plot area.
 */
export function pointX(index, totalPoints, plotBounds) {
  if (totalPoints <= 1) return plotBounds.x1
  const ratio = index / (totalPoints - 1)
  return plotBounds.x1 + ratio * (plotBounds.x2 - plotBounds.x1)
}

/**
 * Build y-coordinate for a value given min/max and plot bounds.
 */
export function pointY(value, minVal, maxVal, plotBounds) {
  if (maxVal === minVal) return (plotBounds.y1 + plotBounds.y2) / 2
  const ratio = (value - minVal) / (maxVal - minVal)
  return plotBounds.y2 - ratio * (plotBounds.y2 - plotBounds.y1)
}
