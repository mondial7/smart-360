<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'
import type { DashboardStats } from '@/types/dashboard'
import type { FeedbackRound } from '@/types/round'
import MyAnalyticsCard from '@/components/MyAnalyticsCard.vue'

const auth = useAuthStore()
const stats = ref<DashboardStats | null>(null)
const activeRounds = ref<FeedbackRound[]>([])
const myRounds = ref<FeedbackRound[]>([])
const submissionStatus = ref<Record<string, boolean>>({})
const loading = ref(true)

onMounted(async () => {
  await loadDashboard()
})

async function loadDashboard() {
  try {
    if (auth.isAdmin) {
      const [statsRes, roundsRes] = await Promise.all([
        apiClient.get('/dashboard/stats'),
        apiClient.get('/rounds')
      ])
      stats.value = statsRes.data
      activeRounds.value = (roundsRes.data || []).filter((r: FeedbackRound) => r.status === 'active')
    } else {
      const [statsRes, roundsRes] = await Promise.all([
        apiClient.get('/dashboard/stats'),
        apiClient.get('/rounds-for-me')
      ])
      stats.value = statsRes.data
      myRounds.value = roundsRes.data || []

      if (myRounds.value && myRounds.value.length > 0) {
        for (const round of myRounds.value) {
          try {
            const checkRes = await apiClient.get(`/submissions/check/${round.id}`)
            submissionStatus.value[round.id] = checkRes.data.submitted
          } catch (err) {
            submissionStatus.value[round.id] = false
          }
        }
      }
    }
  } catch (error) {
    console.error('Failed to load dashboard:', error)
  } finally {
    loading.value = false
  }
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return 'No deadline'
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric'
  })
}

function getDeadlineStatus(deadline: string | null): string {
  if (!deadline) return 'no-deadline'
  const deadlineDate = new Date(deadline)
  const today = new Date()
  const daysLeft = Math.ceil((deadlineDate.getTime() - today.getTime()) / (1000 * 60 * 60 * 24))

  if (daysLeft < 0) return 'overdue'
  if (daysLeft === 0) return 'due-today'
  if (daysLeft <= 3) return 'due-soon'
  return 'on-time'
}

function getRoundStatusText(round: FeedbackRound): string {
  if (submissionStatus.value[round.id]) return 'Submitted'
  if (round.status !== 'active') return 'Closed'
  if (getDeadlineStatus(round.deadline) === 'overdue') return 'Overdue'
  return 'Pending'
}
</script>

<template>
  <div class="dashboard">
    <div class="dashboard__header">
      <h1 class="dashboard__title">Dashboard</h1>
      <p class="dashboard__welcome">Welcome back, {{ auth.user?.name }}</p>
    </div>

    <div v-if="loading" class="dashboard__loading">Loading dashboard...</div>

    <template v-else>
      <!-- Admin Dashboard -->
      <template v-if="auth.isAdmin">
        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-card__value">{{ stats?.totalUsers || 0 }}</div>
            <div class="stat-card__label">Total Users</div>
            <div class="stat-card__breakdown">
              <span>{{ stats?.adminCount || 0 }} admins</span>
              <span>{{ stats?.memberCount || 0 }} members</span>
            </div>
          </div>

          <div class="stat-card">
            <div class="stat-card__value">{{ stats?.totalRounds || 0 }}</div>
            <div class="stat-card__label">Total Rounds</div>
          </div>

          <div class="stat-card stat-card--accent">
            <div class="stat-card__value">{{ stats?.activeRounds || 0 }}</div>
            <div class="stat-card__label">Active Rounds</div>
          </div>

          <div class="stat-card" v-if="stats?.pendingReviews">
            <div class="stat-card__value">{{ stats.pendingReviews }}</div>
            <div class="stat-card__label">Pending Reviews</div>
          </div>
        </div>

        <div class="card dashboard__section" v-if="activeRounds.length > 0">
          <h2 class="dashboard__section-title">Active Rounds</h2>
          <div class="rounds-list">
            <div v-for="round in activeRounds.slice(0, 3)" :key="round.id" class="round-item">
              <div class="round-item__subject">
                <img v-if="round.subject?.photoUrl" :src="round.subject.photoUrl" class="round-item__avatar">
                <div v-else class="round-item__avatar round-item__avatar--placeholder">
                  {{ round.subject?.name.charAt(0) }}
                </div>
                <span class="round-item__name">{{ round.subject?.name }}</span>
              </div>
              <div class="round-item__meta">
                <span class="round-item__deadline">Due {{ formatDate(round.deadline) }}</span>
                <span class="round-item__count">{{ round.reviewers?.length || 0 }} reviewers</span>
              </div>
            </div>
          </div>
          <router-link to="/rounds" class="dashboard__view-all">View all rounds →</router-link>
        </div>

        <div class="empty-state" v-else>
          <div class="empty-state__icon">📋</div>
          <p class="empty-state__text">No active feedback rounds. Create your first round to get started!</p>
        </div>

        <div class="dashboard__section">
          <h2 class="dashboard__section-title">Quick Actions</h2>
          <div class="action-grid">
            <router-link to="/team" class="action-card">
              <span class="action-card__icon">👥</span>
              <span class="action-card__text">Manage Team</span>
            </router-link>
            <router-link to="/rounds" class="action-card">
              <span class="action-card__icon">📋</span>
              <span class="action-card__text">View Rounds</span>
            </router-link>
            <router-link to="/rounds/new" class="action-card action-card--primary">
              <span class="action-card__icon">➕</span>
              <span class="action-card__text">Create Round</span>
            </router-link>
          </div>
        </div>
      </template>

      <!-- Member Dashboard -->
      <template v-else>
        <div class="stats-grid stats-grid--member">
          <div class="stat-card stat-card--accent" v-if="myRounds.filter(r => r.status === 'active' && !submissionStatus[r.id]).length > 0">
            <div class="stat-card__value">
              {{ myRounds.filter(r => r.status === 'active' && !submissionStatus[r.id]).length }}
            </div>
            <div class="stat-card__label">Pending Reviews</div>
            <router-link to="/rounds" class="stat-card__link">Review now →</router-link>
          </div>

          <div class="stat-card">
            <div class="stat-card__value">{{ stats?.mySubmissions || 0 }}</div>
            <div class="stat-card__label">Completed Reviews</div>
          </div>

          <div class="stat-card">
            <div class="stat-card__value">{{ stats?.myFeedbackCount || 0 }}</div>
            <div class="stat-card__label">My Feedback</div>
            <router-link to="/my-feedback" class="stat-card__link">View →</router-link>
          </div>
        </div>

        <MyAnalyticsCard />

        <div v-if="myRounds.length > 0" class="feedback-section">
          <h2 class="dashboard__section-title">Feedback Requests</h2>
          <div class="feedback-list">
            <div v-for="round in myRounds" :key="round.id" class="feedback-item card">
              <div class="feedback-item__header">
                <div class="feedback-item__person">
                  <img v-if="round.subject?.photoUrl" :src="round.subject.photoUrl" class="feedback-item__avatar">
                  <div v-else class="feedback-item__avatar feedback-item__avatar--placeholder">
                    {{ round.subject?.name.charAt(0) }}
                  </div>
                  <div>
                    <h3 class="feedback-item__name">{{ round.subject?.name }}</h3>
                    <span :class="['feedback-item__status', `feedback-item__status--${getRoundStatusText(round).toLowerCase()}`]">
                      {{ getRoundStatusText(round) }}
                    </span>
                  </div>
                </div>
                <div class="feedback-item__deadline">
                  <span class="feedback-item__deadline-label">Due</span>
                  <span :class="['feedback-item__deadline-date', `feedback-item__deadline-date--${getDeadlineStatus(round.deadline)}`]">
                    {{ formatDate(round.deadline) }}
                  </span>
                </div>
              </div>

              <div class="feedback-item__actions">
                <template v-if="!submissionStatus[round.id] && round.status === 'active'">
                  <router-link :to="`/rounds/${round.id}/submit`" class="btn btn--primary">
                    Submit Feedback
                  </router-link>
                </template>
                <template v-else-if="submissionStatus[round.id]">
                  <router-link :to="`/rounds/${round.id}/submission`" class="btn btn--secondary">
                    View Submission
                  </router-link>
                </template>
                <template v-else>
                  <span class="feedback-item__closed">Round Closed</span>
                </template>
              </div>
            </div>
          </div>
        </div>

        <div class="empty-state card" v-else>
          <div class="empty-state__icon">📋</div>
          <h2 class="empty-state__title">No Feedback Requests</h2>
          <p class="empty-state__text">You don't have any pending feedback requests at the moment.</p>
          <router-link to="/team" class="btn btn--primary">View Team</router-link>
        </div>
      </template>
    </template>
  </div>
</template>

<style scoped lang="scss">
.dashboard {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;

  @media (max-width: 767px) {
    padding: 1rem;
  }

  &__header {
    margin-bottom: 2rem;
  }

  &__title {
    font-size: 2rem;
    margin-bottom: 0.5rem;
    color: var(--text-primary);

    @media (max-width: 767px) {
      font-size: 1.5rem;
    }
  }

  &__welcome {
    color: var(--text-secondary);
  }

  &__loading {
    text-align: center;
    color: var(--text-secondary);
    padding: 3rem;
  }

  &__section {
    margin-bottom: 2rem;
  }

  &__section-title {
    font-size: 1.25rem;
    margin-bottom: 1rem;
    color: var(--text-primary);
  }

  &__view-all {
    display: block;
    margin-top: 1rem;
    text-align: center;
    color: var(--color-primary);
    text-decoration: none;
    font-weight: 500;

    &:hover {
      text-decoration: underline;
    }
  }
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;

  @media (max-width: 767px) {
    grid-template-columns: 1fr;
  }

  &--member {
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    max-width: 600px;

    @media (max-width: 767px) {
      grid-template-columns: 1fr;
    }
  }
}

.stat-card {
  background: var(--bg-primary);
  padding: 1.5rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  display: flex;
  flex-direction: column;

  &--accent {
    background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
    color: white;
  }

  &__value {
    font-size: 2.5rem;
    font-weight: 700;
    line-height: 1;

    @media (max-width: 767px) {
      font-size: 2rem;
    }
  }

  &__label {
    font-size: 0.9rem;
    color: var(--text-secondary);
    margin-top: 0.5rem;

    .stat-card--accent & {
      color: rgba(255, 255, 255, 0.8);
    }
  }

  &__breakdown {
    display: flex;
    gap: 1rem;
    margin-top: 0.75rem;
    font-size: 0.8rem;
    color: var(--text-tertiary);
  }

  &__link {
    font-size: 0.85rem;
    color: var(--color-primary);
    text-decoration: none;
    margin-top: 0.5rem;

    .stat-card--accent & {
      color: rgba(255, 255, 255, 0.9);
    }
  }
}

.rounds-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.round-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: var(--bg-secondary);
  border-radius: 8px;

  @media (max-width: 767px) {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
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

    &--placeholder {
      background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
      color: white;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 0.8rem;
      font-weight: 600;
    }
  }

  &__name {
    font-weight: 500;
    color: var(--text-primary);
  }

  &__meta {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0.25rem;

    @media (max-width: 767px) {
      align-items: flex-start;
    }
  }

  &__deadline {
    font-size: 0.85rem;
    color: var(--color-error);
    font-weight: 500;
  }

  &__count {
    font-size: 0.8rem;
    color: var(--text-secondary);
  }
}

.empty-state {
  text-align: center;
  padding: 3rem 2rem;
  background: var(--bg-secondary);
  border-radius: 12px;
  margin-bottom: 2rem;

  &__icon {
    font-size: 3rem;
    margin-bottom: 1rem;
    opacity: 0.5;
  }

  &__title {
    font-size: 1.5rem;
    margin-bottom: 0.5rem;
    color: var(--text-primary);
  }

  &__text {
    color: var(--text-secondary);
    margin-bottom: 1.5rem;
  }
}

.action-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 1rem;

  @media (max-width: 767px) {
    grid-template-columns: repeat(2, 1fr);
  }
}

.action-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 1.5rem;
  background: var(--bg-primary);
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  text-decoration: none;
  color: var(--text-primary);
  transition: transform 0.2s, box-shadow 0.2s;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  }

  &--primary {
    background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
    color: white;
  }

  &__icon {
    font-size: 1.5rem;
  }

  &__text {
    font-size: 0.85rem;
    font-weight: 500;
    text-align: center;
  }
}

.feedback-section {
  margin-top: 2rem;
}

.feedback-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.feedback-item {
  &__header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;

    @media (max-width: 767px) {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.75rem;
    }
  }

  &__person {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  &__avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;

    &--placeholder {
      background: var(--color-primary);
      color: white;
      display: flex;
      align-items: center;
      justify-content: center;
      font-weight: 600;
    }
  }

  &__name {
    margin: 0 0 0.25rem 0;
    font-size: 1rem;
    color: var(--text-primary);
  }

  &__status {
    font-size: 0.85rem;
    font-weight: 500;

    &--pending {
      color: var(--color-warning);
    }

    &--submitted {
      color: var(--color-success);
    }

    &--overdue {
      color: var(--color-error);
    }

    &--closed {
      color: var(--text-tertiary);
    }
  }

  &__deadline {
    text-align: right;

    @media (max-width: 767px) {
      text-align: left;
    }
  }

  &__deadline-label {
    font-size: 0.75rem;
    color: var(--text-tertiary);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  &__deadline-date {
    font-size: 0.9rem;
    font-weight: 500;
    display: block;
    margin-top: 0.25rem;

    &--overdue {
      color: var(--color-error);
    }

    &--due-today,
    &--due-soon {
      color: var(--color-warning);
    }

    &--on-time {
      color: var(--color-success);
    }
  }

  &__actions {
    display: flex;
    justify-content: flex-end;
  }

  &__closed {
    color: var(--text-tertiary);
    font-size: 0.9rem;
    font-style: italic;
  }
}
</style>
