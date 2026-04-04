<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'
import type { Team } from '@/types/team'
import type { User } from '@/types/user'
import type { CreateTeamRoundsRequest } from '@/types/team'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const step = ref(1)
const loading = ref(false)
const team = ref<Team | null>(null)
const selectedSubjectIds = ref<string[]>([])
const deadlines = ref<Record<string, string>>({}) // subjectId -> deadline string

// Errors
const errors = ref<string[]>([])

onMounted(async () => {
  const teamId = route.params.teamId as string
  if (teamId) {
    await loadTeam(teamId)
  }
})

async function loadTeam(id: string) {
  try {
    const response = await apiClient.get(`/teams/${id}`)
    team.value = response.data
    initializeDeadlines()
  } catch (error) {
    console.error('Failed to load team:', error)
    errors.value.push('Failed to load team')
  }
}

function initializeDeadlines() {
  if (!team.value) return

  // Set default deadline (2 weeks from now) for all potential subjects
  const twoWeeks = new Date()
  twoWeeks.setDate(twoWeeks.getDate() + 14)
  const defaultDeadline = twoWeeks.toISOString().slice(0, 16)

  team.value.members.forEach(member => {
    if (member.id !== auth.user?.id) {
      deadlines.value[member.id] = defaultDeadline
    }
  })
}

const availableSubjects = computed(() => {
  if (!team.value) return []
  // All members except the current user (creator)
  return team.value.members.filter(m => m.id !== auth.user?.id)
})

const selectedSubjects = computed(() => {
  return availableSubjects.value.filter(s => selectedSubjectIds.value.includes(s.id))
})

const reviewerCount = computed(() => {
  if (!team.value) return 0
  // All team members except subject and creator
  return team.value.members.length - 2
})

function toggleSubject(id: string) {
  const index = selectedSubjectIds.value.indexOf(id)
  if (index > -1) {
    selectedSubjectIds.value.splice(index, 1)
  } else {
    selectedSubjectIds.value.push(id)
  }
}

function setAllDeadlines(deadline: string) {
  selectedSubjectIds.value.forEach(id => {
    deadlines.value[id] = deadline
  })
}

function validateStep(): boolean {
  errors.value = []

  if (step.value === 2 && selectedSubjectIds.value.length === 0) {
    errors.value.push('Please select at least one team member to receive feedback')
    return false
  }

  if (step.value === 3) {
    for (const subjectId of selectedSubjectIds.value) {
      if (!deadlines.value[subjectId]) {
        errors.value.push('Please set a deadline for all selected subjects')
        return false
      }
      const deadlineDate = new Date(deadlines.value[subjectId])
      if (deadlineDate <= new Date()) {
        errors.value.push('All deadlines must be in the future')
        return false
      }
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
  errors.value = []
}

async function createRounds() {
  if (!team.value || !validateStep()) return

  loading.value = true

  try {
    const subjects = selectedSubjectIds.value.map(id => ({
      subjectId: id,
      deadline: new Date(deadlines.value[id]).toISOString()
    }))

    const request: CreateTeamRoundsRequest = {
      subjects
    }

    const response = await apiClient.post(`/teams/${team.value.id}/rounds/create-batch`, request)

    // Show success message
    const { successCount, failedSubjects } = response.data
    if (failedSubjects && failedSubjects.length > 0) {
      alert(`Created ${successCount} rounds successfully.\n\nFailed: ${failedSubjects.join(', ')}`)
    } else {
      alert(`Successfully created ${successCount} feedback rounds!`)
    }

    router.push('/rounds')
  } catch (error: any) {
    console.error('Failed to create rounds:', error)
    errors.value = [error.response?.data?.error || 'Failed to create rounds']
  } finally {
    loading.value = false
  }
}

function formatDate(dateStr: string): string {
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
      <h1 class="wizard__title">Create Team Round</h1>
      <p class="wizard__subtitle">Step {{ step }} of 4</p>
    </header>

    <!-- Progress Bar -->
    <div class="wizard__progress">
      <div class="wizard__progress-step" :class="{ 'wizard__progress-step--active': step >= 1, 'wizard__progress-step--complete': step > 1 }">1. Team</div>
      <div class="wizard__progress-step" :class="{ 'wizard__progress-step--active': step >= 2, 'wizard__progress-step--complete': step > 2 }">2. Subjects</div>
      <div class="wizard__progress-step" :class="{ 'wizard__progress-step--active': step >= 3, 'wizard__progress-step--complete': step > 3 }">3. Deadlines</div>
      <div class="wizard__progress-step" :class="{ 'wizard__progress-step--active': step >= 4 }">4. Confirm</div>
    </div>

    <!-- Errors -->
    <div v-if="errors.length" class="wizard__errors">
      <p v-for="error in errors" :key="error" class="wizard__error">{{ error }}</p>
    </div>

    <!-- Step 1: Team Info -->
    <div v-if="step === 1" class="wizard__step">
      <h2 class="wizard__step-title">Team Round for {{ team?.name }}</h2>
      <p class="wizard__step-hint">You're creating feedback rounds for your team members. Each member will receive feedback from all other team members (except themselves and you).</p>

      <div v-if="team" class="team-info">
        <div class="team-info__row">
          <span class="team-info__label">Team:</span>
          <span class="team-info__value">{{ team.name }}</span>
        </div>
        <div class="team-info__row">
          <span class="team-info__label">Total Members:</span>
          <span class="team-info__value">{{ team.members.length }}</span>
        </div>
        <div class="team-info__row">
          <span class="team-info__label">Available Subjects:</span>
          <span class="team-info__value">{{ availableSubjects.length }} (excluding you)</span>
        </div>
        <div class="team-info__row">
          <span class="team-info__label">Reviewers per Round:</span>
          <span class="team-info__value">{{ reviewerCount }} (auto-assigned)</span>
        </div>
      </div>
    </div>

    <!-- Step 2: Select Subjects -->
    <div v-if="step === 2" class="wizard__step">
      <h2 class="wizard__step-title">Who should receive feedback?</h2>
      <p class="wizard__step-hint">Select team members to create feedback rounds for. Each will receive feedback from all other team members.</p>

      <div class="user-grid">
        <div
          v-for="user in availableSubjects"
          :key="user.id"
          class="user-card"
          :class="{ 'user-card--selected': selectedSubjectIds.includes(user.id) }"
          @click="toggleSubject(user.id)"
        >
          <img v-if="user.photoUrl" :src="user.photoUrl" :alt="user.name" class="user-card__photo">
          <div v-else class="user-card__photo-placeholder">{{ user.name.charAt(0).toUpperCase() }}</div>
          <div class="user-card__info">
            <span class="user-card__name">{{ user.name }}</span>
            <span class="user-card__email">{{ user.email }}</span>
          </div>
          <span v-if="selectedSubjectIds.includes(user.id)" class="user-card__check">✓</span>
        </div>
      </div>

      <p class="wizard__selection-count">{{ selectedSubjects.length }} member(s) selected</p>
    </div>

    <!-- Step 3: Set Deadlines -->
    <div v-if="step === 3" class="wizard__step">
      <h2 class="wizard__step-title">Set Deadlines</h2>
      <p class="wizard__step-hint">Set individual deadlines for each person's round, or use the same deadline for all.</p>

      <div class="wizard__deadline-actions">
        <label class="wizard__deadline-label">Set same deadline for all:</label>
        <input
          type="datetime-local"
          :value="deadlines[selectedSubjectIds[0]]"
          @change="setAllDeadlines(($event.target as HTMLInputElement).value)"
          class="wizard__datetime-picker"
        >
      </div>

      <div class="wizard__deadline-list">
        <div v-for="subject in selectedSubjects" :key="subject.id" class="deadline-item">
          <div class="deadline-item__subject">
            <img v-if="subject.photoUrl" :src="subject.photoUrl" class="deadline-item__avatar">
            <div v-else class="deadline-item__avatar-placeholder">{{ subject.name.charAt(0) }}</div>
            <span class="deadline-item__name">{{ subject.name }}</span>
          </div>
          <input
            type="datetime-local"
            v-model="deadlines[subject.id]"
            class="deadline-item__input"
          >
        </div>
      </div>
    </div>

    <!-- Step 4: Review & Confirm -->
    <div v-if="step === 4" class="wizard__step">
      <h2 class="wizard__step-title">Review and Create</h2>
      <p class="wizard__step-hint">You're about to create {{ selectedSubjects.length }} feedback rounds.</p>

      <div class="wizard__review">
        <div class="review-section">
          <h4 class="review-section__title">Team</h4>
          <p class="review-section__text">{{ team?.name }}</p>
        </div>

        <div class="review-section">
          <h4 class="review-section__title">Rounds to Create ({{ selectedSubjects.length }})</h4>
          <div class="review-section__rounds">
            <div v-for="subject in selectedSubjects" :key="subject.id" class="round-preview">
              <div class="round-preview__subject">
                <img v-if="subject.photoUrl" :src="subject.photoUrl" class="round-preview__avatar">
                <div v-else class="round-preview__avatar-placeholder">{{ subject.name.charAt(0) }}</div>
                <strong class="round-preview__name">{{ subject.name }}</strong>
              </div>
              <div class="round-preview__details">
                <span class="round-preview__deadline">{{ formatDate(deadlines[subject.id]) }}</span>
                <span class="round-preview__reviewers">{{ reviewerCount }} reviewers</span>
              </div>
            </div>
          </div>
        </div>

        <div class="review-section">
          <h4 class="review-section__title">Auto-assigned Reviewers</h4>
          <p class="review-section__text">All team members except the subject and you will be assigned as reviewers for each round.</p>
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
        :disabled="step === 2 && selectedSubjectIds.length === 0"
      >
        Next
      </button>
      <button
        v-if="step === 4"
        class="btn btn--primary wizard__nav-next"
        @click="createRounds"
        :disabled="loading"
      >
        {{ loading ? 'Creating...' : `Create ${selectedSubjects.length} Rounds` }}
      </button>
    </div>
  </div>
</template>

<style scoped lang="scss">
.wizard {
  max-width: 900px;
  margin: 0 auto;
  padding: 1rem;

  @media (min-width: 768px) {
    padding: 2rem;
  }

  &__header {
    text-align: center;
    margin-bottom: 1.5rem;

    @media (min-width: 768px) {
      margin-bottom: 2rem;
    }
  }

  &__title {
    font-size: 1.5rem;
    color: var(--text-primary);
    margin: 0 0 0.5rem 0;

    @media (min-width: 768px) {
      font-size: 2rem;
    }
  }

  &__subtitle {
    color: var(--text-secondary);
    font-size: 0.85rem;
    margin: 0;

    @media (min-width: 768px) {
      font-size: 0.95rem;
    }
  }

  &__progress {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 2rem;
    position: relative;

    @media (min-width: 768px) {
      flex-direction: row;
      justify-content: space-between;
      margin-bottom: 3rem;
    }

    &::before {
      @media (min-width: 768px) {
        content: '';
        position: absolute;
        top: 15px;
        left: 10%;
        right: 10%;
        height: 2px;
        background: var(--border-color);
        z-index: 0;
      }
    }
  }

  &__progress-step {
    flex: 1;
    text-align: center;
    padding: 0.5rem;
    font-size: 0.8rem;
    color: var(--text-tertiary);
    position: relative;
    z-index: 1;

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }

    &::before {
      @media (min-width: 768px) {
        content: '';
        display: block;
        width: 30px;
        height: 30px;
        margin: 0 auto 0.5rem;
        border-radius: 50%;
        background: var(--bg-primary);
        border: 2px solid var(--border-color);
      }
    }

    &--active {
      color: var(--color-primary);
      font-weight: 500;

      &::before {
        border-color: var(--color-primary);
        background: var(--bg-primary);
      }
    }

    &--complete {
      &::before {
        background: var(--color-primary);
        border-color: var(--color-primary);
      }
    }
  }

  &__errors {
    background: rgba(244, 67, 54, 0.1);
    border: 1px solid var(--color-error);
    border-radius: 8px;
    padding: 1rem;
    margin-bottom: 1.5rem;
    color: var(--color-error);
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
    border-radius: 12px;
    padding: 1.25rem;
    border: 1px solid var(--border-color);
    margin-bottom: 1.5rem;

    @media (min-width: 768px) {
      padding: 2rem;
      margin-bottom: 2rem;
    }
  }

  &__step-title {
    font-size: 1.25rem;
    color: var(--text-primary);
    margin: 0 0 0.5rem 0;

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
    margin-top: 1rem;
    color: var(--text-secondary);
    font-weight: 500;
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__deadline-actions {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-bottom: 1.5rem;
    padding: 1rem;
    background: var(--bg-secondary);
    border-radius: 8px;

    @media (min-width: 768px) {
      flex-direction: row;
      align-items: center;
      gap: 1rem;
      margin-bottom: 2rem;
    }
  }

  &__deadline-label {
    font-weight: 600;
    color: var(--text-primary);
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__datetime-picker {
    padding: 0.5rem;
    font-size: 0.9rem;
    border: 2px solid var(--border-color);
    border-radius: 6px;
    min-width: 200px;
    background: var(--bg-primary);
    color: var(--text-primary);
    min-height: 44px;

    @media (min-width: 768px) {
      font-size: 1rem;
      min-width: 220px;
    }

    &:focus {
      outline: none;
      border-color: var(--color-primary);
    }
  }

  &__deadline-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  &__review {
    background: var(--bg-secondary);
    border-radius: 8px;
    padding: 1.25rem;

    @media (min-width: 768px) {
      padding: 1.5rem;
    }
  }

  &__nav {
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 0.75rem;

    @media (min-width: 768px) {
      flex-direction: row;
      gap: 1rem;
    }
  }

  &__nav-next {
    @media (min-width: 768px) {
      margin-left: auto;
    }
  }
}

.team-info {
  background: var(--bg-secondary);
  border-radius: 8px;
  padding: 1.25rem;

  @media (min-width: 768px) {
    padding: 1.5rem;
  }

  &__row {
    display: flex;
    justify-content: space-between;
    padding: 0.75rem 0;
    border-bottom: 1px solid var(--border-color);

    &:last-child {
      border-bottom: none;
    }
  }

  &__label {
    font-weight: 600;
    color: var(--text-secondary);
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__value {
    color: var(--text-primary);
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }
}

.user-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.75rem;

  @media (min-width: 768px) {
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 1rem;
  }
}

.user-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  border: 2px solid var(--border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  background: var(--bg-primary);
  min-height: 80px;

  @media (min-width: 768px) {
    gap: 1rem;
  }

  &:hover {
    border-color: var(--color-primary);
    background: rgba(102, 126, 234, 0.05);
  }

  &--selected {
    border-color: var(--color-primary);
    background: rgba(102, 126, 234, 0.05);
  }

  &__photo {
    width: 46px;
    height: 46px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;

    @media (min-width: 768px) {
      width: 50px;
      height: 50px;
    }
  }

  &__photo-placeholder {
    width: 46px;
    height: 46px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.2rem;
    font-weight: bold;
    flex-shrink: 0;

    @media (min-width: 768px) {
      width: 50px;
      height: 50px;
      font-size: 1.3rem;
    }
  }

  &__info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  &__name {
    font-weight: 600;
    color: var(--text-primary);
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__email {
    font-size: 0.8rem;
    color: var(--text-secondary);

    @media (min-width: 768px) {
      font-size: 0.85rem;
    }
  }

  &__check {
    color: var(--color-primary);
    font-size: 1.4rem;
    font-weight: bold;
    flex-shrink: 0;

    @media (min-width: 768px) {
      font-size: 1.5rem;
    }
  }
}

.deadline-item {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1rem;
  background: var(--bg-secondary);
  border-radius: 8px;

  @media (min-width: 768px) {
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
  }

  &__subject {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  &__avatar {
    width: 38px;
    height: 38px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;

    @media (min-width: 768px) {
      width: 40px;
      height: 40px;
    }
  }

  &__avatar-placeholder {
    width: 38px;
    height: 38px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1rem;
    font-weight: bold;
    flex-shrink: 0;

    @media (min-width: 768px) {
      width: 40px;
      height: 40px;
      font-size: 1.1rem;
    }
  }

  &__name {
    font-size: 0.9rem;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__input {
    padding: 0.5rem;
    font-size: 0.85rem;
    border: 2px solid var(--border-color);
    border-radius: 6px;
    min-width: 100%;
    background: var(--bg-primary);
    color: var(--text-primary);
    min-height: 44px;

    @media (min-width: 768px) {
      font-size: 0.9rem;
      min-width: 220px;
    }

    &:focus {
      outline: none;
      border-color: var(--color-primary);
    }
  }
}

.review-section {
  margin-bottom: 1.25rem;

  @media (min-width: 768px) {
    margin-bottom: 1.5rem;
  }

  &:last-child {
    margin-bottom: 0;
  }

  &__title {
    color: var(--text-secondary);
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin: 0 0 0.75rem 0;

    @media (min-width: 768px) {
      font-size: 0.85rem;
    }
  }

  &__text {
    font-size: 1rem;
    color: var(--text-primary);
    font-weight: 600;
    margin: 0;

    @media (min-width: 768px) {
      font-size: 1.1rem;
    }
  }

  &__rounds {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }
}

.round-preview {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0.75rem;
  background: var(--bg-primary);
  border-radius: 6px;

  @media (min-width: 768px) {
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
  }

  &__subject {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  &__avatar {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;

    @media (min-width: 768px) {
      width: 40px;
      height: 40px;
    }
  }

  &__avatar-placeholder {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1rem;
    font-weight: bold;
    flex-shrink: 0;

    @media (min-width: 768px) {
      width: 40px;
      height: 40px;
      font-size: 1.1rem;
    }
  }

  &__name {
    font-size: 0.9rem;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__details {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.25rem;
    font-size: 0.8rem;
    color: var(--text-secondary);

    @media (min-width: 768px) {
      align-items: flex-end;
      font-size: 0.85rem;
    }
  }

  &__deadline,
  &__reviewers {
    // Inherit styles from __details
  }
}
</style>
