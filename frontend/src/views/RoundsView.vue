<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'
import type { FeedbackRound } from '@/types/round'
import { PhCheck, PhEye, PhNotePencil } from '@phosphor-icons/vue'

const auth = useAuthStore()
const rounds = ref<FeedbackRound[]>([])
const loading = ref(true)
const error = ref(null)
const submissionStatus = ref<Record<string, boolean>>({})
const consolidationStatus = ref<Record<string, boolean>>({})

onMounted(async () => {
  console.log('RoundsView mounted')
  await loadRounds()
})

async function loadRounds() {
  loading.value = true
  error.value = null
  try {
    // Admins see all rounds, members see rounds where they're reviewers or subject
    const endpoint = auth.isAdmin ? '/rounds' : '/rounds-for-me'
    const response = await apiClient.get(endpoint)
    rounds.value = response.data || []
    console.log('Loaded rounds:', rounds.value)
    
    // Check submission status and consolidation status for each round
    for (const round of rounds.value) {
      try {
        const checkRes = await apiClient.get(`/submissions/check/${round.id}`)
        submissionStatus.value[round.id] = checkRes.data.submitted
      } catch {
        submissionStatus.value[round.id] = false
      }
      
      // Check if consolidation exists
      try {
        await apiClient.get(`/consolidations/${round.id}`)
        consolidationStatus.value[round.id] = true
      } catch {
        consolidationStatus.value[round.id] = false
      }
    }
  } catch (err: any) {
    console.error('Failed to load rounds:', err)
    error.value = err.response?.data?.error || err.message || 'Failed to load rounds'
    rounds.value = []
  } finally {
    loading.value = false
  }
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return 'No deadline'
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function isAssignedReviewer(round: FeedbackRound): boolean {
  return round.reviewers?.some(r => r.reviewerId === auth.user?.id) ?? false
}

function hasSubmitted(roundId: string): boolean {  // Changed from number to string
  return submissionStatus.value[roundId] ?? false
}

async function closeRound(id: string) {  // Changed from number to string
  if (!confirm('Close this round? No more submissions will be accepted.')) return
  
  try {
    await apiClient.post(`/rounds/${id}/close`)
    await loadRounds()
  } catch (error) {
    console.error('Failed to close round:', error)
    alert('Failed to close round')
  }
}
</script>

<template>
  <div class="rounds">
    <header class="rounds__header">
      <div class="rounds__header-text">
        <h1 class="rounds__title">Feedback Rounds</h1>
        <p class="rounds__subtitle">Manage all feedback collection cycles</p>
      </div>
      <router-link v-if="auth.isAdmin" to="/rounds/new" class="btn btn--primary">
        + Create Round
      </router-link>
    </header>

    <div v-if="loading" class="rounds__loading">Loading rounds...</div>

    <div v-else-if="error" class="rounds__error">
      <p class="rounds__error-message">Error: {{ error }}</p>
      <button @click="loadRounds" class="btn btn--secondary">Retry</button>
    </div>

    <div v-else-if="rounds.length === 0" class="rounds__empty">
      <p class="rounds__empty-text">No feedback rounds yet.</p>
      <p v-if="auth.isAdmin" class="rounds__empty-subtext">
        <router-link to="/rounds/new">Create your first round</router-link> to start collecting feedback.
      </p>
    </div>

    <div v-else class="rounds__grid">
      <div v-for="round in rounds" :key="round.id" class="round-card">
        <div class="round-card__header">
          <span
            class="badge"
            :class="{
              'badge--active': round.status === 'active',
              'badge--draft': round.status === 'draft',
              'badge--closed': round.status === 'closed',
              'badge--shared': round.status === 'shared'
            }"
          >
            {{ round.status }}
          </span>
          <span class="round-card__date">Created {{ formatDate(round.createdAt) }}</span>
        </div>

        <div class="round-card__body">
          <div class="round-card__section">
            <span class="round-card__label">Feedback for</span>
            <div class="round-card__subject">
              <img v-if="round.subject?.photoUrl" :src="round.subject.photoUrl" class="round-card__avatar" alt="">
              <div v-else class="round-card__avatar round-card__avatar--placeholder">{{ round.subject?.name.charAt(0) }}</div>
              <router-link
                v-if="auth.isAdmin"
                :to="`/rounds/${round.id}`"
                class="round-card__name round-card__name--link"
              >
                {{ round.subject?.name }}
              </router-link>
              <span v-else class="round-card__name">{{ round.subject?.name }}</span>
            </div>
          </div>

          <div class="round-card__section">
            <span class="round-card__label">{{ round.reviewers?.length || 0 }} reviewers assigned</span>
            <div class="round-card__reviewers">
              <template v-for="reviewer in round.reviewers?.slice(0, 5)" :key="reviewer.id">
                <img
                  v-if="reviewer.reviewer?.photoUrl"
                  :src="reviewer.reviewer.photoUrl"
                  :title="reviewer.reviewer.name"
                  class="round-card__reviewer-avatar"
                  alt=""
                >
                <div
                  v-else
                  :title="reviewer.reviewer?.name || 'Unknown'"
                  class="round-card__reviewer-avatar round-card__reviewer-avatar--placeholder"
                >
                  {{ reviewer.reviewer?.name?.charAt(0) || '?' }}
                </div>
              </template>
              <span v-if="(round.reviewers?.length || 0) > 5" class="round-card__more">+{{ (round.reviewers?.length || 0) - 5 }}</span>
            </div>
          </div>

          <div class="round-card__section">
            <span class="round-card__label">Deadline</span>
            <span
              class="round-card__deadline"
              :class="{ 'round-card__deadline--overdue': round.status === 'active' && round.deadline && new Date(round.deadline) < new Date() }"
            >
              {{ formatDate(round.deadline) }}
            </span>
          </div>
        </div>

        <div v-if="isAssignedReviewer(round) && !hasSubmitted(round.id) && round.status === 'active'" class="round-card__actions">
          <router-link :to="`/rounds/${round.id}/submit`" class="btn btn--primary round-card__btn">
            Submit Feedback
          </router-link>
        </div>

        <div v-else-if="hasSubmitted(round.id)" class="round-card__actions round-card__actions--column">
          <span class="round-card__submitted">
            <PhCheck :size="14" weight="bold" />
            <span>Feedback Submitted</span>
          </span>
          <router-link :to="`/rounds/${round.id}/submission`" class="btn btn--secondary round-card__btn">
            View My Submission
          </router-link>
        </div>

        <div v-else-if="auth.isAdmin && round.status === 'closed'" class="round-card__actions">
          <router-link v-if="!consolidationStatus[round.id]" :to="`/rounds/${round.id}/consolidation`" class="btn btn--primary round-card__btn">
            Consolidate Feedback
          </router-link>
          <router-link v-else :to="`/rounds/${round.id}#consolidation`" class="btn btn--success btn--with-icon round-card__btn">
            <PhEye :size="16" weight="regular" />
            <span>View Feedback</span>
          </router-link>
        </div>

        <div v-else-if="auth.isAdmin && round.status === 'active'" class="round-card__actions round-card__actions--split">
          <router-link :to="`/rounds/${round.id}`" class="btn btn--primary btn--with-icon round-card__btn">
            <PhNotePencil :size="16" weight="regular" />
            <span>Edit</span>
          </router-link>
          <button class="btn btn--danger round-card__btn" @click="closeRound(round.id)">
            Close Round
          </button>
        </div>

        <div v-else-if="auth.isAdmin" class="round-card__actions">
          <router-link :to="`/rounds/${round.id}`" class="btn btn--primary btn--with-icon round-card__btn">
            <PhNotePencil :size="16" weight="regular" />
            <span>Edit</span>
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.rounds {
  padding: 1rem;
  max-width: 1200px;
  margin: 0 auto;

  @media (min-width: 768px) {
    padding: 2rem;
  }

  &__header {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-bottom: 2rem;

    @media (min-width: 768px) {
      flex-direction: row;
      justify-content: space-between;
      align-items: center;
    }
  }

  &__header-text {
    flex: 1;
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
    margin: 0;
  }

  &__loading {
    text-align: center;
    color: var(--text-secondary);
    padding: 3rem 1rem;
  }

  &__error {
    text-align: center;
    padding: 2rem 1rem;
    background: var(--bg-primary);
    border-radius: 12px;
    border: 1px solid var(--color-error);

    @media (min-width: 768px) {
      padding: 3rem;
    }
  }

  &__error-message {
    color: var(--color-error);
    margin-bottom: 1rem;
  }

  &__empty {
    text-align: center;
    padding: 2rem 1rem;
    background: var(--bg-primary);
    border-radius: 12px;
    border: 1px solid var(--border-color);

    @media (min-width: 768px) {
      padding: 3rem;
    }
  }

  &__empty-text {
    color: var(--text-secondary);
    margin: 0 0 0.5rem 0;
  }

  &__empty-subtext {
    margin: 0.5rem 0 0 0;
    font-size: 0.9rem;
    color: var(--text-tertiary);

    a {
      color: var(--color-primary);
      text-decoration: none;

      &:hover {
        text-decoration: underline;
      }
    }
  }

  &__grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 1rem;

    @media (min-width: 640px) {
      grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
      gap: 1.5rem;
    }
  }
}

.round-card {
  background: var(--bg-primary);
  border-radius: 12px;
  border: 1px solid var(--border-color);
  padding: 1.25rem;
  transition: box-shadow 0.2s;

  @media (min-width: 768px) {
    padding: 1.5rem;
  }

  &:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  &__header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  &__date {
    font-size: 0.75rem;
    color: var(--text-tertiary);

    @media (min-width: 768px) {
      font-size: 0.8rem;
    }
  }

  &__body {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  &__section {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  &__label {
    font-size: 0.7rem;
    color: var(--text-tertiary);
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

  &__avatar {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;

    @media (min-width: 768px) {
      width: 36px;
      height: 36px;
    }

    &--placeholder {
      background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
      color: white;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 0.9rem;
      font-weight: 600;
    }
  }

  &__name {
    font-weight: 500;
    color: var(--text-primary);
    font-size: 0.9rem;

    @media (min-width: 768px) {
      font-size: 1rem;
    }

    &--link {
      color: var(--color-primary);
      text-decoration: none;

      &:hover {
        text-decoration: underline;
      }
    }
  }

  &__reviewers {
    display: flex;
    align-items: center;
    gap: 0.25rem;
  }

  &__reviewer-avatar {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    object-fit: cover;
    border: 2px solid var(--bg-primary);
    margin-left: -8px;

    @media (min-width: 768px) {
      width: 28px;
      height: 28px;
    }

    &:first-child {
      margin-left: 0;
    }

    &--placeholder {
      background: var(--bg-tertiary);
      color: var(--text-secondary);
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 0.65rem;
      font-weight: 600;

      @media (min-width: 768px) {
        font-size: 0.7rem;
      }
    }
  }

  &__more {
    margin-left: 4px;
    font-size: 0.75rem;
    color: var(--text-secondary);

    @media (min-width: 768px) {
      font-size: 0.8rem;
    }
  }

  &__deadline {
    font-size: 0.85rem;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }

    &--overdue {
      color: var(--color-error);
      font-weight: 500;
    }
  }

  &__actions {
    margin-top: 1rem;
    padding-top: 1rem;
    border-top: 1px solid var(--border-color);
    display: flex;
    gap: 0.5rem;

    &--split {
      display: flex;
      gap: 0.5rem;
    }

    &--column {
      flex-direction: column;
      text-align: center;
      gap: 0.75rem;
    }
  }

  &__btn {
    flex: 1;
    min-height: 44px;
    font-size: 0.85rem;

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }
  }

  &__submitted {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    color: var(--color-success);
    font-size: 0.85rem;
    font-weight: 500;

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }
  }
}
</style>
