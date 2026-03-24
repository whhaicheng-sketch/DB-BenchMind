<template>
  <div class="page-header">
    <div class="header-left">
      <h1 class="page-title">Templates</h1>
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
  padding-bottom: 6px;
  border-bottom: 1px solid var(--border-light);
  gap: var(--spacing-md);
  flex-wrap: wrap;
}

.page-title {
  font-size: 26px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}
</style>
