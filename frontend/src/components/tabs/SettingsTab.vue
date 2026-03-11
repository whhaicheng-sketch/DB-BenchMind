<template>
  <div class="settings-tab">
    <div class="tab-header">
      <h2>Settings</h2>
    </div>
    <div class="settings-section">
      <h3>Application Settings</h3>
      <div class="setting-item">
        <label>Theme</label>
        <select v-model="theme">
          <option value="dark">Dark</option>
          <option value="light">Light</option>
        </select>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const theme = ref('dark')
</script>

<style scoped>
.settings-tab {
  height: 100%;
}

.tab-header {
  margin-bottom: 20px;
}

.tab-header h2 {
  font-size: 24px;
  font-weight: 600;
}

.settings-section {
  background-color: #2a3a4a;
  padding: 20px;
  border-radius: 8px;
}

.settings-section h3 {
  font-size: 18px;
  margin-bottom: 16px;
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #3a4a5a;
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-item label {
  font-weight: 500;
}

.setting-item select {
  padding: 8px 12px;
  background-color: #1e2a3a;
  color: #e2e8f0;
  border: 1px solid #4a5568;
  border-radius: 4px;
}
</style>
