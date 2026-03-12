<template>
  <div class="page-header">
    <div class="header-left">
      <h1 class="page-title">Templates</h1>
      <p class="page-subtitle">Manage connection-agnostic benchmark scenario templates for future task binding.</p>
    </div>

    <TemplateToolbar @create="$emit('create')" />
  </div>
</template>

<script setup>
import TemplateToolbar from './TemplateToolbar.vue'

defineEmits(['create'])
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding-bottom: 20px;
  border-bottom: 1px solid #2d3748;
  gap: 16px;
  flex-wrap: wrap;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #f7fafc;
}

.page-subtitle {
  font-size: 14px;
  color: #718096;
  max-width: 680px;
}
</style>
