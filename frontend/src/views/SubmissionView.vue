<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'
import { PhNotePencil, PhWarningCircle } from '@phosphor-icons/vue'

const route = useRoute()
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
  { key: 'a', text: 'What does this person do that has the biggest positive impact on the team or product? Where possible, share one concrete example (Situation → Behaviour → Impact).' },
  { key: 'b', text: 'Looking at the last 3–6 months, what\'s currently holding this person back from their next level of impact (skill, habit, or environment)?' },
  { key: 'c', text: 'If this person doubled down on one strength over the next 6 months, what should it be — and what would change for the team?' },
  { key: 'd', text: 'What\'s one concrete experiment or focus area you\'d suggest they try in the next 30–60 days?' }
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
  <div class="submission">
    <div v-if="loading" class="submission__loading">Loading...</div>

    <template v-else-if="round && submission">
      <header class="submission__header">
        <router-link to="/rounds" class="submission__back">← Back to Rounds</router-link>
        <h1 class="submission__title">My Feedback Submission</h1>
        <p class="submission__subtitle">
          For <strong class="submission__subject">{{ round.subject?.name }}</strong> •
          Submitted {{ formatDate(submission.submittedAt) }}
        </p>
      </header>

      <div class="submission__content">
        <div class="submission__meta">
          <div class="submission__meta-info">
            <span class="submission__meta-item">Status: <span class="badge badge--success">Submitted</span></span>
            <span class="submission__meta-item">Round: {{ round.status }}</span>
            <span v-if="submission.updatedAt && submission.updatedAt !== submission.submittedAt" class="submission__meta-item">
              Last updated: {{ formatDate(submission.updatedAt) }}
            </span>
          </div>
          <div v-if="canEdit()" class="submission__meta-actions">
            <button @click="startEditing" class="btn btn--secondary btn--with-icon submission__edit-btn">
              <PhNotePencil :size="16" weight="regular" />
              <span>Edit Feedback</span>
            </button>
          </div>
        </div>

        <div v-if="!editing" class="submission__responses">
          <h2 class="submission__responses-title">Your Feedback Responses</h2>

          <div class="response-item">
            <h3 class="response-item__question">1. What does this person do that has the biggest positive impact on the team or product?</h3>
            <div class="response-item__answer">
              {{ submission.responsesParsed?.a || 'No response provided' }}
            </div>
          </div>

          <div class="response-item">
            <h3 class="response-item__question">2. What's currently holding this person back from their next level of impact?</h3>
            <div class="response-item__answer">
              {{ submission.responsesParsed?.b || 'No response provided' }}
            </div>
          </div>

          <div class="response-item">
            <h3 class="response-item__question">3. If they doubled down on one strength, what should it be — and what would change for the team?</h3>
            <div class="response-item__answer">
              {{ submission.responsesParsed?.c || 'No response provided' }}
            </div>
          </div>

          <div class="response-item">
            <h3 class="response-item__question">4. One concrete experiment or focus area for the next 30–60 days?</h3>
            <div class="response-item__answer">
              {{ submission.responsesParsed?.d || 'No response provided' }}
            </div>
          </div>
        </div>

        <div v-else class="submission__edit">
          <h2 class="submission__edit-title">Edit Your Feedback</h2>

          <div class="submission__edit-form">
            <div v-for="question in questions" :key="question.key" class="edit-question">
              <h3 class="edit-question__label">{{ question.text }}</h3>
              <textarea
                v-model="editResponses[question.key]"
                placeholder="Enter your response..."
                rows="3"
                class="edit-question__textarea"
              ></textarea>
            </div>
          </div>

          <div class="submission__edit-actions">
            <button @click="saveEdits" class="btn btn--primary">Save Changes</button>
            <button @click="cancelEditing" class="btn btn--secondary">Cancel</button>
          </div>
        </div>

        <div v-if="round.status !== 'active'" class="submission__notice">
          <p class="submission__notice-text">
            <PhWarningCircle :size="18" weight="fill" />
            <span>This round is {{ round.status }}. Feedback can only be edited while the round is active.</span>
          </p>
        </div>
      </div>
    </template>

    <div v-else class="submission__error">
      <p class="submission__error-text">Submission not found.</p>
      <router-link to="/rounds" class="btn btn--primary">Back to Rounds</router-link>
    </div>
  </div>
</template>

<style scoped lang="scss">
.submission {
  padding: 1rem;
  max-width: 800px;
  margin: 0 auto;

  @media (min-width: 768px) {
    padding: 2rem;
  }

  &__loading,
  &__error {
    text-align: center;
    padding: 2rem 1rem;

    @media (min-width: 768px) {
      padding: 3rem;
    }
  }

  &__error-text {
    color: var(--color-error);
    margin-bottom: 1.5rem;
  }

  &__header {
    margin-bottom: 2rem;
  }

  &__back {
    display: block;
    margin-bottom: 1rem;
    color: var(--color-primary);
    text-decoration: none;
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }

    &:hover {
      text-decoration: underline;
    }
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
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__subject {
    color: var(--text-primary);
  }

  &__content {
    background: var(--bg-primary);
    border-radius: 12px;
    border: 1px solid var(--border-color);
    overflow: hidden;
  }

  &__meta {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding: 1.25rem;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border-color);

    @media (min-width: 768px) {
      flex-direction: row;
      justify-content: space-between;
      align-items: center;
      padding: 1.5rem;
    }
  }

  &__meta-info {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    color: var(--text-secondary);
    font-size: 0.85rem;

    @media (min-width: 768px) {
      flex-direction: row;
      gap: 1rem;
      align-items: center;
      font-size: 0.9rem;
    }
  }

  &__meta-item {
    display: block;
  }

  &__meta-actions {
    display: flex;
    gap: 0.5rem;
  }

  &__edit-btn {
    font-size: 0.8rem;
    padding: 0.375rem 0.75rem;

    @media (min-width: 768px) {
      font-size: 0.85rem;
    }
  }

  &__responses {
    padding: 1.25rem;

    @media (min-width: 768px) {
      padding: 1.5rem;
    }
  }

  &__responses-title {
    font-size: 1.125rem;
    margin-bottom: 1.5rem;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 1.25rem;
    }
  }

  &__edit {
    padding: 1.25rem;

    @media (min-width: 768px) {
      padding: 1.5rem;
    }
  }

  &__edit-title {
    font-size: 1.125rem;
    margin-bottom: 1.5rem;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 1.25rem;
    }
  }

  &__edit-form {
    margin-bottom: 1.5rem;
  }

  &__edit-actions {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;

    @media (min-width: 768px) {
      flex-direction: row;
      gap: 0.5rem;
      justify-content: flex-end;
    }
  }

  &__notice {
    background: rgba(255, 152, 0, 0.1);
    border: 1px solid var(--color-warning);
    padding: 1rem;
    border-radius: 8px;
    margin: 1rem;
  }

  &__notice-text {
    display: inline-flex;
    align-items: flex-start;
    gap: 0.5rem;
    color: var(--text-primary);
    margin: 0;
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }
}

.response-item {
  margin-bottom: 1.5rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid var(--border-color);

  @media (min-width: 768px) {
    margin-bottom: 2rem;
    padding-bottom: 2rem;
  }

  &:last-child {
    margin-bottom: 0;
    padding-bottom: 0;
    border-bottom: none;
  }

  &__question {
    font-size: 0.95rem;
    margin-bottom: 0.75rem;
    color: var(--color-primary);
    font-weight: 600;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__answer {
    background: var(--bg-secondary);
    padding: 1rem;
    border-radius: 8px;
    border-left: 4px solid var(--color-primary);
    line-height: 1.6;
    color: var(--text-primary);
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }
}

.edit-question {
  margin-bottom: 1.25rem;

  @media (min-width: 768px) {
    margin-bottom: 1.5rem;
  }

  &__label {
    font-size: 0.95rem;
    margin-bottom: 0.5rem;
    color: var(--color-primary);
    font-weight: 600;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__textarea {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid var(--border-color);
    border-radius: 8px;
    font-family: inherit;
    font-size: 0.9rem;
    resize: vertical;
    line-height: 1.5;
    background: var(--bg-primary);
    color: var(--text-primary);
    min-height: 80px;

    @media (min-width: 768px) {
      font-size: 1rem;
    }

    &:focus {
      outline: none;
      border-color: var(--color-primary);
    }
  }
}
</style>
