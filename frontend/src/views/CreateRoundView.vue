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
  <div class="wizard">
    <header class="wizard__header">
      <h1 class="wizard__title">Create Feedback Round</h1>
      <p class="wizard__subtitle">Step {{ step }} of 4</p>
    </header>

    <!-- Progress Bar -->
    <div class="wizard__progress">
      <div class="wizard__progress-step" :class="{ 'wizard__progress-step--active': step >= 1, 'wizard__progress-step--complete': step > 1 }">1. Subject</div>
      <div class="wizard__progress-step" :class="{ 'wizard__progress-step--active': step >= 2, 'wizard__progress-step--complete': step > 2 }">2. Reviewers</div>
      <div class="wizard__progress-step" :class="{ 'wizard__progress-step--active': step >= 3, 'wizard__progress-step--complete': step > 3 }">3. Deadline</div>
      <div class="wizard__progress-step" :class="{ 'wizard__progress-step--active': step >= 4 }">4. Confirm</div>
    </div>

    <!-- Errors -->
    <div v-if="errors.length" class="wizard__errors">
      <p v-for="error in errors" :key="error" class="wizard__error">{{ error }}</p>
    </div>

    <!-- Step 1: Select Subject -->
    <div v-if="step === 1" class="wizard__step">
      <h2 class="wizard__step-title">Who is receiving feedback?</h2>
      <p class="wizard__step-hint">Select the team member who will receive feedback from their peers.</p>

      <div class="user-grid">
        <div
          v-for="user in users.filter(u => u.id !== auth.user?.id)"
          :key="user.id"
          class="user-card"
          :class="{ 'user-card--selected': subjectId === user.id }"
          @click="subjectId = user.id"
        >
          <img v-if="user.photoUrl" :src="user.photoUrl" :alt="user.name" class="user-card__photo">
          <div v-else class="user-card__photo-placeholder">{{ user.name.charAt(0).toUpperCase() }}</div>
          <span class="user-card__name">{{ user.name }}</span>
        </div>
      </div>
    </div>

    <!-- Step 2: Select Reviewers -->
    <div v-if="step === 2" class="wizard__step">
      <h2 class="wizard__step-title">Who will provide feedback?</h2>
      <p class="wizard__step-hint">Select 3-5 peers who will give anonymous feedback. The subject and you (the admin) are automatically excluded.</p>

      <div class="user-grid">
        <div
          v-for="user in availableReviewers()"
          :key="user.id"
          class="user-card"
          :class="{ 'user-card--selected': reviewerIds.includes(user.id) }"
          @click="toggleReviewer(user.id)"
        >
          <img v-if="user.photoUrl" :src="user.photoUrl" :alt="user.name" class="user-card__photo">
          <div v-else class="user-card__photo-placeholder">{{ user.name.charAt(0).toUpperCase() }}</div>
          <span class="user-card__name">{{ user.name }}</span>
          <span v-if="reviewerIds.includes(user.id)" class="user-card__check">✓</span>
        </div>
      </div>

      <p class="wizard__selection-count">{{ reviewers().length }} reviewer(s) selected</p>
    </div>

    <!-- Step 3: Set Deadline -->
    <div v-if="step === 3" class="wizard__step">
      <h2 class="wizard__step-title">When is the deadline?</h2>
      <p class="wizard__step-hint">Set a clear deadline so reviewers know when to submit their feedback.</p>

      <div class="wizard__deadline-input">
        <input
          type="datetime-local"
          v-model="deadline"
          class="wizard__datetime-picker"
        >
      </div>

      <div class="wizard__deadline-preview" v-if="deadline">
        <strong>Selected:</strong> {{ formatDateTimeLocal(deadline) }}
      </div>
    </div>

    <!-- Step 4: Review & Confirm -->
    <div v-if="step === 4" class="wizard__step">
      <h2 class="wizard__step-title">Review and Create</h2>
      <p class="wizard__step-hint">Review the details before creating the feedback round.</p>

      <div class="wizard__review">
        <div class="review-section">
          <h4 class="review-section__title">Feedback Subject</h4>
          <div class="review-section__value">
            <img v-if="subject()?.photoUrl" :src="subject()!.photoUrl" class="review-section__photo">
            <div v-else class="review-section__photo-placeholder">{{ subject()?.name.charAt(0) }}</div>
            {{ subject()?.name }}
          </div>
        </div>

        <div class="review-section">
          <h4 class="review-section__title">Reviewers ({{ reviewers().length }})</h4>
          <div class="review-section__list">
            <span v-for="reviewer in reviewers()" :key="reviewer.id" class="review-section__tag">
              {{ reviewer.name }}
            </span>
          </div>
        </div>

        <div class="review-section">
          <h4 class="review-section__title">Deadline</h4>
          <p class="review-section__text">{{ formatDateTimeLocal(deadline) }}</p>
        </div>

        <div class="review-section">
          <h4 class="review-section__title">Feedback Questions</h4>
          <ol class="review-section__questions">
            <li>What are this person's key strengths?</li>
            <li>What areas could this person improve?</li>
            <li>What specific behaviors or actions have you observed that stood out?</li>
            <li>What advice would you give to help this person grow?</li>
          </ol>
        </div>
      </div>
    </div>

    <!-- Navigation -->
    <div class="wizard__nav">
      <button
        v-if="step > 1"
        class="btn btn--secondary"
        @click="prevStep"
        :disabled="loading"
      >
        Back
      </button>
      <button
        v-if="step < 4"
        class="btn btn--primary wizard__nav-next"
        @click="nextStep"
      >
        Next
      </button>
      <button
        v-if="step === 4"
        class="btn btn--primary wizard__nav-next"
        @click="createRound"
        :disabled="loading"
      >
        {{ loading ? 'Creating...' : 'Create Feedback Round' }}
      </button>
    </div>
  </div>
</template>

<style scoped lang="scss">
.wizard {
  padding: 1rem;
  max-width: 900px;
  margin: 0 auto;

  @media (min-width: 768px) {
    padding: 2rem;
  }

  &__header {
    margin-bottom: 1.5rem;

    @media (min-width: 768px) {
      margin-bottom: 2rem;
    }
  }

  &__title {
    font-size: 1.5rem;
    margin: 0 0 0.5rem 0;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 2rem;
    }
  }

  &__subtitle {
    color: var(--text-secondary);
    margin: 0;
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__progress {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 1.5rem;
    padding: 1rem;
    background: var(--bg-primary);
    border-radius: 12px;
    border: 1px solid var(--border-color);

    @media (min-width: 768px) {
      flex-direction: row;
      justify-content: space-between;
      margin-bottom: 2rem;
    }
  }

  &__progress-step {
    font-size: 0.8rem;
    color: var(--text-tertiary);
    font-weight: 500;
    padding: 0.25rem 0;

    @media (min-width: 768px) {
      font-size: 0.85rem;
      padding: 0;
    }

    &--active {
      color: var(--color-primary);
    }

    &--complete {
      color: var(--color-success);
    }
  }

  &__errors {
    background: rgba(244, 67, 54, 0.1);
    color: var(--color-error);
    padding: 1rem;
    border-radius: 8px;
    margin-bottom: 1.5rem;
    border: 1px solid var(--color-error);
  }

  &__error {
    margin: 0;
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__step {
    background: var(--bg-primary);
    padding: 1.25rem;
    border-radius: 12px;
    border: 1px solid var(--border-color);
    margin-bottom: 1.5rem;

    @media (min-width: 768px) {
      padding: 2rem;
    }
  }

  &__step-title {
    margin: 0 0 0.5rem 0;
    font-size: 1.25rem;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 1.5rem;
    }
  }

  &__step-hint {
    color: var(--text-secondary);
    margin: 0 0 1.5rem 0;
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__selection-count {
    text-align: center;
    color: var(--text-secondary);
    margin-top: 1rem;
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__deadline-input {
    display: flex;
    justify-content: center;
  }

  &__datetime-picker {
    padding: 1rem;
    font-size: 0.9rem;
    border: 2px solid var(--border-color);
    border-radius: 8px;
    min-width: 260px;
    background: var(--bg-primary);
    color: var(--text-primary);
    min-height: 44px;

    @media (min-width: 768px) {
      font-size: 1rem;
      min-width: 280px;
    }

    &:focus {
      outline: none;
      border-color: var(--color-primary);
      box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
    }
  }

  &__deadline-preview {
    text-align: center;
    margin-top: 1rem;
    padding: 1rem;
    background: var(--bg-secondary);
    border-radius: 8px;
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__review {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;

    @media (min-width: 768px) {
      gap: 1.5rem;
    }
  }

  &__nav {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;

    @media (min-width: 768px) {
      flex-direction: row;
      justify-content: space-between;
      gap: 1rem;
    }
  }

  &__nav-next {
    @media (min-width: 768px) {
      margin-left: auto;
    }
  }
}

.user-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 0.75rem;

  @media (min-width: 768px) {
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 1rem;
  }
}

.user-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem;
  border: 2px solid var(--border-color);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
  min-height: 110px;
  background: var(--bg-primary);

  &:hover {
    border-color: var(--color-primary);
  }

  &--selected {
    border-color: var(--color-primary);
    background: rgba(102, 126, 234, 0.05);
  }

  &__photo {
    width: 44px;
    height: 44px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;

    @media (min-width: 768px) {
      width: 48px;
      height: 48px;
    }
  }

  &__photo-placeholder {
    width: 44px;
    height: 44px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.1rem;
    font-weight: 600;
    flex-shrink: 0;

    @media (min-width: 768px) {
      width: 48px;
      height: 48px;
      font-size: 1.25rem;
    }
  }

  &__name {
    font-size: 0.85rem;
    text-align: center;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }
  }

  &__check {
    position: absolute;
    top: 8px;
    right: 8px;
    background: var(--color-primary);
    color: white;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.7rem;

    @media (min-width: 768px) {
      width: 20px;
      height: 20px;
      font-size: 0.75rem;
    }
  }
}

.review-section {
  &__title {
    color: var(--text-secondary);
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin: 0 0 0.5rem 0;

    @media (min-width: 768px) {
      font-size: 0.85rem;
    }
  }

  &__value {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    font-size: 1rem;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 1.1rem;
    }
  }

  &__photo {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;

    @media (min-width: 768px) {
      width: 36px;
      height: 36px;
    }
  }

  &__photo-placeholder {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.9rem;
    font-weight: 600;
    flex-shrink: 0;

    @media (min-width: 768px) {
      width: 36px;
      height: 36px;
      font-size: 1rem;
    }
  }

  &__list {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  &__tag {
    background: rgba(33, 150, 243, 0.1);
    color: var(--color-info);
    padding: 0.375rem 0.75rem;
    border-radius: 16px;
    font-size: 0.8rem;

    @media (min-width: 768px) {
      font-size: 0.85rem;
    }
  }

  &__text {
    margin: 0;
    color: var(--text-primary);
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__questions {
    margin: 0;
    padding-left: 1.25rem;
    color: var(--text-primary);
    font-size: 0.9rem;

    @media (min-width: 768px) {
      padding-left: 1.5rem;
      font-size: 1rem;
    }

    li {
      margin-bottom: 0.5rem;
    }
  }
}
</style>
