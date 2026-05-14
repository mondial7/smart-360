<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import apiClient from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import type { FeedbackRound, RoundTemplate } from '@/types/round'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const roundId = route.params.id as string  // Changed from parseInt to string

const round = ref<FeedbackRound | null>(null)
const template = ref<RoundTemplate | null>(null)
const loading = ref(true)
const submitting = ref(false)
const alreadySubmitted = ref(false)

const isSelf = computed(() => !!round.value && round.value.subjectId === auth.user?.id)

// Question prompts come from the round's template. We render whatever the
// template defines — including its `key` — so a template can ship with a
// custom number or ordering of questions without code changes here.
const questions = computed(() => {
  if (!template.value) return []
  return template.value.questions.map(q => ({
    key: q.key,
    text: isSelf.value ? q.selfText : q.peerText
  }))
})

const responses = ref<Record<string, string>>({})

type Relationship = 'manager' | 'report' | 'peer' | 'cross_functional'
type Frequency = 'daily' | 'weekly' | 'monthly' | 'rarely'

const relationship = ref<Relationship | ''>('')
const interactionFrequency = ref<Frequency | ''>('')

const relationshipOptions: Array<{ value: Relationship; label: string }> = [
  { value: 'manager', label: 'I manage them' },
  { value: 'report', label: 'They manage me' },
  { value: 'peer', label: 'We are peers / teammates' },
  { value: 'cross_functional', label: 'We collaborate across teams' }
]

const frequencyOptions: Array<{ value: Frequency; label: string }> = [
  { value: 'daily', label: 'Daily — we work together most days' },
  { value: 'weekly', label: 'Weekly — we sync at least once a week' },
  { value: 'monthly', label: 'Monthly — we connect occasionally' },
  { value: 'rarely', label: 'Rarely — limited direct interaction' }
]

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

      // Load the template referenced by the round — falls back to the default
      // slug if the round predates configurable templates.
      const templateKey = round.value?.templateId || 'default'
      const templateRes = await apiClient.get(`/templates/${templateKey}`)
      template.value = templateRes.data

      // Seed an empty response slot for every question key defined by the template.
      const slots: Record<string, string> = {}
      for (const q of template.value?.questions || []) {
        slots[q.key] = ''
      }
      responses.value = slots
    }
  } catch (error) {
    console.error('Failed to load round:', error)
  } finally {
    loading.value = false
  }
}

async function submitFeedback() {
  if (!isSelf.value) {
    if (!relationship.value) {
      alert('Please pick your relationship to this person before submitting.')
      return
    }
    if (!interactionFrequency.value) {
      alert('Please pick how often you interact with this person before submitting.')
      return
    }
  }
  for (const q of questions.value) {
    if (!responses.value[q.key].trim()) {
      alert('Please answer all questions before submitting.')
      return
    }
  }

  submitting.value = true
  try {
    const payload: Record<string, unknown> = {
      roundId,
      responses: JSON.stringify(responses.value)
    }
    if (!isSelf.value) {
      payload.relationship = relationship.value
      payload.interactionFrequency = interactionFrequency.value
    }
    await apiClient.post('/submissions', payload)
    alert(isSelf.value
      ? 'Self-assessment submitted. Thanks for taking the time to reflect.'
      : 'Feedback submitted successfully! Thank you for your input.')
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
        <h1 class="submit__title">{{ isSelf ? 'Your Self-Assessment' : 'Submit Feedback' }}</h1>
        <p v-if="isSelf" class="submit__subtitle">
          This is your half of the 360. Your answers will be compared against peer feedback so you can see where you and your team are aligned, and where the gaps are.
        </p>
        <p v-else class="submit__subtitle">
          Anonymous feedback for <strong class="submit__subject">{{ round.subject?.name }}</strong>
        </p>
        <p class="submit__deadline">Deadline: {{ formatDate(round.deadline) }}</p>
      </header>

      <div class="submit__banner">
        <span class="submit__banner-icon">{{ isSelf ? '🪞' : '🔒' }}</span>
        <span class="submit__banner-text">
          {{ isSelf
            ? 'Be candid — the value of a self-assessment is the gap between how you see yourself and how others see you.'
            : 'Your feedback is completely anonymous. The subject will never know who wrote what.' }}
        </span>
      </div>

      <form @submit.prevent="submitFeedback" class="submit__form">
        <fieldset v-if="!isSelf" class="submit__context">
          <legend class="submit__context-legend">Your vantage point</legend>
          <p class="submit__context-hint">
            This helps the consolidation weight signals — a daily peer's view carries different evidentiary weight than a one-off contact.
          </p>
          <div class="submit__context-row">
            <label for="relationship" class="submit__context-label">Your relationship</label>
            <select id="relationship" v-model="relationship" required class="submit__context-select">
              <option value="" disabled>Choose…</option>
              <option v-for="opt in relationshipOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
            </select>
          </div>
          <div class="submit__context-row">
            <label for="frequency" class="submit__context-label">How often you interact</label>
            <select id="frequency" v-model="interactionFrequency" required class="submit__context-select">
              <option value="" disabled>Choose…</option>
              <option v-for="opt in frequencyOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
            </select>
          </div>
        </fieldset>

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
            {{ submitting ? 'Submitting...' : (isSelf ? 'Submit Self-Assessment' : 'Submit Feedback') }}
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

.submit__context {
  background: var(--bg-primary);
  padding: 1.25rem;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  display: grid;
  gap: 0.75rem;
}

.submit__context-legend {
  font-weight: 600;
  font-size: 1rem;
  color: var(--text-primary);
  padding: 0 0.4rem;
}

.submit__context-hint {
  margin: 0;
  font-size: 0.85rem;
  color: var(--text-secondary);
  line-height: 1.4;
}

.submit__context-row {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;

  @media (min-width: 640px) {
    flex-direction: row;
    align-items: center;
    gap: 0.75rem;
  }
}

.submit__context-label {
  font-weight: 500;
  color: var(--text-primary);
  font-size: 0.95rem;

  @media (min-width: 640px) {
    flex: 0 0 14rem;
  }
}

.submit__context-select {
  flex: 1;
  padding: 0.55rem 0.7rem;
  border: 2px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-primary);
  font-size: 0.95rem;
  color: var(--text-primary);
  font-family: inherit;

  &:focus {
    outline: none;
    border-color: var(--color-primary);
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
