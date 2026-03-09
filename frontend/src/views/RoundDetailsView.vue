<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'
import type { FeedbackRound } from '@/types/round'

const route = useRoute()
const auth = useAuthStore()
const roundId = parseInt(route.params.id as string)

const round = ref<FeedbackRound | null>(null)
const submissions = ref<any[]>([])
const consolidation = ref<any>(null)
const loading = ref(true)
const activeTab = ref('submissions')

onMounted(async () => {
  await loadData()
})

async function loadData() {
  try {
    // Load round details
    const roundRes = await apiClient.get(`/rounds/${roundId}`)
    round.value = roundRes.data
    
    // Load submissions
    const subRes = await apiClient.get(`/rounds/${roundId}/submissions`)
    submissions.value = subRes.data
    
    // Try to load consolidation
    try {
      const consRes = await apiClient.get(`/consolidations/${roundId}`)
      consolidation.value = consRes.data
    } catch {
      // No consolidation yet
    }
  } catch (error) {
    console.error('Failed to load data:', error)
  } finally {
    loading.value = false
  }
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return 'Not set'
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function getQuestionText(key: string): string {
  const questions: Record<string, string> = {
    a: 'What are this person\'s key strengths?',
    b: 'What areas could this person improve?',
    c: 'What specific behaviors or actions have you observed that stood out?',
    d: 'What advice would you give to help this person grow?'
  }
  return questions[key] || key
}

function parseResponses(responsesStr: string): Record<string, string> {
  try {
    return JSON.parse(responsesStr)
  } catch {
    return {}
  }
}
</script>

<template>
  <div class="round-details">
    <div v-if="loading" class="loading">Loading...</div>
    
    <template v-else-if="round">
      <header class="page-header">
        <router-link to="/rounds" class="back-link">← Back to Rounds</router-link>
        <h1>Round Details</h1>
        <div class="round-meta">
          <div class="meta-item">
            <span class="label">Subject:</span>
            <div class="subject">
              <img v-if="round.subject?.photoUrl" :src="round.subject.photoUrl" class="avatar">
              <div v-else class="avatar-placeholder">{{ round.subject?.name.charAt(0) }}</div>
              <span>{{ round.subject?.name }}</span>
            </div>
          </div>
          <div class="meta-item">
            <span class="label">Status:</span>
            <span :class="['status-badge', round.status]">{{ round.status }}</span>
          </div>
          <div class="meta-item">
            <span class="label">Deadline:</span>
            <span :class="{ overdue: round.status === 'active' && round.deadline && new Date(round.deadline) < new Date() }">
              {{ formatDate(round.deadline) }}
            </span>
          </div>
          <div class="meta-item">
            <span class="label">Reviewers:</span>
            <span>{{ round.reviewers?.length || 0 }} assigned</span>
          </div>
          <div class="meta-item">
            <span class="label">Submissions:</span>
            <span>{{ submissions.length }} received</span>
          </div>
        </div>
      </header>

      <!-- Tab Navigation -->
      <div class="tabs">
        <button 
          :class="['tab', { active: activeTab === 'submissions' }]"
          @click="activeTab = 'submissions'"
        >
          Raw Submissions ({{ submissions.length }})
        </button>
        <button 
          v-if="consolidation"
          :class="['tab', { active: activeTab === 'consolidation' }]"
          @click="activeTab = 'consolidation'"
        >
          Consolidated Feedback
        </button>
      </div>

      <!-- Raw Submissions Tab -->
      <div v-if="activeTab === 'submissions'" class="tab-content">
        <div v-if="submissions.length === 0" class="empty-state">
          <p>No submissions received yet.</p>
          <p v-if="round.status === 'active'" class="hint">
            {{ round.reviewers?.length || 0 }} reviewers assigned, deadline is {{ formatDate(round.deadline) }}
          </p>
        </div>
        
        <div v-else class="submissions-grid">
          <div v-for="submission in submissions" :key="submission.id" class="submission-card">
            <div class="submission-header">
              <span class="reviewer-id">Reviewer #{{ submission.id }}</span>
              <span class="submitted-at">Submitted {{ formatDate(submission.submittedAt) }}</span>
            </div>
            
            <div class="responses">
              <div v-for="(response, key) in parseResponses(submission.responses)" :key="key" class="response-item">
                <h4 class="question">{{ getQuestionText(key) }}</h4>
                <p class="answer">{{ response }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Consolidated Feedback Tab -->
      <div v-if="activeTab === 'consolidation' && consolidation" class="tab-content">
        <div class="consolidation-header">
          <div class="meta">
            <span>Generated {{ formatDate(consolidation.createdAt) }}</span>
            <span v-if="consolidation.sharedAt" class="shared-badge">
              ✓ Shared {{ formatDate(consolidation.sharedAt) }}
            </span>
            <span v-else class="not-shared">Not yet shared</span>
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
          <section v-if="consolidation.adminNotes" class="section">
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
.round-details {
  padding: 2rem;
  max-width: 1200px;
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
  margin-bottom: 1rem;
}

.round-meta {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  background: white;
  padding: 1.5rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.label {
  font-size: 0.75rem;
  color: #888;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.subject {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.avatar, .avatar-placeholder {
  width: 32px;
  height: 32px;
  border-radius: 50%;
}

.avatar {
  object-fit: cover;
}

.avatar-placeholder {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8rem;
  font-weight: 600;
}

.status-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: capitalize;
}

.status-badge.active {
  background: #e8f5e9;
  color: #4caf50;
}

.status-badge.closed {
  background: #ffebee;
  color: #c62828;
}

.status-badge.shared {
  background: #e3f2fd;
  color: #1976d2;
}

.overdue {
  color: #f44336;
  font-weight: 500;
}

.tabs {
  display: flex;
  gap: 1rem;
  margin-bottom: 2rem;
  border-bottom: 2px solid #e0e0e0;
}

.tab {
  padding: 1rem 1.5rem;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  font-size: 1rem;
  font-weight: 500;
  color: #666;
  transition: all 0.2s;
}

.tab.active {
  color: #667eea;
  border-bottom-color: #667eea;
}

.tab:hover {
  color: #667eea;
}

.tab-content {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  padding: 2rem;
}

.empty-state {
  text-align: center;
  padding: 3rem;
  color: #666;
}

.hint {
  font-size: 0.9rem;
  color: #888;
  margin-top: 0.5rem;
}

.submissions-grid {
  display: grid;
  gap: 1.5rem;
}

.submission-card {
  border: 1px solid #e0e0e0;
  border-radius: 12px;
  padding: 1.5rem;
}

.submission-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #e0e0e0;
}

.reviewer-id {
  font-weight: 500;
  color: #333;
}

.submitted-at {
  font-size: 0.85rem;
  color: #888;
}

.responses {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.response-item .question {
  font-size: 0.9rem;
  color: #667eea;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.response-item .answer {
  color: #444;
  line-height: 1.5;
  white-space: pre-wrap;
}

.consolidation-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
  padding-bottom: 1rem;
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

.not-shared {
  color: #ff9800;
  font-weight: 500;
}

.consolidation-body {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.section {
  padding-bottom: 2rem;
  border-bottom: 1px solid #eee;
}

.section:last-child {
  border-bottom: none;
  padding-bottom: 0;
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

.admin-note-display {
  background: #f5f5f5;
  padding: 1rem;
  border-radius: 8px;
  font-style: italic;
  color: #666;
}

.btn-primary {
  padding: 0.75rem 1.5rem;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  text-decoration: none;
  display: inline-block;
}
</style>
