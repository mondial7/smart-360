<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import apiClient from '@/api/client'
import type { UserWithFeedbackStats } from '@/types/user'

const auth = useAuthStore()
const teamMembers = ref<UserWithFeedbackStats[]>([])
const loading = ref(true)
const updating = ref<string | null>(null)

// Sorting state
const sortColumn = ref<keyof UserWithFeedbackStats>('lastFeedbackReceived')
const sortDirection = ref<'asc' | 'desc'>('asc')

onMounted(async () => {
  await loadTeam()
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

// Sorting logic
const sortedTeamMembers = computed(() => {
  const members = [...teamMembers.value]

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

function formatDateTime(dateStr: string | null): string {
  if (!dateStr) return 'Never'
  return new Date(dateStr).toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
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
  <div class="team-page">
    <header class="page-header">
      <h1>Team</h1>
      <p>Organization roster with feedback insights</p>
    </header>

    <div v-if="loading" class="loading">Loading team members...</div>

    <div v-else class="table-container">
      <table class="team-table">
        <thead>
          <tr>
            <th @click="sortBy('name')" class="sortable">
              <div class="th-content">
                Team Member
                <span class="sort-indicator" v-if="sortColumn === 'name'">
                  {{ sortDirection === 'asc' ? '↑' : '↓' }}
                </span>
              </div>
            </th>
            <th @click="sortBy('role')" class="sortable">
              <div class="th-content">
                Role
                <span class="sort-indicator" v-if="sortColumn === 'role'">
                  {{ sortDirection === 'asc' ? '↑' : '↓' }}
                </span>
              </div>
            </th>
            <th @click="sortBy('lastFeedbackReceived')" class="sortable">
              <div class="th-content">
                Last Feedback
                <span class="sort-indicator" v-if="sortColumn === 'lastFeedbackReceived'">
                  {{ sortDirection === 'asc' ? '↑' : '↓' }}
                </span>
              </div>
            </th>
            <th @click="sortBy('activeRoundsAsSubject')" class="sortable">
              <div class="th-content">
                Active Rounds
                <span class="sort-indicator" v-if="sortColumn === 'activeRoundsAsSubject'">
                  {{ sortDirection === 'asc' ? '↑' : '↓' }}
                </span>
              </div>
            </th>
            <th @click="sortBy('pendingReviews')" class="sortable">
              <div class="th-content">
                Pending Reviews
                <span class="sort-indicator" v-if="sortColumn === 'pendingReviews'">
                  {{ sortDirection === 'asc' ? '↑' : '↓' }}
                </span>
              </div>
            </th>
            <th @click="sortBy('totalFeedbackReceived')" class="sortable">
              <div class="th-content">
                Total Feedback
                <span class="sort-indicator" v-if="sortColumn === 'totalFeedbackReceived'">
                  {{ sortDirection === 'asc' ? '↑' : '↓' }}
                </span>
              </div>
            </th>
            <th v-if="auth.isAdmin">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="member in sortedTeamMembers" :key="member.id">
            <!-- Team Member Column -->
            <td class="member-cell">
              <div class="member-info">
                <img v-if="member.photoUrl" :src="member.photoUrl" :alt="member.name" class="member-photo">
                <div v-else class="member-photo-placeholder">{{ member.name.charAt(0).toUpperCase() }}</div>
                <div class="member-details">
                  <div class="member-name">{{ member.name }}</div>
                  <div class="member-email">{{ member.email }}</div>
                </div>
              </div>
            </td>

            <!-- Role Column -->
            <td>
              <span class="role-badge" :class="member.role">{{ member.role }}</span>
            </td>

            <!-- Last Feedback Column -->
            <td>
              <div class="feedback-date">
                <div class="date-main">{{ formatDate(member.lastFeedbackReceived) }}</div>
                <div class="date-relative" v-if="member.lastFeedbackReceived">
                  {{ getRelativeTime(member.lastFeedbackReceived) }}
                </div>
              </div>
            </td>

            <!-- Active Rounds Column -->
            <td class="center">
              <span class="metric-badge" :class="{ 'has-value': member.activeRoundsAsSubject > 0 }">
                {{ member.activeRoundsAsSubject }}
              </span>
            </td>

            <!-- Pending Reviews Column -->
            <td class="center">
              <span class="metric-badge warning" :class="{ 'has-value': member.pendingReviews > 0 }">
                {{ member.pendingReviews }}
              </span>
            </td>

            <!-- Total Feedback Column -->
            <td class="center">
              <span class="metric-badge">{{ member.totalFeedbackReceived }}</span>
            </td>

            <!-- Actions Column -->
            <td v-if="auth.isAdmin">
              <button
                v-if="member.id !== auth.user?.id"
                class="role-toggle-btn"
                :class="{ 'loading': updating === member.id }"
                @click="toggleRole(member)"
                :disabled="updating === member.id"
              >
                {{ member.role === 'admin' ? 'Demote' : 'Promote' }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="sortedTeamMembers.length === 0" class="empty-state">
        No team members found.
      </div>
    </div>
  </div>
</template>

<style scoped>
.team-page {
  padding: 2rem;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 2rem;
}

.page-header h1 {
  font-size: 2rem;
  margin-bottom: 0.5rem;
}

.page-header p {
  color: #666;
}

.loading {
  text-align: center;
  color: #666;
  padding: 3rem;
}

.table-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  overflow: hidden;
}

.team-table {
  width: 100%;
  border-collapse: collapse;
}

.team-table thead {
  background: #f8f9fa;
  border-bottom: 2px solid #e0e0e0;
}

.team-table th {
  padding: 1rem;
  text-align: left;
  font-weight: 600;
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: #555;
}

.team-table th.sortable {
  cursor: pointer;
  user-select: none;
  transition: background 0.2s;
}

.team-table th.sortable:hover {
  background: #eff0f2;
}

.th-content {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.sort-indicator {
  font-size: 0.9rem;
  color: #667eea;
}

.team-table tbody tr {
  border-bottom: 1px solid #f0f0f0;
  transition: background 0.15s;
}

.team-table tbody tr:hover {
  background: #f9fafb;
}

.team-table td {
  padding: 1rem;
  vertical-align: middle;
}

.team-table td.center {
  text-align: center;
}

.member-cell {
  min-width: 250px;
}

.member-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.member-photo {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
}

.member-photo-placeholder {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.1rem;
  font-weight: 600;
  flex-shrink: 0;
}

.member-details {
  min-width: 0;
}

.member-name {
  font-weight: 500;
  font-size: 0.95rem;
  margin-bottom: 0.15rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.member-email {
  color: #666;
  font-size: 0.85rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.role-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: capitalize;
}

.role-badge.admin {
  background: #e3f2fd;
  color: #1976d2;
}

.role-badge.member {
  background: #f3e5f5;
  color: #7b1fa2;
}

.feedback-date {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.date-main {
  font-size: 0.9rem;
  color: #333;
}

.date-relative {
  font-size: 0.75rem;
  color: #888;
}

.metric-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 32px;
  padding: 0.25rem 0.5rem;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  background: #f5f5f5;
  color: #666;
}

.metric-badge.has-value {
  background: #e3f2fd;
  color: #1976d2;
}

.metric-badge.warning.has-value {
  background: #fff3e0;
  color: #f57c00;
}

.role-toggle-btn {
  font-size: 0.8rem;
  padding: 0.4rem 0.85rem;
  border: 1px solid #ddd;
  border-radius: 6px;
  background: white;
  color: #666;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.role-toggle-btn:hover {
  background: #f5f5f5;
  border-color: #667eea;
  color: #667eea;
}

.role-toggle-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.role-toggle-btn.loading {
  opacity: 0.6;
}

.empty-state {
  text-align: center;
  padding: 3rem;
  color: #666;
}

/* Responsive Design */
@media (max-width: 1024px) {
  .team-table {
    font-size: 0.9rem;
  }

  .team-table th,
  .team-table td {
    padding: 0.75rem;
  }

  .member-photo,
  .member-photo-placeholder {
    width: 36px;
    height: 36px;
  }
}

@media (max-width: 768px) {
  .team-page {
    padding: 1rem;
  }

  .table-container {
    overflow-x: auto;
  }

  .team-table {
    min-width: 900px;
  }
}
</style>
