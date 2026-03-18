<template>
  <div class="impact-analysis-page">
    <!-- Toolbar -->
    <ImpactAnalysisToolbar
      :connections="store.mysqlConnections"
      :selected-connection-id="store.selectedConnectionId"
      :selected-connection="store.selectedConnection"
      :config="store.config"
      :is-analyzing="store.isAnalyzing"
      :session="store.session"
      @update:connection-id="store.selectConnection"
      @update:config="store.updateConfig"
      @start="handleStart"
      @stop="handleStop"
      @reset="handleReset"
    />

    <!-- Main Content Grid -->
    <div class="page-content">
      <!-- Left Column: Summary + Chart -->
      <div class="left-column">
        <!-- Summary Cards -->
        <ImpactSummaryCards :summary-data="store.summaryData" />

        <!-- Trend Chart -->
        <ImpactTrendChart
          :trend-data="store.recentTrendData"
          :events="store.events"
        />
      </div>

      <!-- Right Column: Cluster Status + Event Stream -->
      <div class="right-column">
        <!-- Cluster Status Panel -->
        <ClusterStatusPanel :cluster-status="store.clusterStatus" />

        <!-- Event Stream -->
        <ImpactEventStream :events="store.recentEvents" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue'
import { useImpactAnalysisStore } from '../store/impactAnalysis'
import ImpactAnalysisToolbar from '../components/ImpactAnalysisToolbar.vue'
import ImpactSummaryCards from '../components/ImpactSummaryCards.vue'
import ImpactTrendChart from '../components/ImpactTrendChart.vue'
import ClusterStatusPanel from '../components/ClusterStatusPanel.vue'
import ImpactEventStream from '../components/ImpactEventStream.vue'

const store = useImpactAnalysisStore()

onMounted(async () => {
  await store.initialize()
})

onUnmounted(() => {
  store.stopRealtimeUpdates()
})

async function handleStart() {
  await store.startAnalysis()
}

async function handleStop() {
  await store.stopAnalysis()
}

function handleReset() {
  store.resetSession()
}
</script>

<style scoped>
.impact-analysis-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.page-content {
  display: grid;
  grid-template-columns: 1fr 380px;
  gap: 20px;
  flex: 1;
  overflow: hidden;
  min-height: 0;
}

.left-column {
  display: flex;
  flex-direction: column;
  gap: 20px;
  overflow: hidden;
}

.right-column {
  display: flex;
  flex-direction: column;
  gap: 20px;
  overflow: hidden;
}

/* Ensure event stream fills remaining space */
.right-column > *:last-child {
  flex: 1;
  min-height: 0;
}

@media (max-width: 1200px) {
  .page-content {
    grid-template-columns: 1fr;
  }

  .right-column {
    flex-direction: row;
    flex-wrap: wrap;
  }

  .right-column > * {
    flex: 1;
    min-width: 300px;
  }
}
</style>
