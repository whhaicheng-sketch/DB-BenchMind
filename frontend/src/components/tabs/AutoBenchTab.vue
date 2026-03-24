<template>
  <section class="autobench-page">
    <header class="page-header">
      <div>
        <h1 class="page-title">AutoBench</h1>
        <p class="page-subtitle">Automated database test suites will be orchestrated from this standalone workspace.</p>
      </div>
    </header>

    <div class="placeholder-card">
      <h2>New Module</h2>
      <p>AutoBench is delivered as a separate page and does not modify the existing Performance Analysis workflow.</p>
    </div>
  </section>
</template>

<style scoped>
.autobench-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.page-title {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
}

.page-subtitle {
  margin-top: 8px;
  color: var(--text-secondary);
  max-width: 720px;
  line-height: 1.5;
}

.placeholder-card {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 24px;
  box-shadow: var(--shadow-sm);
}

.placeholder-card h2 {
  font-size: 18px;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.placeholder-card p {
  color: var(--text-secondary);
  line-height: 1.6;
}
</style>
