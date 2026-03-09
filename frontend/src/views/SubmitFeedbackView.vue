<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import apiClient from '@/api/client'
import type { FeedbackRound } from '@/types/round'

const route = useRoute()
const router = useRouter()
const roundId = parseInt(route.params.id as string)

const round = ref<FeedbackRound | null>(null)
const loading = ref(true)
const submitting = ref(false)
const alreadySubmitted = ref(false)

const questions = [
  { key: 'a', text: 'What are this person\'s key strengths?' },
  { key: 'b', text: 'What areas could this person improve?' },
  { key: 'c', text: 'What specific behaviors or actions have you observed that stood out?' },
  { key: 'd', text: 'What advice would you give to help this person grow?' }
]

const responses = ref<Record<string, string>>({
  a: '',
  b: '',
  c: '',
  d: ''
})

onMounted(async () => {
  await loadRound()
})

async function loadRound() {
  try {
    // Check if already submitted
    const checkRes = await apiClient.get(`/submissions/check/${roundId}`)
    alreadySubmitted.value = checkRes.data.submitted

    if (!alreadySubmitted.value) {
      // Load round details
      const roundRes = await apiClient.get(`/rounds/${roundId}`)
      round.value = roundRes.data
    }
  } catch (error) {
    console.error('Failed to load round:', error)
  } finally {
    loading.value = false
  }
}

async function submitFeedback() {
  // Validate all questions answered
  for (const q of questions) {
    if (!responses.value[q.key].trim()) {
      alert('Please answer all questions before submitting.')
      return
    }
  }

  submitting.value = true
  try {
    await apiClient.post('/submissions', {
      roundId,
      responses: responses.value
    })
    alert('Feedback submitted successfully! Thank you for your input.')
    router.push('/rounds')
  } catch (error: any) {
    alert(error.response?.data?.error || 'Failed to submit feedback. Please try again.')
  } finally {
    submitting.value = false
  }
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return 'No deadline'
  return new Date(dateStr).toLocaleDateString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<template>
  <div class="submission-page">
    <div v-if="loading" class="loading">Loading...</div>

    <div v-else-if="alreadySubmitted" class="submitted-state">
      <h2>Already Submitted</h2>
      <p>You have already submitted feedback for this round.</p>
      <router-link to="/rounds" class="btn-primary">Back to Rounds</router-link>
    </div>

    <template v-else-if="round">
      <header class="page-header">
        <h1>Submit Feedback</h1>
        <p class="subtitle">
          Anonymous feedback for <strong>{{ round.subject?.name }}</strong>
        </p>
        <p class="deadline">Deadline: {{ formatDate(round.deadline) }}</p>
      </header>

      <div class="anonymity-banner">
        <span class="icon">🔒</span>
        <span>Your feedback is completely anonymous. The subject will never know who wrote what.</span>
      </div>

      <form @submit.prevent="submitFeedback" class="feedback-form">
        <div v-for="question in questions" :key="question.key" class="question-card">
          <label :for="question.key">{{ question.text }}</label>
          <textarea
            :id="question.key"
            v-model="responses[question.key]"
            rows="4"
            placeholder="Type your answer here..."
            required
          ></textarea>
        </div>

        <div class="form-actions">
          <router-link to="/rounds" class="btn-secondary">Cancel</router-link>
          <button type="submit" class="btn-primary" :disabled="submitting">
            {{ submitting ? 'Submitting...' : 'Submit Feedback' }}
          </button>
        </div>
      </form>
    </template>

    <div v-else class="error-state">
      <p>Failed to load feedback round.</p>
      <router-link to="/rounds" class="btn-primary">Back to Rounds</router-link>
    </div>
  </div>
</template>

<style scoped>
.submission-page {
  padding: 2rem;
  max-width: 800px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 2rem;
}

.page-header h1 {
  font-size: 2rem;
  margin-bottom: 0.5rem;
}

.subtitle {
  color: #666;
  font-size: 1.1rem;
}

.subtitle strong {
  color: #667eea;
}

.deadline {
  color: #f44336;
  font-weight: 500;
  margin-top: 0.5rem;
}

.anonymity-banner {
  background: #e3f2fd;
  border-left: 4px solid #1976d2;
  padding: 1rem;
  border-radius: 0 8px 8px 0;
  margin-bottom: 2rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  color: #1976d2;
}

.icon {
  font-size: 1.25rem;
}

.feedback-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.question-card {
  background: white;
  padding: 1.5rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}

.question-card label {
  display: block;
  font-weight: 500;
  margin-bottom: 0.75rem;
  color: #333;
  font-size: 1.1rem;
}

.question-card textarea {
  width: 100%;
  padding: 0.75rem;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 1rem;
  resize: vertical;
  font-family: inherit;
}

.question-card textarea:focus {
  outline: none;
  border-color: #667eea;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1rem;
}

.btn-primary, .btn-secondary {
  padding: 0.875rem 2rem;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  text-decoration: none;
  display: inline-block;
  text-align: center;
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

.loading, .submitted-state, .error-state {
  text-align: center;
  padding: 3rem;
}

.submitted-state {
  background: #f5f5f5;
  border-radius: 12px;
}

.submitted-state h2 {
  margin-bottom: 1rem;
  color: #4caf50;
}

.submitted-state p {
  color: #666;
  margin-bottom: 1.5rem;
}
</style>
