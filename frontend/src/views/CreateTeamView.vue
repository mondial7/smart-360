<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import apiClient from '@/api/client'
import type { User } from '@/types/user'
import type { CreateTeamRequest } from '@/types/team'
import { PhCheck } from '@phosphor-icons/vue'

const router = useRouter()

const step = ref(1)
const loading = ref(false)
const users = ref<User[]>([])

// Form data
const teamName = ref('')
const teamAdminId = ref('')
const memberIds = ref<string[]>([])

// Validation errors
const errors = ref<string[]>([])

onMounted(async () => {
  await loadUsers()
})

async function loadUsers() {
  try {
    const response = await apiClient.get('/users')
    // Filter out users already in teams if needed
    users.value = response.data
  } catch (error) {
    console.error('Failed to load users:', error)
  }
}

const availableUsers = computed(() => {
  // Users not yet in a team
  return users.value.filter(u => !u.teamId)
})

const selectedAdmin = computed(() => {
  return users.value.find(u => u.id === teamAdminId.value)
})

const selectedMembers = computed(() => {
  return users.value.filter(u => memberIds.value.includes(u.id))
})

function toggleMember(userId: string) {
  const index = memberIds.value.indexOf(userId)
  if (index > -1) {
    memberIds.value.splice(index, 1)
  } else {
    memberIds.value.push(userId)
  }
}

function validateStep(): boolean {
  errors.value = []

  if (step.value === 1 && !teamName.value.trim()) {
    errors.value.push('Please enter a team name')
    return false
  }

  if (step.value === 2 && !teamAdminId.value) {
    errors.value.push('Please select a team admin')
    return false
  }

  if (step.value === 3 && memberIds.value.length === 0) {
    errors.value.push('Please select at least one team member')
    return false
  }

  if (step.value === 3 && !memberIds.value.includes(teamAdminId.value)) {
    errors.value.push('Team admin must be included in team members')
    return false
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

async function createTeam() {
  if (!validateStep()) return

  loading.value = true

  // Ensure team admin is in member list
  if (!memberIds.value.includes(teamAdminId.value)) {
    memberIds.value.push(teamAdminId.value)
  }

  const request: CreateTeamRequest = {
    name: teamName.value.trim(),
    teamAdminId: teamAdminId.value,
    memberIds: memberIds.value
  }

  try {
    await apiClient.post('/teams', request)
    router.push('/teams')
  } catch (error: any) {
    errors.value = [error.response?.data?.error || 'Failed to create team']
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="wizard">
    <header class="wizard__header">
      <h1 class="wizard__title">Create Team</h1>
      <p class="wizard__subtitle">Step {{ step }} of 4</p>
    </header>

    <!-- Progress Bar -->
    <div class="wizard__progress">
      <div class="wizard__progress-step" :class="{ 'wizard__progress-step--active': step >= 1, 'wizard__progress-step--complete': step > 1 }">1. Name</div>
      <div class="wizard__progress-step" :class="{ 'wizard__progress-step--active': step >= 2, 'wizard__progress-step--complete': step > 2 }">2. Admin</div>
      <div class="wizard__progress-step" :class="{ 'wizard__progress-step--active': step >= 3, 'wizard__progress-step--complete': step > 3 }">3. Members</div>
      <div class="wizard__progress-step" :class="{ 'wizard__progress-step--active': step >= 4 }">4. Confirm</div>
    </div>

    <!-- Errors -->
    <div v-if="errors.length" class="wizard__errors">
      <p v-for="error in errors" :key="error" class="wizard__error">{{ error }}</p>
    </div>

    <!-- Step 1: Team Name -->
    <div v-if="step === 1" class="wizard__step">
      <h2 class="wizard__step-title">What's the team name?</h2>
      <p class="wizard__step-hint">Choose a name that describes this team.</p>

      <input
        v-model="teamName"
        type="text"
        placeholder="e.g., Engineering Team, Product Team"
        class="wizard__team-input"
        @keyup.enter="nextStep"
        autofocus
      >
    </div>

    <!-- Step 2: Select Team Admin -->
    <div v-if="step === 2" class="wizard__step">
      <h2 class="wizard__step-title">Who will be the team admin?</h2>
      <p class="wizard__step-hint">Select one person to manage this team.</p>

      <div class="user-grid">
        <div
          v-for="user in availableUsers"
          :key="user.id"
          class="user-card"
          :class="{ 'user-card--selected': teamAdminId === user.id }"
          @click="teamAdminId = user.id"
        >
          <img v-if="user.photoUrl" :src="user.photoUrl" :alt="user.name" class="user-card__photo">
          <div v-else class="user-card__photo-placeholder">{{ user.name.charAt(0).toUpperCase() }}</div>
          <div class="user-card__info">
            <span class="user-card__name">{{ user.name }}</span>
            <span class="user-card__email">{{ user.email }}</span>
          </div>
          <PhCheck v-if="teamAdminId === user.id" class="user-card__check" :size="16" weight="bold" />
        </div>
      </div>
    </div>

    <!-- Step 3: Select Members -->
    <div v-if="step === 3" class="wizard__step">
      <h2 class="wizard__step-title">Who should be on this team?</h2>
      <p class="wizard__step-hint">Select all team members. The team admin will be included automatically.</p>

      <div class="user-grid">
        <div
          v-for="user in availableUsers"
          :key="user.id"
          class="user-card"
          :class="{ 'user-card--selected': memberIds.includes(user.id), 'user-card--disabled': user.id === teamAdminId }"
          @click="toggleMember(user.id)"
        >
          <img v-if="user.photoUrl" :src="user.photoUrl" :alt="user.name" class="user-card__photo">
          <div v-else class="user-card__photo-placeholder">{{ user.name.charAt(0).toUpperCase() }}</div>
          <div class="user-card__info">
            <span class="user-card__name">{{ user.name }}</span>
            <span class="user-card__email">{{ user.email }}</span>
            <span v-if="user.id === teamAdminId" class="user-card__badge">Team Admin</span>
          </div>
          <PhCheck v-if="memberIds.includes(user.id) || user.id === teamAdminId" class="user-card__check" :size="16" weight="bold" />
        </div>
      </div>

      <p class="wizard__selection-count">{{ memberIds.length + (memberIds.includes(teamAdminId) ? 0 : 1) }} member(s) selected</p>
    </div>

    <!-- Step 4: Review & Confirm -->
    <div v-if="step === 4" class="wizard__step">
      <h2 class="wizard__step-title">Review and Create</h2>
      <p class="wizard__step-hint">Please review the team details before creating.</p>

      <div class="wizard__review">
        <div class="review-section">
          <h4 class="review-section__title">Team Name</h4>
          <p class="review-section__text">{{ teamName }}</p>
        </div>

        <div class="review-section">
          <h4 class="review-section__title">Team Admin</h4>
          <div v-if="selectedAdmin" class="review-section__preview">
            <img v-if="selectedAdmin.photoUrl" :src="selectedAdmin.photoUrl" class="review-section__avatar">
            <div v-else class="review-section__avatar-placeholder">{{ selectedAdmin.name.charAt(0) }}</div>
            <div>
              <div class="review-section__name">{{ selectedAdmin.name }}</div>
              <div class="review-section__email">{{ selectedAdmin.email }}</div>
            </div>
          </div>
        </div>

        <div class="review-section">
          <h4 class="review-section__title">Team Members ({{ selectedMembers.length + (memberIds.includes(teamAdminId) ? 0 : 1) }})</h4>
          <div class="review-section__list">
            <div v-for="member in selectedMembers" :key="member.id" class="review-member">
              <img v-if="member.photoUrl" :src="member.photoUrl" class="review-member__avatar">
              <div v-else class="review-member__avatar-placeholder">{{ member.name.charAt(0) }}</div>
              <span class="review-member__name">{{ member.name }}</span>
            </div>
            <div v-if="selectedAdmin && !memberIds.includes(teamAdminId)" class="review-member">
              <img v-if="selectedAdmin.photoUrl" :src="selectedAdmin.photoUrl" class="review-member__avatar">
              <div v-else class="review-member__avatar-placeholder">{{ selectedAdmin.name.charAt(0) }}</div>
              <span class="review-member__name">{{ selectedAdmin.name }}</span>
            </div>
          </div>
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
        :disabled="(step === 1 && !teamName.trim()) || (step === 2 && !teamAdminId) || (step === 3 && memberIds.length === 0)"
      >
        Next
      </button>
      <button
        v-if="step === 4"
        class="btn btn--primary wizard__nav-next"
        @click="createTeam"
        :disabled="loading"
      >
        {{ loading ? 'Creating...' : 'Create Team' }}
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

  &__team-input {
    width: 100%;
    padding: 0.875rem;
    font-size: 1rem;
    border: 2px solid var(--border-color);
    border-radius: 8px;
    transition: border-color 0.2s;
    background: var(--bg-primary);
    color: var(--text-primary);
    min-height: 44px;

    @media (min-width: 768px) {
      padding: 1rem;
      font-size: 1.1rem;
    }

    &:focus {
      outline: none;
      border-color: var(--color-primary);
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

  &--disabled {
    opacity: 0.6;
    cursor: default;
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

  &__badge {
    display: inline-block;
    padding: 0.2rem 0.5rem;
    background: var(--color-primary);
    color: white;
    border-radius: 4px;
    font-size: 0.7rem;
    font-weight: 500;
    margin-top: 0.25rem;

    @media (min-width: 768px) {
      font-size: 0.75rem;
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

  &__preview {
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

  &__list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
}

.review-member {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem;
  background: var(--bg-primary);
  border-radius: 6px;

  &__avatar {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;

    @media (min-width: 768px) {
      width: 30px;
      height: 30px;
    }
  }

  &__avatar-placeholder {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.85rem;
    font-weight: bold;
    flex-shrink: 0;

    @media (min-width: 768px) {
      width: 30px;
      height: 30px;
      font-size: 0.9rem;
    }
  }

  &__name {
    font-size: 0.9rem;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }
}
</style>
