<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'
import type { FeedbackRound } from '@/types/round'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const roundId = route.params.roundId as string  // Changed from parseInt to string
const round = ref<FeedbackRound | null>(null)
const consolidation = ref<any>(null)
const loading = ref(true)
const generating = ref(false)
const sharing = ref(false)
const adminNotes = ref('')

onMounted(async () => {
  await loadData()
})

async function loadData() {
  try {
    // Load round details
    const roundRes = await apiClient.get(`/rounds/${roundId}`)
    round.value = roundRes.data
    
    // Try to load existing consolidation
    try {
      const consRes = await apiClient.get(`/consolidations/${roundId}`)
      consolidation.value = consRes.data
      adminNotes.value = consRes.data.adminNotes || ''
    } catch {
      // No consolidation yet
    }
  } catch (error) {
    console.error('Failed to load data:', error)
  } finally {
    loading.value = false
  }
}

async function generateConsolidation() {
  generating.value = true
  try {
    const res = await apiClient.post(`/rounds/${roundId}/consolidate`)
    consolidation.value = res.data
    adminNotes.value = res.data.adminNotes || ''
  } catch (error: any) {
    if (error.response?.status === 409) {
      alert('Consolidation already exists')
      await loadData()
    } else {
      alert(error.response?.data?.error || 'Failed to generate consolidation')
    }
  } finally {
    generating.value = false
  }
}

async function saveNotes() {
  if (!consolidation.value) return
  try {
    await apiClient.put(`/consolidations/${consolidation.value.id}/notes`, {
      adminNotes: adminNotes.value
    })
    alert('Notes saved')
  } catch (error) {
    alert('Failed to save notes')
  }
}

async function shareConsolidation() {
  if (!consolidation.value) return
  if (!confirm('Share this consolidated feedback with the subject?')) return
  
  sharing.value = true
  try {
    await apiClient.post(`/consolidations/${consolidation.value.id}/share`)
    alert('Consolidation shared successfully!')
    await loadData()
  } catch (error) {
    alert('Failed to share consolidation')
  } finally {
    sharing.value = false
  }
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return 'Not shared yet'
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<template>
  <div class="consolidation-page">
    <div v-if="loading" class="loading">Loading...</div>
    
    <template v-else-if="round">
      <header class="page-header">
        <router-link to="/rounds" class="back-link">← Back to Rounds</router-link>
        <h1>Feedback Consolidation</h1>
        <p class="subtitle">
          For <strong>{{ round.subject?.name }}</strong> • 
          {{ round.reviewers?.length || 0 }} reviewers • 
          Status: <span :class="['status-badge', round.status]">{{ round.status }}</span>
        </p>
      </header>

      <!-- Generate Button -->
      <div v-if="!consolidation && round.status === 'closed'" class="generate-section">
        <p>This round is closed and ready for AI consolidation.</p>
        <button 
          class="btn-primary" 
          @click="generateConsolidation"
          :disabled="generating"
        >
          {{ generating ? 'Generating...' : '🤖 Generate AI Consolidation' }}
        </button>
      </div>

      <div v-else-if="!consolidation && round.status !== 'closed'" class="info-section">
        <p>Round must be closed before generating consolidation.</p>
        <button class="btn-secondary" @click="router.push('/rounds')">
          Close Round First
        </button>
      </div>

      <!-- Consolidation View -->
      <div v-else-if="consolidation" class="consolidation-content">
        <div class="consolidation-header">
          <div class="meta">
            <span>Generated {{ formatDate(consolidation.createdAt) }}</span>
            <span v-if="consolidation.sharedAt" class="shared-badge">
              ✓ Shared {{ formatDate(consolidation.sharedAt) }}
            </span>
          </div>
          
          <div v-if="!consolidation.sharedAt" class="actions">
            <button 
              class="btn-primary" 
              @click="shareConsolidation"
              :disabled="sharing"
            >
              {{ sharing ? 'Sharing...' : '📤 Share with Subject' }}
            </button>
          </div>
        </div>

        <div class="consolidation-body">
          <!-- Executive Summary -->
          <section class="section">
            <h2>Executive Summary</h2>
            <p class="summary-text">{{ consolidation.executiveSummary }}</p>
          </section>

          <!-- Strengths -->
          <section class="section">
            <h2>💪 Key Strengths</h2>
            <ul class="styled-list positive">
              <li v-for="(strength, i) in consolidation.strengths" :key="i">
                {{ strength }}
              </li>
            </ul>
          </section>

          <!-- Areas for Improvement -->
          <section class="section">
            <h2>📈 Areas for Improvement</h2>
            <ul class="styled-list improvement">
              <li v-for="(area, i) in consolidation.areasForImprovement" :key="i">
                {{ area }}
              </li>
            </ul>
          </section>

          <!-- Actionable Insights -->
          <section class="section">
            <h2>🎯 Actionable Insights</h2>
            <ul class="styled-list action">
              <li v-for="(insight, i) in consolidation.actionableInsights" :key="i">
                {{ insight }}
              </li>
            </ul>
          </section>

          <!-- Question Summaries -->
          <section class="section question-summaries">
            <h2>📋 Detailed Question Analysis</h2>
            <div class="question-cards">
              <div class="question-card">
                <h4>1. Key Strengths</h4>
                <p>{{ consolidation.questionSummaries?.a }}</p>
              </div>
              <div class="question-card">
                <h4>2. Areas to Improve</h4>
                <p>{{ consolidation.questionSummaries?.b }}</p>
              </div>
              <div class="question-card">
                <h4>3. Observed Behaviors</h4>
                <p>{{ consolidation.questionSummaries?.c }}</p>
              </div>
              <div class="question-card">
                <h4>4. Growth Advice</h4>
                <p>{{ consolidation.questionSummaries?.d }}</p>
              </div>
            </div>
          </section>

          <!-- Admin Notes -->
          <section v-if="!consolidation.sharedAt" class="section admin-notes">
            <h2>📝 Admin Notes (Optional)</h2>
            <p class="hint">Add context or guidance for the subject. These notes will be included when shared.</p>
            <textarea
              v-model="adminNotes"
              rows="4"
              placeholder="Add your notes here..."
            ></textarea>
            <button class="btn-secondary" @click="saveNotes">
              Save Notes
            </button>
          </section>
          
          <section v-else-if="consolidation.adminNotes" class="section">
            <h2>📝 Admin Notes</h2>
            <p class="admin-note-display">{{ consolidation.adminNotes }}</p>
          </section>
        </div>
      </div>
    </template>

    <div v-else class="error-state">
      <p>Round not found.</p>
      <router-link to="/rounds" class="btn-primary">Back to Rounds</router-link>
    </div>
  </div>
</template>

<style scoped>
.consolidation-page {
  padding: 2rem;
  max-width: 900px;
  margin: 0 auto;
}

.loading, .error-state {
  text-align: center;
  padding: 3rem;
}

.page-header {
  margin-bottom: 2rem;
}

.back-link {
  display: block;
  margin-bottom: 1rem;
  color: #667eea;
  text-decoration: none;
}

.page-header h1 {
  font-size: 2rem;
  margin-bottom: 0.5rem;
}

.subtitle {
  color: #666;
}

.subtitle strong {
  color: #333;
}

.status-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: capitalize;
}

.status-badge.closed {
  background: #ffebee;
  color: #c62828;
}

.status-badge.shared {
  background: #e3f2fd;
  color: #1976d2;
}

.generate-section, .info-section {
  text-align: center;
  padding: 3rem;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}

.generate-section p, .info-section p {
  color: #666;
  margin-bottom: 1.5rem;
}

.consolidation-content {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  overflow: hidden;
}

.consolidation-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  background: #f8f9fa;
  border-bottom: 1px solid #e0e0e0;
}

.meta {
  display: flex;
  gap: 1rem;
  align-items: center;
  color: #666;
  font-size: 0.9rem;
}

.shared-badge {
  color: #4caf50;
  font-weight: 500;
}

.consolidation-body {
  padding: 1.5rem;
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

.section h2 {
  font-size: 1.25rem;
  margin-bottom: 1rem;
  color: #333;
}

.summary-text {
  font-size: 1.1rem;
  line-height: 1.6;
  color: #444;
}

.styled-list {
  list-style: none;
  padding: 0;
}

.styled-list li {
  padding: 1rem;
  margin-bottom: 0.75rem;
  border-radius: 8px;
  position: relative;
  padding-left: 2.5rem;
}

.styled-list li::before {
  position: absolute;
  left: 0.75rem;
  font-size: 1.25rem;
}

.styled-list.positive li {
  background: #e8f5e9;
}

.styled-list.positive li::before {
  content: '✓';
  color: #4caf50;
}

.styled-list.improvement li {
  background: #fff3e0;
}

.styled-list.improvement li::before {
  content: '↑';
  color: #ff9800;
}

.styled-list.action li {
  background: #e3f2fd;
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
}

.question-card h4 {
  color: #667eea;
  font-size: 0.9rem;
  margin-bottom: 0.5rem;
}

.question-card p {
  color: #666;
  font-size: 0.9rem;
  line-height: 1.5;
}

.admin-notes .hint {
  color: #888;
  font-size: 0.9rem;
  margin-bottom: 1rem;
}

.admin-notes textarea {
  width: 100%;
  padding: 0.75rem;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-family: inherit;
  font-size: 1rem;
  resize: vertical;
  margin-bottom: 1rem;
}

.admin-note-display {
  background: #f5f5f5;
  padding: 1rem;
  border-radius: 8px;
  font-style: italic;
  color: #666;
}

.btn-primary, .btn-secondary {
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  text-decoration: none;
  display: inline-block;
}

.btn-primary {
  background: #667eea;
  color: white;
  border: none;
}

.btn-primary:hover:not(:disabled) {
  background: #5a6fd6;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: white;
  color: #666;
  border: 1px solid #ddd;
}

.btn-secondary:hover {
  background: #f5f5f5;
}
</style>
