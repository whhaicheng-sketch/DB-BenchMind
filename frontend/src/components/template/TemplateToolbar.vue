<template>
  <div class="toolbar">
    <button class="btn btn-primary" @click="$emit('create')">New Template</button>
  </div>
</template>

<script setup>
defineEmits(['create'])
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

</style>
