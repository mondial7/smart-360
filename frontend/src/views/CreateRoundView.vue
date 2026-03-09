<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'
import type { User } from '@/types/user'
import type { CreateRoundRequest } from '@/types/round'

const router = useRouter()
const auth = useAuthStore()

const step = ref(1)
const loading = ref(false)
const users = ref<User[]>([])

// Form data
const subjectId = ref<number | null>(null)
const reviewerIds = ref<number[]>([])
const deadline = ref('')

// Validation errors
const errors = ref<string[]>([])

onMounted(async () => {
  await loadUsers()
  // Set default deadline to 2 weeks from now
  const twoWeeks = new Date()
  twoWeeks.setDate(twoWeeks.getDate() + 14)
  deadline.value = twoWeeks.toISOString().slice(0, 16)
})

async function loadUsers() {
  try {
    const response = await apiClient.get('/users')
    users.value = response.data
  } catch (error) {
    console.error('Failed to load users:', error)
  }
}

const subject = () => users.value.find(u => u.id === subjectId.value)
const reviewers = () => users.value.filter(u => reviewerIds.value.includes(u.id) && u.id !== subjectId.value)
const availableReviewers = () => users.value.filter(u => u.id !== subjectId.value && u.id !== auth.user?.id)

function toggleReviewer(id: number) {
  const index = reviewerIds.value.indexOf(id)
  if (index > -1) {
    reviewerIds.value.splice(index, 1)
  } else {
    reviewerIds.value.push(id)
  }
}

function validateStep(): boolean {
  errors.value = []
  
  if (step.value === 1 && !subjectId.value) {
    errors.value.push('Please select a feedback subject')
    return false
  }
  
  if (step.value === 2 && reviewerIds.value.length === 0) {
    errors.value.push('Please select at least one reviewer')
    return false
  }
  
  if (step.value === 3) {
    if (!deadline.value) {
      errors.value.push('Please set a deadline')
      return false
    }
    const deadlineDate = new Date(deadline.value)
    if (deadlineDate <= new Date()) {
      errors.value.push('Deadline must be in the future')
      return false
    }
  }
  
  return true
}

function nextStep() {
  if (validateStep()) {
    step.value++
  }
}

function prevStep() {
  step.value--
}

async function createRound() {
  if (!validateStep()) return
  
  loading.value = true
  
  // Filter out subject from reviewers
  const validReviewerIds = reviewerIds.value.filter(id => id !== subjectId.value)
  
  const request: CreateRoundRequest = {
    subjectId: subjectId.value!,
    reviewerIds: validReviewerIds,
    deadline: new Date(deadline.value).toISOString()
  }
  
  try {
    await apiClient.post('/rounds', request)
    router.push('/rounds')
  } catch (error: any) {
    errors.value = [error.response?.data?.error || 'Failed to create round']
  } finally {
    loading.value = false
  }
}

function formatDateTimeLocal(dateStr: string): string {
  return new Date(dateStr).toLocaleString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<template>
  <div class="wizard-page">
    <header class="page-header">
      <h1>Create Feedback Round</h1>
      <p>Step {{ step }} of 4</p>
    </header>

    <!-- Progress Bar -->
    <div class="progress-bar">
      <div class="progress-step" :class="{ active: step >= 1, complete: step > 1 }">1. Subject</div>
      <div class="progress-step" :class="{ active: step >= 2, complete: step > 2 }">2. Reviewers</div>
      <div class="progress-step" :class="{ active: step >= 3, complete: step > 3 }">3. Deadline</div>
      <div class="progress-step" :class="{ active: step >= 4 }">4. Confirm</div>
    </div>

    <!-- Errors -->
    <div v-if="errors.length" class="errors">
      <p v-for="error in errors" :key="error">{{ error }}</p>
    </div>

    <!-- Step 1: Select Subject -->
    <div v-if="step === 1" class="wizard-step">
      <h2>Who is receiving feedback?</h2>
      <p class="hint">Select the team member who will receive feedback from their peers.</p>
      
      <div class="user-grid">
        <div 
          v-for="user in users.filter(u => u.id !== auth.user?.id)" 
          :key="user.id"
          class="user-card"
          :class="{ selected: subjectId === user.id }"
          @click="subjectId = user.id"
        >
          <img v-if="user.photoUrl" :src="user.photoUrl" :alt="user.name" class="user-photo">
          <div v-else class="user-photo-placeholder">{{ user.name.charAt(0).toUpperCase() }}</div>
          <span class="user-name">{{ user.name }}</span>
        </div>
      </div>
    </div>

    <!-- Step 2: Select Reviewers -->
    <div v-if="step === 2" class="wizard-step">
      <h2>Who will provide feedback?</h2>
      <p class="hint">Select 3-5 peers who will give anonymous feedback. The subject and you (the admin) are automatically excluded.</p>
      
      <div class="user-grid">
        <div 
          v-for="user in availableReviewers()" 
          :key="user.id"
          class="user-card"
          :class="{ selected: reviewerIds.includes(user.id) }"
          @click="toggleReviewer(user.id)"
        >
          <img v-if="user.photoUrl" :src="user.photoUrl" :alt="user.name" class="user-photo">
          <div v-else class="user-photo-placeholder">{{ user.name.charAt(0).toUpperCase() }}</div>
          <span class="user-name">{{ user.name }}</span>
          <span v-if="reviewerIds.includes(user.id)" class="checkmark">✓</span>
        </div>
      </div>
      
      <p class="selection-count">{{ reviewers().length }} reviewer(s) selected</p>
    </div>

    <!-- Step 3: Set Deadline -->
    <div v-if="step === 3" class="wizard-step">
      <h2>When is the deadline?</h2>
      <p class="hint">Set a clear deadline so reviewers know when to submit their feedback.</p>
      
      <div class="deadline-input">
        <input 
          type="datetime-local" 
          v-model="deadline"
          class="datetime-picker"
        >
      </div>
      
      <div class="deadline-preview" v-if="deadline">
        <strong>Selected:</strong> {{ formatDateTimeLocal(deadline) }}
      </div>
    </div>

    <!-- Step 4: Review & Confirm -->
    <div v-if="step === 4" class="wizard-step">
      <h2>Review and Create</h2>
      <p class="hint">Review the details before creating the feedback round.</p>
      
      <div class="review-card">
        <div class="review-section">
          <h4>Feedback Subject</h4>
          <div class="review-value">
            <img v-if="subject()?.photoUrl" :src="subject()!.photoUrl" class="review-photo">
            <div v-else class="review-photo-placeholder">{{ subject()?.name.charAt(0) }}</div>
            {{ subject()?.name }}
          </div>
        </div>
        
        <div class="review-section">
          <h4>Reviewers ({{ reviewers().length }})</h4>
          <div class="review-list">
            <span v-for="reviewer in reviewers()" :key="reviewer.id" class="review-tag">
              {{ reviewer.name }}
            </span>
          </div>
        </div>
        
        <div class="review-section">
          <h4>Deadline</h4>
          <p>{{ formatDateTimeLocal(deadline) }}</p>
        </div>
        
        <div class="review-section questions">
          <h4>Feedback Questions</h4>
          <ol>
            <li>What are this person's key strengths?</li>
            <li>What areas could this person improve?</li>
            <li>What specific behaviors or actions have you observed that stood out?</li>
            <li>What advice would you give to help this person grow?</li>
          </ol>
        </div>
      </div>
    </div>

    <!-- Navigation -->
    <div class="wizard-nav">
      <button 
        v-if="step > 1" 
        class="btn-secondary" 
        @click="prevStep"
        :disabled="loading"
      >
        Back
      </button>
      <button 
        v-if="step < 4" 
        class="btn-primary" 
        @click="nextStep"
      >
        Next
      </button>
      <button 
        v-if="step === 4" 
        class="btn-primary" 
        @click="createRound"
        :disabled="loading"
      >
        {{ loading ? 'Creating...' : 'Create Feedback Round' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.wizard-page {
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

.progress-bar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 2rem;
  padding: 1rem;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}

.progress-step {
  font-size: 0.85rem;
  color: #999;
  font-weight: 500;
}

.progress-step.active {
  color: #667eea;
}

.progress-step.complete {
  color: #4caf50;
}

.errors {
  background: #ffebee;
  color: #c62828;
  padding: 1rem;
  border-radius: 8px;
  margin-bottom: 1.5rem;
}

.errors p {
  margin: 0;
}

.wizard-step {
  background: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  margin-bottom: 1.5rem;
}

.wizard-step h2 {
  margin-bottom: 0.5rem;
}

.hint {
  color: #666;
  margin-bottom: 1.5rem;
}

.user-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 1rem;
}

.user-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem;
  border: 2px solid #e0e0e0;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
}

.user-card:hover {
  border-color: #667eea;
}

.user-card.selected {
  border-color: #667eea;
  background: #f5f7ff;
}

.user-photo {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  object-fit: cover;
}

.user-photo-placeholder {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
  font-weight: 600;
}

.user-name {
  font-size: 0.9rem;
  text-align: center;
}

.checkmark {
  position: absolute;
  top: 8px;
  right: 8px;
  background: #667eea;
  color: white;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
}

.selection-count {
  text-align: center;
  color: #666;
  margin-top: 1rem;
}

.deadline-input {
  display: flex;
  justify-content: center;
}

.datetime-picker {
  padding: 1rem;
  font-size: 1rem;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  min-width: 280px;
}

.deadline-preview {
  text-align: center;
  margin-top: 1rem;
  padding: 1rem;
  background: #f5f7fa;
  border-radius: 8px;
}

.review-card {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.review-section h4 {
  color: #666;
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 0.5rem;
}

.review-value {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 1.1rem;
}

.review-photo {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  object-fit: cover;
}

.review-photo-placeholder {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1rem;
  font-weight: 600;
}

.review-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.review-tag {
  background: #e3f2fd;
  color: #1976d2;
  padding: 0.4rem 0.8rem;
  border-radius: 16px;
  font-size: 0.85rem;
}

.questions ol {
  margin: 0;
  padding-left: 1.5rem;
}

.questions li {
  margin-bottom: 0.5rem;
  color: #444;
}

.wizard-nav {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
}

.btn-primary, .btn-secondary {
  padding: 0.875rem 2rem;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #667eea;
  color: white;
  border: none;
  margin-left: auto;
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

.btn-secondary:hover:not(:disabled) {
  background: #f5f5f5;
}
</style>
