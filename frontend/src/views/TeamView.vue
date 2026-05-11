<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'
import type { UserWithFeedbackStats } from '@/types/user'
import type { Team } from '@/types/team'
import { PhArrowUp, PhArrowDown } from '@phosphor-icons/vue'

const auth = useAuthStore()
const teamMembers = ref<UserWithFeedbackStats[]>([])
const teams = ref<Team[]>([])
const loading = ref(true)
const updating = ref<string | null>(null)

// Filter state
const selectedTeamId = ref<string>('')

// Sorting state
const sortColumn = ref<keyof UserWithFeedbackStats>('lastFeedbackReceived')
const sortDirection = ref<'asc' | 'desc'>('asc')

onMounted(async () => {
  await Promise.all([loadTeam(), loadTeams()])
})

async function loadTeam() {
  try {
    // Use new endpoint with feedback stats
    const response = await apiClient.get('/users/with-feedback-stats')
    teamMembers.value = response.data
  } catch (error) {
    console.error('Failed to load team:', error)
  } finally {
    loading.value = false
  }
}

async function loadTeams() {
  try {
    const response = await apiClient.get('/teams')
    teams.value = response.data
  } catch (error) {
    console.error('Failed to load teams:', error)
  }
}

function getTeamName(teamId: string | null | undefined): string {
  if (!teamId) return 'Unassigned'
  const team = teams.value.find(t => t.id === teamId)
  return team?.name || 'Unknown'
}

// Filter and sort logic
const sortedTeamMembers = computed(() => {
  let members = [...teamMembers.value]

  // Apply team filter
  if (selectedTeamId.value) {
    members = members.filter(m => m.teamId === selectedTeamId.value)
  }

  members.sort((a, b) => {
    let aVal = a[sortColumn.value]
    let bVal = b[sortColumn.value]

    // Handle null values (push to end)
    if (aVal === null || aVal === undefined) return 1
    if (bVal === null || bVal === undefined) return -1

    // Handle date strings
    if (sortColumn.value === 'lastFeedbackReceived' ||
        sortColumn.value === 'createdAt' ||
        sortColumn.value === 'lastLogin') {
      aVal = aVal ? new Date(aVal as string).getTime() : 0
      bVal = bVal ? new Date(bVal as string).getTime() : 0
    }

    // Handle strings (case insensitive)
    if (typeof aVal === 'string' && typeof bVal === 'string') {
      aVal = aVal.toLowerCase()
      bVal = bVal.toLowerCase()
    }

    // Compare
    if (aVal < bVal) return sortDirection.value === 'asc' ? -1 : 1
    if (aVal > bVal) return sortDirection.value === 'asc' ? 1 : -1
    return 0
  })

  return members
})

function sortBy(column: keyof UserWithFeedbackStats) {
  if (sortColumn.value === column) {
    // Toggle direction if same column
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
  } else {
    // New column, default to ascending
    sortColumn.value = column
    sortDirection.value = 'asc'
  }
}

async function toggleRole(member: UserWithFeedbackStats) {
  if (!auth.isAdmin || updating.value === member.id) return

  const newRole = member.role === 'admin' ? 'member' : 'admin'
  updating.value = member.id

  try {
    await apiClient.put(`/users/${member.id}/role`, { role: newRole })
    member.role = newRole
  } catch (error) {
    console.error('Failed to update role:', error)
    alert('Failed to update role. Please try again.')
  } finally {
    updating.value = null
  }
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return 'Never'
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

function getRelativeTime(dateStr: string | null): string {
  if (!dateStr) return 'Never'

  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

  if (diffDays === 0) return 'Today'
  if (diffDays === 1) return 'Yesterday'
  if (diffDays < 7) return `${diffDays} days ago`
  if (diffDays < 30) return `${Math.floor(diffDays / 7)} weeks ago`
  if (diffDays < 365) return `${Math.floor(diffDays / 30)} months ago`
  return `${Math.floor(diffDays / 365)} years ago`
}
</script>

<template>
  <div class="team">
    <header class="team__header">
      <h1 class="team__title">Team</h1>
      <p class="team__subtitle">Organization roster with feedback insights</p>
    </header>

    <!-- Team Filter -->
    <div v-if="!loading && teams.length > 0" class="team__filters">
      <select v-model="selectedTeamId" class="team__filter-select">
        <option value="">All Teams</option>
        <option v-for="team in teams" :key="team.id" :value="team.id">
          {{ team.name }}
        </option>
      </select>
    </div>

    <div v-if="loading" class="team__loading">Loading team members...</div>

    <div v-else class="team__table-wrapper">
      <table class="team-table">
        <thead class="team-table__head">
          <tr>
            <th @click="sortBy('name')" class="team-table__th team-table__th--sortable">
              <div class="team-table__th-content">
                Team Member
                <span class="team-table__sort" v-if="sortColumn === 'name'">
                  <PhArrowUp v-if="sortDirection === 'asc'" :size="14" weight="bold" />
                  <PhArrowDown v-else :size="14" weight="bold" />
                </span>
              </div>
            </th>
            <th @click="sortBy('role')" class="team-table__th team-table__th--sortable">
              <div class="team-table__th-content">
                Role
                <span class="team-table__sort" v-if="sortColumn === 'role'">
                  <PhArrowUp v-if="sortDirection === 'asc'" :size="14" weight="bold" />
                  <PhArrowDown v-else :size="14" weight="bold" />
                </span>
              </div>
            </th>
            <th class="team-table__th">
              <div class="team-table__th-content">Team</div>
            </th>
            <th @click="sortBy('lastFeedbackReceived')" class="team-table__th team-table__th--sortable">
              <div class="team-table__th-content">
                Last Feedback
                <span class="team-table__sort" v-if="sortColumn === 'lastFeedbackReceived'">
                  <PhArrowUp v-if="sortDirection === 'asc'" :size="14" weight="bold" />
                  <PhArrowDown v-else :size="14" weight="bold" />
                </span>
              </div>
            </th>
            <th @click="sortBy('activeRoundsAsSubject')" class="team-table__th team-table__th--sortable">
              <div class="team-table__th-content">
                Active Rounds
                <span class="team-table__sort" v-if="sortColumn === 'activeRoundsAsSubject'">
                  <PhArrowUp v-if="sortDirection === 'asc'" :size="14" weight="bold" />
                  <PhArrowDown v-else :size="14" weight="bold" />
                </span>
              </div>
            </th>
            <th @click="sortBy('pendingReviews')" class="team-table__th team-table__th--sortable">
              <div class="team-table__th-content">
                Pending Reviews
                <span class="team-table__sort" v-if="sortColumn === 'pendingReviews'">
                  <PhArrowUp v-if="sortDirection === 'asc'" :size="14" weight="bold" />
                  <PhArrowDown v-else :size="14" weight="bold" />
                </span>
              </div>
            </th>
            <th @click="sortBy('totalFeedbackReceived')" class="team-table__th team-table__th--sortable">
              <div class="team-table__th-content">
                Total Feedback
                <span class="team-table__sort" v-if="sortColumn === 'totalFeedbackReceived'">
                  <PhArrowUp v-if="sortDirection === 'asc'" :size="14" weight="bold" />
                  <PhArrowDown v-else :size="14" weight="bold" />
                </span>
              </div>
            </th>
            <th v-if="auth.isAdmin" class="team-table__th">Actions</th>
          </tr>
        </thead>
        <tbody class="team-table__body">
          <tr v-for="member in sortedTeamMembers" :key="member.id" class="team-table__row">
            <!-- Team Member Column -->
            <td class="team-table__td team-table__td--member">
              <div class="member">
                <img v-if="member.photoUrl" :src="member.photoUrl" :alt="member.name" class="member__photo">
                <div v-else class="member__photo member__photo--placeholder">{{ member.name.charAt(0).toUpperCase() }}</div>
                <div class="member__details">
                  <div class="member__name">{{ member.name }}</div>
                  <div class="member__email">{{ member.email }}</div>
                </div>
              </div>
            </td>

            <!-- Role Column -->
            <td class="team-table__td">
              <span class="badge" :class="{
                'badge--admin': member.role === 'admin',
                'badge--member': member.role === 'member',
                'badge--team-admin': member.role === 'team_admin'
              }">
                {{ member.role }}
              </span>
            </td>

            <!-- Team Column -->
            <td class="team-table__td">
              <span class="badge badge--team" :class="{ 'badge--unassigned': !member.teamId }">
                {{ getTeamName(member.teamId) }}
              </span>
            </td>

            <!-- Last Feedback Column -->
            <td class="team-table__td">
              <div class="feedback-date">
                <div class="feedback-date__main">{{ formatDate(member.lastFeedbackReceived) }}</div>
                <div class="feedback-date__relative" v-if="member.lastFeedbackReceived">
                  {{ getRelativeTime(member.lastFeedbackReceived) }}
                </div>
              </div>
            </td>

            <!-- Active Rounds Column -->
            <td class="team-table__td team-table__td--center">
              <span class="metric" :class="{ 'metric--active': member.activeRoundsAsSubject > 0 }">
                {{ member.activeRoundsAsSubject }}
              </span>
            </td>

            <!-- Pending Reviews Column -->
            <td class="team-table__td team-table__td--center">
              <span class="metric metric--warning" :class="{ 'metric--active': member.pendingReviews > 0 }">
                {{ member.pendingReviews }}
              </span>
            </td>

            <!-- Total Feedback Column -->
            <td class="team-table__td team-table__td--center">
              <span class="metric">{{ member.totalFeedbackReceived }}</span>
            </td>

            <!-- Actions Column -->
            <td v-if="auth.isAdmin" class="team-table__td">
              <button
                v-if="member.id !== auth.user?.id"
                class="team-table__action-btn"
                :class="{ 'team-table__action-btn--loading': updating === member.id }"
                @click="toggleRole(member)"
                :disabled="updating === member.id"
              >
                {{ member.role === 'admin' ? 'Demote' : 'Promote' }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="sortedTeamMembers.length === 0" class="team__empty">
        No team members found.
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.team {
  padding: 1rem;
  max-width: 1400px;
  margin: 0 auto;

  @media (min-width: 768px) {
    padding: 2rem;
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
    margin: 0;
  }

  &__filters {
    margin-bottom: 1.5rem;
  }

  &__filter-select {
    padding: 0.5rem 1rem;
    border: 2px solid var(--border-color);
    border-radius: 8px;
    font-size: 0.9rem;
    background: var(--bg-primary);
    color: var(--text-primary);
    cursor: pointer;
    transition: border-color 0.2s;
    min-height: 44px;

    @media (min-width: 768px) {
      font-size: 0.95rem;
    }

    &:focus {
      outline: none;
      border-color: var(--color-primary);
    }
  }

  &__loading {
    text-align: center;
    color: var(--text-secondary);
    padding: 3rem 1rem;
  }

  &__table-wrapper {
    background: var(--bg-primary);
    border-radius: 12px;
    border: 1px solid var(--border-color);
    overflow: hidden;

    @media (max-width: 768px) {
      overflow-x: auto;
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
}

.team-table {
  width: 100%;
  border-collapse: collapse;

  @media (max-width: 768px) {
    min-width: 900px;
    font-size: 0.85rem;
  }

  @media (min-width: 769px) and (max-width: 1024px) {
    font-size: 0.9rem;
  }

  &__head {
    background: var(--bg-secondary);
    border-bottom: 2px solid var(--border-color);
  }

  &__th {
    padding: 0.75rem;
    text-align: left;
    font-weight: 600;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: var(--text-secondary);

    @media (min-width: 768px) {
      padding: 1rem;
      font-size: 0.85rem;
    }

    &--sortable {
      cursor: pointer;
      user-select: none;
      transition: background 0.2s;

      &:hover {
        background: var(--bg-tertiary);
      }
    }
  }

  &__th-content {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  &__sort {
    font-size: 0.9rem;
    color: var(--color-primary);
  }

  &__body {
  }

  &__row {
    border-bottom: 1px solid var(--border-color);
    transition: background 0.15s;

    &:hover {
      background: var(--bg-secondary);
    }
  }

  &__td {
    padding: 0.75rem;
    vertical-align: middle;
    color: var(--text-primary);

    @media (min-width: 768px) {
      padding: 1rem;
    }

    &--member {
      min-width: 220px;

      @media (min-width: 768px) {
        min-width: 250px;
      }
    }

    &--center {
      text-align: center;
    }
  }

  &__action-btn {
    font-size: 0.75rem;
    padding: 0.4rem 0.75rem;
    border: 1px solid var(--border-color);
    border-radius: 6px;
    background: var(--bg-primary);
    color: var(--text-secondary);
    cursor: pointer;
    transition: all 0.2s;
    white-space: nowrap;
    min-height: 44px;

    @media (min-width: 768px) {
      font-size: 0.8rem;
      padding: 0.4rem 0.85rem;
    }

    &:hover {
      background: var(--bg-secondary);
      border-color: var(--color-primary);
      color: var(--color-primary);
    }

    &:disabled {
      opacity: 0.6;
      cursor: not-allowed;
    }

    &--loading {
      opacity: 0.6;
    }
  }
}

.member {
  display: flex;
  align-items: center;
  gap: 0.5rem;

  @media (min-width: 768px) {
    gap: 0.75rem;
  }

  &__photo {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;

    @media (min-width: 768px) {
      width: 40px;
      height: 40px;
    }

    &--placeholder {
      background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
      color: white;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 1rem;
      font-weight: 600;

      @media (min-width: 768px) {
        font-size: 1.1rem;
      }
    }
  }

  &__details {
    min-width: 0;
    flex: 1;
  }

  &__name {
    font-weight: 500;
    font-size: 0.85rem;
    margin-bottom: 0.15rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 0.95rem;
    }
  }

  &__email {
    color: var(--text-secondary);
    font-size: 0.75rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;

    @media (min-width: 768px) {
      font-size: 0.85rem;
    }
  }
}

.badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.7rem;
  font-weight: 500;
  text-transform: capitalize;

  @media (min-width: 768px) {
    font-size: 0.75rem;
  }

  &--admin {
    background: rgba(25, 118, 210, 0.1);
    color: #1976d2;
  }

  &--member {
    background: rgba(123, 31, 162, 0.1);
    color: #7b1fa2;
  }

  &--team-admin {
    background: rgba(46, 125, 50, 0.1);
    color: #2e7d32;
  }

  &--team {
    background: rgba(25, 118, 210, 0.1);
    color: #1976d2;
  }

  &--unassigned {
    background: var(--bg-tertiary);
    color: var(--text-tertiary);
  }
}

.feedback-date {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;

  &__main {
    font-size: 0.85rem;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 0.9rem;
    }
  }

  &__relative {
    font-size: 0.7rem;
    color: var(--text-tertiary);

    @media (min-width: 768px) {
      font-size: 0.75rem;
    }
  }
}

.metric {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  padding: 0.25rem 0.4rem;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 500;
  background: var(--bg-tertiary);
  color: var(--text-secondary);

  @media (min-width: 768px) {
    min-width: 32px;
    padding: 0.25rem 0.5rem;
    font-size: 0.9rem;
  }

  &--active {
    background: rgba(25, 118, 210, 0.1);
    color: #1976d2;
  }

  &--warning {
    &.metric--active {
      background: rgba(245, 124, 0, 0.1);
      color: #f57c00;
    }
  }
}
</style>
