<template>
  <div class="templates-tab">
    <div class="tab-header">
      <h2>Templates</h2>
    </div>
    <TemplateList
      v-model="selectedTemplateId"
      :disabled="isRunning"
      db-type=""
      @template-selected="handleTemplateSelected"
    />
    <div v-if="selectedTemplate" class="template-details">
      <h3>{{ selectedTemplate.name }}</h3>
      <p>{{ selectedTemplate.description }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useTemplateStore } from '../../stores/template'
import { useBenchmarkStore } from '../../stores/benchmark'
import TemplateList from '../template/TemplateList.vue'

const templateStore = useTemplateStore()
const benchmarkStore = useBenchmarkStore()

const selectedTemplateId = ref('')

const selectedTemplate = computed(() => templateStore.selectedTemplate)
const isRunning = computed(() => benchmarkStore.isRunning)

const handleTemplateSelected = (template) => {
  console.log('Selected template:', template)
}
</script>

<style scoped>
.templates-tab {
  height: 100%;
}

.tab-header {
  margin-bottom: 20px;
}

.tab-header h2 {
  font-size: 24px;
  font-weight: 600;
}

.template-details {
  margin-top: 20px;
  padding: 16px;
  background-color: #2a3a4a;
  border-radius: 8px;
}

.template-details h3 {
  font-size: 18px;
  margin-bottom: 8px;
}

.template-details p {
  color: #a0aec0;
}
</style>
