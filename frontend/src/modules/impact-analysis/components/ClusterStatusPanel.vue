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
  background-color: #1a2332;
  border: 1px solid #2a3a4a;
  border-radius: 8px;
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
  color: #e2e8f0;
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
  background-color: rgba(72, 187, 120, 0.2);
  color: #68d391;
}

.status-badge.warning {
  background-color: rgba(236, 201, 75, 0.2);
  color: #f6e05e;
}

.status-badge.critical {
  background-color: rgba(245, 101, 101, 0.2);
  color: #fc8181;
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
  background-color: rgba(0, 0, 0, 0.2);
  border-radius: 6px;
}

.item-icon {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  background-color: rgba(255, 255, 255, 0.05);
  border-radius: 6px;
}

.item-icon.primary {
  background-color: rgba(66, 153, 225, 0.2);
}

.item-icon.secondary {
  background-color: rgba(160, 174, 192, 0.2);
}

.item-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.item-label {
  font-size: 11px;
  color: #718096;
  text-transform: uppercase;
}

.item-value {
  font-size: 13px;
  font-weight: 500;
  color: #e2e8f0;
  font-family: 'SF Mono', Monaco, monospace;
}

.node-status {
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
}

.node-status.online {
  background-color: rgba(72, 187, 120, 0.2);
  color: #68d391;
}

.node-status.offline {
  background-color: rgba(245, 101, 101, 0.2);
  color: #fc8181;
}

.node-status.unknown {
  background-color: rgba(160, 174, 192, 0.2);
  color: #a0aec0;
}

.nodes-detail {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #2a3a4a;
}

.nodes-title {
  font-size: 12px;
  font-weight: 600;
  color: #a0aec0;
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
  background-color: rgba(0, 0, 0, 0.2);
  border-radius: 4px;
}

.node-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.node-ip {
  font-size: 12px;
  color: #e2e8f0;
  font-family: 'SF Mono', Monaco, monospace;
}

.node-role {
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
}

.node-role.primary {
  background-color: rgba(66, 153, 225, 0.2);
  color: #63b3ed;
}

.node-role.secondary {
  background-color: rgba(160, 174, 192, 0.2);
  color: #a0aec0;
}

.node-role.unknown {
  background-color: rgba(160, 174, 192, 0.1);
  color: #718096;
}

.node-status-badge {
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
}

.node-status-badge.online {
  background-color: rgba(72, 187, 120, 0.2);
  color: #68d391;
}

.node-status-badge.offline {
  background-color: rgba(245, 101, 101, 0.2);
  color: #fc8181;
}

.node-status-badge.unknown {
  background-color: rgba(160, 174, 192, 0.2);
  color: #a0aec0;
}
</style>
