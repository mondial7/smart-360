<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import apiClient from '@/api/client'
import type { User } from '@/types/user'
import type { CreateTeamRequest } from '@/types/team'

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
  <div class="wizard-page">
    <header class="page-header">
      <h1>Create Team</h1>
      <p>Step {{ step }} of 4</p>
    </header>

    <!-- Progress Bar -->
    <div class="progress-bar">
      <div class="progress-step" :class="{ active: step >= 1, complete: step > 1 }">1. Name</div>
      <div class="progress-step" :class="{ active: step >= 2, complete: step > 2 }">2. Admin</div>
      <div class="progress-step" :class="{ active: step >= 3, complete: step > 3 }">3. Members</div>
      <div class="progress-step" :class="{ active: step >= 4 }">4. Confirm</div>
    </div>

    <!-- Errors -->
    <div v-if="errors.length" class="errors">
      <p v-for="error in errors" :key="error">{{ error }}</p>
    </div>

    <!-- Step 1: Team Name -->
    <div v-if="step === 1" class="wizard-step">
      <h2>What's the team name?</h2>
      <p class="hint">Choose a name that describes this team.</p>

      <input
        v-model="teamName"
        type="text"
        placeholder="e.g., Engineering Team, Product Team"
        class="team-name-input"
        @keyup.enter="nextStep"
        autofocus
      >
    </div>

    <!-- Step 2: Select Team Admin -->
    <div v-if="step === 2" class="wizard-step">
      <h2>Who will be the team admin?</h2>
      <p class="hint">Select one person to manage this team.</p>

      <div class="user-grid">
        <div
          v-for="user in availableUsers"
          :key="user.id"
          class="user-card"
          :class="{ selected: teamAdminId === user.id }"
          @click="teamAdminId = user.id"
        >
          <img v-if="user.photoUrl" :src="user.photoUrl" :alt="user.name" class="user-photo">
          <div v-else class="user-photo-placeholder">{{ user.name.charAt(0).toUpperCase() }}</div>
          <div class="user-info">
            <span class="user-name">{{ user.name }}</span>
            <span class="user-email">{{ user.email }}</span>
          </div>
          <span v-if="teamAdminId === user.id" class="checkmark">✓</span>
        </div>
      </div>
    </div>

    <!-- Step 3: Select Members -->
    <div v-if="step === 3" class="wizard-step">
      <h2>Who should be on this team?</h2>
      <p class="hint">Select all team members. The team admin will be included automatically.</p>

      <div class="user-grid">
        <div
          v-for="user in availableUsers"
          :key="user.id"
          class="user-card"
          :class="{ selected: memberIds.includes(user.id), disabled: user.id === teamAdminId }"
          @click="toggleMember(user.id)"
        >
          <img v-if="user.photoUrl" :src="user.photoUrl" :alt="user.name" class="user-photo">
          <div v-else class="user-photo-placeholder">{{ user.name.charAt(0).toUpperCase() }}</div>
          <div class="user-info">
            <span class="user-name">{{ user.name }}</span>
            <span class="user-email">{{ user.email }}</span>
            <span v-if="user.id === teamAdminId" class="badge">Team Admin</span>
          </div>
          <span v-if="memberIds.includes(user.id) || user.id === teamAdminId" class="checkmark">✓</span>
        </div>
      </div>

      <p class="selection-count">{{ memberIds.length + (memberIds.includes(teamAdminId) ? 0 : 1) }} member(s) selected</p>
    </div>

    <!-- Step 4: Review & Confirm -->
    <div v-if="step === 4" class="wizard-step">
      <h2>Review and Create</h2>
      <p class="hint">Please review the team details before creating.</p>

      <div class="review-card">
        <div class="review-section">
          <h4>Team Name</h4>
          <p>{{ teamName }}</p>
        </div>

        <div class="review-section">
          <h4>Team Admin</h4>
          <div v-if="selectedAdmin" class="user-preview">
            <img v-if="selectedAdmin.photoUrl" :src="selectedAdmin.photoUrl" class="mini-avatar">
            <div v-else class="mini-avatar-placeholder">{{ selectedAdmin.name.charAt(0) }}</div>
            <div>
              <div class="name">{{ selectedAdmin.name }}</div>
              <div class="email">{{ selectedAdmin.email }}</div>
            </div>
          </div>
        </div>

        <div class="review-section">
          <h4>Team Members ({{ selectedMembers.length + (memberIds.includes(teamAdminId) ? 0 : 1) }})</h4>
          <div class="members-list">
            <div v-for="member in selectedMembers" :key="member.id" class="member-item">
              <img v-if="member.photoUrl" :src="member.photoUrl" class="tiny-avatar">
              <div v-else class="tiny-avatar-placeholder">{{ member.name.charAt(0) }}</div>
              <span>{{ member.name }}</span>
            </div>
            <div v-if="selectedAdmin && !memberIds.includes(teamAdminId)" class="member-item">
              <img v-if="selectedAdmin.photoUrl" :src="selectedAdmin.photoUrl" class="tiny-avatar">
              <div v-else class="tiny-avatar-placeholder">{{ selectedAdmin.name.charAt(0) }}</div>
              <span>{{ selectedAdmin.name }}</span>
            </div>
          </div>
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
        :disabled="(step === 1 && !teamName.trim()) || (step === 2 && !teamAdminId) || (step === 3 && memberIds.length === 0)"
      >
        Next
      </button>
      <button
        v-if="step === 4"
        class="btn-primary"
        @click="createTeam"
        :disabled="loading"
      >
        {{ loading ? 'Creating...' : 'Create Team' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.wizard-page {
  max-width: 900px;
  margin: 0 auto;
  padding: 2rem;
}

.page-header {
  text-align: center;
  margin-bottom: 2rem;
}

.page-header h1 {
  font-size: 2rem;
  color: #333;
  margin-bottom: 0.5rem;
}

.page-header p {
  color: #666;
  font-size: 0.95rem;
}

.progress-bar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 3rem;
  position: relative;
}

.progress-bar::before {
  content: '';
  position: absolute;
  top: 15px;
  left: 10%;
  right: 10%;
  height: 2px;
  background: #e0e0e0;
  z-index: 0;
}

.progress-step {
  flex: 1;
  text-align: center;
  padding: 0.5rem;
  font-size: 0.9rem;
  color: #999;
  position: relative;
  z-index: 1;
}

.progress-step::before {
  content: '';
  display: block;
  width: 30px;
  height: 30px;
  margin: 0 auto 0.5rem;
  border-radius: 50%;
  background: white;
  border: 2px solid #e0e0e0;
}

.progress-step.active {
  color: #667eea;
  font-weight: 500;
}

.progress-step.active::before {
  border-color: #667eea;
  background: white;
}

.progress-step.complete::before {
  background: #667eea;
  border-color: #667eea;
}

.errors {
  background: #fee;
  border: 1px solid #fcc;
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 1.5rem;
  color: #c33;
}

.wizard-step {
  background: white;
  border-radius: 12px;
  padding: 2rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  margin-bottom: 2rem;
}

.wizard-step h2 {
  font-size: 1.5rem;
  color: #333;
  margin-bottom: 0.5rem;
}

.hint {
  color: #666;
  margin-bottom: 1.5rem;
}

.team-name-input {
  width: 100%;
  padding: 1rem;
  font-size: 1.1rem;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  transition: border-color 0.2s;
}

.team-name-input:focus {
  outline: none;
  border-color: #667eea;
}

.user-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
}

.user-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.user-card:hover {
  border-color: #667eea;
  background: #f5f7ff;
}

.user-card.selected {
  border-color: #667eea;
  background: #f5f7ff;
}

.user-card.disabled {
  opacity: 0.6;
  cursor: default;
  border-color: #667eea;
  background: #f5f7ff;
}

.user-photo {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  object-fit: cover;
}

.user-photo-placeholder {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.3rem;
  font-weight: bold;
}

.user-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.user-name {
  font-weight: 600;
  color: #333;
}

.user-email {
  font-size: 0.85rem;
  color: #666;
}

.badge {
  display: inline-block;
  padding: 0.2rem 0.5rem;
  background: #667eea;
  color: white;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
  margin-top: 0.25rem;
}

.checkmark {
  color: #667eea;
  font-size: 1.5rem;
  font-weight: bold;
}

.selection-count {
  text-align: center;
  margin-top: 1rem;
  color: #666;
  font-weight: 500;
}

.review-card {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 1.5rem;
}

.review-section {
  margin-bottom: 1.5rem;
}

.review-section:last-child {
  margin-bottom: 0;
}

.review-section h4 {
  color: #666;
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 0.75rem;
}

.review-section > p {
  font-size: 1.1rem;
  color: #333;
  font-weight: 600;
}

.user-preview {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.mini-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
}

.mini-avatar-placeholder {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.1rem;
  font-weight: bold;
}

.name {
  font-weight: 600;
  color: #333;
}

.email {
  font-size: 0.85rem;
  color: #666;
}

.members-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.member-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem;
  background: white;
  border-radius: 6px;
}

.tiny-avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  object-fit: cover;
}

.tiny-avatar-placeholder {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.9rem;
  font-weight: bold;
}

.wizard-nav {
  display: flex;
  justify-content: center;
  gap: 1rem;
}

.btn-primary,
.btn-secondary {
  padding: 0.75rem 2rem;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #667eea;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #5568d3;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: white;
  color: #667eea;
  border: 2px solid #667eea;
}

.btn-secondary:hover:not(:disabled) {
  background: #f5f7ff;
}
</style>
