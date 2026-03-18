<template>
  <div class="cluster-status-panel">
    <div class="panel-header">
      <h3 class="panel-title">Cluster Status</h3>
      <span class="status-badge" :class="clusterHealthClass">
        {{ clusterHealthLabel }}
      </span>
    </div>

    <div class="status-grid">
      <!-- VIP -->
      <div class="status-item">
        <div class="item-icon">🌐</div>
        <div class="item-content">
          <span class="item-label">VIP</span>
          <span class="item-value">{{ clusterStatus?.vip || '--' }}</span>
        </div>
      </div>

      <!-- Current Primary -->
      <div class="status-item">
        <div class="item-icon primary">👑</div>
        <div class="item-content">
          <span class="item-label">Current Primary</span>
          <span class="item-value">{{ clusterStatus?.currentPrimary || '--' }}</span>
        </div>
        <span class="node-status" :class="primaryStatusClass">
          {{ primaryStatusLabel }}
        </span>
      </div>

      <!-- Current Secondary -->
      <div class="status-item">
        <div class="item-icon secondary">🔄</div>
        <div class="item-content">
          <span class="item-label">Current Secondary</span>
          <span class="item-value">{{ clusterStatus?.currentSecondary || '--' }}</span>
        </div>
        <span class="node-status" :class="secondaryStatusClass">
          {{ secondaryStatusLabel }}
        </span>
      </div>

      <!-- Last Role Switch Time -->
      <div class="status-item">
        <div class="item-icon">⏰</div>
        <div class="item-content">
          <span class="item-label">Last Role Switch</span>
          <span class="item-value">{{ formatLastSwitchTime }}</span>
        </div>
      </div>
    </div>

    <!-- Nodes Detail -->
    <div v-if="clusterStatus?.nodes && clusterStatus.nodes.length > 0" class="nodes-detail">
      <h4 class="nodes-title">Node Details</h4>
      <div class="nodes-list">
        <div
          v-for="node in clusterStatus.nodes"
          :key="node.nodeId"
          class="node-item"
        >
          <div class="node-info">
            <span class="node-ip">{{ node.ip }}:{{ node.port }}</span>
            <span class="node-role" :class="node.role">{{ node.role }}</span>
          </div>
          <span class="node-status-badge" :class="node.status">
            {{ node.status }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { NodeStatus, NodeRole } from '../constants'
import { formatDateTime } from '../types'

const props = defineProps({
  clusterStatus: {
    type: Object,
    default: () => null
  }
})

const primaryStatusClass = computed(() => {
  const status = props.clusterStatus?.primaryStatus
  if (status === NodeStatus.ONLINE) return 'online'
  if (status === NodeStatus.OFFLINE) return 'offline'
  return 'unknown'
})

const primaryStatusLabel = computed(() => {
  return props.clusterStatus?.primaryStatus || 'Unknown'
})

const secondaryStatusClass = computed(() => {
  const status = props.clusterStatus?.secondaryStatus
  if (status === NodeStatus.ONLINE) return 'online'
  if (status === NodeStatus.OFFLINE) return 'offline'
  return 'unknown'
})

const secondaryStatusLabel = computed(() => {
  return props.clusterStatus?.secondaryStatus || 'Unknown'
})

const clusterHealthClass = computed(() => {
  const primary = props.clusterStatus?.primaryStatus
  const secondary = props.clusterStatus?.secondaryStatus

  if (primary === NodeStatus.ONLINE && secondary === NodeStatus.ONLINE) {
    return 'healthy'
  }
  if (primary === NodeStatus.OFFLINE) {
    return 'critical'
  }
  return 'warning'
})

const clusterHealthLabel = computed(() => {
  const primary = props.clusterStatus?.primaryStatus
  const secondary = props.clusterStatus?.secondaryStatus

  if (primary === NodeStatus.ONLINE && secondary === NodeStatus.ONLINE) {
    return 'Healthy'
  }
  if (primary === NodeStatus.OFFLINE) {
    return 'Critical'
  }
  return 'Degraded'
})

const formatLastSwitchTime = computed(() => {
  if (!props.clusterStatus?.lastRoleSwitchTime) return 'Never'
  return formatDateTime(props.clusterStatus.lastRoleSwitchTime)
})
</script>

<style scoped>
.cluster-status-panel {
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 16px 20px;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.panel-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.status-badge {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
}

.status-badge.healthy {
  background-color: var(--success-bg);
  color: var(--success);
}

.status-badge.warning {
  background-color: var(--warning-bg);
  color: var(--warning);
}

.status-badge.critical {
  background-color: var(--danger-bg);
  color: var(--danger);
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background-color: var(--bg-secondary);
  border-radius: var(--radius-sm);
}

.item-icon {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  background-color: var(--bg-hover);
  border-radius: var(--radius-sm);
}

.item-icon.primary {
  background-color: var(--primary-light);
}

.item-icon.secondary {
  background-color: var(--bg-secondary);
}

.item-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.item-label {
  font-size: 11px;
  color: var(--text-muted);
  text-transform: uppercase;
}

.item-value {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  font-family: var(--font-family-mono);
}

.node-status {
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
}

.node-status.online {
  background-color: var(--success-bg);
  color: var(--success);
}

.node-status.offline {
  background-color: var(--danger-bg);
  color: var(--danger);
}

.node-status.unknown {
  background-color: var(--bg-secondary);
  color: var(--text-muted);
}

.nodes-detail {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--border-light);
}

.nodes-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  margin: 0 0 12px 0;
  text-transform: uppercase;
}

.nodes-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.node-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background-color: var(--bg-secondary);
  border-radius: var(--radius-sm);
}

.node-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.node-ip {
  font-size: 12px;
  color: var(--text-primary);
  font-family: var(--font-family-mono);
}

.node-role {
  padding: 2px 6px;
  border-radius: var(--radius-xs);
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
}

.node-role.primary {
  background-color: var(--primary-light);
  color: var(--primary);
}

.node-role.secondary {
  background-color: var(--bg-secondary);
  color: var(--text-muted);
}

.node-role.unknown {
  background-color: var(--bg-secondary);
  color: var(--text-muted);
}

.node-status-badge {
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
}

.node-status-badge.online {
  background-color: var(--success-bg);
  color: var(--success);
}

.node-status-badge.offline {
  background-color: var(--danger-bg);
  color: var(--danger);
}

.node-status-badge.unknown {
  background-color: var(--bg-secondary);
  color: var(--text-muted);
}
</style>
