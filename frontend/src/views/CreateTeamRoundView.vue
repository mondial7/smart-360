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
  <div class="wizard-page">
    <header class="page-header">
      <h1>Create Team Round</h1>
      <p>Step {{ step }} of 4</p>
    </header>

    <!-- Progress Bar -->
    <div class="progress-bar">
      <div class="progress-step" :class="{ active: step >= 1, complete: step > 1 }">1. Team</div>
      <div class="progress-step" :class="{ active: step >= 2, complete: step > 2 }">2. Subjects</div>
      <div class="progress-step" :class="{ active: step >= 3, complete: step > 3 }">3. Deadlines</div>
      <div class="progress-step" :class="{ active: step >= 4 }">4. Confirm</div>
    </div>

    <!-- Errors -->
    <div v-if="errors.length" class="errors">
      <p v-for="error in errors" :key="error">{{ error }}</p>
    </div>

    <!-- Step 1: Team Info -->
    <div v-if="step === 1" class="wizard-step">
      <h2>Team Round for {{ team?.name }}</h2>
      <p class="hint">You're creating feedback rounds for your team members. Each member will receive feedback from all other team members (except themselves and you).</p>

      <div v-if="team" class="team-info-card">
        <div class="info-row">
          <span class="label">Team:</span>
          <span class="value">{{ team.name }}</span>
        </div>
        <div class="info-row">
          <span class="label">Total Members:</span>
          <span class="value">{{ team.members.length }}</span>
        </div>
        <div class="info-row">
          <span class="label">Available Subjects:</span>
          <span class="value">{{ availableSubjects.length }} (excluding you)</span>
        </div>
        <div class="info-row">
          <span class="label">Reviewers per Round:</span>
          <span class="value">{{ reviewerCount }} (auto-assigned)</span>
        </div>
      </div>
    </div>

    <!-- Step 2: Select Subjects -->
    <div v-if="step === 2" class="wizard-step">
      <h2>Who should receive feedback?</h2>
      <p class="hint">Select team members to create feedback rounds for. Each will receive feedback from all other team members.</p>

      <div class="user-grid">
        <div
          v-for="user in availableSubjects"
          :key="user.id"
          class="user-card"
          :class="{ selected: selectedSubjectIds.includes(user.id) }"
          @click="toggleSubject(user.id)"
        >
          <img v-if="user.photoUrl" :src="user.photoUrl" :alt="user.name" class="user-photo">
          <div v-else class="user-photo-placeholder">{{ user.name.charAt(0).toUpperCase() }}</div>
          <div class="user-info">
            <span class="user-name">{{ user.name }}</span>
            <span class="user-email">{{ user.email }}</span>
          </div>
          <span v-if="selectedSubjectIds.includes(user.id)" class="checkmark">✓</span>
        </div>
      </div>

      <p class="selection-count">{{ selectedSubjects.length }} member(s) selected</p>
    </div>

    <!-- Step 3: Set Deadlines -->
    <div v-if="step === 3" class="wizard-step">
      <h2>Set Deadlines</h2>
      <p class="hint">Set individual deadlines for each person's round, or use the same deadline for all.</p>

      <div class="deadline-actions">
        <label>Set same deadline for all:</label>
        <input
          type="datetime-local"
          :value="deadlines[selectedSubjectIds[0]]"
          @change="setAllDeadlines(($event.target as HTMLInputElement).value)"
          class="datetime-picker"
        >
      </div>

      <div class="deadline-list">
        <div v-for="subject in selectedSubjects" :key="subject.id" class="deadline-item">
          <div class="subject-info">
            <img v-if="subject.photoUrl" :src="subject.photoUrl" class="mini-avatar">
            <div v-else class="mini-avatar-placeholder">{{ subject.name.charAt(0) }}</div>
            <span>{{ subject.name }}</span>
          </div>
          <input
            type="datetime-local"
            v-model="deadlines[subject.id]"
            class="datetime-picker-small"
          >
        </div>
      </div>
    </div>

    <!-- Step 4: Review & Confirm -->
    <div v-if="step === 4" class="wizard-step">
      <h2>Review and Create</h2>
      <p class="hint">You're about to create {{ selectedSubjects.length }} feedback rounds.</p>

      <div class="review-card">
        <div class="review-section">
          <h4>Team</h4>
          <p>{{ team?.name }}</p>
        </div>

        <div class="review-section">
          <h4>Rounds to Create ({{ selectedSubjects.length }})</h4>
          <div class="rounds-preview">
            <div v-for="subject in selectedSubjects" :key="subject.id" class="round-preview-item">
              <div class="subject">
                <img v-if="subject.photoUrl" :src="subject.photoUrl" class="mini-avatar">
                <div v-else class="mini-avatar-placeholder">{{ subject.name.charAt(0) }}</div>
                <strong>{{ subject.name }}</strong>
              </div>
              <div class="details">
                <span class="deadline">{{ formatDate(deadlines[subject.id]) }}</span>
                <span class="reviewers">{{ reviewerCount }} reviewers</span>
              </div>
            </div>
          </div>
        </div>

        <div class="review-section">
          <h4>Auto-assigned Reviewers</h4>
          <p>All team members except the subject and you will be assigned as reviewers for each round.</p>
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
        :disabled="step === 2 && selectedSubjectIds.length === 0"
      >
        Next
      </button>
      <button
        v-if="step === 4"
        class="btn-primary"
        @click="createRounds"
        :disabled="loading"
      >
        {{ loading ? 'Creating...' : `Create ${selectedSubjects.length} Rounds` }}
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

.team-info-card {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 1.5rem;
}

.info-row {
  display: flex;
  justify-content: space-between;
  padding: 0.75rem 0;
  border-bottom: 1px solid #e0e0e0;
}

.info-row:last-child {
  border-bottom: none;
}

.info-row .label {
  font-weight: 600;
  color: #666;
}

.info-row .value {
  color: #333;
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

.deadline-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 2rem;
  padding: 1rem;
  background: #f8f9fa;
  border-radius: 8px;
}

.deadline-actions label {
  font-weight: 600;
  color: #333;
}

.datetime-picker {
  padding: 0.5rem;
  font-size: 1rem;
  border: 2px solid #e0e0e0;
  border-radius: 6px;
  min-width: 220px;
}

.datetime-picker:focus {
  outline: none;
  border-color: #667eea;
}

.deadline-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.deadline-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: #f8f9fa;
  border-radius: 8px;
}

.subject-info {
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

.datetime-picker-small {
  padding: 0.5rem;
  font-size: 0.9rem;
  border: 2px solid #e0e0e0;
  border-radius: 6px;
  min-width: 220px;
}

.datetime-picker-small:focus {
  outline: none;
  border-color: #667eea;
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

.rounds-preview {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.round-preview-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem;
  background: white;
  border-radius: 6px;
}

.round-preview-item .subject {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.round-preview-item .details {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.25rem;
  font-size: 0.85rem;
  color: #666;
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
