<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'
import type { FeedbackRound } from '@/types/round'
import {
  PhNotePencil,
  PhArrowLeft,
  PhTrophy,
  PhTrendUp,
  PhTarget,
  PhFileText,
  PhClipboardText,
  PhCheck,
  PhArrowUp,
  PhArrowRight
} from '@phosphor-icons/vue'

const route = useRoute()
const auth = useAuthStore()
const roundId = route.params.id as string  // Changed from parseInt to string

const round = ref<FeedbackRound | null>(null)
const submissions = ref<any[]>([])
const consolidation = ref<any>(null)
const loading = ref(true)
const activeTab = ref('submissions')
const showEditModal = ref(false)
const editForm = ref({
  subjectId: '',
  deadline: '',
  status: '',
  reviewerIds: [] as string[]
})
const users = ref<any[]>([])
const editLoading = ref(false)
const selectedReviewer = ref('')

// Computed properties
const availableUsers = computed(() => {
  return users.value.filter(user => 
    user.id !== auth.user?.id && 
    !editForm.value.reviewerIds.includes(user.id) &&
    user.id !== round.value?.subjectId
  )
})

onMounted(async () => {
  await loadData()
  
  // Check if URL hash indicates consolidation tab should be active
  if (window.location.hash === '#consolidation') {
    activeTab.value = 'consolidation'
  }
})

async function loadData() {
  try {
    // Load round details
    const roundRes = await apiClient.get(`/rounds/${roundId}`)
    round.value = roundRes.data
    
    // Load users for edit form
    const usersRes = await apiClient.get('/users')
    users.value = usersRes.data
    
    // Load submissions
    const subRes = await apiClient.get(`/submissions/round/${roundId}`)
    submissions.value = subRes.data
    
    // Try to load consolidation
    try {
      const consRes = await apiClient.get(`/consolidations/${roundId}`)
      consolidation.value = consRes.data
      // Parse consolidation fields
      if (consolidation.value.strengths && typeof consolidation.value.strengths === 'string') {
        consolidation.value.strengths = JSON.parse(consolidation.value.strengths)
      }
      if (consolidation.value.areasForImprovement && typeof consolidation.value.areasForImprovement === 'string') {
        consolidation.value.areasForImprovement = JSON.parse(consolidation.value.areasForImprovement)
      }
      if (consolidation.value.actionableInsights && typeof consolidation.value.actionableInsights === 'string') {
        consolidation.value.actionableInsights = JSON.parse(consolidation.value.actionableInsights)
      }
      if (consolidation.value.questionSummaries && typeof consolidation.value.questionSummaries === 'string') {
        consolidation.value.questionSummaries = JSON.parse(consolidation.value.questionSummaries)
      }
    } catch (error) {
      console.error('Error loading consolidation:', error)
      // No consolidation yet
    }
  } catch (error) {
    console.error('Failed to load data:', error)
  } finally {
    loading.value = false
  }
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return 'Not set'
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function getQuestionText(key: string): string {
  const questions: Record<string, string> = {
    a: 'What are this person\'s key strengths?',
    b: 'What areas could this person improve?',
    c: 'What specific behaviors or actions have you observed that stood out?',
    d: 'What advice would you give to help this person grow?'
  }
  return questions[key] || key
}

function parseResponses(responsesStr: string): Record<string, string> {
  try {
    return JSON.parse(responsesStr)
  } catch {
    return {}
  }
}

function openEditModal() {
  if (!round.value) return
  
  // Filter out subject from existing reviewers
  const existingReviewers = round.value.reviewers?.map(r => r.reviewerId).filter(id => id !== round.value?.subjectId) || []
  
  editForm.value = {
    subjectId: round.value.subjectId || '',
    deadline: round.value.deadline ? new Date(round.value.deadline).toISOString().slice(0, 16) : '',
    status: round.value.status || '',
    reviewerIds: existingReviewers
  }
  showEditModal.value = true
}

async function updateRound() {
  if (!round.value) return
  
  editLoading.value = true
  try {
    const updateData: any = {}
    
    if (editForm.value.subjectId && editForm.value.subjectId !== round.value.subjectId) {
      updateData.subjectId = editForm.value.subjectId
    }
    
    if (editForm.value.deadline) {
      updateData.deadline = new Date(editForm.value.deadline).toISOString()
    }
    
    if (editForm.value.status && editForm.value.status !== round.value.status) {
      updateData.status = editForm.value.status
    }
    
    // Handle reviewer updates
    const currentReviewerIds = round.value.reviewers?.map(r => r.reviewerId) || []
    const newReviewerIds = editForm.value.reviewerIds
    
    // Find reviewers to add and remove
    const toAdd = newReviewerIds.filter(id => !currentReviewerIds.includes(id))
    const toRemove = currentReviewerIds.filter(id => !newReviewerIds.includes(id))
    
    if (Object.keys(updateData).length === 0 && toAdd.length === 0 && toRemove.length === 0) {
      showEditModal.value = false
      return
    }
    
    // Update basic round info if needed
    if (Object.keys(updateData).length > 0) {
      await apiClient.put(`/rounds/${roundId}`, updateData)
    }
    
    // Add new reviewers
    for (const reviewerId of toAdd) {
      try {
        await apiClient.post(`/rounds/${roundId}/reviewers`, { reviewerIds: [reviewerId] })
      } catch (error) {
        console.error('Failed to add reviewer:', error)
      }
    }
    
    // Remove reviewers (we'd need to implement this endpoint)
    for (const reviewerId of toRemove) {
      try {
        await apiClient.delete(`/rounds/${roundId}/reviewers/${reviewerId}`)
      } catch (error) {
        console.error('Failed to remove reviewer:', error)
        // For now, we'll ignore if the endpoint doesn't exist
      }
    }
    
    // Reload data to show updated round
    await loadData()
    showEditModal.value = false
  } catch (error: any) {
    console.error('Failed to update round:', error)
    alert(error.response?.data?.error || 'Failed to update round')
  } finally {
    editLoading.value = false
  }
}

function addReviewer(reviewerId: string) {
  if (!editForm.value.reviewerIds.includes(reviewerId)) {
    editForm.value.reviewerIds.push(reviewerId)
  }
}

function removeReviewer(reviewerId: string) {
  const index = editForm.value.reviewerIds.indexOf(reviewerId)
  if (index > -1) {
    editForm.value.reviewerIds.splice(index, 1)
  }
  // Reset the select dropdown
  selectedReviewer.value = ''
}

function getUserById(userId: string) {
  return users.value.find(user => user.id === userId)
}
</script>

<template>
  <div class="round-details">
    <div v-if="loading" class="round-details__loading">Loading...</div>

    <template v-else-if="round">
      <header class="round-details__header">
        <router-link to="/rounds" class="round-details__back">
          <PhArrowLeft :size="14" weight="bold" />
          <span>Back to Rounds</span>
        </router-link>
        <div class="round-details__header-content">
          <h1 class="round-details__title">Round Details</h1>
          <button
            v-if="auth.isAdmin || round.createdById === auth.user?.id"
            @click="openEditModal"
            class="btn btn--primary btn--with-icon round-details__edit-btn"
          >
            <PhNotePencil :size="16" weight="regular" />
            <span>Edit Round</span>
          </button>
        </div>
        <div class="round-details__meta">
          <div class="round-details__meta-item">
            <span class="round-details__meta-label">Subject:</span>
            <div class="round-details__subject">
              <img v-if="round.subject?.photoUrl" :src="round.subject.photoUrl" class="round-details__avatar">
              <div v-else class="round-details__avatar-placeholder">{{ round.subject?.name.charAt(0) }}</div>
              <span>{{ round.subject?.name }}</span>
            </div>
          </div>
          <div class="round-details__meta-item">
            <span class="round-details__meta-label">Status:</span>
            <span :class="['badge', `badge--${round.status}`]">{{ round.status }}</span>
          </div>
          <div class="round-details__meta-item">
            <span class="round-details__meta-label">Deadline:</span>
            <span :class="{ 'round-details__overdue': round.status === 'active' && round.deadline && new Date(round.deadline) < new Date() }">
              {{ formatDate(round.deadline) }}
            </span>
          </div>
          <div class="round-details__meta-item">
            <span class="round-details__meta-label">Reviewers:</span>
            <span>{{ round.reviewers?.length || 0 }} assigned</span>
          </div>
          <div class="round-details__meta-item">
            <span class="round-details__meta-label">Submissions:</span>
            <span>{{ submissions.length }} received</span>
          </div>
        </div>
      </header>

      <!-- Tab Navigation -->
      <div class="round-details__tabs">
        <button
          :class="['round-details__tab', { 'round-details__tab--active': activeTab === 'submissions' }]"
          @click="activeTab = 'submissions'"
        >
          Raw Submissions ({{ submissions.length }})
        </button>
        <button
          v-if="consolidation"
          :class="['round-details__tab', { 'round-details__tab--active': activeTab === 'consolidation' }]"
          @click="activeTab = 'consolidation'"
        >
          Consolidated Feedback
        </button>
      </div>

      <!-- Raw Submissions Tab -->
      <div v-if="activeTab === 'submissions'" class="round-details__tab-content">
        <div v-if="submissions.length === 0" class="round-details__empty">
          <p class="round-details__empty-text">No submissions received yet.</p>
          <p v-if="round.status === 'active'" class="round-details__empty-hint">
            {{ round.reviewers?.length || 0 }} reviewers assigned, deadline is {{ formatDate(round.deadline) }}
          </p>
        </div>

        <div v-else class="round-details__submissions">
          <div v-for="submission in submissions" :key="submission.id" class="submission">
            <div class="submission__header">
              <span class="submission__reviewer-id">Reviewer #{{ submission.id }}</span>
              <span class="submission__submitted-at">Submitted {{ formatDate(submission.submittedAt) }}</span>
            </div>

            <div class="submission__responses">
              <div v-for="(response, key) in parseResponses(submission.responses)" :key="key" class="submission__response">
                <h4 class="submission__question">{{ getQuestionText(key) }}</h4>
                <p class="submission__answer">{{ response }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Consolidated Feedback Tab -->
      <div v-if="activeTab === 'consolidation' && consolidation" class="round-details__tab-content">
        <div class="round-details__consolidation-header">
          <div class="round-details__consolidation-meta">
            <span>Generated {{ formatDate(consolidation.createdAt) }}</span>
            <span v-if="consolidation.sharedAt" class="round-details__shared-badge">
              ✓ Shared {{ formatDate(consolidation.sharedAt) }}
            </span>
            <span v-else class="round-details__not-shared">Not yet shared</span>
          </div>
        </div>

        <div class="round-details__consolidation-body">
          <!-- Executive Summary -->
          <section class="consolidation-section">
            <h2 class="consolidation-section__title">
              <PhFileText :size="20" weight="duotone" />
              <span>Executive Summary</span>
            </h2>
            <p class="consolidation-section__summary">{{ consolidation.executiveSummary }}</p>
          </section>

          <!-- Strengths -->
          <section class="consolidation-section">
            <h2 class="consolidation-section__title">
              <PhTrophy :size="20" weight="duotone" />
              <span>Key Strengths</span>
            </h2>
            <ul class="consolidation-list consolidation-list--positive">
              <li v-for="(strength, i) in consolidation.strengths" :key="i">
                <PhCheck class="consolidation-list__icon" :size="18" weight="bold" />
                <span>{{ strength }}</span>
              </li>
            </ul>
          </section>

          <!-- Areas for Improvement -->
          <section class="consolidation-section">
            <h2 class="consolidation-section__title">
              <PhTrendUp :size="20" weight="duotone" />
              <span>Areas for Improvement</span>
            </h2>
            <ul class="consolidation-list consolidation-list--improvement">
              <li v-for="(area, i) in consolidation.areasForImprovement" :key="i">
                <PhArrowUp class="consolidation-list__icon" :size="18" weight="bold" />
                <span>{{ area }}</span>
              </li>
            </ul>
          </section>

          <!-- Actionable Insights -->
          <section class="consolidation-section">
            <h2 class="consolidation-section__title">
              <PhTarget :size="20" weight="duotone" />
              <span>Actionable Insights</span>
            </h2>
            <ul class="consolidation-list consolidation-list--action">
              <li v-for="(insight, i) in consolidation.actionableInsights" :key="i">
                <PhArrowRight class="consolidation-list__icon" :size="18" weight="bold" />
                <span>{{ insight }}</span>
              </li>
            </ul>
          </section>

          <!-- Question Summaries -->
          <section class="consolidation-section">
            <h2 class="consolidation-section__title">
              <PhClipboardText :size="20" weight="duotone" />
              <span>Detailed Question Analysis</span>
            </h2>
            <div class="questions">
              <div class="question-card">
                <h4 class="question-card__title">1. Key Strengths</h4>
                <p class="question-card__text">{{ consolidation.questionSummaries?.a }}</p>
              </div>
              <div class="question-card">
                <h4 class="question-card__title">2. Areas to Improve</h4>
                <p class="question-card__text">{{ consolidation.questionSummaries?.b }}</p>
              </div>
              <div class="question-card">
                <h4 class="question-card__title">3. Observed Behaviors</h4>
                <p class="question-card__text">{{ consolidation.questionSummaries?.c }}</p>
              </div>
              <div class="question-card">
                <h4 class="question-card__title">4. Growth Advice</h4>
                <p class="question-card__text">{{ consolidation.questionSummaries?.d }}</p>
              </div>
            </div>
          </section>

          <!-- Admin Notes -->
          <section v-if="consolidation.adminNotes" class="consolidation-section">
            <h2 class="consolidation-section__title">
              <PhNotePencil :size="20" weight="duotone" />
              <span>Admin Notes</span>
            </h2>
            <p class="consolidation-section__admin-notes">{{ consolidation.adminNotes }}</p>
          </section>
        </div>
      </div>
    </template>

    <div v-else class="round-details__error">
      <p class="round-details__error-text">Round not found.</p>
      <router-link to="/rounds" class="btn btn--primary">Back to Rounds</router-link>
    </div>

    <!-- Edit Modal -->
    <div v-if="showEditModal" class="modal-overlay" @click.self="showEditModal = false">
      <div class="modal">
        <div class="modal__header">
          <h2 class="modal__title">Edit Round</h2>
          <button @click="showEditModal = false" class="modal__close">×</button>
        </div>

        <form @submit.prevent="updateRound" class="modal__form">
          <div class="modal__field">
            <label for="subjectId" class="modal__label">Subject</label>
            <select id="subjectId" v-model="editForm.subjectId" required class="modal__select">
              <option value="">Select a subject</option>
              <option
                v-for="user in users.filter(u => u.id !== auth.user?.id)"
                :key="user.id"
                :value="user.id"
              >
                {{ user.name }}
              </option>
            </select>
          </div>

          <div class="modal__field">
            <label for="deadline" class="modal__label">Deadline</label>
            <input
              id="deadline"
              type="datetime-local"
              v-model="editForm.deadline"
              :min="new Date().toISOString().slice(0, 16)"
              class="modal__input"
            >
          </div>

          <div class="modal__field">
            <label for="status" class="modal__label">Status</label>
            <select id="status" v-model="editForm.status" required class="modal__select">
              <option value="draft">Draft</option>
              <option value="active">Active</option>
              <option value="closed">Closed</option>
              <option value="shared">Shared</option>
            </select>
          </div>


          <div class="modal__field">
            <label class="modal__label">Reviewers</label>
            <div class="reviewer-mgmt">
              <!-- Show warning if subject was previously assigned as reviewer -->
              <div v-if="round.value?.reviewers?.some(r => r.reviewerId === round.value?.subjectId)" class="reviewer-mgmt__warning">
                ⚠️ The subject ({{ getUserById(round.value?.subjectId || '')?.name }}) was previously assigned as a reviewer and has been removed.
              </div>

              <!-- Current reviewers -->
              <div v-if="editForm.reviewerIds.length > 0" class="reviewer-mgmt__current">
                <div class="reviewer-mgmt__list">
                  <div
                    v-for="reviewerId in editForm.reviewerIds"
                    :key="reviewerId"
                    class="reviewer-tag"
                  >
                    <span class="reviewer-tag__name">{{ getUserById(reviewerId)?.name || reviewerId }}</span>
                    <button
                      type="button"
                      @click="removeReviewer(reviewerId)"
                      class="reviewer-tag__remove"
                    >
                      ×
                    </button>
                  </div>
                </div>
              </div>

              <!-- Add new reviewers -->
              <div class="reviewer-mgmt__add">
                <select
                  v-model="selectedReviewer"
                  @change="addReviewer(selectedReviewer)"
                  class="reviewer-mgmt__select"
                >
                  <option value="">Add reviewer...</option>
                  <option
                    v-for="user in availableUsers"
                    :key="user.id"
                    :value="user.id"
                  >
                    {{ user.name }}
                  </option>
                </select>
              </div>
            </div>
          </div>

          <div class="modal__actions">
            <button type="button" @click="showEditModal = false" class="btn btn--secondary">
              Cancel
            </button>
            <button type="submit" :disabled="editLoading" class="btn btn--primary">
              {{ editLoading ? 'Saving...' : 'Save Changes' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.round-details {
  padding: 1rem;
  max-width: 1200px;
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
    margin-bottom: 1.5rem;

    @media (min-width: 768px) {
      margin-bottom: 2rem;
    }
  }

  &__back {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
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

  &__header-content {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-bottom: 1rem;

    @media (min-width: 768px) {
      flex-direction: row;
      justify-content: space-between;
      align-items: center;
    }
  }

  &__title {
    font-size: 1.5rem;
    margin: 0;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 2rem;
    }
  }

  &__edit-btn {
    font-size: 0.8rem;
    padding: 0.5rem 1rem;
    align-self: flex-start;

    @media (min-width: 768px) {
      font-size: 0.9rem;
      align-self: auto;
    }
  }

  &__meta {
    display: grid;
    grid-template-columns: 1fr;
    gap: 1rem;
    background: var(--bg-primary);
    padding: 1.25rem;
    border-radius: 12px;
    border: 1px solid var(--border-color);

    @media (min-width: 768px) {
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      padding: 1.5rem;
    }
  }

  &__meta-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  &__meta-label {
    font-size: 0.7rem;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.5px;

    @media (min-width: 768px) {
      font-size: 0.75rem;
    }
  }

  &__subject {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  &__avatar,
  &__avatar-placeholder {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  &__avatar {
    object-fit: cover;
  }

  &__avatar-placeholder {
    background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.8rem;
    font-weight: 600;
  }

  &__overdue {
    color: var(--color-error);
    font-weight: 500;
  }

  &__tabs {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1.5rem;
    border-bottom: 2px solid var(--border-color);
    overflow-x: auto;

    @media (min-width: 768px) {
      gap: 1rem;
      margin-bottom: 2rem;
    }
  }

  &__tab {
    padding: 0.75rem 1rem;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    cursor: pointer;
    font-size: 0.85rem;
    font-weight: 500;
    color: var(--text-secondary);
    transition: all 0.2s;
    white-space: nowrap;
    min-height: 44px;

    @media (min-width: 768px) {
      padding: 1rem 1.5rem;
      font-size: 1rem;
    }

    &--active {
      color: var(--color-primary);
      border-bottom-color: var(--color-primary);
    }

    &:hover {
      color: var(--color-primary);
    }
  }

  &__tab-content {
    background: var(--bg-primary);
    border-radius: 12px;
    border: 1px solid var(--border-color);
    padding: 1.25rem;

    @media (min-width: 768px) {
      padding: 2rem;
    }
  }

  &__empty {
    text-align: center;
    padding: 2rem 1rem;
    color: var(--text-secondary);

    @media (min-width: 768px) {
      padding: 3rem;
    }
  }

  &__empty-text {
    margin: 0 0 0.5rem 0;
  }

  &__empty-hint {
    font-size: 0.85rem;
    color: var(--text-tertiary);
    margin: 0.5rem 0 0 0;

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }
  }

  &__submissions {
    display: grid;
    gap: 1.25rem;

    @media (min-width: 768px) {
      gap: 1.5rem;
    }
  }

  &__consolidation-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
    padding-bottom: 1rem;
    border-bottom: 1px solid var(--border-color);

    @media (min-width: 768px) {
      margin-bottom: 2rem;
    }
  }

  &__consolidation-meta {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    align-items: flex-start;
    color: var(--text-secondary);
    font-size: 0.85rem;

    @media (min-width: 768px) {
      flex-direction: row;
      gap: 1rem;
      align-items: center;
      font-size: 0.9rem;
    }
  }

  &__shared-badge {
    color: var(--color-success);
    font-weight: 500;
  }

  &__not-shared {
    color: var(--color-warning);
    font-weight: 500;
  }

  &__consolidation-body {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;

    @media (min-width: 768px) {
      gap: 2rem;
    }
  }
}

.submission {
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 1.25rem;

  @media (min-width: 768px) {
    padding: 1.5rem;
  }

  &__header {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 1.25rem;
    padding-bottom: 1rem;
    border-bottom: 1px solid var(--border-color);

    @media (min-width: 768px) {
      flex-direction: row;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 1.5rem;
    }
  }

  &__reviewer-id {
    font-weight: 500;
    color: var(--text-primary);
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__submitted-at {
    font-size: 0.8rem;
    color: var(--text-tertiary);

    @media (min-width: 768px) {
      font-size: 0.85rem;
    }
  }

  &__responses {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;

    @media (min-width: 768px) {
      gap: 1.5rem;
    }
  }

  &__response {
    // No specific styles needed, children styled below
  }

  &__question {
    font-size: 0.85rem;
    color: var(--color-primary);
    margin: 0 0 0.5rem 0;
    font-weight: 500;

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }
  }

  &__answer {
    color: var(--text-primary);
    line-height: 1.5;
    white-space: pre-wrap;
    margin: 0;
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }
}

.consolidation-section {
  padding-bottom: 1.5rem;
  border-bottom: 1px solid var(--border-color);

  @media (min-width: 768px) {
    padding-bottom: 2rem;
  }

  &:last-child {
    border-bottom: none;
    padding-bottom: 0;
  }

  &__title {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 1.125rem;
    margin: 0 0 1rem 0;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 1.25rem;
    }
  }

  &__summary {
    font-size: 1rem;
    line-height: 1.6;
    color: var(--text-primary);
    margin: 0;

    @media (min-width: 768px) {
      font-size: 1.1rem;
    }
  }

  &__admin-notes {
    background: var(--bg-secondary);
    padding: 1rem;
    border-radius: 8px;
    font-style: italic;
    color: var(--text-secondary);
    margin: 0;
  }
}

.consolidation-list {
  list-style: none;
  padding: 0;
  margin: 0;

  li {
    display: flex;
    align-items: flex-start;
    gap: 0.6rem;
    padding: 1rem;
    margin-bottom: 0.75rem;
    border-radius: 8px;
    color: var(--text-primary);
  }

  &__icon {
    flex-shrink: 0;
    margin-top: 0.15rem;
  }

  &--positive li {
    background: rgba(76, 175, 80, 0.1);
  }

  &--positive &__icon {
    color: var(--color-success);
  }

  &--improvement li {
    background: rgba(255, 152, 0, 0.1);
  }

  &--improvement &__icon {
    color: var(--color-warning);
  }

  &--action li {
    background: rgba(33, 150, 243, 0.1);
  }

  &--action &__icon {
    color: var(--color-info);
  }
}

.questions {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1rem;

  @media (min-width: 768px) {
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  }
}

.question-card {
  padding: 1.25rem;
  background: var(--bg-secondary);
  border-radius: 8px;
  border: 1px solid var(--border-color);

  @media (min-width: 768px) {
    padding: 1.5rem;
  }

  &__title {
    color: var(--color-primary);
    font-size: 0.85rem;
    margin: 0 0 0.5rem 0;
    font-weight: 600;

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }
  }

  &__text {
    color: var(--text-secondary);
    font-size: 0.85rem;
    line-height: 1.5;
    margin: 0;

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }
  }
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal {
  background: var(--bg-primary);
  border-radius: 12px;
  width: 100%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;

  &__header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.25rem;
    border-bottom: 1px solid var(--border-color);

    @media (min-width: 768px) {
      padding: 1.5rem;
    }
  }

  &__title {
    margin: 0;
    font-size: 1.25rem;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 1.5rem;
    }
  }

  &__close {
    background: none;
    border: none;
    font-size: 1.5rem;
    cursor: pointer;
    color: var(--text-secondary);
    padding: 0;
    min-width: 44px;
    min-height: 44px;
    display: flex;
    align-items: center;
    justify-content: center;

    &:hover {
      color: var(--text-primary);
    }
  }

  &__form {
    padding: 1.25rem;

    @media (min-width: 768px) {
      padding: 1.5rem;
    }
  }

  &__field {
    margin-bottom: 1.5rem;
  }

  &__label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
    color: var(--text-primary);
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__input,
  &__select {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid var(--border-color);
    border-radius: 6px;
    font-size: 0.9rem;
    background: var(--bg-primary);
    color: var(--text-primary);
    min-height: 44px;

    @media (min-width: 768px) {
      font-size: 1rem;
    }

    &:focus {
      outline: none;
      border-color: var(--color-primary);
      box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
    }
  }

  &__actions {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-top: 1.5rem;

    @media (min-width: 768px) {
      flex-direction: row;
      gap: 1rem;
      justify-content: flex-end;
      margin-top: 2rem;
    }
  }
}

.reviewer-mgmt {
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 1rem;
  background: var(--bg-secondary);

  &__warning {
    background: rgba(255, 152, 0, 0.1);
    border: 1px solid var(--color-warning);
    color: var(--text-primary);
    padding: 0.75rem;
    border-radius: 4px;
    font-size: 0.8rem;
    margin-bottom: 1rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;

    @media (min-width: 768px) {
      font-size: 0.85rem;
    }
  }

  &__current {
    margin-bottom: 1rem;
  }

  &__list {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  &__add {
    // No specific styles needed
  }

  &__select {
    width: 100%;
    padding: 0.5rem;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-size: 0.85rem;
    min-height: 44px;

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }

    &:focus {
      outline: none;
      border-color: var(--color-primary);
    }
  }
}

.reviewer-tag {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--color-primary);
  color: white;
  padding: 0.375rem 0.625rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 500;

  @media (min-width: 768px) {
    font-size: 0.8rem;
  }

  &__name {
    max-width: 100px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  &__remove {
    background: none;
    border: none;
    color: white;
    cursor: pointer;
    font-size: 1rem;
    padding: 0;
    min-width: 20px;
    min-height: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    transition: background 0.2s;

    &:hover {
      background: rgba(255, 255, 255, 0.2);
    }
  }
}
</style>
