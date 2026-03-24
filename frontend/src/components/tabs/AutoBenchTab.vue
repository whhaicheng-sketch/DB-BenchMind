<script setup>
import { computed, ref } from 'vue'
import {
  buildLocalPlanPreview,
  connectionFilterOptions,
  createAutoBenchWizardDraft,
  describeSelectedProfiles,
  filterPlaceholderConnections,
  placeholderConnections,
  policySummaryItems,
  profileOptions,
  toggleDraftConnectionSelection,
  toggleDraftProfileSelection,
  validateAutoBenchWizardDraft
} from './autobenchWizardDraft.mjs'

const draft = ref(createAutoBenchWizardDraft())
const activeConnectionFilter = ref('all')

const wizardPlaceholder = {
  description: 'This Wizard stays local-only in T2.3 and prepares a static draft shape for later orchestration tasks.'
}

const monitorPlaceholder = {
  description: 'This area will later show suite status, stage progress, and item-level runtime visibility.',
  items: [
    'Suite status placeholder',
    'Stage progress placeholder',
    'Item progress placeholder'
  ]
}

const reportPlaceholder = {
  description: 'This area will later summarize suite outcomes and expose report entry points.',
  items: [
    'Summary metrics placeholder',
    'Report entry placeholder',
    'Export surface placeholder'
  ]
}

const wizardValidation = computed(() => validateAutoBenchWizardDraft(draft.value))
const filteredConnections = computed(() => filterPlaceholderConnections(placeholderConnections, activeConnectionFilter.value))
const selectedProfileSummary = computed(() => describeSelectedProfiles(draft.value.selectedProfiles))
const planPreview = computed(() => buildLocalPlanPreview(draft.value, placeholderConnections))

function toggleConnectionSelection(connectionId) {
  draft.value = toggleDraftConnectionSelection(draft.value, connectionId)
}

function toggleProfileSelection(profileId) {
  draft.value = toggleDraftProfileSelection(draft.value, profileId)
}
</script>

<template>
  <section class="autobench-page">
    <header class="page-header">
      <div>
        <h1 class="page-title">AutoBench</h1>
        <p class="page-subtitle">
          Standalone suite workspace for future automated database testing flows. This page currently provides a local
          Wizard draft only.
        </p>
      </div>
      <button class="placeholder-action" type="button" disabled>Create Suite (later task)</button>
    </header>

    <div class="autobench-grid">
      <section class="autobench-section autobench-wizard" aria-labelledby="autobench-wizard-title">
        <div class="section-header">
          <h2 id="autobench-wizard-title">Wizard</h2>
          <p>{{ wizardPlaceholder.description }}</p>
        </div>

        <div class="wizard-groups">
          <section class="wizard-group" aria-labelledby="autobench-connections-title">
            <div class="wizard-group-header">
              <h3 id="autobench-connections-title">Selected Connections</h3>
              <span class="wizard-chip">{{ draft.selectedConnectionIds.length }} selected</span>
            </div>
            <p class="wizard-group-copy">Static placeholder targets only. No connection store or backend data is used here.</p>
            <div class="filter-block">
              <h4>Database Type Filter</h4>
              <div class="filter-options">
                <button
                  v-for="option in connectionFilterOptions"
                  :key="option.id"
                  type="button"
                  :class="['filter-pill', { active: activeConnectionFilter === option.id }]"
                  @click="activeConnectionFilter = option.id"
                >
                  {{ option.label }}
                </button>
              </div>
            </div>
            <label
              v-for="connection in filteredConnections"
              :key="connection.id"
              class="wizard-option-card"
            >
              <input
                type="checkbox"
                :checked="draft.selectedConnectionIds.includes(connection.id)"
                @change="toggleConnectionSelection(connection.id)"
              >
              <span class="wizard-option-copy">
                <strong>{{ connection.label }}</strong>
                <small class="wizard-option-meta">{{ connection.databaseType }}</small>
                <small>{{ connection.detail }}</small>
              </span>
            </label>
            <p v-if="wizardValidation.connectionError" class="wizard-validation">{{ wizardValidation.connectionError }}</p>
          </section>

          <section class="wizard-group" aria-labelledby="autobench-profiles-title">
            <div class="wizard-group-header">
              <h3 id="autobench-profiles-title">Profiles</h3>
              <span class="wizard-chip">Default order preserved</span>
            </div>
            <p class="wizard-group-copy">First-stage profiles stay local and ordered as test, cpu_bound, io_bound.</p>
            <div class="filter-block">
              <h4>Profile Scope</h4>
              <p class="wizard-group-copy compact">Local-only profile metadata. This is not yet a suite plan preview.</p>
            </div>
            <label
              v-for="profile in profileOptions"
              :key="profile.id"
              class="wizard-option-card wizard-option-card-profile"
            >
              <input
                type="checkbox"
                :checked="draft.selectedProfiles.includes(profile.id)"
                @change="toggleProfileSelection(profile.id)"
              >
              <span class="wizard-option-copy">
                <strong>{{ profile.label }}</strong>
                <small class="wizard-option-meta">{{ profile.scope }}</small>
                <small>{{ profile.description }}</small>
              </span>
            </label>
            <p class="wizard-order">Selected order: {{ selectedProfileSummary }}</p>
            <p v-if="wizardValidation.profileError" class="wizard-validation">{{ wizardValidation.profileError }}</p>
          </section>

          <section class="wizard-group" aria-labelledby="autobench-policy-title">
            <div class="wizard-group-header">
              <h3>Plan Preview</h3>
              <span class="wizard-chip">{{ planPreview.totalItems }} local items</span>
            </div>
            <p class="wizard-group-copy">Preview only. This is a local orchestration sketch and does not create or run a suite.</p>
            <div v-if="planPreview.totalItems === 0" class="preview-empty">
              Select at least one connection and one profile to build the local preview.
            </div>
            <ul v-else class="preview-list">
              <li v-for="item in planPreview.items" :key="item.id" class="preview-row">
                <span>#{{ item.order }}</span>
                <strong>{{ item.connectionLabel }}</strong>
                <small>{{ item.databaseType }}</small>
                <small>{{ item.profileId }}</small>
              </li>
            </ul>
          </section>

          <section class="wizard-group" aria-labelledby="autobench-policy-title">
            <div class="wizard-group-header">
              <h3 id="autobench-policy-title">Execution Policy</h3>
              <span class="wizard-chip">Read-only stage one defaults</span>
            </div>
            <p class="wizard-group-copy">This round only exposes the default orchestration policy and does not offer a strategy editor.</p>
            <dl class="policy-summary">
              <div v-for="item in policySummaryItems" :key="item.label" class="policy-row">
                <dt>{{ item.label }}</dt>
                <dd>{{ item.value }}</dd>
              </div>
            </dl>
          </section>
        </div>
      </section>

      <section class="autobench-section autobench-monitor" aria-labelledby="autobench-monitor-title">
        <div class="section-header">
          <h2 id="autobench-monitor-title">Monitor</h2>
          <p>{{ monitorPlaceholder.description }}</p>
        </div>
        <ul class="placeholder-list">
          <li v-for="item in monitorPlaceholder.items" :key="item">{{ item }}</li>
        </ul>
      </section>

      <section class="autobench-section autobench-report" aria-labelledby="autobench-report-title">
        <div class="section-header">
          <h2 id="autobench-report-title">Report</h2>
          <p>{{ reportPlaceholder.description }}</p>
        </div>
        <ul class="placeholder-list">
          <li v-for="item in reportPlaceholder.items" :key="item">{{ item }}</li>
        </ul>
      </section>
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
  flex-wrap: wrap;
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

.placeholder-action {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 10px 14px;
  background: var(--bg-secondary);
  color: var(--text-muted);
  cursor: not-allowed;
  box-shadow: none;
}

.autobench-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(0, 1fr) minmax(0, 1fr);
  gap: 16px;
}

.autobench-section {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 24px;
  box-shadow: var(--shadow-sm);
  min-height: 220px;
}

.section-header h2 {
  font-size: 18px;
  color: var(--text-primary);
}

.section-header p {
  margin-top: 8px;
  color: var(--text-secondary);
  line-height: 1.6;
}

.wizard-groups {
  margin-top: 18px;
  display: grid;
  gap: 18px;
}

.wizard-group {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 16px;
  background: var(--bg-secondary);
}

.wizard-group-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.wizard-group-header h3 {
  font-size: 15px;
  color: var(--text-primary);
}

.wizard-chip {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 999px;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  font-size: 12px;
}

.wizard-group-copy {
  margin-top: 8px;
  color: var(--text-secondary);
  line-height: 1.5;
}

.wizard-group-copy.compact {
  margin-top: 6px;
}

.wizard-option-card {
  margin-top: 12px;
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
  background: var(--bg-primary);
  cursor: pointer;
}

.wizard-option-card-profile {
  align-items: center;
}

.filter-block {
  margin-top: 14px;
}

.filter-block h4 {
  font-size: 13px;
  color: var(--text-primary);
}

.filter-options {
  margin-top: 10px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.filter-pill {
  border: 1px solid var(--border-color);
  border-radius: 999px;
  background: var(--bg-primary);
  color: var(--text-secondary);
  padding: 6px 10px;
  cursor: pointer;
}

.filter-pill.active {
  color: var(--primary);
  border-color: var(--primary);
}

.wizard-option-copy {
  display: grid;
  gap: 4px;
}

.wizard-option-copy strong {
  color: var(--text-primary);
  font-size: 14px;
}

.wizard-option-copy small {
  color: var(--text-secondary);
  line-height: 1.4;
}

.wizard-option-meta {
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.wizard-order {
  margin-top: 12px;
  color: var(--text-secondary);
  line-height: 1.5;
}

.wizard-validation {
  margin-top: 10px;
  color: var(--danger);
  font-size: 13px;
}

.preview-empty {
  margin-top: 12px;
  color: var(--text-secondary);
  line-height: 1.5;
}

.preview-list {
  margin-top: 12px;
  display: grid;
  gap: 10px;
  padding-left: 0;
  list-style: none;
}

.preview-row {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto auto;
  gap: 10px;
  align-items: center;
  padding: 10px 12px;
  border-radius: var(--radius-md);
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
}

.preview-row strong {
  color: var(--text-primary);
}

.preview-row small,
.preview-row span {
  color: var(--text-secondary);
}

.policy-summary {
  margin-top: 14px;
  display: grid;
  gap: 10px;
}

.policy-row {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding-bottom: 10px;
  border-bottom: 1px dashed var(--border-color);
}

.policy-row dt {
  color: var(--text-secondary);
}

.policy-row dd {
  color: var(--text-primary);
  font-weight: 600;
  text-align: right;
}

.placeholder-list {
  margin-top: 16px;
  padding-left: 18px;
  color: var(--text-secondary);
  display: grid;
  gap: 10px;
}

@media (max-width: 1080px) {
  .autobench-grid {
    grid-template-columns: 1fr;
  }
}
</style>
