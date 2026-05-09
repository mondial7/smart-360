<template>
  <div class="audit">
    <div class="audit__header">
      <h1 class="audit__title">Audit Log</h1>
      <p class="audit__subtitle">Track all changes and activities in the system</p>
    </div>

    <!-- Filters -->
    <div class="audit__filters">
      <div class="audit__filter-group">
        <label class="audit__filter-label">Action Type</label>
        <select v-model="filterAction" @change="fetchAuditLogs" class="audit__filter-select">
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

      <div class="audit__filter-group">
        <label class="audit__filter-label">Round Subject</label>
        <input
          v-model="filterSubject"
          @input="fetchAuditLogs"
          type="text"
          placeholder="Search by subject name..."
          class="audit__filter-input"
        />
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="audit__loading">Loading audit logs...</div>

    <!-- Error State -->
    <div v-else-if="error" class="audit__error">{{ error }}</div>

    <!-- Empty State -->
    <div v-else-if="logs.length === 0" class="audit__empty">
      <p>No audit logs found.</p>
    </div>

    <!-- Audit Logs -->
    <div v-else class="audit__list">
      <div
        v-for="log in filteredLogs"
        :key="log.id"
        class="log-card"
        :class="`log-card--${getActionColor(log.action)}`"
      >
        <div class="log-card__header">
          <span class="badge badge--with-icon" :class="`badge--${getActionColor(log.action)}`">
            <component :is="getActionIcon(log.action)" :size="14" weight="bold" />
            <span>{{ formatAction(log.action) }}</span>
          </span>
          <span class="log-card__timestamp">{{ formatDate(log.createdAt) }}</span>
        </div>

        <div class="log-card__body">
          <p class="log-card__description">{{ log.description }}</p>

          <div class="log-card__details">
            <div class="log-card__detail">
              <span class="log-card__label">Actor:</span>
              <span class="log-card__value">{{ log.actorName }} ({{ log.actorEmail }})</span>
            </div>
            <div class="log-card__detail">
              <span class="log-card__label">Round:</span>
              <span class="log-card__value">{{ log.roundSubject }}</span>
            </div>
            <div v-if="log.oldValue || log.newValue" class="log-card__detail">
              <span class="log-card__label">Change:</span>
              <span class="log-card__value">
                <span v-if="log.oldValue" class="log-card__old-value">{{ log.oldValue }}</span>
                <PhArrowRight v-if="log.oldValue && log.newValue" class="log-card__arrow" :size="14" weight="bold" />
                <span v-if="log.newValue" class="log-card__new-value">{{ log.newValue }}</span>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="audit__pagination">
      <button
        @click="previousPage"
        :disabled="currentPage === 1"
        class="btn btn--secondary audit__page-btn"
      >
        Previous
      </button>

      <span class="audit__page-info">
        Page {{ currentPage }} of {{ totalPages }} ({{ total }} total)
      </span>

      <button
        @click="nextPage"
        :disabled="currentPage === totalPages"
        class="btn btn--secondary audit__page-btn"
      >
        Next
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, markRaw, type Component } from 'vue'
import apiClient from '@/api/client'
import {
  PhSparkle,
  PhArrowsClockwise,
  PhUser,
  PhCalendar,
  PhPlus,
  PhMinus,
  PhPaperPlaneTilt,
  PhNotePencil,
  PhArrowRight
} from '@phosphor-icons/vue'

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

function getActionIcon(action: string): Component {
  if (action.includes('created')) return markRaw(PhSparkle)
  if (action.includes('status_changed')) return markRaw(PhArrowsClockwise)
  if (action.includes('subject_changed')) return markRaw(PhUser)
  if (action.includes('deadline_changed')) return markRaw(PhCalendar)
  if (action.includes('added')) return markRaw(PhPlus)
  if (action.includes('removed')) return markRaw(PhMinus)
  if (action.includes('shared')) return markRaw(PhPaperPlaneTilt)
  return markRaw(PhNotePencil)
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

<style scoped lang="scss">
.audit {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1rem;

  @media (min-width: 768px) {
    padding: 2rem;
  }

  &__header {
    margin-bottom: 2rem;
  }

  &__title {
    font-size: 1.5rem;
    font-weight: 600;
    margin: 0 0 0.5rem 0;
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
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-bottom: 2rem;
    padding: 1.25rem;
    background: var(--bg-primary);
    border-radius: 8px;
    border: 1px solid var(--border-color);

    @media (min-width: 768px) {
      flex-direction: row;
      padding: 1.5rem;
    }
  }

  &__filter-group {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  &__filter-label {
    font-size: 0.8rem;
    font-weight: 500;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 0.875rem;
    }
  }

  &__filter-select,
  &__filter-input {
    padding: 0.5rem;
    border: 1px solid var(--border-color);
    border-radius: 6px;
    font-size: 0.85rem;
    background: var(--bg-primary);
    color: var(--text-primary);
    min-height: 44px;

    @media (min-width: 768px) {
      font-size: 0.875rem;
    }

    &:focus {
      outline: none;
      border-color: var(--color-primary);
      box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
    }
  }

  &__loading,
  &__error,
  &__empty {
    text-align: center;
    padding: 2rem 1rem;
    color: var(--text-secondary);

    @media (min-width: 768px) {
      padding: 3rem;
    }
  }

  &__error {
    color: var(--color-error);
  }

  &__list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  &__pagination {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    gap: 1rem;
    margin-top: 2rem;
    padding: 1.25rem;
    background: var(--bg-primary);
    border-radius: 8px;
    border: 1px solid var(--border-color);

    @media (min-width: 768px) {
      flex-direction: row;
      padding: 1.5rem;
    }
  }

  &__page-btn {
    min-width: 100px;

    &:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }
  }

  &__page-info {
    font-size: 0.8rem;
    color: var(--text-secondary);

    @media (min-width: 768px) {
      font-size: 0.875rem;
    }
  }
}

.log-card {
  background: var(--bg-primary);
  border-radius: 8px;
  border: 1px solid var(--border-color);
  border-left: 4px solid var(--border-color);
  overflow: hidden;
  transition: box-shadow 0.2s;

  &:hover {
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  }

  &--green { border-left-color: #10b981; }
  &--blue { border-left-color: #3b82f6; }
  &--cyan { border-left-color: #06b6d4; }
  &--orange { border-left-color: #f97316; }
  &--purple { border-left-color: #a855f7; }

  &__header {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    padding: 1rem;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border-color);

    @media (min-width: 768px) {
      flex-direction: row;
      justify-content: space-between;
      align-items: center;
      padding: 1rem 1.5rem;
    }
  }

  &__timestamp {
    font-size: 0.8rem;
    color: var(--text-tertiary);

    @media (min-width: 768px) {
      font-size: 0.875rem;
    }
  }

  &__body {
    padding: 1.25rem;

    @media (min-width: 768px) {
      padding: 1.5rem;
    }
  }

  &__description {
    font-size: 0.9rem;
    margin: 0 0 1rem 0;
    color: var(--text-primary);

    @media (min-width: 768px) {
      font-size: 1rem;
    }
  }

  &__details {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  &__detail {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    font-size: 0.8rem;

    @media (min-width: 768px) {
      flex-direction: row;
      gap: 0.5rem;
      font-size: 0.875rem;
    }
  }

  &__label {
    font-weight: 500;
    color: var(--text-secondary);
    min-width: 60px;
  }

  &__value {
    color: var(--text-primary);
  }

  &__old-value {
    text-decoration: line-through;
    color: var(--text-tertiary);
  }

  &__arrow {
    color: var(--text-tertiary);
    margin: 0 0.25rem;
  }

  &__new-value {
    font-weight: 500;
    color: var(--color-success);
  }
}

.badge {
  font-size: 0.8rem;
  font-weight: 500;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  background: var(--bg-tertiary);
  color: var(--text-primary);
  display: inline-block;

  @media (min-width: 768px) {
    font-size: 0.875rem;
  }

  &--green { background: rgba(16, 185, 129, 0.1); color: #065f46; }
  &--blue { background: rgba(59, 130, 246, 0.1); color: #1e40af; }
  &--cyan { background: rgba(6, 182, 212, 0.1); color: #155e75; }
  &--orange { background: rgba(249, 115, 22, 0.1); color: #9a3412; }
  &--purple { background: rgba(168, 85, 247, 0.1); color: #6b21a8; }
}
</style>
