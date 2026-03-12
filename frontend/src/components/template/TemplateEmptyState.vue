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
  gap: 10px;
}

.empty-icon {
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(66, 153, 225, 0.16);
  color: #93c5fd;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.empty-title {
  font-size: 20px;
  color: #f8fafc;
}

.empty-description {
  color: #94a3b8;
  line-height: 1.6;
  font-size: 13px;
}

.empty-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  justify-content: center;
  margin-top: 6px;
}

.btn {
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 10px 14px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.btn-primary {
  background: #3182ce;
  border-color: #3182ce;
  color: #fff;
}

.btn-secondary {
  background: #111827;
  color: #cbd5e0;
}
</style>
