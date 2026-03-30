<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'

const auth = useAuthStore()

const consolidations = ref<any[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const expandedIds = ref<Set<string>>(new Set())

onMounted(async () => {
  await loadConsolidations()
})

async function loadConsolidations() {
  try {
    loading.value = true
    error.value = null
    const response = await apiClient.get('/dashboard/my-consolidations')
    consolidations.value = response.data || []

    // Parse JSON fields for each consolidation
    consolidations.value.forEach(parseConsolidationFields)
  } catch (err: any) {
    console.error('Failed to load consolidations:', err)
    error.value = err.response?.data?.error || 'Failed to load feedback'
  } finally {
    loading.value = false
  }
}

function parseConsolidationFields(consolidation: any) {
  try {
    if (consolidation.strengths && typeof consolidation.strengths === 'string') {
      consolidation.strengths = JSON.parse(consolidation.strengths)
    }
    if (consolidation.areasForImprovement && typeof consolidation.areasForImprovement === 'string') {
      consolidation.areasForImprovement = JSON.parse(consolidation.areasForImprovement)
    }
    if (consolidation.actionableInsights && typeof consolidation.actionableInsights === 'string') {
      consolidation.actionableInsights = JSON.parse(consolidation.actionableInsights)
    }
    if (consolidation.questionSummaries && typeof consolidation.questionSummaries === 'string') {
      consolidation.questionSummaries = JSON.parse(consolidation.questionSummaries)
    }
  } catch (error) {
    console.error('Error parsing consolidation fields:', error)
  }
}

function toggleExpanded(id: string) {
  if (expandedIds.value.has(id)) {
    expandedIds.value.delete(id)
  } else {
    expandedIds.value.add(id)
  }
}

function isExpanded(id: string): boolean {
  return expandedIds.value.has(id)
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return 'Unknown date'
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'long',
    day: 'numeric',
    year: 'numeric'
  })
}
</script>

<template>
  <div class="my-feedback-page">
    <header class="page-header">
      <h1>My Feedback</h1>
      <p>View your consolidated feedback and development insights</p>
    </header>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Loading your feedback...</p>
    </div>

    <div v-else-if="error" class="error-state">
      <p class="error-message">{{ error }}</p>
      <button @click="loadConsolidations" class="btn-secondary">Try Again</button>
    </div>

    <div v-else-if="consolidations.length === 0" class="empty-state">
      <p>No feedback has been shared with you yet.</p>
      <p class="subtext">When an administrator shares consolidated feedback, it will appear here.</p>
    </div>

    <div v-else class="feedback-list">
      <div
        v-for="consolidation in consolidations"
        :key="consolidation.id"
        class="consolidation-card"
      >
        <div class="card-header" @click="toggleExpanded(consolidation.id)">
          <div class="header-content">
            <h2>Feedback Received</h2>
            <p class="date">Shared {{ formatDate(consolidation.sharedAt) }}</p>
          </div>
          <button class="expand-btn" :class="{ expanded: isExpanded(consolidation.id) }">
            {{ isExpanded(consolidation.id) ? '−' : '+' }}
          </button>
        </div>

        <div v-if="isExpanded(consolidation.id)" class="card-body">
          <!-- Executive Summary -->
          <section class="section">
            <h3>📄 Executive Summary</h3>
            <p class="summary-text">{{ consolidation.executiveSummary }}</p>
          </section>

          <!-- Strengths -->
          <section class="section">
            <h3>💪 Key Strengths</h3>
            <ul class="styled-list positive">
              <li v-for="(strength, i) in consolidation.strengths" :key="i">
                {{ strength }}
              </li>
            </ul>
          </section>

          <!-- Areas for Improvement -->
          <section class="section">
            <h3>📈 Areas for Improvement</h3>
            <ul class="styled-list improvement">
              <li v-for="(area, i) in consolidation.areasForImprovement" :key="i">
                {{ area }}
              </li>
            </ul>
          </section>

          <!-- Actionable Insights -->
          <section class="section">
            <h3>🎯 Actionable Insights</h3>
            <ul class="styled-list action">
              <li v-for="(insight, i) in consolidation.actionableInsights" :key="i">
                {{ insight }}
              </li>
            </ul>
          </section>

          <!-- Question Summaries -->
          <section class="section question-summaries">
            <h3>📋 Detailed Question Analysis</h3>
            <div class="question-cards">
              <div class="question-card">
                <h4>1. Key Strengths</h4>
                <p>{{ consolidation.questionSummaries?.a || 'No summary available' }}</p>
              </div>
              <div class="question-card">
                <h4>2. Areas to Improve</h4>
                <p>{{ consolidation.questionSummaries?.b || 'No summary available' }}</p>
              </div>
              <div class="question-card">
                <h4>3. Observed Behaviors</h4>
                <p>{{ consolidation.questionSummaries?.c || 'No summary available' }}</p>
              </div>
              <div class="question-card">
                <h4>4. Growth Advice</h4>
                <p>{{ consolidation.questionSummaries?.d || 'No summary available' }}</p>
              </div>
            </div>
          </section>

          <!-- Admin Notes -->
          <section v-if="consolidation.adminNotes" class="section admin-notes">
            <h3>📝 Additional Notes</h3>
            <div class="admin-note-display">{{ consolidation.adminNotes }}</div>
          </section>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.my-feedback-page {
  padding: 2rem;
  max-width: 900px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 2rem;
}

.page-header h1 {
  font-size: 2rem;
  margin-bottom: 0.5rem;
}

.page-header p {
  color: #666;
}

.loading-state, .error-state, .empty-state {
  background: white;
  padding: 3rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  text-align: center;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.error-state p {
  color: #d32f2f;
  margin-bottom: 1rem;
}

.empty-state p {
  color: #666;
  margin-bottom: 0.5rem;
}

.subtext {
  font-size: 0.9rem;
  color: #999;
}

.feedback-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.consolidation-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  overflow: hidden;
  transition: box-shadow 0.2s;
}

.consolidation-card:hover {
  box-shadow: 0 4px 12px rgba(0,0,0,0.12);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  cursor: pointer;
  user-select: none;
}

.header-content h2 {
  font-size: 1.25rem;
  margin: 0 0 0.25rem 0;
}

.date {
  font-size: 0.9rem;
  opacity: 0.9;
  margin: 0;
}

.expand-btn {
  background: rgba(255, 255, 255, 0.2);
  border: 2px solid rgba(255, 255, 255, 0.5);
  color: white;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  font-size: 1.5rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.expand-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: scale(1.1);
}

.expand-btn.expanded {
  transform: rotate(180deg);
}

.card-body {
  padding: 2rem;
  animation: slideDown 0.3s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.section {
  margin-bottom: 2rem;
  padding-bottom: 2rem;
  border-bottom: 1px solid #eee;
}

.section:last-child {
  margin-bottom: 0;
  padding-bottom: 0;
  border-bottom: none;
}

.section h3 {
  font-size: 1.125rem;
  margin-bottom: 1rem;
  color: #333;
}

.summary-text {
  font-size: 1.05rem;
  line-height: 1.7;
  color: #444;
}

.styled-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.styled-list li {
  padding: 1rem;
  margin-bottom: 0.75rem;
  border-radius: 8px;
  position: relative;
  padding-left: 2.5rem;
  line-height: 1.5;
}

.styled-list li::before {
  position: absolute;
  left: 0.75rem;
  font-size: 1.25rem;
}

.styled-list.positive li {
  background: #e8f5e9;
  border-left: 3px solid #4caf50;
}

.styled-list.positive li::before {
  content: '✓';
  color: #4caf50;
}

.styled-list.improvement li {
  background: #fff3e0;
  border-left: 3px solid #ff9800;
}

.styled-list.improvement li::before {
  content: '↑';
  color: #ff9800;
}

.styled-list.action li {
  background: #e3f2fd;
  border-left: 3px solid #2196f3;
}

.styled-list.action li::before {
  content: '→';
  color: #2196f3;
}

.question-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1rem;
}

.question-card {
  padding: 1.5rem;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
}

.question-card h4 {
  color: #667eea;
  font-size: 0.9rem;
  margin-bottom: 0.75rem;
  font-weight: 600;
}

.question-card p {
  color: #555;
  font-size: 0.95rem;
  line-height: 1.6;
  margin: 0;
}

.admin-notes h3 {
  color: #667eea;
}

.admin-note-display {
  background: #f5f5f5;
  padding: 1.25rem;
  border-radius: 8px;
  font-style: italic;
  color: #555;
  line-height: 1.6;
  border-left: 4px solid #667eea;
}

.btn-secondary {
  background: white;
  color: #666;
  border: 1px solid #ddd;
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background: #f5f5f5;
  border-color: #bbb;
}
</style>
