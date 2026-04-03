<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import apiClient from '@/api/client'
import type { FeedbackRound } from '@/types/round'

const route = useRoute()
const router = useRouter()
const roundId = route.params.id as string  // Changed from parseInt to string

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
      responses: JSON.stringify(responses.value)
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
  <div class="submit">
    <div v-if="loading" class="submit__loading">Loading...</div>

    <div v-else-if="alreadySubmitted" class="submit__submitted">
      <h2 class="submit__submitted-title">Already Submitted</h2>
      <p class="submit__submitted-text">You have already submitted feedback for this round.</p>
      <router-link to="/rounds" class="btn btn--primary">Back to Rounds</router-link>
    </div>

    <template v-else-if="round">
      <header class="submit__header">
        <h1 class="submit__title">Submit Feedback</h1>
        <p class="submit__subtitle">
          Anonymous feedback for <strong class="submit__subject">{{ round.subject?.name }}</strong>
        </p>
        <p class="submit__deadline">Deadline: {{ formatDate(round.deadline) }}</p>
      </header>

      <div class="submit__banner">
        <span class="submit__banner-icon">🔒</span>
        <span class="submit__banner-text">Your feedback is completely anonymous. The subject will never know who wrote what.</span>
      </div>

      <form @submit.prevent="submitFeedback" class="submit__form">
        <div v-for="question in questions" :key="question.key" class="question">
          <label :for="question.key" class="question__label">{{ question.text }}</label>
          <textarea
            :id="question.key"
            v-model="responses[question.key]"
            rows="4"
            placeholder="Type your answer here..."
            required
            class="question__textarea"
          ></textarea>
        </div>

        <div class="submit__actions">
          <router-link to="/rounds" class="btn btn--secondary">Cancel</router-link>
          <button type="submit" class="btn btn--primary" :disabled="submitting">
            {{ submitting ? 'Submitting...' : 'Submit Feedback' }}
          </button>
        </div>
      </form>
    </template>

    <div v-else class="submit__error">
      <p class="submit__error-text">Failed to load feedback round.</p>
      <router-link to="/rounds" class="btn btn--primary">Back to Rounds</router-link>
    </div>
  </div>
</template>

<style scoped lang="scss">
.submit {
  padding: 1rem;
  max-width: 800px;
  margin: 0 auto;

  @media (min-width: 768px) {
    padding: 2rem;
  }

  &__loading,
  &__submitted,
  &__error {
    text-align: center;
    padding: 2rem 1rem;

    @media (min-width: 768px) {
      padding: 3rem;
    }
  }

  &__submitted {
    background: var(--bg-secondary);
    border-radius: 12px;
    border: 1px solid var(--border-color);
  }

  &__submitted-title {
    margin-bottom: 1rem;
    color: var(--color-success);
    font-size: 1.5rem;

    @media (min-width: 768px) {
      font-size: 1.75rem;
    }
  }

  &__submitted-text {
    color: var(--text-secondary);
    margin-bottom: 1.5rem;
  }

  &__error-text {
    color: var(--color-error);
    margin-bottom: 1.5rem;
  }

  &__header {
    margin-bottom: 2rem;
  }

  &__title {
    font-size: 1.5rem;
    margin-bottom: 0.5rem;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 2rem;
    }
  }

  &__subtitle {
    color: var(--text-secondary);
    font-size: 1rem;
    margin: 0;

    @media (min-width: 768px) {
      font-size: 1.1rem;
    }
  }

  &__subject {
    color: var(--color-primary);
  }

  &__deadline {
    color: var(--color-error);
    font-weight: 500;
    margin-top: 0.5rem;
  }

  &__banner {
    background: rgba(33, 150, 243, 0.1);
    border-left: 4px solid #1976d2;
    padding: 1rem;
    border-radius: 0 8px 8px 0;
    margin-bottom: 2rem;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    color: #1976d2;
  }

  &__banner-icon {
    font-size: 1.125rem;
    flex-shrink: 0;

    @media (min-width: 768px) {
      font-size: 1.25rem;
    }
  }

  &__banner-text {
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__form {
    display: flex;
    flex-direction: column;
    gap: 1rem;

    @media (min-width: 768px) {
      gap: 1.5rem;
    }
  }

  &__actions {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-top: 1rem;

    @media (min-width: 768px) {
      flex-direction: row;
      justify-content: flex-end;
      gap: 1rem;
    }
  }
}

.question {
  background: var(--bg-primary);
  padding: 1.25rem;
  border-radius: 12px;
  border: 1px solid var(--border-color);

  @media (min-width: 768px) {
    padding: 1.5rem;
  }

  &__label {
    display: block;
    font-weight: 500;
    margin-bottom: 0.75rem;
    color: var(--text-primary);
    font-size: 1rem;

    @media (min-width: 768px) {
      font-size: 1.1rem;
    }
  }

  &__textarea {
    width: 100%;
    padding: 0.75rem;
    border: 2px solid var(--border-color);
    border-radius: 8px;
    font-size: 0.95rem;
    resize: vertical;
    font-family: inherit;
    background: var(--bg-primary);
    color: var(--text-primary);
    min-height: 100px;

    @media (min-width: 768px) {
      font-size: 1rem;
    }

    &:focus {
      outline: none;
      border-color: var(--color-primary);
    }

    &::placeholder {
      color: var(--text-tertiary);
    }
  }
}
</style>
