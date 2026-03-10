<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const roundId = route.params.id as string
const round = ref<any>(null)
const submission = ref<any>(null)
const loading = ref(true)
const editing = ref(false)
const editResponses = ref<Record<string, string>>({
  a: '',
  b: '',
  c: '',
  d: ''
})

const questions = [
  { key: 'a', text: 'What are this person\'s key strengths?' },
  { key: 'b', text: 'What areas could this person improve?' },
  { key: 'c', text: 'What specific behaviors or actions have you observed that stood out?' },
  { key: 'd', text: 'What advice would you give to help this person grow?' }
]

onMounted(async () => {
  await loadData()
})

async function loadData() {
  try {
    // Load round details
    const roundRes = await apiClient.get(`/rounds/${roundId}`)
    round.value = roundRes.data
    
    // Load user's submission
    const subRes = await apiClient.get(`/submissions/round/${roundId}`)
    const submissions = subRes.data || []
    
    // Find current user's submission
    const userSubmission = submissions.find((sub: any) => sub.reviewerId === auth.user?.id)
    if (userSubmission) {
      submission.value = userSubmission
      // Parse the responses
      try {
        submission.value.responsesParsed = JSON.parse(userSubmission.responses)
      } catch (error) {
        console.error('Error parsing responses:', error)
        submission.value.responsesParsed = {}
      }
    }
  } catch (error) {
    console.error('Failed to load data:', error)
  } finally {
    loading.value = false
  }
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function startEditing() {
  if (!submission.value) return
  
  editing.value = true
  editResponses.value = {...submission.value.responsesParsed}
}

function cancelEditing() {
  editing.value = false
  editResponses.value = { a: '', b: '', c: '', d: '' }
}

async function saveEdits() {
  // Validate all questions answered
  for (const q of questions) {
    if (!editResponses.value[q.key].trim()) {
      alert('Please answer all questions before saving.')
      return
    }
  }

  try {
    await apiClient.put(`/submissions/${submission.value.id}`, {
      responses: JSON.stringify(editResponses.value)
    })
    
    // Update local submission data
    submission.value.responses = JSON.stringify(editResponses.value)
    submission.value.responsesParsed = {...editResponses.value}
    submission.value.updatedAt = new Date().toISOString()
    
    editing.value = false
    alert('Feedback updated successfully!')
  } catch (error: any) {
    alert(error.response?.data?.error || 'Failed to update feedback. Please try again.')
  }
}

function canEdit() {
  return round.value?.status === 'active' && !submission.value?.sharedAt
}
</script>

<template>
  <div class="submission-view">
    <div v-if="loading" class="loading">Loading...</div>
    
    <template v-else-if="round && submission">
      <header class="page-header">
        <router-link to="/rounds" class="back-link">← Back to Rounds</router-link>
        <h1>My Feedback Submission</h1>
        <p class="subtitle">
          For <strong>{{ round.subject?.name }}</strong> • 
          Submitted {{ formatDate(submission.submittedAt) }}
        </p>
      </header>

      <div class="submission-content">
        <div class="submission-header">
          <div class="meta">
            <span>Status: <span class="status-badge submitted">Submitted</span></span>
            <span>Round: {{ round.status }}</span>
            <span v-if="submission.updatedAt && submission.updatedAt !== submission.submittedAt">
              Last updated: {{ formatDate(submission.updatedAt) }}
            </span>
          </div>
          <div v-if="canEdit()" class="actions">
            <button @click="startEditing" class="edit-btn">✏️ Edit Feedback</button>
          </div>
        </div>

        <div v-if="!editing" class="questions-section">
          <h2>Your Feedback Responses</h2>
          
          <div class="question-item">
            <h3>1. What are this person's key strengths?</h3>
            <div class="response">
              {{ submission.responsesParsed?.a || 'No response provided' }}
            </div>
          </div>

          <div class="question-item">
            <h3>2. What areas could this person improve?</h3>
            <div class="response">
              {{ submission.responsesParsed?.b || 'No response provided' }}
            </div>
          </div>

          <div class="question-item">
            <h3>3. What specific behaviors or actions have you observed that stood out?</h3>
            <div class="response">
              {{ submission.responsesParsed?.c || 'No response provided' }}
            </div>
          </div>

          <div class="question-item">
            <h3>4. What advice would you give to help this person grow?</h3>
            <div class="response">
              {{ submission.responsesParsed?.d || 'No response provided' }}
            </div>
          </div>
        </div>

        <div v-else class="edit-section">
          <h2>Edit Your Feedback</h2>
          
          <div class="edit-form">
            <div v-for="question in questions" :key="question.key" class="question-edit">
              <h3>{{ question.text }}</h3>
              <textarea
                v-model="editResponses[question.key]"
                placeholder="Enter your response..."
                rows="3"
                class="response-textarea"
              ></textarea>
            </div>
          </div>

          <div class="edit-actions">
            <button @click="saveEdits" class="btn-primary">Save Changes</button>
            <button @click="cancelEditing" class="btn-secondary">Cancel</button>
          </div>
        </div>

        <div v-if="round.status !== 'active'" class="notice">
          <p>⚠️ This round is {{ round.status }}. Feedback can only be edited while the round is active.</p>
        </div>
      </div>
    </template>

    <div v-else class="error-state">
      <p>Submission not found.</p>
      <router-link to="/rounds" class="btn-primary">Back to Rounds</router-link>
    </div>
  </div>
</template>

<style scoped>
.submission-view {
  padding: 2rem;
  max-width: 800px;
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

.status-badge.submitted {
  background: #e8f5e9;
  color: #4caf50;
}

.submission-content {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  overflow: hidden;
}

.submission-header {
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

.actions {
  display: flex;
  gap: 0.5rem;
}

.edit-btn {
  background: none;
  border: 1px solid #667eea;
  color: #667eea;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s;
}

.edit-btn:hover {
  background: #667eea;
  color: white;
}

.questions-section {
  padding: 1.5rem;
}

.questions-section h2 {
  font-size: 1.25rem;
  margin-bottom: 1.5rem;
  color: #333;
}

.question-item {
  margin-bottom: 2rem;
  padding-bottom: 2rem;
  border-bottom: 1px solid #eee;
}

.question-item:last-child {
  margin-bottom: 0;
  padding-bottom: 0;
  border-bottom: none;
}

.question-item h3 {
  font-size: 1rem;
  margin-bottom: 0.75rem;
  color: #667eea;
  font-weight: 600;
}

.response {
  background: #f8f9fa;
  padding: 1rem;
  border-radius: 8px;
  border-left: 4px solid #667eea;
  line-height: 1.6;
  color: #444;
}

.edit-section {
  padding: 1.5rem;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
  margin-bottom: 1rem;
}

.edit-section h2 {
  font-size: 1.25rem;
  margin-bottom: 1.5rem;
  color: #333;
}

.edit-form {
  margin-bottom: 1.5rem;
}

.question-edit {
  margin-bottom: 1.5rem;
}

.question-edit h3 {
  font-size: 1rem;
  margin-bottom: 0.5rem;
  color: #667eea;
  font-weight: 600;
}

.response-textarea {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-family: inherit;
  font-size: 1rem;
  resize: vertical;
  line-height: 1.5;
}

.edit-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
}

.notice {
  background: #fff3cd;
  border: 1px solid #ffeaa7;
  color: #856404;
  padding: 1rem;
  border-radius: 8px;
  margin-top: 1rem;
}

.btn-primary {
  display: inline-block;
  padding: 0.75rem 1.5rem;
  background: #667eea;
  color: white;
  text-decoration: none;
  border-radius: 8px;
  font-weight: 500;
  transition: background 0.2s;
}

.btn-primary:hover {
  background: #5a6fd6;
}

.btn-secondary {
  display: inline-block;
  padding: 0.75rem 1.5rem;
  background: white;
  color: #666;
  text-decoration: none;
  border: 1px solid #ddd;
  border-radius: 8px;
  font-weight: 500;
  transition: background 0.2s;
}

.btn-secondary:hover {
  background: #f5f5f5;
}
</style>
