<template>
  <div class="audit-log-view">
    <div class="header">
      <h1>Audit Log</h1>
      <p class="subtitle">Track all changes and activities in the system</p>
    </div>

    <!-- Filters -->
    <div class="filters">
      <div class="filter-group">
        <label>Action Type</label>
        <select v-model="filterAction" @change="fetchAuditLogs">
          <option value="">All Actions</option>
          <option value="round.created">Round Created</option>
          <option value="round.status_changed">Status Changed</option>
          <option value="round.subject_changed">Subject Changed</option>
          <option value="round.deadline_changed">Deadline Changed</option>
          <option value="reviewer.added">Reviewer Added</option>
          <option value="reviewer.removed">Reviewer Removed</option>
          <option value="consolidation.shared">Consolidation Shared</option>
        </select>
      </div>

      <div class="filter-group">
        <label>Round Subject</label>
        <input
          v-model="filterSubject"
          @input="fetchAuditLogs"
          type="text"
          placeholder="Search by subject name..."
        />
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading">Loading audit logs...</div>

    <!-- Error State -->
    <div v-else-if="error" class="error">{{ error }}</div>

    <!-- Empty State -->
    <div v-else-if="logs.length === 0" class="empty-state">
      <p>No audit logs found.</p>
    </div>

    <!-- Audit Logs -->
    <div v-else class="audit-logs">
      <div
        v-for="log in filteredLogs"
        :key="log.id"
        class="audit-card"
        :class="`action-${getActionColor(log.action)}`"
      >
        <div class="card-header">
          <span class="action-badge" :class="`badge-${getActionColor(log.action)}`">
            {{ getActionIcon(log.action) }} {{ formatAction(log.action) }}
          </span>
          <span class="timestamp">{{ formatDate(log.createdAt) }}</span>
        </div>

        <div class="card-body">
          <p class="description">{{ log.description }}</p>

          <div class="details">
            <div class="detail-row">
              <span class="label">Actor:</span>
              <span class="value">{{ log.actorName }} ({{ log.actorEmail }})</span>
            </div>
            <div class="detail-row">
              <span class="label">Round:</span>
              <span class="value">{{ log.roundSubject }}</span>
            </div>
            <div v-if="log.oldValue || log.newValue" class="detail-row">
              <span class="label">Change:</span>
              <span class="value">
                <span v-if="log.oldValue" class="old-value">{{ log.oldValue }}</span>
                <span v-if="log.oldValue && log.newValue" class="arrow">→</span>
                <span v-if="log.newValue" class="new-value">{{ log.newValue }}</span>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="pagination">
      <button
        @click="previousPage"
        :disabled="currentPage === 1"
        class="page-btn"
      >
        Previous
      </button>

      <span class="page-info">
        Page {{ currentPage }} of {{ totalPages }} ({{ total }} total)
      </span>

      <button
        @click="nextPage"
        :disabled="currentPage === totalPages"
        class="page-btn"
      >
        Next
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import apiClient from '@/api/client'

interface AuditLog {
  id: string
  action: string
  actorId: string
  actorName: string
  actorEmail: string
  roundId: string
  roundSubject: string
  description: string
  oldValue?: string
  newValue?: string
  createdAt: string
}

interface AuditLogsResponse {
  logs: AuditLog[]
  total: number
  page: number
  limit: number
}

const logs = ref<AuditLog[]>([])
const loading = ref(true)
const error = ref('')
const total = ref(0)
const currentPage = ref(1)
const limit = 50

const filterAction = ref('')
const filterSubject = ref('')

const totalPages = computed(() => Math.ceil(total.value / limit))

const filteredLogs = computed(() => {
  return logs.value.filter(log => {
    const matchesAction = !filterAction.value || log.action === filterAction.value
    const matchesSubject = !filterSubject.value ||
      log.roundSubject.toLowerCase().includes(filterSubject.value.toLowerCase())
    return matchesAction && matchesSubject
  })
})

async function fetchAuditLogs() {
  try {
    loading.value = true
    error.value = ''
    const response = await apiClient.get<AuditLogsResponse>(`/audit-logs?page=${currentPage.value}&limit=${limit}`)
    logs.value = response.data.logs || []
    total.value = response.data.total || 0
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to load audit logs'
    console.error('Error fetching audit logs:', err)
  } finally {
    loading.value = false
  }
}

function nextPage() {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    fetchAuditLogs()
  }
}

function previousPage() {
  if (currentPage.value > 1) {
    currentPage.value--
    fetchAuditLogs()
  }
}

function getActionColor(action: string): string {
  if (action.includes('created')) return 'green'
  if (action.includes('status_changed')) return 'blue'
  if (action.includes('added')) return 'cyan'
  if (action.includes('removed')) return 'orange'
  if (action.includes('shared')) return 'purple'
  return 'gray'
}

function getActionIcon(action: string): string {
  if (action.includes('created')) return '✨'
  if (action.includes('status_changed')) return '🔄'
  if (action.includes('subject_changed')) return '👤'
  if (action.includes('deadline_changed')) return '📅'
  if (action.includes('added')) return '➕'
  if (action.includes('removed')) return '➖'
  if (action.includes('shared')) return '📤'
  return '📝'
}

function formatAction(action: string): string {
  return action
    .replace(/_/g, ' ')
    .replace(/\./g, ' ')
    .split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

function formatDate(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`

  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined
  })
}

onMounted(() => {
  fetchAuditLogs()
})
</script>

<style scoped>
.audit-log-view {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
}

.header {
  margin-bottom: 2rem;
}

.header h1 {
  font-size: 2rem;
  font-weight: 600;
  margin: 0 0 0.5rem 0;
}

.subtitle {
  color: #666;
  margin: 0;
}

.filters {
  display: flex;
  gap: 1rem;
  margin-bottom: 2rem;
  padding: 1.5rem;
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.filter-group {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.filter-group label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
}

.filter-group select,
.filter-group input {
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.875rem;
}

.filter-group select:focus,
.filter-group input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.loading,
.error,
.empty-state {
  text-align: center;
  padding: 3rem;
  color: #666;
}

.error {
  color: #dc2626;
}

.audit-logs {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.audit-card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border-left: 4px solid #d1d5db;
  overflow: hidden;
  transition: box-shadow 0.2s;
}

.audit-card:hover {
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.audit-card.action-green { border-left-color: #10b981; }
.audit-card.action-blue { border-left-color: #3b82f6; }
.audit-card.action-cyan { border-left-color: #06b6d4; }
.audit-card.action-orange { border-left-color: #f97316; }
.audit-card.action-purple { border-left-color: #a855f7; }

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  background: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
}

.action-badge {
  font-size: 0.875rem;
  font-weight: 500;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  background: #e5e7eb;
  color: #374151;
}

.badge-green { background: #d1fae5; color: #065f46; }
.badge-blue { background: #dbeafe; color: #1e40af; }
.badge-cyan { background: #cffafe; color: #155e75; }
.badge-orange { background: #fed7aa; color: #9a3412; }
.badge-purple { background: #e9d5ff; color: #6b21a8; }

.timestamp {
  font-size: 0.875rem;
  color: #6b7280;
}

.card-body {
  padding: 1.5rem;
}

.description {
  font-size: 1rem;
  margin: 0 0 1rem 0;
  color: #111827;
}

.details {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.detail-row {
  display: flex;
  gap: 0.5rem;
  font-size: 0.875rem;
}

.detail-row .label {
  font-weight: 500;
  color: #6b7280;
  min-width: 60px;
}

.detail-row .value {
  color: #374151;
}

.old-value {
  text-decoration: line-through;
  color: #9ca3af;
}

.arrow {
  color: #9ca3af;
  margin: 0 0.25rem;
}

.new-value {
  font-weight: 500;
  color: #10b981;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  margin-top: 2rem;
  padding: 1.5rem;
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.page-btn {
  padding: 0.5rem 1rem;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: background 0.2s;
}

.page-btn:hover:not(:disabled) {
  background: #2563eb;
}

.page-btn:disabled {
  background: #d1d5db;
  cursor: not-allowed;
}

.page-info {
  font-size: 0.875rem;
  color: #6b7280;
}
</style>
