<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'
import type { FeedbackRound } from '@/types/round'

const route = useRoute()
const auth = useAuthStore()
const roundId = route.params.id as string  // Changed from parseInt to string

const round = ref<FeedbackRound | null>(null)
const submissions = ref<any[]>([])
const consolidation = ref<any>(null)
const loading = ref(true)
const activeTab = ref('submissions')
const showEditModal = ref(false)
const editForm = ref({
  subjectId: '',
  deadline: '',
  status: '',
  reviewerIds: [] as string[]
})
const users = ref<any[]>([])
const editLoading = ref(false)
const selectedReviewer = ref('')

// Computed properties
const availableUsers = computed(() => {
  return users.value.filter(user => 
    user.id !== auth.user?.id && 
    !editForm.value.reviewerIds.includes(user.id) &&
    user.id !== round.value?.subjectId
  )
})

onMounted(async () => {
  await loadData()
  
  // Check if URL hash indicates consolidation tab should be active
  if (window.location.hash === '#consolidation') {
    activeTab.value = 'consolidation'
  }
})

async function loadData() {
  try {
    // Load round details
    const roundRes = await apiClient.get(`/rounds/${roundId}`)
    round.value = roundRes.data
    
    // Load users for edit form
    const usersRes = await apiClient.get('/users')
    users.value = usersRes.data
    
    // Load submissions
    const subRes = await apiClient.get(`/submissions/round/${roundId}`)
    submissions.value = subRes.data
    
    // Try to load consolidation
    try {
      const consRes = await apiClient.get(`/consolidations/${roundId}`)
      consolidation.value = consRes.data
      // Parse consolidation fields
      if (consolidation.value.strengths && typeof consolidation.value.strengths === 'string') {
        consolidation.value.strengths = JSON.parse(consolidation.value.strengths)
      }
      if (consolidation.value.areasForImprovement && typeof consolidation.value.areasForImprovement === 'string') {
        consolidation.value.areasForImprovement = JSON.parse(consolidation.value.areasForImprovement)
      }
      if (consolidation.value.actionableInsights && typeof consolidation.value.actionableInsights === 'string') {
        consolidation.value.actionableInsights = JSON.parse(consolidation.value.actionableInsights)
      }
      if (consolidation.value.questionSummaries && typeof consolidation.value.questionSummaries === 'string') {
        consolidation.value.questionSummaries = JSON.parse(consolidation.value.questionSummaries)
      }
    } catch (error) {
      console.error('Error loading consolidation:', error)
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

function openEditModal() {
  if (!round.value) return
  
  // Filter out subject from existing reviewers
  const existingReviewers = round.value.reviewers?.map(r => r.reviewerId).filter(id => id !== round.value?.subjectId) || []
  
  editForm.value = {
    subjectId: round.value.subjectId || '',
    deadline: round.value.deadline ? new Date(round.value.deadline).toISOString().slice(0, 16) : '',
    status: round.value.status || '',
    reviewerIds: existingReviewers
  }
  showEditModal.value = true
}

async function updateRound() {
  if (!round.value) return
  
  editLoading.value = true
  try {
    const updateData: any = {}
    
    if (editForm.value.subjectId && editForm.value.subjectId !== round.value.subjectId) {
      updateData.subjectId = editForm.value.subjectId
    }
    
    if (editForm.value.deadline) {
      updateData.deadline = new Date(editForm.value.deadline).toISOString()
    }
    
    if (editForm.value.status && editForm.value.status !== round.value.status) {
      updateData.status = editForm.value.status
    }
    
    // Handle reviewer updates
    const currentReviewerIds = round.value.reviewers?.map(r => r.reviewerId) || []
    const newReviewerIds = editForm.value.reviewerIds
    
    // Find reviewers to add and remove
    const toAdd = newReviewerIds.filter(id => !currentReviewerIds.includes(id))
    const toRemove = currentReviewerIds.filter(id => !newReviewerIds.includes(id))
    
    if (Object.keys(updateData).length === 0 && toAdd.length === 0 && toRemove.length === 0) {
      showEditModal.value = false
      return
    }
    
    // Update basic round info if needed
    if (Object.keys(updateData).length > 0) {
      await apiClient.put(`/rounds/${roundId}`, updateData)
    }
    
    // Add new reviewers
    for (const reviewerId of toAdd) {
      try {
        await apiClient.post(`/rounds/${roundId}/reviewers`, { reviewerIds: [reviewerId] })
      } catch (error) {
        console.error('Failed to add reviewer:', error)
      }
    }
    
    // Remove reviewers (we'd need to implement this endpoint)
    for (const reviewerId of toRemove) {
      try {
        await apiClient.delete(`/rounds/${roundId}/reviewers/${reviewerId}`)
      } catch (error) {
        console.error('Failed to remove reviewer:', error)
        // For now, we'll ignore if the endpoint doesn't exist
      }
    }
    
    // Reload data to show updated round
    await loadData()
    showEditModal.value = false
  } catch (error: any) {
    console.error('Failed to update round:', error)
    alert(error.response?.data?.error || 'Failed to update round')
  } finally {
    editLoading.value = false
  }
}

function addReviewer(reviewerId: string) {
  if (!editForm.value.reviewerIds.includes(reviewerId)) {
    editForm.value.reviewerIds.push(reviewerId)
  }
}

function removeReviewer(reviewerId: string) {
  const index = editForm.value.reviewerIds.indexOf(reviewerId)
  if (index > -1) {
    editForm.value.reviewerIds.splice(index, 1)
  }
  // Reset the select dropdown
  selectedReviewer.value = ''
}

function getUserById(userId: string) {
  return users.value.find(user => user.id === userId)
}
</script>

<template>
  <div class="round-details">
    <div v-if="loading" class="loading">Loading...</div>
    
    <template v-else-if="round">
      <header class="page-header">
        <router-link to="/rounds" class="back-link">← Back to Rounds</router-link>
        <div class="header-content">
          <h1>Round Details</h1>
          <button 
            v-if="auth.isAdmin || round.createdById === auth.user?.id" 
            @click="openEditModal" 
            class="edit-btn"
          >
            ✏️ Edit Round
          </button>
        </div>
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

    <!-- Edit Modal -->
    <div v-if="showEditModal" class="modal-overlay" @click.self="showEditModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h2>Edit Round</h2>
          <button @click="showEditModal = false" class="close-btn">×</button>
        </div>
        
        <form @submit.prevent="updateRound" class="edit-form">
          <div class="form-group">
            <label for="subjectId">Subject</label>
            <select id="subjectId" v-model="editForm.subjectId" required>
              <option value="">Select a subject</option>
              <option 
                v-for="user in users.filter(u => u.id !== auth.user?.id)" 
                :key="user.id" 
                :value="user.id"
              >
                {{ user.name }}
              </option>
            </select>
          </div>
          
          <div class="form-group">
            <label for="deadline">Deadline</label>
            <input 
              id="deadline" 
              type="datetime-local" 
              v-model="editForm.deadline"
              :min="new Date().toISOString().slice(0, 16)"
            >
          </div>
          
          <div class="form-group">
            <label for="status">Status</label>
            <select id="status" v-model="editForm.status" required>
              <option value="draft">Draft</option>
              <option value="active">Active</option>
              <option value="closed">Closed</option>
              <option value="shared">Shared</option>
            </select>
          </div>
          
          <div class="form-group">
            <label>Reviewers</label>
            <div class="reviewer-management">
              <!-- Show warning if subject was previously assigned as reviewer -->
              <div v-if="round.value?.reviewers?.some(r => r.reviewerId === round.value?.subjectId)" class="subject-reviewer-warning">
                ⚠️ The subject ({{ getUserById(round.value?.subjectId || '')?.name }}) was previously assigned as a reviewer and has been removed.
              </div>
              
              <!-- Current reviewers -->
              <div v-if="editForm.reviewerIds.length > 0" class="current-reviewers">
                <div class="reviewer-list">
                  <div 
                    v-for="reviewerId in editForm.reviewerIds" 
                    :key="reviewerId"
                    class="reviewer-tag"
                  >
                    <span>{{ getUserById(reviewerId)?.name || reviewerId }}</span>
                    <button 
                      type="button" 
                      @click="removeReviewer(reviewerId)"
                      class="remove-btn"
                    >
                      ×
                    </button>
                  </div>
                </div>
              </div>
              
              <!-- Add new reviewers -->
              <div class="add-reviewer">
                <select 
                  v-model="selectedReviewer" 
                  @change="addReviewer(selectedReviewer)"
                >
                  <option value="">Add reviewer...</option>
                  <option 
                    v-for="user in availableUsers" 
                    :key="user.id" 
                    :value="user.id"
                  >
                    {{ user.name }}
                  </option>
                </select>
              </div>
            </div>
          </div>
          
          <div class="form-actions">
            <button type="button" @click="showEditModal = false" class="cancel-btn">
              Cancel
            </button>
            <button type="submit" :disabled="editLoading" class="save-btn">
              {{ editLoading ? 'Saving...' : 'Save Changes' }}
            </button>
          </div>
        </form>
      </div>
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

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.edit-btn {
  padding: 0.5rem 1rem;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}

.edit-btn:hover {
  background: #5a6fd6;
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

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  padding: 0;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e0e0e0;
}

.modal-header h2 {
  margin: 0;
  font-size: 1.5rem;
  color: #333;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #666;
  padding: 0;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  color: #333;
}

.edit-form {
  padding: 1.5rem;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: #333;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 1rem;
  box-sizing: border-box;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.2);
}

.form-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  margin-top: 2rem;
}

.cancel-btn {
  padding: 0.75rem 1.5rem;
  background: #f5f5f5;
  color: #666;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 1rem;
  cursor: pointer;
  transition: background 0.2s;
}

.cancel-btn:hover {
  background: #e0e0e0;
}

.save-btn {
  padding: 0.75rem 1.5rem;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  cursor: pointer;
  transition: background 0.2s;
}

.save-btn:hover:not(:disabled) {
  background: #5a6fd6;
}

.save-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Reviewer Management Styles */
.reviewer-management {
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  padding: 1rem;
  background: #f8f9fa;
}

.current-reviewers {
  margin-bottom: 1rem;
}

.reviewer-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.reviewer-tag {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: #667eea;
  color: white;
  padding: 0.25rem 0.5rem;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
}

.reviewer-tag span {
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.remove-btn {
  background: none;
  border: none;
  color: white;
  cursor: pointer;
  font-size: 1rem;
  padding: 0;
  width: 16px;
  height: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: background 0.2s;
}

.remove-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

.add-reviewer select {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: white;
  font-size: 0.9rem;
}

.add-reviewer select:focus {
  outline: none;
  border-color: #667eea;
}

.subject-reviewer-warning {
  background: #fff3cd;
  border: 1px solid #ffeaa7;
  color: #856404;
  padding: 0.75rem;
  border-radius: 4px;
  font-size: 0.85rem;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
</style>
