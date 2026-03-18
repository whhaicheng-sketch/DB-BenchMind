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
          <h1 class="page-title">Impact Analysis</h1>
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
  gap: var(--spacing-md);
  color: var(--text-muted);
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border-color);
  border-top-color: var(--primary);
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
  margin-bottom: var(--spacing-lg);
}

.page-title {
  font-size: var(--font-size-title);
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 var(--spacing-xs) 0;
}

.page-subtitle {
  font-size: var(--font-size-md);
  color: var(--text-muted);
  margin: 0;
  line-height: 1.5;
}
</style>
