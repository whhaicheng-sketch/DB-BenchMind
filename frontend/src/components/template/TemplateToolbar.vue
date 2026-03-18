<template>
  <div class="toolbar">
    <button class="btn btn-primary" @click="$emit('create')">
      <span class="btn-icon">+</span>
      <span>New Template</span>
    </button>
  </div>
</template>

<script setup>
defineEmits(['create'])
</script>

<style scoped>
.toolbar {
  display: flex;
  gap: var(--spacing-sm);
  flex-wrap: wrap;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 8px 16px;
  font-size: var(--font-size-sm);
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  background-color: var(--bg-primary);
  color: var(--text-primary);
}

.btn:hover {
  border-color: var(--border-dark);
  background-color: var(--bg-secondary);
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

.btn-icon {
  font-size: var(--font-size-md);
  font-weight: 600;
}
</style>
