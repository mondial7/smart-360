<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'
import type { FeedbackRound } from '@/types/round'

const auth = useAuthStore()
const rounds = ref<FeedbackRound[]>([])
const loading = ref(true)
const error = ref(null)
const submissionStatus = ref<Record<string, boolean>>({})
const consolidationStatus = ref<Record<string, boolean>>({})

onMounted(async () => {
  console.log('RoundsView mounted')
  await loadRounds()
})

async function loadRounds() {
  loading.value = true
  error.value = null
  try {
    // Admins see all rounds, members see rounds where they're reviewers or subject
    const endpoint = auth.isAdmin ? '/rounds' : '/rounds-for-me'
    const response = await apiClient.get(endpoint)
    rounds.value = response.data || []
    console.log('Loaded rounds:', rounds.value)
    
    // Check submission status and consolidation status for each round
    for (const round of rounds.value) {
      try {
        const checkRes = await apiClient.get(`/submissions/check/${round.id}`)
        submissionStatus.value[round.id] = checkRes.data.submitted
      } catch {
        submissionStatus.value[round.id] = false
      }
      
      // Check if consolidation exists
      try {
        const consRes = await apiClient.get(`/consolidations/${round.id}`)
        consolidationStatus.value[round.id] = true
      } catch {
        consolidationStatus.value[round.id] = false
      }
    }
  } catch (err: any) {
    console.error('Failed to load rounds:', err)
    error.value = err.response?.data?.error || err.message || 'Failed to load rounds'
    rounds.value = []
  } finally {
    loading.value = false
  }
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return 'No deadline'
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function getStatusColor(status: string): string {
  const colors: Record<string, string> = {
    draft: '#ff9800',
    active: '#4caf50',
    closed: '#f44336',
    shared: '#2196f3'
  }
  return colors[status] || '#666'
}

function isAssignedReviewer(round: FeedbackRound): boolean {
  return round.reviewers?.some(r => r.reviewerId === auth.user?.id) ?? false
}

function hasSubmitted(roundId: string): boolean {  // Changed from number to string
  return submissionStatus.value[roundId] ?? false
}

async function closeRound(id: string) {  // Changed from number to string
  if (!confirm('Close this round? No more submissions will be accepted.')) return
  
  try {
    await apiClient.post(`/rounds/${id}/close`)
    await loadRounds()
  } catch (error) {
    console.error('Failed to close round:', error)
    alert('Failed to close round')
  }
}
</script>

<template>
  <div class="rounds-page">
    <header class="page-header">
      <div>
        <h1>Feedback Rounds</h1>
        <p>Manage all feedback collection cycles</p>
      </div>
      <router-link v-if="auth.isAdmin" to="/rounds/new" class="create-btn">
        + Create Round
      </router-link>
    </header>

    <div v-if="loading" class="loading">Loading rounds...</div>
    
    <div v-else-if="error" class="error-state">
      <p>Error: {{ error }}</p>
      <button @click="loadRounds" class="retry-btn">Retry</button>
    </div>
    
    <div v-else-if="rounds.length === 0" class="empty-state">
      <p>No feedback rounds yet.</p>
      <p v-if="auth.isAdmin" class="subtext">
        <router-link to="/rounds/new">Create your first round</router-link> to start collecting feedback.
      </p>
    </div>
    
    <div v-else class="rounds-list">
      <div v-for="round in rounds" :key="round.id" class="round-card">
        <div class="round-header">
          <span class="status-badge" :style="{ backgroundColor: getStatusColor(round.status) + '20', color: getStatusColor(round.status) }">
            {{ round.status }}
          </span>
          <span class="date">Created {{ formatDate(round.createdAt) }}</span>
        </div>
        
        <div class="round-body">
          <div class="subject-section">
            <span class="label">Feedback for</span>
            <div class="subject">
              <img v-if="round.subject?.photoUrl" :src="round.subject.photoUrl" class="mini-avatar">
              <div v-else class="mini-avatar-placeholder">{{ round.subject?.name.charAt(0) }}</div>
              <router-link 
                v-if="auth.isAdmin" 
                :to="`/rounds/${round.id}`" 
                class="name-link"
              >
                {{ round.subject?.name }}
              </router-link>
              <span v-else class="name">{{ round.subject?.name }}</span>
            </div>
          </div>
          
          <div class="reviewers-section">
            <span class="label">{{ round.reviewers?.length || 0 }} reviewers assigned</span>
            <div class="reviewer-list">
              <template v-for="reviewer in round.reviewers?.slice(0, 5)" :key="reviewer.id">
                <img 
                  v-if="reviewer.reviewer?.photoUrl"
                  :src="reviewer.reviewer.photoUrl"
                  :title="reviewer.reviewer.name"
                  class="reviewer-avatar"
                >
                <div 
                  v-else
                  :title="reviewer.reviewer?.name || 'Unknown'"
                  class="reviewer-avatar-placeholder"
                >
                  {{ reviewer.reviewer?.name?.charAt(0) || '?' }}
                </div>
              </template>
              <span v-if="(round.reviewers?.length || 0) > 5" class="more">+{{ (round.reviewers?.length || 0) - 5 }}</span>
            </div>
          </div>
          
          <div class="deadline-section">
            <span class="label">Deadline</span>
            <span :class="['deadline', { overdue: round.status === 'active' && round.deadline && new Date(round.deadline) < new Date() }]">
              {{ formatDate(round.deadline) }}
            </span>
          </div>
        </div>
        
        <div v-if="isAssignedReviewer(round) && !hasSubmitted(round.id) && round.status === 'active'" class="round-actions single">
          <router-link :to="`/rounds/${round.id}/submit`" class="submit-btn">
            Submit Feedback
          </router-link>
        </div>
        
        <div v-else-if="hasSubmitted(round.id)" class="round-actions single submitted">
          <span class="submitted-badge">✓ Feedback Submitted</span>
          <router-link :to="`/rounds/${round.id}/submission`" class="view-submission-btn">
            View My Submission
          </router-link>
        </div>
        
        <div v-else-if="auth.isAdmin && round.status === 'closed'" class="round-actions single">
          <router-link v-if="!consolidationStatus[round.id]" :to="`/rounds/${round.id}/consolidation`" class="consolidate-btn">
            Consolidate Feedback
          </router-link>
          <router-link v-else :to="`/rounds/${round.id}#consolidation`" class="view-feedback-btn">
            👁️ View Feedback
          </router-link>
        </div>
        
        <div v-else-if="auth.isAdmin && round.status === 'active'" class="round-actions">
          <router-link :to="`/rounds/${round.id}`" class="edit-btn">
            ✏️ Edit
          </router-link>
          <button class="close-btn" @click="closeRound(round.id)">
            Close Round
          </button>
        </div>
        
        <div v-else-if="auth.isAdmin" class="round-actions single">
          <router-link :to="`/rounds/${round.id}`" class="edit-btn">
            ✏️ Edit
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.rounds-page {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.page-header h1 {
  font-size: 2rem;
  margin-bottom: 0.5rem;
}

.page-header p {
  color: #666;
}

.create-btn {
  padding: 0.75rem 1.5rem;
  background: #667eea;
  color: white;
  text-decoration: none;
  border-radius: 8px;
  font-weight: 500;
  transition: background 0.2s;
}

.create-btn:hover {
  background: #5a6fd6;
}

.loading {
  text-align: center;
  color: #666;
  padding: 3rem;
}

.error-state {
  text-align: center;
  padding: 3rem;
  background: #ffebee;
  border-radius: 12px;
  color: #c62828;
}

.error-state p {
  margin-bottom: 1rem;
}

.retry-btn {
  padding: 0.5rem 1.5rem;
  background: #c62828;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
}

.retry-btn:hover {
  background: #a02222;
}

.empty-state {
  text-align: center;
  padding: 3rem;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}

.empty-state p {
  color: #666;
}

.subtext {
  margin-top: 0.5rem;
  font-size: 0.9rem;
}

.subtext a {
  color: #667eea;
  text-decoration: none;
}

.rounds-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 1.5rem;
}

.round-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  padding: 1.5rem;
}

.round-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.status-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: capitalize;
}

.date {
  font-size: 0.8rem;
  color: #888;
}

.round-body {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.label {
  font-size: 0.75rem;
  color: #888;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  display: block;
  margin-bottom: 0.25rem;
}

.subject {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.avatar, .avatar-placeholder {
  width: 36px;
  height: 36px;
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
  font-size: 0.9rem;
  font-weight: 600;
}

.name {
  font-weight: 500;
}

.name-link {
  font-weight: 500;
  color: #667eea;
  text-decoration: none;
}

.name-link:hover {
  text-decoration: underline;
}

.reviewer-list {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.reviewer-avatar, .reviewer-avatar-placeholder {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 2px solid white;
  margin-left: -8px;
}

.reviewer-avatar:first-child, .reviewer-avatar-placeholder:first-child {
  margin-left: 0;
}

.reviewer-avatar {
  object-fit: cover;
}

.reviewer-avatar-placeholder {
  background: #e0e0e0;
  color: #666;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.7rem;
  font-weight: 600;
}

.more {
  margin-left: 4px;
  font-size: 0.8rem;
  color: #666;
}

.deadline {
  font-size: 0.9rem;
  color: #333;
}

.deadline.overdue {
  color: #f44336;
  font-weight: 500;
}

.round-actions {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #eee;
  display: flex;
  gap: 0.5rem;
}

.round-actions.single {
  display: block;
}

.close-btn {
  flex: 1;
  padding: 0.5rem;
  border: 1px solid #f44336;
  background: transparent;
  color: #f44336;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.85rem;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #ffebee;
}

.submit-btn, .edit-btn, .consolidate-btn {
  flex: 1;
  display: block;
  padding: 0.5rem;
  text-decoration: none;
  border-radius: 6px;
  text-align: center;
  font-size: 0.85rem;
  transition: background 0.2s;
}

.submit-btn, .edit-btn {
  background: #667eea;
  color: white;
}

.submit-btn:hover, .edit-btn:hover {
  background: #5a6fd6;
}

.consolidate-btn {
  background: #2196f3;
  color: white;
}

.consolidate-btn:hover {
  background: #1976d2;
}

.submitted {
  text-align: center;
}

.submitted-badge {
  color: #4caf50;
  font-size: 0.85rem;
  font-weight: 500;
  display: block;
  margin-bottom: 0.5rem;
}

.view-submission-btn {
  display: inline-block;
  padding: 0.4rem 0.8rem;
  background: #2196f3;
  color: white;
  text-decoration: none;
  border-radius: 4px;
  font-size: 0.8rem;
  transition: background 0.2s;
}

.view-submission-btn:hover {
  background: #1976d2;
}

.view-feedback-btn {
  display: block;
  width: 100%;
  padding: 0.5rem;
  background: #4caf50;
  color: white;
  text-decoration: none;
  border-radius: 6px;
  text-align: center;
  font-size: 0.85rem;
  transition: background 0.2s;
}

.view-feedback-btn:hover {
  background: #45a049;
}

.submitted {
  text-align: center;
}

.submitted-badge {
  color: #4caf50;
  font-size: 0.85rem;
  font-weight: 500;
  display: block;
  margin-bottom: 0.5rem;
}

.edit-btn {
  display: block;
  width: 100%;
  padding: 0.5rem;
  background: #667eea;
  color: white;
  text-decoration: none;
  border-radius: 6px;
  text-align: center;
  font-size: 0.85rem;
  transition: background 0.2s;
}

.edit-btn:hover {
  background: #5a6fd6;
}
</style>
