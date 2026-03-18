<template>
  <div class="impact-analysis-tab">
    <!-- Loading State -->
    <div v-if="store.loading && !store.mysqlConnections.length" class="loading-state">
      <div class="spinner"></div>
      <span>Loading Impact Analysis...</span>
    </div>

    <!-- No MySQL Cluster Connection State -->
    <ImpactEmptyState
      v-else-if="store.pageState === 'no_connection'"
      @navigate="handleNavigate"
    />

    <!-- Ready / Analyzing / Completed State -->
    <div v-else class="tab-content-wrapper">
      <!-- Page Header -->
      <div class="page-header">
        <div class="header-left">
          <h2 class="page-title">Impact Analysis</h2>
          <p class="page-subtitle">
            Real-time analysis for MySQL high availability PoC. Monitor RTO, business interruption, and data consistency during failover scenarios.
          </p>
        </div>
      </div>

      <!-- Main Page Component -->
      <ImpactAnalysisPage />
    </div>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue'
import { useImpactAnalysisStore } from '../../modules/impact-analysis/store/impactAnalysis'
import ImpactEmptyState from '../../modules/impact-analysis/components/ImpactEmptyState.vue'
import ImpactAnalysisPage from '../../modules/impact-analysis/pages/ImpactAnalysisPage.vue'

const store = useImpactAnalysisStore()

onMounted(async () => {
  await store.initialize()
})

onUnmounted(() => {
  store.stopRealtimeUpdates()
})

function handleNavigate(tab) {
  // Navigation is handled by ImpactEmptyState via appStore
  console.log('Navigate to:', tab)
}
</script>

<style scoped>
.impact-analysis-tab {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 16px;
  color: #a0aec0;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #2a3a4a;
  border-top-color: #4299e1;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.tab-content-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.page-header {
  margin-bottom: 20px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #e2e8f0;
  margin: 0 0 8px 0;
}

.page-subtitle {
  font-size: 14px;
  color: #718096;
  margin: 0;
  line-height: 1.5;
}
</style>
