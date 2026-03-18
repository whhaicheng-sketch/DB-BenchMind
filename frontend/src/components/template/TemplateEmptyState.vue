<template>
  <div class="empty-state">
    <div class="empty-icon">Templates</div>
    <h3 class="empty-title">{{ title }}</h3>
    <p class="empty-description">{{ description }}</p>
    <div v-if="primaryLabel || secondaryLabel" class="empty-actions">
      <button v-if="primaryLabel" class="btn btn-primary" @click="$emit('primary')">{{ primaryLabel }}</button>
      <button v-if="secondaryLabel" class="btn btn-secondary" @click="$emit('secondary')">{{ secondaryLabel }}</button>
    </div>
  </div>
</template>

<script setup>
defineProps({
  title: {
    type: String,
    required: true
  },
  description: {
    type: String,
    required: true
  },
  primaryLabel: {
    type: String,
    default: ''
  },
  secondaryLabel: {
    type: String,
    default: ''
  }
})

defineEmits(['primary', 'secondary'])
</script>

<style scoped>
.empty-state {
  max-width: 420px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-sm);
}

.empty-icon {
  padding: 8px 16px;
  border-radius: 20px;
  background-color: var(--primary-light);
  color: var(--primary);
  font-size: var(--font-size-xs);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.empty-title {
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.empty-description {
  color: var(--text-muted);
  line-height: 1.6;
  font-size: var(--font-size-sm);
  margin: 0;
}

.empty-actions {
  display: flex;
  gap: var(--spacing-sm);
  flex-wrap: wrap;
  justify-content: center;
  margin-top: var(--spacing-sm);
}

.btn {
  display: inline-flex;
  align-items: center;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 8px 16px;
  font-size: var(--font-size-sm);
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

.btn-secondary {
  background-color: var(--bg-primary);
  color: var(--text-secondary);
}

.btn-secondary:hover {
  background-color: var(--bg-secondary);
  border-color: var(--border-dark);
  color: var(--text-primary);
}
</style>
