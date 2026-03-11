<template>
  <div class="history-tab">
    <div class="tab-header">
      <h2>History</h2>
      <button class="btn btn-primary" @click="refreshHistory">
        Refresh
      </button>
    </div>
    <div class="history-list">
      <div v-if="histories.length === 0" class="empty-state">
        No benchmark history found
      </div>
      <div v-else v-for="item in histories" :key="item.id" class="history-item">
        <div class="history-info">
          <div class="history-name">{{ item.name }}</div>
          <div class="history-date">{{ formatDate(item.start_time) }}</div>
        </div>
        <div class="history-metrics">
          <span>TPM: {{ item.tpm || 'N/A' }}</span>
          <span>TPS: {{ item.tps || 'N/A' }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const histories = ref([])

onMounted(() => {
  refreshHistory()
})

const refreshHistory = async () => {
  // TODO: Fetch from store
}

const formatDate = (date) => {
  if (!date) return ''
  return new Date(date).toLocaleString()
}
</script>

<style scoped>
.history-tab {
  height: 100%;
}

.tab-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.tab-header h2 {
  font-size: 24px;
  font-weight: 600;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.history-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background-color: #2a3a4a;
  border-radius: 8px;
}

.history-name {
  font-weight: 500;
}

.history-date {
  font-size: 13px;
  color: #a0aec0;
}

.history-metrics {
  display: flex;
  gap: 16px;
  color: #a0aec0;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: #718096;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
}

.btn-primary {
  background-color: #4299e1;
  color: white;
}
</style>
