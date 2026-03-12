<template>
  <div class="toolbar">
    <button class="btn btn-primary" @click="$emit('create')">New Template</button>
    <button class="btn btn-secondary" @click="templateStore.placeholderAction('import')">Import</button>
    <button class="btn btn-secondary" @click="templateStore.placeholderAction('export')">Export</button>
  </div>
</template>

<script setup>
import { useTemplateStore } from '../../stores/template'

defineEmits(['create'])

const templateStore = useTemplateStore()
</script>

<style scoped>
.toolbar {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.btn {
  border: 1px solid #3a4a5a;
  border-radius: 8px;
  padding: 10px 16px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-primary {
  background: #3182ce;
  border-color: #3182ce;
  color: #f7fafc;
}

.btn-primary:hover {
  background: #2b6cb0;
}

.btn-secondary {
  background: #1a202c;
  color: #cbd5e0;
}

.btn-secondary:hover {
  border-color: #4299e1;
  color: #f7fafc;
}
</style>
