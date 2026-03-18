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
  padding-bottom: var(--spacing-lg);
  border-bottom: 1px solid var(--border-light);
  gap: var(--spacing-md);
  flex-wrap: wrap;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  font-size: var(--font-size-title);
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.page-subtitle {
  font-size: var(--font-size-md);
  color: var(--text-muted);
  max-width: 680px;
  margin: 0;
}
</style>
