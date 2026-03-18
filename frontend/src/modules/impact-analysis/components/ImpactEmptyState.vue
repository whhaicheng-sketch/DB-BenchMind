<template>
  <div class="impact-empty-state">
    <div class="empty-icon">
      <span class="icon-large">⚡</span>
    </div>
    <h2 class="empty-title">No Impact Analysis Scenario Configured</h2>
    <p class="empty-description">
      To perform high availability PoC analysis, you need to create a MySQL cluster connection first.
    </p>
    <div class="empty-actions">
      <button class="btn btn-primary" @click="goToConnections">
        Create MySQL Cluster Connection
      </button>
    </div>
    <div class="empty-help">
      <h4>What you need:</h4>
      <ul>
        <li>A MySQL high availability cluster (MHA, MGR, etc.)</li>
        <li>VIP or access endpoint</li>
        <li>Primary and Secondary node IPs</li>
      </ul>
    </div>
  </div>
</template>

<script setup>
import { useAppStore } from '../../../stores/app'

const emit = defineEmits(['navigate'])

const appStore = useAppStore()

function goToConnections() {
  appStore.setActiveTab('connections')
  emit('navigate', 'connections')
}
</script>

<style scoped>
.impact-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  padding: 40px;
  text-align: center;
}

.empty-icon {
  margin-bottom: 24px;
}

.icon-large {
  font-size: 64px;
  opacity: 0.7;
}

.empty-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
}

.empty-description {
  font-size: 15px;
  color: var(--text-muted);
  max-width: 400px;
  margin-bottom: 24px;
  line-height: 1.6;
}

.empty-actions {
  margin-bottom: 32px;
}

.btn {
  padding: 10px 20px;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-primary {
  background-color: var(--primary);
  border-color: var(--primary);
  color: white;
}

.btn-primary:hover {
  background-color: var(--primary-hover);
  border-color: var(--primary-hover);
}

.empty-help {
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 20px 28px;
  text-align: left;
  max-width: 400px;
}

.empty-help h4 {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
}

.empty-help ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.empty-help li {
  font-size: 13px;
  color: var(--text-muted);
  padding: 6px 0;
  padding-left: 20px;
  position: relative;
}

.empty-help li::before {
  content: '•';
  position: absolute;
  left: 0;
  color: var(--primary);
}
</style>
